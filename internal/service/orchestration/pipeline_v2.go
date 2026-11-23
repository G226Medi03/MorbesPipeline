package orchestration

import (
	"context"
	"fmt"
	"strings"

	llmparser "github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/patient"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	presolve "github.com/G226Medi03/MorbesPipeline/internal/service/resolve"
	source "github.com/G226Medi03/SourceEntity"
	clog "github.com/G226Medi03/log-sqlite"
)

func (self *Service) run_v2_from(ctx context.Context, id solve.CaseID, from_node string) error {
	return self.runPipelineNodes(ctx, id, from_node)
}

func (self *Service) run_case_extraction_node(ctx context.Context, id solve.CaseID, state *pipelineState) error {
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

func (self *Service) run_resolve_node(ctx context.Context, id solve.CaseID, state *pipelineState) (err error) {
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
	extraction := state.Extraction
	patientDrafts := extraction.PatientDescriptions()
	if !looks_like_patient_description(state.Case.Problem) {
		patientDrafts = nil
	}
	resolved := ResolvedCaseExtraction{
		Actors:              extraction.Actors(),
		Statuses:            make([]repo.IEStatus, 0, len(extraction.Statuses())),
		Actions:             make([]repo.IEAction, 0, len(extraction.Actions())),
		Relations:           extraction.Relations(),
		PatientDescriptions: []patient.PatientDescription{},
	}
	entityStore, err := self.store.MedicalEntities()
	if err != nil {
		return resolved, err
	}
	sourceID := state.ExplanationSourceID
	if sourceID.ToI64() <= 0 {
		sourceID = state.Case.SourceID
	}
	type infusionActor struct {
		name    string
		sources []source.Source
	}
	infusionActors := make([]infusionActor, 0)
	draftSources := func(draftSourceID source.SourceId) []source.Source {
		if draftSourceID.ToI64() <= 0 {
			draftSourceID = sourceID
		}
		if draftSourceID.ToI64() <= 0 {
			return nil
		}
		return []source.Source{{SourceID: draftSourceID}}
	}

	for index, draft := range extraction.Actors() {
		actor, err := presolve.Actor(draft)
		if err != nil {
			clog.Warn("case extraction actor skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			continue
		}
		if actor == nil {
			continue
		}
		sources := draftSources(sourceID)
		if _, err := entityStore.InsertActor(actor, nil, sourceIDs(sources)); err != nil {
			clog.Warn("case extraction actor skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			continue
		}
		if strings.Contains(strings.TrimSpace(actor.ActorName()), "투여") {
			infusionActors = append(infusionActors, infusionActor{
				name:    strings.TrimSpace(actor.ActorName()),
				sources: sources,
			})
		}
	}

	for index, draft := range extraction.Statuses() {
		draftSourceID := draft.SourceID
		if draftSourceID.ToI64() <= 0 {
			draftSourceID = sourceID
		}
		status, err := llmparser.StatusFromDTO(draft.Value)
		if err != nil {
			clog.Warn("case extraction status skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			resolved.Statuses = append(resolved.Statuses, nil)
			continue
		}
		if draftSourceID.ToI64() > 0 {
			if _, err := entityStore.InsertStatus(status, sourceIDs(draftSources(draftSourceID))); err != nil {
				clog.Warn("case extraction status save failed", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			}
		}
		resolved.Statuses = append(resolved.Statuses, status)
	}

	for _, descriptionDraft := range patientDrafts {
		description := patient.PatientDescription{
			Choice:    strings.TrimSpace(descriptionDraft.Choice),
			Knowledge: trim_string_list(descriptionDraft.Knowledge),
		}
		for index, draft := range descriptionDraft.Description {
			draftSourceID := draft.SourceID
			if draftSourceID.ToI64() <= 0 {
				draftSourceID = descriptionDraft.SourceID
			}
			if draftSourceID.ToI64() <= 0 {
				draftSourceID = sourceID
			}
			status, err := llmparser.StatusFromDTO(draft.Value)
			if err != nil {
				clog.Warn("case description status skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
				continue
			}
			if draftSourceID.ToI64() > 0 {
				if _, err := entityStore.InsertStatus(status, sourceIDs(draftSources(draftSourceID))); err != nil {
					clog.Warn("case description status save failed", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
				}
			}
			description.Description = append(description.Description, status)
		}
		if len(description.Description) > 0 || len(description.Knowledge) > 0 {
			resolved.PatientDescriptions = append(resolved.PatientDescriptions, description)
		}
	}

	for index, draft := range extraction.Actions() {
		value, err := llmparser.ActionFromDTO(draft.Value)
		if err != nil {
			clog.Warn("case extraction action skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			resolved.Actions = append(resolved.Actions, nil)
			continue
		}
		sources := draftSources(draft.SourceID)
		if _, err := entityStore.InsertAction(value, nil, sourceIDs(sources)); err != nil {
			clog.Warn("case extraction action save failed", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
		}
		resolved.Actions = append(resolved.Actions, value)
	}

	for index, draft := range extraction.MedicalProcedureOptions() {
		indications := make([]repo.IEStatus, 0, len(draft.Indications))
		for _, statusDTO := range draft.Indications {
			if status, err := llmparser.StatusFromDTO(statusDTO); err == nil {
				indications = append(indications, status)
			}
		}
		if len(indications) == 0 {
			for _, status := range resolved.Statuses {
				if status != nil {
					indications = append(indications, status)
				}
			}
		}
		effects := make([]repo.IEStatus, 0, len(draft.ExpectedEffects))
		for _, statusDTO := range draft.ExpectedEffects {
			if status, err := llmparser.StatusFromDTO(statusDTO); err == nil {
				effects = append(effects, status)
			}
		}
		optionSources := draftSources(draft.SourceID)
		procedure, err := llmparser.ProcedureFromDTO(&draft.MedicalProcedure)
		if err != nil {
			clog.Warn("case extraction procedure skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			continue
		}
		if err := self.store.SaveProcedure(procedure, optionSources); err != nil {
			clog.Warn("case extraction procedure save failed", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			continue
		}
		saved, err := self.store.SaveMedicalProcedureOptionByValues(
			procedure,
			draft.Category,
			indications,
			effects,
			optionSources,
		)
		if err != nil {
			clog.Warn("case extraction option skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("index", index), clog.AttrOf("error", err))
			continue
		}
		resolved.MedicalProcedureOptions = append(resolved.MedicalProcedureOptions, saved)
	}

	for _, actor := range infusionActors {
		if err := presolve.EnsureInfusionOption(self.store, actor.name, actor.sources); err != nil {
			clog.Warn("case extraction infusion option skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("name", actor.name), clog.AttrOf("error", err))
		}
	}

	for _, relation := range extraction.Relations() {
		status, statusOK := statusAt(resolved.Statuses, relation.CauseStatusIndex)
		action, actionOK := actionAt(resolved.Actions, relation.ActionIndex)
		result, resultOK := statusAt(resolved.Statuses, relation.ResultStatusIndex)
		relationSourceID := relation.SourceID
		if relationSourceID.ToI64() <= 0 {
			relationSourceID = sourceID
		}
		if statusOK && actionOK {
			if err := entityStore.LinkStatusAction(status, action, relationSourceID); err != nil {
				clog.Warn("case extraction status-action link skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("error", err))
			}
		}
		if actionOK && resultOK {
			if err := self.store.LinkActionResultStatusValues(action, result, relationSourceID); err != nil {
				clog.Warn("case extraction action-result link skip", clog.AttrOf("case", state.Case.ID), clog.AttrOf("error", err))
			}
		}
	}

	return resolved, nil
}

func statusAt(values []repo.IEStatus, index int) (repo.IEStatus, bool) {
	if index < 0 || index >= len(values) || values[index] == nil {
		return nil, false
	}
	return values[index], true
}

func actionAt(values []repo.IEAction, index int) (repo.IEAction, bool) {
	if index < 0 || index >= len(values) || values[index] == nil {
		return nil, false
	}
	return values[index], true
}
func (self *Service) load_case_extraction(id solve.CaseID, state *pipelineState) error {
	var output caseExtractionStepOutput
	if err := self.load_step_output(id, NodeCaseExtraction, &output); err != nil {
		return err
	}
	state.Extraction = output.CaseExtraction
	state.QuestionType = solve.NormalizeQuestionType(output.QuestionType)
	if state.QuestionType == "" {
		state.QuestionType = solve.NormalizeQuestionType(state.Case.QuestionType)
	}
	return nil
}

func (self *Service) load_resolve(id solve.CaseID, state *pipelineState) error {
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
