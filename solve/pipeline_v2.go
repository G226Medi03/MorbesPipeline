package csolve

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	"github.com/G226Medi03/MedicalEntityParser/parser"
	"github.com/G226Medi03/MorbesPipeline/entityconv"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/logs"
	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/medicalprocedureoption"
	"github.com/G226Medi03/MorbesPipeline/patient"
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
	"github.com/G226Medi03/MorbesPipeline/trait"
	source "github.com/G226Medi03/SourceEntity"
)

func (self *Service) run_v2_from(ctx context.Context, id domain.CaseID, from_node string) error {
	return self.runPipelineNodes(ctx, id, from_node)
}
func (self *Service) run_case_extraction_node(ctx context.Context, id domain.CaseID, state *pipelineState) error {
	step_id, err := self.store.StartStep(id, NodeCaseExtraction, 0)
	if err != nil {
		return err
	}
	if err := self.set_stage(id, "case_extraction"); err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}

	if err := self.extractionPipeline(ctx, id, state); err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	output := caseExtractionStepOutput{QuestionType: string(state.QuestionType), CaseExtraction: state.Extraction}
	return self.store.FinishStep(step_id, "succeeded", encode_json(output), "", 0)
}

func (self *Service) run_resolve_node(ctx context.Context, id domain.CaseID, state *pipelineState) (err error) {
	step_id, err := self.store.StartStep(id, NodeResolve, 0)
	if err != nil {
		return err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("resolve panic: %v", recovered)
			_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
			_ = self.store.SetFailed(id, err)
		}
	}()
	if err := self.set_stage(id, "resolve"); err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	resolved, err := self.resolve_case_extraction(state)
	if err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	state.Resolved = resolved
	state.Statuses = resolved.Statuses
	state.Actions = resolved.Actions
	state.PatientDescriptions = resolved.PatientDescriptions
	return self.store.FinishStep(step_id, "succeeded", encode_json(resolveStepOutput{Resolved: resolved}), "", 0)
}

