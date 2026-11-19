package store

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

func (s *StoreContainer) ResetSteps(caseID domain.CaseID) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Step.Reset(s.ctx, s.solveExecutor(), caseID)
}

func (s *StoreContainer) StartStep(caseID domain.CaseID, node string, promptVersion int) (ids.SolveStepID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Step.Start(s.ctx, s.solveExecutor(), caseID, node, promptVersion)
}

func (s *StoreContainer) FinishStep(id ids.SolveStepID, state, outputJSON, message string, callID ids.LLMCallID) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Step.Finish(s.ctx, s.solveExecutor(), id, state, outputJSON, message, callID)
}

func (s *StoreContainer) ListSteps(caseID domain.CaseID) ([]domain.Step, error) {
	return s.Solve.Step.List(s.ctx, s.solveExecutor(), caseID)
}

func (s *StoreContainer) GetStep(caseID domain.CaseID, node string) (domain.Step, bool, error) {
	return s.Solve.Step.Get(s.ctx, s.solveExecutor(), caseID, node)
}

func (s *StoreContainer) MarkStaleFrom(caseID domain.CaseID, fromSeq int) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Step.MarkStaleFrom(s.ctx, s.solveExecutor(), caseID, fromSeq)
}

func (s *StoreContainer) MarkDownstreamStale(caseID domain.CaseID, fromNode string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Step.MarkDownstreamStale(s.ctx, s.solveExecutor(), caseID, fromNode)
}
