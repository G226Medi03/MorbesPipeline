package store

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
)

func (s *StoreContainer) AppendLLMCall(call model.LLMCallInput) (int64, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Call.Append(s.ctx, s.solveExecutor(), call)
}

func (s *StoreContainer) ListLLMCalls(caseID domain.CaseID) ([]model.LLMCall, error) {
	return s.Solve.Call.List(s.ctx, s.solveExecutor(), caseID)
}

func (s *StoreContainer) DeleteLLMCallsByStage(caseID domain.CaseID, stage string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Call.DeleteByStage(s.ctx, s.solveExecutor(), caseID, stage)
}
