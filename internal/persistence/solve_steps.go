package store

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

func (s *Store) ResetSteps(caseID domain.CaseID) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Step.Reset(s.ctx, s.solveExecutor(), caseID)
}

func (s *Store) StartStep(caseID domain.CaseID, node string, promptVersion int) (ids.SolveStepID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Step.Start(s.ctx, s.solveExecutor(), caseID, node, promptVersion)
}

func (s *Store) FinishStep(id ids.SolveStepID, state, outputJSON, message string, callID ids.LLMCallID) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Step.Finish(s.ctx, s.solveExecutor(), id, state, outputJSON, message, callID)
}

func (s *Store) ListSteps(caseID domain.CaseID) ([]domain.Step, error) {
	return s.SolveRepository.Step.List(s.ctx, s.solveExecutor(), caseID)
}

func (s *Store) GetStep(caseID domain.CaseID, node string) (domain.Step, bool, error) {
	return s.SolveRepository.Step.Get(s.ctx, s.solveExecutor(), caseID, node)
}

func (s *Store) MarkStaleFrom(caseID domain.CaseID, fromSeq int) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Step.MarkStaleFrom(s.ctx, s.solveExecutor(), caseID, fromSeq)
}

func (s *Store) MarkDownstreamStale(caseID domain.CaseID, fromNode string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Step.MarkDownstreamStale(s.ctx, s.solveExecutor(), caseID, fromNode)
}
