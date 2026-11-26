package store

import (
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/model"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *SolveService) AppendLLMCall(call model.LLMCallInput) (int64, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Call.Append(s.store.Context(), s.store.SolveExecutor(), call)
}

func (s *SolveService) ListLLMCalls(caseID solve.CaseID) ([]model.LLMCall, error) {
	calls, err := s.store.SolveRepo().Call.List(s.store.Context(), s.store.SolveExecutor(), caseID)
	if err != nil {
		return nil, err
	}
	if len(calls) == 0 {
		return calls, nil
	}
	ids := make([]source.RespId, 0, len(calls))
	for _, call := range calls {
		if call.RespID.ToI64() > 0 {
			ids = append(ids, call.RespID)
		}
	}
	if len(ids) == 0 {
		return calls, nil
	}
	responses, err := s.store.SourceRepo().GetLlmResponses(s.store.Context(), s.store.SourceExecutor(), ids)
	if err != nil {
		return nil, err
	}
	for index := range calls {
		response, found := responses[calls[index].RespID]
		if !found {
			continue
		}
		calls[index].Model = response.Model
		calls[index].Prompt = response.Prompt
		calls[index].RawOutput = string(response.Output)
	}
	return calls, nil
}

func (s *SolveService) DeleteLLMCallsByStage(caseID solve.CaseID, stage string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Call.DeleteByStage(s.store.Context(), s.store.SolveExecutor(), caseID, stage)
}
