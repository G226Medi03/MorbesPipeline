package csolve

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	morbesllm "github.com/G226Medi03/MorbesLLMService"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	source "github.com/G226Medi03/SourceEntity"
)

func (self *Service) run_v2_actor_extraction(ctx context.Context, id domain.CaseID, state *pipelineState) ([]ActorDraft, error) {
	if !state.Prompts.present[PromptActor] {
		return nil, nil
	}
	started := time.Now()
	result, err := self.llm.ExtractActors(
		ctx,
		state.Case.Problem,
		"",
		"",
		extraction_refered_source(state),
		self.cache_mode,
	)
	prompt, raw := response_texts(result.Response)
	version := prompt_version_or(state.Prompts.versions[PromptActor], result.PromptVersion)
	self.record_v2_extraction_call(id, seqActor, "actor_extract", "actor", version, prompt, raw, result.Response, result.Actors, started, err)
	return result.Actors, err
}

func (self *Service) run_v2_status_extraction(ctx context.Context, id domain.CaseID, state *pipelineState) ([]StatusDraft, error) {
	started := time.Now()
	statuses, response, err := self.llm.ExtractStatuses(
		ctx,
		problem_stem(state.Case.Problem),
		"",
		"",
		extraction_refered_source(state),
		self.cache_mode,
	)
	prompt, raw := response_texts(response)
	var drafts []StatusDraft
	if err == nil {
		keys := make([]string, 0, len(statuses))
		for key := range statuses {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			drafts = append(drafts, StatusDraft{
				Key:      key,
				Value:    statuses[key],
				SourceID: extraction_source_id(response, state),
			})
		}
	}
	self.record_v2_extraction_call(id, seqStatus, "status_extract", "status", state.Prompts.versions[PromptStatus], prompt, raw, response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_action_extraction(ctx context.Context, id domain.CaseID, state *pipelineState) ([]ActionDraft, error) {
	started := time.Now()
	actions, response, err := self.llm.ExtractActions(
		ctx,
		problem_stem(state.Case.Problem),
		"",
		"",
		extraction_refered_source(state),
		self.cache_mode,
	)
	prompt, raw := response_texts(response)
	var drafts []ActionDraft
	if err == nil {
		keys := make([]string, 0, len(actions))
		for key := range actions {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			drafts = append(drafts, ActionDraft{
				Key:      key,
				Value:    actions[key],
				SourceID: extraction_source_id(response, state),
			})
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
	if !state.Prompts.present[PromptPatientDescription] {
		return nil, nil
	}
	status_map := status_map_from_drafts(statuses)
	started := time.Now()
	result, err := self.llm.ExtractPatientDescription(
		ctx,
		problem_stem(state.Case.Problem),
		"",
		"",
		status_map,
		extraction_refered_source(state),
		self.cache_mode,
	)
	var drafts []PatientDescriptionDraft
	if err == nil {
		for index, item := range result.Descriptions {
			source_id := item.SourceID
			if source_id.ToI64() <= 0 {
				source_id = extraction_source_id(result.Response, state)
			}
			draft := PatientDescriptionDraft{
				Choice:    strings.TrimSpace(item.Choice),
				Knowledge: trim_string_list(item.Knowledge),
				SourceID:  source_id,
			}
			for status_index, status_value := range item.Description {
				draft.Description = append(draft.Description, StatusDraft{
					Key:      fmt.Sprintf("p%d_%d", index+1, status_index+1),
					Value:    status_value,
					SourceID: draft.SourceID,
				})
			}
			if len(draft.Description) > 0 || len(draft.Knowledge) > 0 {
				drafts = append(drafts, draft)
			}
		}
	}
	prompt, raw := response_texts(result.Response)
	version := prompt_version_or(state.Prompts.versions[PromptPatientDescription], result.PromptVersion)
	self.record_v2_extraction_call(id, seqPatientDescription, "patient_description_extract", "patient_description", version, prompt, raw, result.Response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_medical_procedure_option_extraction(
	ctx context.Context,
	id domain.CaseID,
	state *pipelineState,
	statuses []StatusDraft,
	patient_descriptions []PatientDescriptionDraft,
) ([]OptionDraft, error) {
	if !state.Prompts.present[PromptMedicalProcedureOption] {
		return nil, nil
	}
	status_map := status_map_from_drafts(statuses)
	started := time.Now()
	result, err := self.llm.ExtractMedicalProcedureOption(
		ctx,
		state.Case.Problem,
		"",
		"",
		status_map,
		patient_descriptions_for_service(patient_descriptions),
		extraction_refered_source(state),
		self.cache_mode,
	)
	var drafts []OptionDraft
	if err == nil {
		for _, item := range result.Options {
			source_id := item.SourceID
			if source_id.ToI64() <= 0 {
				source_id = extraction_source_id(result.Response, state)
			}
			drafts = append(drafts, OptionDraft{
				Key:              strings.TrimSpace(item.Key),
				Category:         strings.TrimSpace(item.Category),
				Traits:           item.Traits,
				MedicalProcedure: item.MedicalProcedure,
				Indications:      trim_string_list(item.Indications),
				ExpectedEffects:  trim_string_list(item.ExpectedEffects),
				SourceID:         source_id,
			})
		}
	}
	prompt, raw := response_texts(result.Response)
	version := prompt_version_or(state.Prompts.versions[PromptMedicalProcedureOption], result.PromptVersion)
	self.record_v2_extraction_call(id, seqMedicalProcedureOption, "medical_procedure_option_extract", "medical_procedure_option", version, prompt, raw, result.Response, drafts, started, err)
	return drafts, err
}

func (self *Service) run_v2_relation_extraction(
	ctx context.Context,
	id domain.CaseID,
	state *pipelineState,
	statuses []StatusDraft,
	actions []ActionDraft,
) ([]RelationDraft, error) {
	if !state.Prompts.present[PromptRelation] {
		return nil, fmt.Errorf("relation prompt가 비어 있습니다")
	}
	status_map := status_map_from_drafts(statuses)
	action_map := action_map_from_drafts(actions)
	started := time.Now()
	result, err := self.llm.ExtractRelations(
		ctx,
		problem_stem(state.Case.Problem),
		"",
		"",
		status_map,
		action_map,
		state.ExplanationSourceID,
		self.cache_mode,
	)
	var drafts []RelationDraft
	if err == nil {
		for _, steps := range result.Relations {
			for index := 1; index+1 < len(steps); index += 2 {
				drafts = append(drafts, RelationDraft{
					CauseStatusKey:  steps[index-1],
					ActionKey:       steps[index],
					ResultStatusKey: steps[index+1],
					SourceID:        result.Ref.SourceID,
				})
			}
		}
	}
	prompt, _ := response_texts(result.Response)
	version := prompt_version_or(state.Prompts.versions[PromptRelation], result.PromptVersion)
	self.record_v2_relation_call(id, prompt, result.Response, result.Relations, version, started, err)
	return drafts, err
}

func response_texts(response *source.LlmResponse) (string, string) {
	if response == nil {
		return "", ""
	}
	return response.Prompt, response.OutputText()
}

func prompt_version_or(fallback, version int) int {
	if version > 0 {
		return version
	}
	return fallback
}

func status_map_from_drafts(statuses []StatusDraft) map[string]repo.IEStatus {
	result := make(map[string]repo.IEStatus, len(statuses))
	for _, draft := range statuses {
		result[draft.Key] = draft.Value
	}
	return result
}

func action_map_from_drafts(actions []ActionDraft) map[string]repo.IEAction {
	result := make(map[string]repo.IEAction, len(actions))
	for _, draft := range actions {
		result[draft.Key] = draft.Value
	}
	return result
}

func patient_descriptions_for_service(drafts []PatientDescriptionDraft) []morbesllm.PatientDescription {
	result := make([]morbesllm.PatientDescription, 0, len(drafts))
	for _, draft := range drafts {
		item := morbesllm.PatientDescription{
			Choice:    draft.Choice,
			Knowledge: draft.Knowledge,
			SourceID:  draft.SourceID,
		}
		for _, status_draft := range draft.Description {
			item.Description = append(item.Description, status_draft.Value)
		}
		result = append(result, item)
	}
	return result
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
	prompt string,
	response *source.LlmResponse,
	relations [][]string,
	prompt_version int,
	started time.Time,
	call_err error,
) {
	model := ""
	if response != nil {
		model = response.Model
	}
	call := LLMCallInput{
		CaseID:        case_id,
		Seq:           seqRelation,
		Stage:         "relation_extract",
		Role:          "relation",
		Model:         model,
		PromptVersion: prompt_version,
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
