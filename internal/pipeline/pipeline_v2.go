package csolve

import (
	"context"
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"

	"github.com/G226Medi03/MedicalEntityParser/parser"
	"github.com/G226Medi03/MorbesPipeline/internal/support/logs"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/patient"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
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
		actor, err := self.resolve_actor(draft)
		if err != nil {
			logs.Warn("case extraction actor skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
			continue
		}
		if actor == nil {
			continue
		}
		sources := draft_sources(draft.SourceID)
		if _, err := entityStore.InsertActor(actor, nil, sourceIDs(sources)); err != nil {
			logs.Warn("case extraction actor skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
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
		status, err := parseStatusDraft(draft)
		if err != nil {
			panic(fmt.Errorf("status parse failed key=%s kind=%s value=%s source=%d: %w", draft.Key, draft.Kind, string(draft.Value), draft_source_id, err))
		}
		if draft_source_id.ToI64() > 0 {
			if _, err := entityStore.InsertStatus(status, sourceIDs(draft_sources(draft_source_id))); err != nil {
				panic(fmt.Errorf("status save failed key=%s kind=%s value=%s source=%d: %w", draft.Key, draft.Kind, string(draft.Value), draft_source_id, err))
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
			status, err := parseStatusDraft(draft)
			if err != nil {
				panic(fmt.Errorf("status parse failed key=%s kind=%s value=%s source=%d: %w", draft.Key, draft.Kind, string(draft.Value), draft_source_id, err))
			}
			if draft_source_id.ToI64() > 0 {
				if _, err := entityStore.InsertStatus(status, sourceIDs(draft_sources(draft_source_id))); err != nil {
					panic(fmt.Errorf("status save failed key=%s kind=%s value=%s source=%d: %w", draft.Key, draft.Kind, string(draft.Value), draft_source_id, err))
				}
			}
			resolved.Statuses[draft.Key] = status
		}
	}

	for _, draft := range state.Extraction.Actions {
		variant := draft.Variant
		if strings.TrimSpace(variant) == "" {
			variant = "cause_status"
		}
		normalized, norm_err := parser.NormalizeActionValue(variant, draft.Value)
		if norm_err != nil {
			logs.Warn("case extraction action skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", norm_err))
			continue
		}
		encoded, err := json.Marshal(normalized)
		if err != nil {
			continue
		}
		value, err := parseActionDraft(ActionDraft{Key: draft.Key, Variant: variant, Value: encoded})
		if err != nil {
			logs.Warn("case extraction action skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
			continue
		}
		sources := draft_sources(draft.SourceID)
		if _, err := entityStore.InsertAction(value, nil, sourceIDs(sources)); err != nil {
			logs.Warn("case extraction action skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
			continue
		}
		resolved.Actions[draft.Key] = value
	}

	for _, draft := range state.Extraction.MedicalProcedureOptions {
		indications := make([]repo.IEStatus, 0, len(draft.Indications))
		for _, key := range draft.Indications {
			if status, ok := resolved.Statuses[key]; ok {
				indications = append(indications, status)
			}
		}
		if len(indications) == 0 {
			for _, status := range resolved.Statuses {
				indications = append(indications, status)
			}
		}
		effects := make([]repo.IEStatus, 0, len(draft.ExpectedEffects))
		for _, key := range draft.ExpectedEffects {
			if status, ok := resolved.Statuses[key]; ok {
				effects = append(effects, status)
			}
		}
		option_sources := draft_sources(draft.SourceID)
		procedure, err := repo.UnmarshalOrParseProcedure(draft.MedicalProcedure)
		if err != nil {
			logs.Warn("case extraction procedure skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
			continue
		}
		optionTraits := trait.NormalizeTraits(draft.Traits)
		procedureID, err := self.store.SaveProcedure(procedure, trait.IntrinsicTraits(optionTraits), option_sources)
		if err != nil {
			logs.Warn("case extraction procedure skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
			continue
		}
		saved, err := self.store.SaveMedicalProcedureOptionByValues(
			procedureID,
			draft.Category,
			trait.IntrinsicTraits(optionTraits),
			trait.ContextTraits(optionTraits),
			indications,
			effects,
			option_sources,
		)
		if err != nil {
			logs.Warn("case extraction option skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("key", draft.Key), logs.AttrOf("error", err))
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
			if err := entityStore.LinkStatusAction(status, action, relation_source_id); err != nil {
				logs.Warn("case extraction status-action link skip", logs.AttrOf("case", state.Case.ID), logs.AttrOf("error", err))
			}
		}
		if action_ok && result_ok {
			if err := self.store.LinkActionResultStatusValues(action, result, relation_source_id); err != nil {
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
				status, err = parseStatusDraft(draft)
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

func (self *Service) resolve_actor(draft ActorDraft) (repo.IEActor, error) {
	name := strings.TrimSpace(draft.Name)
	if name == "" {
		return nil, nil
	}
	if repo.ParseActorKind(draft.Kind) == repo.ActorKindDevice {
		return repo.NewDevice(name), nil
	}
	return repo.NewUnknownActor(name), nil
}

func (self *Service) ensure_infusion_option(name string, sources []source.SourceRef) error {
	name = strings.TrimSpace(name)
	if name == "" || len(sources) == 0 {
		return nil
	}
	procedure := repo.Infuse{
		Material: repo.NewUnknownActor(name),
		Route:    repo.InfuseRouteUnknown,
	}
	procedureID, err := self.store.SaveProcedure(procedure, nil, sources)
	if err != nil {
		return err
	}
	_, err = self.store.SaveMedicalProcedureOptionByValues(
		procedureID,
		"treatment",
		nil,
		nil,
		nil,
		nil,
		sources,
	)
	return err
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

func sourceIDs(refs []source.SourceRef) []source.SourceId {
	result := make([]source.SourceId, 0, len(refs))
	for _, ref := range refs {
		result = append(result, ref.SourceID)
	}
	return result
}
