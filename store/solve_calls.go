package store

import (
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
	"github.com/G226Medi03/MorbesPipeline/solve/model"
)

func (s *Store) AppendLLMCall(call model.LLMCallInput) (entity.LLMCallID, error) {
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
