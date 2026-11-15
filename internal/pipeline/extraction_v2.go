package csolve

import (
	"context"
	"encoding/json"
	"fmt"
	morbesllm "github.com/G226Medi03/MorbesLLmPipeline"
	entityjson "github.com/G226Medi03/MorbesPipeline/internal/domain/entityjson"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"sort"
	"strings"
	"sync"
	"time"
)

func (self *Service) run_v2_actor_extraction(ctx context.Context, id domain.CaseID, state *pipelineState) ([]ActorDraft, error) {
	template := state.Prompts.templates[PromptActor]
	if strings.TrimSpace(template) == "" {
		return nil, nil
	}
	prompt := RenderPrompt(template, map[string]string{
		"problem":     state.Case.Problem,
		"answer":      "",
		"explanation": "",
	})
	started := time.Now()
	raw, response, err := self.generate_extraction(ctx, state, prompt)
	var drafts []ActorDraft
	if err == nil {
		items, parse_err := morbesllm.ParseActorItemsResponse(raw)
		if parse_err != nil {
			err = parse_err
		} else {
			for _, item := range items {
				if item == nil {
					continue
				}
				name := strings.TrimSpace(item.ActorName())
				if name == "" {
					continue
				}
				kind := canonical_actor_kind(item.Kind())
				if strings.TrimSpace(item.Kind()) == "" {
					kind = resolve_actor_kind_from_name(name)
				}
				if kind == "" {
					kind = "unknown_actor"
				}
				drafts = append(drafts, ActorDraft{
					Key:      fmt.Sprintf("o%d", len(drafts)+1),
					Kind:     kind,
					Name:     name,
					SourceID: extraction_source_id(response, state),
				})
			}
		}
	}
	self.record_v2_extraction_call(id, seqActor, "actor_extract", "actor", state.Prompts.versions[PromptActor], prompt, raw, response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_status_extraction(ctx context.Context, id domain.CaseID, state *pipelineState) ([]StatusDraft, error) {
	prompt := RenderPrompt(state.Prompts.templates[PromptStatus], map[string]string{
		"problem":     problem_stem(state.Case.Problem),
		"answer":      "",
		"explanation": "",
	})
	prompt += problem_only_instruction(PromptStatus)
	started := time.Now()
	raw, response, err := self.generate_extraction(ctx, state, prompt)
	var drafts []StatusDraft
	if err == nil {
		statuses, parse_err := morbesllm.ParseStatusesResponse(raw)
		if parse_err != nil {
			err = parse_err
		} else {
			keys := make([]string, 0, len(statuses))
			for key := range statuses {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				status := statuses[key]
				value, marshal_err := entityjson.EncodeStatus(status)
				if marshal_err != nil {
					continue
				}
				drafts = append(drafts, StatusDraft{
					Key:      key,
					Kind:     status.Kind(),
					Value:    value,
					SourceID: extraction_source_id(response, state),
				})
			}
		}
	}
	self.record_v2_extraction_call(id, seqStatus, "status_extract", "status", state.Prompts.versions[PromptStatus], prompt, raw, response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_action_extraction(ctx context.Context, id domain.CaseID, state *pipelineState) ([]ActionDraft, error) {
	prompt := RenderPrompt(state.Prompts.templates[PromptAction], map[string]string{
		"problem":     problem_stem(state.Case.Problem),
		"answer":      "",
		"explanation": "",
	})
	prompt += problem_only_instruction(PromptAction)
	started := time.Now()
	raw, response, err := self.generate_extraction(ctx, state, prompt)
	var drafts []ActionDraft
	if err == nil {
		actions, parse_err := morbesllm.ParseActionsResponse(raw)
		if parse_err != nil {
			err = parse_err
		} else {
			keys := make([]string, 0, len(actions))
			for key := range actions {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				action := actions[key]
				encoded, marshal_err := entityjson.EncodeAction(action)
				if marshal_err != nil {
					continue
				}
				drafts = append(drafts, ActionDraft{
					Key:      key,
					Variant:  action.Kind(),
					Value:    encoded,
					SourceID: extraction_source_id(response, state),
				})
			}
		}
	}
	self.record_v2_extraction_call(id, seqAction, "action_extract", "action", state.Prompts.versions[PromptAction], prompt, raw, response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_patient_description_extraction(
	ctx context.Context,
	id domain.CaseID,
	state *pipelineState,
	statuses []StatusDraft,
) ([]PatientDescriptionDraft, error) {
	if !looks_like_patient_description(state.Case.Problem) {
		return nil, nil
	}
	template := state.Prompts.templates[PromptPatientDescription]
	if strings.TrimSpace(template) == "" {
		return nil, nil
	}
	prompt := RenderPrompt(template, map[string]string{
		"problem":     problem_stem(state.Case.Problem),
		"answer":      "",
		"explanation": "",
		"status_json": status_drafts_json(statuses),
	})
	prompt += problem_only_instruction(PromptPatientDescription)
	started := time.Now()
	raw, response, err := self.generate_extraction(ctx, state, prompt)
	var drafts []PatientDescriptionDraft
	if err == nil {
		var parsed struct {
			PatientDescriptions []struct {
				Choice      string            `json:"choice"`
				Description []json.RawMessage `json:"description"`
				Knowledge   []string          `json:"knowledge"`
			} `json:"patientDescriptions"`
		}
		if parse_err := decode_json_object(raw, &parsed); parse_err != nil {
			err = parse_err
		} else {
			for index, item := range parsed.PatientDescriptions {
				draft := PatientDescriptionDraft{
					Choice:    strings.TrimSpace(item.Choice),
					Knowledge: trim_string_list(item.Knowledge),
					SourceID:  extraction_source_id(response, state),
				}
				for status_index, raw_status := range item.Description {
					status_draft, ok := status_draft_from_raw(raw_status, fmt.Sprintf("p%d_%d", index+1, status_index+1))
					if !ok {
						continue
					}
					status_draft.SourceID = draft.SourceID
					draft.Description = append(draft.Description, status_draft)
				}
				if len(draft.Description) > 0 || len(draft.Knowledge) > 0 {
					drafts = append(drafts, draft)
				}
			}
		}
	}
	self.record_v2_extraction_call(id, seqPatientDescription, "patient_description_extract", "patient_description", state.Prompts.versions[PromptPatientDescription], prompt, raw, response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_medical_procedure_option_extraction(
	ctx context.Context,
	id domain.CaseID,
	state *pipelineState,
	statuses []StatusDraft,
	patient_descriptions []PatientDescriptionDraft,
) ([]OptionDraft, error) {
	template := state.Prompts.templates[PromptMedicalProcedureOption]
	if strings.TrimSpace(template) == "" {
		return nil, nil
	}
	patient_json, _ := json.Marshal(patient_descriptions)
	choices_json, _ := json.Marshal(problem_choices(state.Case.Problem))
	problem_text := state.Case.Problem
	if len(choices_json) > 2 {
		problem_text += "\n\n[선택지 JSON]\n" + string(choices_json)
	}
	prompt := RenderPrompt(template, map[string]string{
		"problem":                  problem_text,
		"answer":                   "",
		"explanation":              "",
		"status_json":              status_drafts_json(statuses),
		"patient_description_json": string(patient_json),
	})
	prompt += problem_only_instruction(PromptMedicalProcedureOption)
	started := time.Now()
	raw, response, err := self.generate_extraction(ctx, state, prompt)
	var drafts []OptionDraft
	if err == nil {
		var parsed struct {
			MedicalProcedureOptions []struct {
				ID               string          `json:"id"`
				Category         string          `json:"category"`
				Traits           []string        `json:"traits"`
				MedicalProcedure json.RawMessage `json:"medical_procedure"`
				Indications      []string        `json:"indications"`
				ExpectedEffects  []string        `json:"expected_effects"`
			} `json:"medical_procedure_options"`
		}
		if parse_err := decode_json_object(raw, &parsed); parse_err != nil {
			err = parse_err
		} else {
			for _, item := range parsed.MedicalProcedureOptions {
				if len(item.MedicalProcedure) == 0 || string(item.MedicalProcedure) == "null" {
					continue
				}
				key := strings.TrimSpace(item.ID)
				if key == "" {
					key = fmt.Sprintf("m%d", len(drafts)+1)
				}
				drafts = append(drafts, OptionDraft{
					Key:              key,
					Category:         strings.TrimSpace(item.Category),
					Traits:           item.Traits,
					MedicalProcedure: item.MedicalProcedure,
					Indications:      trim_string_list(item.Indications),
					ExpectedEffects:  trim_string_list(item.ExpectedEffects),
					SourceID:         extraction_source_id(response, state),
				})
			}
		}
	}
	self.record_v2_extraction_call(id, seqMedicalProcedureOption, "medical_procedure_option_extract", "medical_procedure_option", state.Prompts.versions[PromptMedicalProcedureOption], prompt, raw, response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_relation_extraction(
	ctx context.Context,
	id domain.CaseID,
	state *pipelineState,
	statuses []StatusDraft,
	actions []ActionDraft,
) ([]RelationDraft, error) {
	template := state.Prompts.templates[PromptRelation]
	if strings.TrimSpace(template) == "" {
		return nil, fmt.Errorf("relation prompt가 비어 있습니다")
	}
	prompt, err := build_relation_prompt_from_drafts(
		template,
		problem_stem(state.Case.Problem),
		statuses,
		actions,
	)
	if err != nil {
		return nil, err
	}
	prompt += problem_only_instruction(PromptRelation)
	started := time.Now()
	relations, relation_source, ask, err := self.llm.ExtractRelations(
		ctx, prompt, state.ExplanationSourceID, self.cache_mode,
	)
	var drafts []RelationDraft
	if err == nil {
		for _, steps := range relations {
			for index := 1; index+1 < len(steps); index += 2 {
				drafts = append(drafts, RelationDraft{
					CauseStatusKey:  steps[index-1],
					ActionKey:       steps[index],
					ResultStatusKey: steps[index+1],
					SourceID:        relation_source.SourceID,
				})
			}
		}
	}
	self.record_v2_relation_call(id, state, prompt, ask, relations, started, err)
	return drafts, err
}

func (self *Service) generate_extraction(ctx context.Context, state *pipelineState, prompt string) (string, *source.LlmResponse, error) {
	if self.llm == nil {
		return "", nil, fmt.Errorf("extraction LLM is not configured")
	}
	refered_source_id := extraction_refered_source(state)
	response, err := self.llm.AskRole(ctx, morbesllm.RoleExtract, prompt, refered_source_id, self.cache_mode)
	if err != nil {
		return "", nil, err
	}
	return response.OutputText(), response, nil
}

func (self *Service) record_v2_extraction_call(
	case_id domain.CaseID,
	seq int,
	stage string,
	role string,
	prompt_version int,
	prompt string,
	raw string,
	response *source.LlmResponse,
	parsed any,
	started time.Time,
	call_err error,
) {
	model := ""
	if response != nil {
		model = response.Model
	}
	if model == "" && self.llm != nil {
		model = self.llm.Model(morbesllm.RoleExtract)
	}
	call := LLMCallInput{
		CaseID:        case_id,
		Seq:           seq,
		Stage:         stage,
		Role:          role,
		Model:         model,
		PromptVersion: prompt_version,
		Prompt:        prompt,
		RawOutput:     raw,
		ParsedOutput:  encode_json(parsed),
		State:         state_text(call_err),
		Error:         error_text(call_err),
		DurationMS:    time.Since(started).Milliseconds(),
	}
	if response != nil {
		call.SourceID = response.SourceID
		call.RespID = response.RespID
	}
	self.record_call(call)
}

func (self *Service) record_v2_relation_call(
	case_id domain.CaseID,
	state *pipelineState,
	prompt string,
	response *source.LlmResponse,
	relations [][]string,
	started time.Time,
	call_err error,
) {
	call := LLMCallInput{
		CaseID:        case_id,
		Seq:           seqRelation,
		Stage:         "relation_extract",
		Role:          "relation",
		Model:         self.llm.Model(morbesllm.RoleExtract),
		PromptVersion: state.Prompts.versions[PromptRelation],
		Prompt:        prompt,
		ParsedOutput:  encode_json(relations),
		State:         state_text(call_err),
		Error:         error_text(call_err),
		DurationMS:    time.Since(started).Milliseconds(),
	}
	if response != nil {
		call.RawOutput = response.OutputText()
		call.SourceID = response.SourceID
		call.RespID = response.RespID
	}
	self.record_call(call)
}

func extraction_refered_source(state *pipelineState) source.SourceId {
	if state.ExplanationSourceID.ToI64() > 0 {
		return state.ExplanationSourceID
	}
	return state.Case.SourceID
}

func extraction_source_id(response *source.LlmResponse, state *pipelineState) source.SourceId {
	if response != nil && response.SourceID.ToI64() > 0 {
		return response.SourceID
	}
	return extraction_refered_source(state)
}

func status_drafts_json(statuses []StatusDraft) string {
	items := make([]map[string]any, 0, len(statuses))
	for _, draft := range statuses {
		items = append(items, map[string]any{
			"id":    draft.Key,
			"kind":  draft.Kind,
			"value": draft.Value,
		})
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func status_draft_from_raw(raw json.RawMessage, fallback_key string) (StatusDraft, bool) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return StatusDraft{}, false
	}
	kind := ""
	if value, ok := payload["kind"]; ok {
		_ = json.Unmarshal(value, &kind)
	}
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return StatusDraft{}, false
	}
	delete(payload, "kind")
	if value, ok := payload["value"]; ok && len(payload) == 1 {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(value, &nested); err == nil {
			payload = nested
		}
	}
	value, err := json.Marshal(payload)
	if err != nil {
		return StatusDraft{}, false
	}
	return StatusDraft{Key: fallback_key, Kind: kind, Value: value}, true
}

func run_parallel(functions ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(functions))
	for _, function := range functions {
		go func(run func()) {
			defer wg.Done()
			run()
		}(function)
	}
	wg.Wait()
}
