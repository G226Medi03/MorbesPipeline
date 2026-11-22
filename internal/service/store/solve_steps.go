package store

import (
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
)

func (s *SolveService) ResetSteps(caseID solve.CaseID) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Step.Reset(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) StartStep(caseID solve.CaseID, node string, promptVersion int) (int64, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Step.Start(s.store.Context(), s.store.SolveExecutor(), caseID, node, promptVersion)
}

func (s *SolveService) FinishStep(id int64, state, outputJSON, message string, callID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Step.Finish(s.store.Context(), s.store.SolveExecutor(), id, state, outputJSON, message, callID)
}

func (s *SolveService) ListSteps(caseID solve.CaseID) ([]solve.Step, error) {
	return s.store.SolveRepo().Step.List(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) GetStep(caseID solve.CaseID, node string) (solve.Step, bool, error) {
	return s.store.SolveRepo().Step.Get(s.store.Context(), s.store.SolveExecutor(), caseID, node)
}

func (s *SolveService) MarkStaleFrom(caseID solve.CaseID, fromSeq int) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Step.MarkStaleFrom(s.store.Context(), s.store.SolveExecutor(), caseID, fromSeq)
}

func (s *SolveService) MarkDownstreamStale(caseID solve.CaseID, fromNode string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Step.MarkDownstreamStale(s.store.Context(), s.store.SolveExecutor(), caseID, fromNode)
}
