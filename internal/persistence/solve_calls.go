package store

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

func (s *Store) AppendLLMCall(call model.LLMCallInput) (ids.LLMCallID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Call.Append(s.ctx, s.solveExecutor(), call)
}

func (s *Store) ListLLMCalls(caseID domain.CaseID) ([]model.LLMCall, error) {
	return s.SolveRepository.Call.List(s.ctx, s.solveExecutor(), caseID)
}

func (s *Store) DeleteLLMCallsByStage(caseID domain.CaseID, stage string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Call.DeleteByStage(s.ctx, s.solveExecutor(), caseID, stage)
}
