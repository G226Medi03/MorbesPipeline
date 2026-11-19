package solve

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
)

func (s *SolveService) AppendLLMCall(call model.LLMCallInput) (int64, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Call.Append(s.store.Context(), s.store.SolveExecutor(), call)
}

func (s *SolveService) ListLLMCalls(caseID domain.CaseID) ([]model.LLMCall, error) {
	return s.store.SolveRepository().Call.List(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) DeleteLLMCallsByStage(caseID domain.CaseID, stage string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Call.DeleteByStage(s.store.Context(), s.store.SolveExecutor(), caseID, stage)
}
