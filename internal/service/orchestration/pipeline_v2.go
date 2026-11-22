package csolve

import (
	"context"
	"fmt"
	"strings"

	llmparser "github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/patient"
	"github.com/G226Medi03/MorbesPipeline/internal/service/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/trait"
	presolve "github.com/G226Medi03/MorbesPipeline/internal/service/resolve"
	source "github.com/G226Medi03/SourceEntity"
	clog "github.com/G226Medi03/log-sqlite"
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
	if err := self.runLLMExtraction(ctx, id, state); err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	output := caseExtractionStepOutput{QuestionType: string(state.QuestionType), CaseExtraction: state.Extraction}
	return self.store.FinishStep(step_id, "succeeded", marshal_json(output), "", 0)
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
	return self.store.FinishStep(step_id, "succeeded", marshal_json(resolveStepOutput{Resolved: resolved}), "", 0)
}

func (self *Service) resolve_case_extraction(state *pipelineState) (ResolvedCaseExtraction, error) {
	if !looks_like_patient_description(state.Case.Problem) {
		state.Extraction.PatientDescription = nil
		state.Extraction.PatientDescriptions = nil
	}
	resolved := ResolvedCaseExtraction{
		Actors:              state.Extraction.Actors,
		Statuses:            map[string]repo.IEStatus{},
		Actions:             map[string]repo.IEAction{},
		Relations:           state.Extraction.Relations,
		PatientDescriptions: []patient.PatientDescription{},
	}
	entityStore, err := self.store.Entities()
	if err != nil {
		return resolved, err
	}
	source_id := state.ExplanationSourceID
	if source_id.ToI64() <= 0 {
		source_id = state.Case.SourceID
	}
	type infusion_actor struct {
		name    string
		sources []source.Source
	}
	infusion_actors := make([]infusion_actor, 0)
	draft_sources := func(draft_source_id source.SourceId) []source.Source {
		if draft_source_id.ToI64() <= 0 {
			draft_source_id = source_id
		}
		if draft_source_id.ToI64() <= 0 {
			return nil
		}
		return []source.Source{{SourceID: draft_source_id}}
	}

	for index, draft := range state.Extraction.Actors {
		actor, err := presolve.Actor(draft)
		if err != nil {
			clog.Warn("case extraction actor skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			continue
		}
		if actor == nil {
			continue
		}
		sources := draft_sources(source_id)
		if _, err := entityStore.InsertActor(actor, nil, sourceIDs(sources)); err != nil {
			clog.Warn("case extraction actor skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			continue
		}
		if strings.Contains(strings.TrimSpace(draft.Name), "투여") {
			infusion_actors = append(infusion_actors, infusion_actor{
				name:    strings.TrimSpace(draft.Name),
				sources: sources,
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
		status, err := llmparser.StatusFromDTO(draft.Value)
		if err != nil {
			panic(fmt.Errorf("status DTO conversion failed key=%s: %w", draft.Key, err))
		}
		if draft_source_id.ToI64() > 0 {
			if _, err := entityStore.InsertStatus(status, sourceIDs(draft_sources(draft_source_id))); err != nil {
				panic(fmt.Errorf("status save failed key=%s kind=%s source=%d: %w", draft.Key, status.Kind(), draft_source_id, err))
			}
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
			status, err := llmparser.StatusFromDTO(draft.Value)
			if err != nil {
				continue
			}
			if draft_source_id.ToI64() > 0 {
				if _, err := entityStore.InsertStatus(status, sourceIDs(draft_sources(draft_source_id))); err != nil {
					panic(fmt.Errorf("status save failed key=%s kind=%s source=%d: %w", draft.Key, status.Kind(), draft_source_id, err))
				}
			}
			resolved.Statuses[draft.Key] = status
		}
	}

	for _, draft := range state.Extraction.Actions {
		value, err := llmparser.ActionFromDTO(draft.Value)
		if err != nil {
			clog.Warn("case extraction action skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("key", draft.Key), clog.AttrOf("error", err))
			continue
		}
		sources := draft_sources(draft.SourceID)
		if _, err := entityStore.InsertAction(value, nil, sourceIDs(sources)); err != nil {
			clog.Warn("case extraction action skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("key", draft.Key), clog.AttrOf("error", err))
			continue
		}
		resolved.Actions[draft.Key] = value
	}

	for _, draft := range state.Extraction.MedicalProcedureOptions {
		indications := make([]repo.IEStatus, 0, len(draft.Indications))
		for _, statusDTO := range draft.Indications {
			if status, err := llmparser.StatusFromDTO(statusDTO); err == nil {
				indications = append(indications, status)
			}
		}
		if len(indications) == 0 {
			for _, status := range resolved.Statuses {
				indications = append(indications, status)
			}
		}
		effects := make([]repo.IEStatus, 0, len(draft.ExpectedEffects))
		for _, statusDTO := range draft.ExpectedEffects {
			if status, err := llmparser.StatusFromDTO(statusDTO); err == nil {
				effects = append(effects, status)
			}
		}
		option_sources := draft_sources(draft.SourceID)
		procedure, err := llmparser.ProcedureFromDTO(&draft.MedicalProcedure)
		if err != nil {
			clog.Warn("case extraction procedure skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("key", draft.Key), clog.AttrOf("error", err))
			continue
		}
		optionTraits := trait.NormalizeTraits(draft.Traits)
		if _, err := self.store.SaveProcedure(procedure, trait.IntrinsicTraits(optionTraits), option_sources); err != nil {
			clog.Warn("case extraction procedure skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("key", draft.Key), clog.AttrOf("error", err))
			continue
		}
		saved, err := self.store.SaveMedicalProcedureOptionByValues(
			procedure,
			draft.Category,
			trait.IntrinsicTraits(optionTraits),
			trait.ContextTraits(optionTraits),
			indications,
			effects,
			option_sources,
		)
		if err != nil {
			clog.Warn("case extraction option skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("key", draft.Key), clog.AttrOf("error", err))
			continue
		}
		resolved.MedicalProcedureOptions = append(resolved.MedicalProcedureOptions, saved)
	}

	for _, actor := range infusion_actors {
		if err := presolve.EnsureInfusionOption(self.store, actor.name, actor.sources); err != nil {
			clog.Warn("case extraction infusion option skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("name", actor.name), clog.AttrOf("error", err))
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
			if err := entityStore.LinkStatusAction(status, action, relation_source_id); err != nil {
				clog.Warn("case extraction status-action link skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("error", err))
			}
		}
		if action_ok && result_ok {
			if err := self.store.LinkActionResultStatusValues(action, result, relation_source_id); err != nil {
				clog.Warn("case extraction action-result link skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("error", err))
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
				status, err = llmparser.StatusFromDTO(draft.Value)
				if err != nil {
					continue
				}
				if draft_source_id.ToI64() > 0 {
					if _, err := entityStore.InsertStatus(status, sourceIDs(draft_sources(draft_source_id))); err != nil {
						continue
					}
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

func sourceIDs(refs []source.Source) []source.SourceId {
	result := make([]source.SourceId, 0, len(refs))
	for _, ref := range refs {
		result = append(result, ref.SourceID)
	}
	return result
}