func (self *Service) resolve_case_extraction(state *pipelineState) (ResolvedCaseExtraction, error) {
	if !looks_like_patient_description(state.Case.Problem) {
		state.Extraction.PatientDescription = nil
		state.Extraction.PatientDescriptions = nil
	}
	resolved := ResolvedCaseExtraction{
		Actors:              state.Extraction.Actors,
		Statuses:            map[string]entity.Status{},
		Actions:             map[string]entity.Action{},
		Relations:           state.Extraction.Relations,
		PatientDescriptions: []patient.PatientDescription{},
	}
	source_id := state.ExplanationSourceID
	if source_id.ToI64() <= 0 {
		source_id = state.Case.SourceID
	}
	type infusion_actor struct {
		name    string
		sources []source.SourceRef
	}
	infusion_actors := make([]infusion_actor, 0)
	draft_sources := func(draft_source_id source.SourceId) []source.SourceRef {
		if draft_source_id.ToI64() <= 0 {
			draft_source_id = source_id
		}
		if draft_source_id.ToI64() <= 0 {
			return nil
		}
		return []source.SourceRef{{SourceID: draft_source_id}}
	}

	for _, draft := range state.Extraction.Actors {
		actorID, err := self.resolve_actor(draft, draft_sources(draft.SourceID))
		if err != nil {
			logs.Warn("case extraction actor skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
			continue
		}
		if actorID > 0 && strings.Contains(strings.TrimSpace(draft.Name), "투여") {
			infusion_actors = append(infusion_actors, infusion_actor{
				name:    strings.TrimSpace(draft.Name),
				sources: draft_sources(draft.SourceID),
			})
		}
	}

	patient_drafts := make([]PatientDescriptionDraft, 0, len(state.Extraction.PatientDescriptions)+1)
	if state.Extraction.PatientDescription != nil {
		patient_drafts = append(patient_drafts, *state.Extraction.PatientDescription)
	}
	patient_drafts = append(patient_drafts, state.Extraction.PatientDescriptions...)
	if !looks_like_patient_description(state.Case.Problem) {
		patient_drafts = nil
	}

	for _, draft := range state.Extraction.Statuses {
		draft_source_id := draft.SourceID
		if draft_source_id.ToI64() <= 0 {
			draft_source_id = source_id
		}
		status, ok := self.resolve_status_draft(draft, draft_sources(draft_source_id), draft_source_id)
		if !ok {
			continue
		}
		resolved.Statuses[draft.Key] = status
	}
	for _, description_draft := range patient_drafts {
		for _, draft := range description_draft.Description {
			if _, exists := resolved.Statuses[draft.Key]; exists {
				continue
			}
			draft_source_id := draft.SourceID
			if draft_source_id.ToI64() <= 0 {
				draft_source_id = description_draft.SourceID
			}
			if draft_source_id.ToI64() <= 0 {
				draft_source_id = source_id
			}
			status, ok := self.resolve_status_draft(draft, draft_sources(draft_source_id), draft_source_id)
			if !ok {
				continue
			}
			resolved.Statuses[draft.Key] = status
		}
	}

	for _, draft := range state.Extraction.Actions {
		normalized, norm_err := parser.NormalizeActionValue(draft.Variant, draft.Value)
		if norm_err != nil {
			logs.Warn("case extraction action skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", norm_err))
			continue
		}
		value := entity.NewAction(draft.Variant, normalized, draft_sources(draft.SourceID))
		if value.Variant == "" {
			value.Variant = "cause_status"
		}
		if value.Variant == "cause_status" && self.canonical != nil {
			canonical, canonical_err := self.canonical.CanonicalizeCauseStatusValue(value.Value, value.Sources)
			if canonical_err != nil {
				logs.Warn("case extraction action canonicalize skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", canonical_err))
				continue
			}
			value.Value = canonical
		}
		saved, save_err := self.store.SaveAction(value)
		if save_err != nil {
			logs.Warn("case extraction action skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", save_err))
			continue
		}
		resolved.Actions[draft.Key] = saved
	}

	for _, draft := range state.Extraction.MedicalProcedureOptions {
		indication_ids := make([]entity.StatusID, 0, len(draft.Indications))
		for _, key := range draft.Indications {
			if status, ok := resolved.Statuses[key]; ok {
				indication_ids = append(indication_ids, status.Id())
			}
		}
		if len(indication_ids) == 0 {
			for _, status := range resolved.Statuses {
				if status.Id() > 0 {
					indication_ids = append(indication_ids, status.Id())
				}
			}
		}
		expected_effect_ids := make([]entity.StatusID, 0, len(draft.ExpectedEffects))
		for _, key := range draft.ExpectedEffects {
			if status, ok := resolved.Statuses[key]; ok {
				expected_effect_ids = append(expected_effect_ids, status.Id())
			}
		}
		option_sources := draft_sources(draft.SourceID)
		procedure_id, procedure_err := self.canonical.CanonicalizeProcedure(
			draft.MedicalProcedure,
			trait.IntrinsicTraits(draft.Traits),
			option_sources,
		)
		if procedure_err != nil {
			logs.Warn("case extraction procedure skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", procedure_err))
			continue
		}
		saved, save_err := self.store.SaveMedicalProcedureOption(medicalprocedureoption.NewMedicalProcedureOption(
			0,
			procedure_id,
			draft.Category,
			entityconv.EntityTraits(trait.IntrinsicTraits(draft.Traits)),
			entityconv.EntityTraits(trait.ContextTraits(draft.Traits)),
			indication_ids,
			expected_effect_ids,
			option_sources,
			nil,
			nil,
		))
		if save_err != nil {
			logs.Warn("case extraction option skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", save_err))
			continue
		}
		resolved.MedicalProcedureOptions = append(resolved.MedicalProcedureOptions, saved)
	}

	for _, actor := range infusion_actors {
		if err := self.ensure_infusion_option(actor.name, actor.sources); err != nil {
			logs.Warn("case extraction infusion option skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("name", actor.name), logs.AttrOf("error", err))
		}
	}

	for _, relation := range state.Extraction.Relations {
		status, status_ok := resolved.Statuses[relation.CauseStatusKey]
		action, action_ok := resolved.Actions[relation.ActionKey]
		result, result_ok := resolved.Statuses[relation.ResultStatusKey]
		relation_source_id := relation.SourceID
		if relation_source_id.ToI64() <= 0 {
			relation_source_id = source_id
		}
		if status_ok && action_ok {
			if err := self.store.LinkStatusAction(status.Id(), action.Id(), relation_source_id); err != nil {
				logs.Warn("case extraction status-action link skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("error", err))
			}
		}
		if action_ok && result_ok {
			if err := self.store.LinkActionResultStatus(action.Id(), result.Id(), relation_source_id); err != nil {
				logs.Warn("case extraction action-result link skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("error", err))
			}
		}
	}

	for _, description_draft := range patient_drafts {
		description := patient.PatientDescription{
			Choice:    strings.TrimSpace(description_draft.Choice),
			Knowledge: trim_string_list(description_draft.Knowledge),
		}
		for _, draft := range description_draft.Description {
			status, ok := resolved.Statuses[draft.Key]
			if !ok {
				draft_source_id := draft.SourceID
				if draft_source_id.ToI64() <= 0 {
					draft_source_id = description_draft.SourceID
				}
				if draft_source_id.ToI64() <= 0 {
					draft_source_id = source_id
				}
				status, ok = self.resolve_status_draft(draft, draft_sources(draft_source_id), draft_source_id)
				if !ok {
					continue
				}
				resolved.Statuses[draft.Key] = status
			}
			description.Description = append(description.Description, status)
		}
		if len(description.Description) > 0 || len(description.Knowledge) > 0 {
			resolved.PatientDescriptions = append(resolved.PatientDescriptions, description)
		}
	}
	return resolved, nil
}

func (self *Service) resolve_actor(draft ActorDraft, sources []source.SourceRef) (entity.ActorID, error) {
	name := strings.TrimSpace(draft.Name)
	if name == "" {
		return 0, nil
	}
	if eactor.ParseKind(draft.Kind) == eactor.KindDevice {
		saved, err := self.store.SaveDevice(eactor.Device{Name: name}, sources)
		return saved.ID, err
	}
	saved, err := self.store.SaveUnknownActor(eactor.UnknownActor{Name: name}, sources)
	return saved.ID, err
}

func (self *Service) ensure_infusion_option(name string, sources []source.SourceRef) error {
	name = strings.TrimSpace(name)
	if name == "" || len(sources) == 0 {
		return nil
	}
	raw := map[string]any{
		"kind": "Infuse",
		"infuse": map[string]any{
			"material": map[string]any{"kind": "other_actor", "name": name},
			"route":    "unknown",
		},
	}
	procedureID, err := self.canonical.CanonicalizeProcedure(raw, nil, sources)
	if err != nil {
		return err
	}
	_, err = self.store.SaveMedicalProcedureOption(medicalprocedureoption.NewMedicalProcedureOption(
		0,
		procedureID,
		"treatment",
		nil,
		nil,
		nil,
		nil,
		sources,
		nil,
		nil,
	))
	return err
}

func (self *Service) resolve_status_draft(draft StatusDraft, sources []source.SourceRef, source_id source.SourceId) (entity.Status, bool) {
	status, err := parse_status_draft(draft)
	if err != nil {
		panic(fmt.Errorf("status parse failed key=%s kind=%s value=%s source=%d: %w", draft.Key, draft.Kind, string(draft.Value), source_id, err))
	}
	if source_id.ToI64() > 0 {
		saved, save_err := self.canonical.CanonicalizeStatusEntity(status.Variant, status.Value, sources)
		if save_err != nil {
			panic(fmt.Errorf("status canonicalize failed key=%s kind=%s value=%s source=%d: %w", draft.Key, draft.Kind, string(draft.Value), source_id, save_err))
		}
		return saved, true
	}
	status.Sources = sources
	return status, true
}

func parse_status_draft(draft StatusDraft) (entity.Status, error) {
	payload := map[string]any{}
	if len(draft.Value) > 0 {
		if err := json.Unmarshal(draft.Value, &payload); err != nil {
			return entity.NewStatus("", nil, nil), err
		}
	}
	payload["kind"] = draft.Kind
	raw, err := json.Marshal(payload)
	if err != nil {
		return entity.NewStatus("", nil, nil), err
	}
	return patient.ParseStatus(raw)
}

func (self *Service) load_case_extraction(id domain.CaseID, state *pipelineState) error {
	var output caseExtractionStepOutput
	if err := self.load_step_output(id, NodeCaseExtraction, &output); err != nil {
		return err
	}
	state.Extraction = output.CaseExtraction
	state.QuestionType = domain.NormalizeQuestionType(output.QuestionType)
	if state.QuestionType == "" {
		state.QuestionType = domain.NormalizeQuestionType(state.Case.QuestionType)
	}
	return nil
}

func (self *Service) load_resolve(id domain.CaseID, state *pipelineState) error {
	var output resolveStepOutput
	if err := self.load_step_output(id, NodeResolve, &output); err != nil {
		return err
	}
	state.Resolved = output.Resolved
	state.Statuses = output.Resolved.Statuses
	state.Actions = output.Resolved.Actions
	state.PatientDescriptions = output.Resolved.PatientDescriptions
	return nil
}
