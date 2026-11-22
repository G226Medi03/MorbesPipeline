package store

import (
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/model"
)

func (s *SolveService) AppendLLMCall(call model.LLMCallInput) (int64, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Call.Append(s.store.Context(), s.store.SolveExecutor(), call)
}

func (s *SolveService) ListLLMCalls(caseID solve.CaseID) ([]model.LLMCall, error) {
	return s.store.SolveRepo().Call.List(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) DeleteLLMCallsByStage(caseID solve.CaseID, stage string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Call.DeleteByStage(s.store.Context(), s.store.SolveExecutor(), caseID, stage)
}
