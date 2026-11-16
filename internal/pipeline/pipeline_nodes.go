package csolve

import (
	"context"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/external/pipelinecore"
)

const (
	solveStateKey  = "solve_state"
	solveFromKey   = "solve_from_node"
	solveIDKey     = "solve_case_id"
	solveLoadNode  = core.NodeName("CaseLoadNode")
	explainNode    = core.NodeName("CaseExplainNode")
	extractionNode = core.NodeName("LLMCaseExtractionNode")
	resolveNode    = core.NodeName("DraftResolveNode")
	causeNode      = core.NodeName("GraphCauseTraceNode")
)

func (self *Service) runPipelineNodes(ctx context.Context, id domain.CaseID, fromNode string) error {
	pipeline := newSolveNodePipeline(self)
	state := core.NewState()
	state.Set(solveIDKey, id)
	state.Set(solveFromKey, fromNode)
	start := nodeForStage(fromNode)
	if start == "" {
		return pipeline.Run(ctx, state)
	}
	return pipeline.RunFrom(ctx, state, start)
}

func newSolveNodePipeline(runtime *Service) core.Pipeline {
	return core.Pipeline{
		Name: "SolvePipeline",
		Nodes: []core.Node{
			core.NewNode(solveLoadNode, "prepare", core.KindPrepare, core.CacheInherit, runtime.caseLoadNode, false),
			core.NewNode(explainNode, NodeExplain, core.KindLLM, core.CacheInherit, runtime.caseExplainNode, false),
			core.NewNode(extractionNode, NodeCaseExtraction, core.KindLLM, core.CacheInherit, runtime.caseExtractionNode, false),
			core.NewNode(resolveNode, NodeResolve, core.KindResolve, core.CacheInherit, runtime.caseResolveNode, false),
			core.NewNode(causeNode, NodeCause, core.KindGraph, core.CacheInherit, runtime.caseCauseNode, false),
		},
	}
}

func nodeForStage(stage string) core.NodeName {
	switch domain.Stage(stage) {
	case "":
		return ""
	case domain.StageExplain:
		return explainNode
	case domain.StageCaseExtraction:
		return extractionNode
	case domain.StageResolve:
		return resolveNode
	case domain.StageCause:
		return causeNode
	default:
		return core.NodeName(stage)
	}
}

func solveState(state *core.State) (*pipelineState, domain.CaseID, string, error) {
	raw, ok := state.Get(solveStateKey)
	if !ok {
		return nil, "", "", fmt.Errorf("solve state가 없습니다")
	}
	solve, ok := raw.(*pipelineState)
	if !ok {
		return nil, "", "", fmt.Errorf("solve state 타입이 올바르지 않습니다: %T", raw)
	}
	idRaw, _ := state.Get(solveIDKey)
	id, _ := idRaw.(domain.CaseID)
	fromRaw, _ := state.Get(solveFromKey)
	from, _ := fromRaw.(string)
	return solve, id, from, nil
}

func (self *Service) caseLoadNode(ctx context.Context, state *core.State) error {
	id, _ := state.Get(solveIDKey)
	caseID, _ := id.(domain.CaseID)
	fromRaw, _ := state.Get(solveFromKey)
	fromNode, _ := fromRaw.(string)

	item, found, err := self.store.Get(caseID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("solve case 를 찾지 못했습니다: %s", caseID)
	}
	if item.ExamQuestionID > 0 {
		if self.question_store == nil {
			return fmt.Errorf("exam question store 가 없습니다")
		}
		question, questionFound, questionErr := self.question_store.GetQuestion(ctx, item.ExamQuestionID)
		if questionErr != nil {
			return questionErr
		}
		if !questionFound {
			return fmt.Errorf("exam question 을 찾지 못했습니다: %d", item.ExamQuestionID)
		}
		item.Problem = question.Problem
		item.Answer = question.Answer
		item.Explanation = question.Explanation
	}
	prompts, err := self.load_prompts()
	if err != nil {
		_ = self.store.SetFailed(caseID, err)
		return err
	}
	if self.llm == nil {
		return fmt.Errorf("extraction LLM 이 없습니다")
	}
	if fromNode == "" {
		if err := self.store.ResetSteps(caseID); err != nil {
			return err
		}
	} else {
		if !domain.IsStage(fromNode) {
			return fmt.Errorf("알 수 없는 resume node 입니다: %s", fromNode)
		}
		if err := self.store.MarkDownstreamStale(caseID, fromNode); err != nil {
			return err
		}
	}
	if err := self.store.SetRunning(caseID); err != nil {
		return err
	}
	state.Set(solveStateKey, &pipelineState{Case: item, Prompts: prompts})
	return nil
}

func (self *Service) caseExplainNode(ctx context.Context, state *core.State) error {
	solve, id, fromNode, err := solveState(state)
	if err != nil {
		return err
	}
	if fromNode != "" && StepSeq(fromNode) > StepSeq(NodeExplain) {
		return self.load_explain(id, solve)
	}
	if strings.TrimSpace(solve.Case.Explanation) != "" {
		if err := self.record_provided_explanation(id, solve); err != nil {
			_ = self.store.SetFailed(id, err)
			return err
		}
		return nil
	}
	if err := self.run_explain_node(ctx, id, solve); err != nil {
		_ = self.store.SetFailed(id, err)
		return err
	}
	return nil
}

func (self *Service) caseExtractionNode(ctx context.Context, state *core.State) error {
	solve, id, fromNode, err := solveState(state)
	if err != nil {
		return err
	}
	if fromNode != "" && StepSeq(fromNode) > StepSeq(NodeCaseExtraction) {
		return self.load_case_extraction(id, solve)
	}
	if err := self.run_case_extraction_node(ctx, id, solve); err != nil {
		_ = self.store.SetFailed(id, err)
		return err
	}
	return nil
}

func (self *Service) caseResolveNode(ctx context.Context, state *core.State) error {
	solve, id, fromNode, err := solveState(state)
	if err != nil {
		return err
	}
	if fromNode != "" && StepSeq(fromNode) > StepSeq(NodeResolve) {
		return self.load_resolve(id, solve)
	}
	if err := self.run_resolve_node(ctx, id, solve); err != nil {
		_ = self.store.SetFailed(id, err)
		return err
	}
	return nil
}

func (self *Service) caseCauseNode(ctx context.Context, state *core.State) error {
	solve, id, fromNode, err := solveState(state)
	if err != nil {
		return err
	}
	if fromNode != "" && StepSeq(fromNode) > StepSeq(NodeCause) {
		return self.load_cause(id, solve)
	}
	if err := self.run_cause_node(ctx, id, solve); err != nil {
		_ = self.store.SetFailed(id, err)
		return err
	}
	return nil
}
