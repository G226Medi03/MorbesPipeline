package solve

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
)

func (s *SolveService) ResetSteps(caseID domain.CaseID) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Step.Reset(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) StartStep(caseID domain.CaseID, node string, promptVersion int) (int64, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Step.Start(s.store.Context(), s.store.SolveExecutor(), caseID, node, promptVersion)
}

func (s *SolveService) FinishStep(id int64, state, outputJSON, message string, callID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Step.Finish(s.store.Context(), s.store.SolveExecutor(), id, state, outputJSON, message, callID)
}

func (s *SolveService) ListSteps(caseID domain.CaseID) ([]domain.Step, error) {
	return s.store.SolveRepository().Step.List(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) GetStep(caseID domain.CaseID, node string) (domain.Step, bool, error) {
	return s.store.SolveRepository().Step.Get(s.store.Context(), s.store.SolveExecutor(), caseID, node)
}

func (s *SolveService) MarkStaleFrom(caseID domain.CaseID, fromSeq int) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Step.MarkStaleFrom(s.store.Context(), s.store.SolveExecutor(), caseID, fromSeq)
}

func (s *SolveService) MarkDownstreamStale(caseID domain.CaseID, fromNode string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Step.MarkDownstreamStale(s.store.Context(), s.store.SolveExecutor(), caseID, fromNode)
}
