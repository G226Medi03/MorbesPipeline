package csolve

import (
	"context"

	clog "github.com/G226Medi03/log-sqlite"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/PipeLineCore"
)

const (
	extractQuestionTypeNode = core.NodeName("LLMQuestionTypeNode")
	extractParallelNode     = core.NodeName("LLMParallelExtractionNode")
	extractPatientNode      = core.NodeName("LLMPatientDescriptionNode")
	extractRelationNode     = core.NodeName("LLMRelationNode")
	extractOptionNode       = core.NodeName("LLMOptionExtractionNode")
	extractAssembleNode     = core.NodeName("CaseExtractionAssembleNode")

	extractActorsKey    = "extract_actors"
	extractStatusesKey  = "extract_statuses"
	extractActionsKey   = "extract_actions"
	extractPatientsKey  = "extract_patients"
	extractRelationsKey = "extract_relations"
	extractOptionsKey   = "extract_options"
)

func (self *Service) extractionPipeline(ctx context.Context, id domain.CaseID, state *pipelineState) error {
	pipeline := core.Pipeline{
		Name: "CaseExtractionPipeline",
		Nodes: []core.Node{
			core.NewNode(extractQuestionTypeNode, NodeCaseExtraction, core.KindLLM, core.CacheInherit, self.extractQuestionTypeNode, false),
			core.NewNode(extractParallelNode, NodeCaseExtraction, core.KindLLM, core.CacheInherit, self.extractParallelNode, false),
			core.NewNode(extractPatientNode, NodeCaseExtraction, core.KindLLM, core.CacheInherit, self.extractPatientNode, true),
			core.NewNode(extractRelationNode, NodeCaseExtraction, core.KindLLM, core.CacheInherit, self.extractRelationNode, false),
			core.NewNode(extractOptionNode, NodeCaseExtraction, core.KindLLM, core.CacheInherit, self.extractOptionNode, false),
			core.NewNode(extractAssembleNode, NodeCaseExtraction, core.KindTransform, core.CacheInherit, self.extractAssembleNode, false),
		},
	}
	pipelineState := core.NewState()
	pipelineState.Set(solveIDKey, id)
	pipelineState.Set(solveStateKey, state)
	return pipeline.Run(ctx, pipelineState)
}

func extractionSolveState(state *core.State) (*pipelineState, domain.CaseID, error) {
	solve, id, _, err := solveState(state)
	return solve, id, err
}

func (self *Service) extractQuestionTypeNode(ctx context.Context, state *core.State) error {
	solve, id, err := extractionSolveState(state)
	if err != nil {
		return err
	}
	self.run_question_type(ctx, id, solve)
	return nil
}

func (self *Service) extractParallelNode(ctx context.Context, state *core.State) error {
	solve, id, err := extractionSolveState(state)
	if err != nil {
		return err
	}
	var actors []ActorDraft
	var statuses []StatusDraft
	var actions []ActionDraft
	var actorErr, statusErr, actionErr error
	run_parallel(
		func() { actors, actorErr = self.run_v2_actor_extraction(ctx, id, solve) },
		func() { statuses, statusErr = self.run_v2_status_extraction(ctx, id, solve) },
		func() { actions, actionErr = self.run_v2_action_extraction(ctx, id, solve) },
	)
	if statusErr != nil {
		return statusErr
	}
	if actionErr != nil {
		return actionErr
	}
	if actorErr != nil {
		clog.Warn("actor extraction skip", clog.AttrOf("case", id), clog.AttrOf("error", actorErr))
		actors = nil
	}
	state.Set(extractActorsKey, actors)
	state.Set(extractStatusesKey, statuses)
	state.Set(extractActionsKey, actions)
	return nil
}

func (self *Service) extractPatientNode(ctx context.Context, state *core.State) error {
	solve, id, err := extractionSolveState(state)
	if err != nil {
		return err
	}
	statuses, _ := state.Get(extractStatusesKey)
	descriptions, err := self.run_v2_patient_description_extraction(ctx, id, solve, statuses.([]StatusDraft))
	if err != nil {
		clog.Warn("patient description extraction skip", clog.AttrOf("case", id), clog.AttrOf("error", err))
		descriptions = nil
	}
	state.Set(extractPatientsKey, descriptions)
	return nil
}

func (self *Service) extractRelationNode(ctx context.Context, state *core.State) error {
	solve, id, err := extractionSolveState(state)
	if err != nil {
		return err
	}
	statuses, _ := state.Get(extractStatusesKey)
	actions, _ := state.Get(extractActionsKey)
	relations, err := self.run_v2_relation_extraction(ctx, id, solve, statuses.([]StatusDraft), actions.([]ActionDraft))
	if err != nil {
		return err
	}
	state.Set(extractRelationsKey, relations)
	return nil
}

func (self *Service) extractOptionNode(ctx context.Context, state *core.State) error {
	solve, id, err := extractionSolveState(state)
	if err != nil {
		return err
	}
	statuses, _ := state.Get(extractStatusesKey)
	patients, _ := state.Get(extractPatientsKey)
	options, err := self.run_v2_medical_procedure_option_extraction(ctx, id, solve, statuses.([]StatusDraft), patients.([]PatientDescriptionDraft))
	if err != nil {
		return err
	}
	state.Set(extractOptionsKey, options)
	return nil
}

func (self *Service) extractAssembleNode(_ context.Context, state *core.State) error {
	solve, _, err := extractionSolveState(state)
	if err != nil {
		return err
	}
	actors, _ := state.Get(extractActorsKey)
	statuses, _ := state.Get(extractStatusesKey)
	actions, _ := state.Get(extractActionsKey)
	relations, _ := state.Get(extractRelationsKey)
	options, _ := state.Get(extractOptionsKey)
	patients, _ := state.Get(extractPatientsKey)
	extraction := CaseExtraction{
		Actors:                  actors.([]ActorDraft),
		Statuses:                statuses.([]StatusDraft),
		Actions:                 actions.([]ActionDraft),
		Relations:               relations.([]RelationDraft),
		MedicalProcedureOptions: options.([]OptionDraft),
		PatientDescriptions:     patients.([]PatientDescriptionDraft),
	}
	solve.Extraction = extraction
	return nil
}
