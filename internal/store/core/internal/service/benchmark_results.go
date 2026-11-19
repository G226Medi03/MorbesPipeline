package service

import (
	"context"
	"sort"
	"strings"

	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
	sourcedomain "github.com/G226Medi03/SourceEntity"
)

type (
	ProblemResultFilter = storerepo.ProblemResultFilter
	ProblemResultRow    = storerepo.ProblemResultRow
)

func (s *BenchmarkService) ensure_llm_problem_results() error {
	return s.store.BenchmarkRepo().Init(s.store.Context(), s.store.SolveExecutor())
}

func (s *BenchmarkService) GetProblemResultByID(ctx context.Context, id int64) (ProblemResultRow, bool, error) {
	row, found, err := s.store.BenchmarkRepo().GetByID(ctx, s.store.SolveExecutor(), id)
	if err != nil || !found {
		return ProblemResultRow{}, found, err
	}
	responses, err := s.store.SourceRepo().GetLlmResponses(
		ctx, s.store.SourceExecutor(), []sourcedomain.RespId{sourcedomain.RespId(row.LLMResponseID)},
	)
	if err != nil {
		return ProblemResultRow{}, false, err
	}
	response, ok := responses[sourcedomain.RespId(row.LLMResponseID)]
	if !ok {
		return ProblemResultRow{}, false, nil
	}
	return composeProblemResult(row, response), true, nil
}

func (s *BenchmarkService) ListProblemResults(ctx context.Context, filter ProblemResultFilter) ([]ProblemResultRow, error) {
	responseIDs, err := s.problem_result_response_ids(ctx, filter)
	if err != nil {
		return nil, err
	}
	if has_problem_result_response_filter(filter) && len(responseIDs) == 0 {
		return []ProblemResultRow{}, nil
	}
	rows, err := s.store.BenchmarkRepo().List(ctx, s.store.SolveExecutor(), storerepo.BenchmarkFilter{
		Correct:     filter.Correct,
		ResponseIDs: responseIDs,
		Order:       filter.Order,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
	})
	if err != nil {
		return nil, err
	}
	responseMap, err := s.problem_result_responses(ctx, rows)
	if err != nil {
		return nil, err
	}
	result := make([]ProblemResultRow, 0, len(rows))
	for _, row := range rows {
		response, ok := responseMap[sourcedomain.RespId(row.LLMResponseID)]
		if !ok {
			continue
		}
		result = append(result, composeProblemResult(row, response))
	}
	return result, nil
}

func (s *BenchmarkService) CountProblemResults(ctx context.Context, filter ProblemResultFilter) (int, error) {
	responseIDs, err := s.problem_result_response_ids(ctx, filter)
	if err != nil {
		return 0, err
	}
	if has_problem_result_response_filter(filter) && len(responseIDs) == 0 {
		return 0, nil
	}
	return s.store.BenchmarkRepo().Count(ctx, s.store.SolveExecutor(), storerepo.BenchmarkFilter{
		Correct:     filter.Correct,
		ResponseIDs: responseIDs,
	})
}

func (s *BenchmarkService) ListProblemResultModels(ctx context.Context) ([]string, error) {
	responseIDs, err := s.store.BenchmarkRepo().ListResponseIDs(ctx, s.store.SolveExecutor())
	if err != nil {
		return nil, err
	}
	responseMap, err := s.problem_result_response_ids_map(ctx, responseIDs)
	if err != nil {
		return nil, err
	}
	modelSet := make(map[string]struct{})
	for _, response := range responseMap {
		model := strings.TrimSpace(response.Model)
		if model != "" {
			modelSet[model] = struct{}{}
		}
	}
	models := make([]string, 0, len(modelSet))
	for model := range modelSet {
		models = append(models, model)
	}
	sort.Strings(models)
	return models, nil
}

func (s *BenchmarkService) problem_result_response_ids(ctx context.Context, filter ProblemResultFilter) ([]sourcedomain.RespId, error) {
	model := strings.TrimSpace(filter.Model)
	var modelIDs []sourcedomain.RespId
	if model != "" {
		var err error
		modelIDs, err = s.store.SourceRepo().LlmResponseIDsByModel(ctx, s.store.SourceExecutor(), model)
		if err != nil {
			return nil, err
		}
	}
	var errorIDs []sourcedomain.RespId
	if filter.Error != nil {
		var err error
		errorIDs, err = s.store.SourceRepo().LlmResponseIDsByError(ctx, s.store.SourceExecutor(), *filter.Error)
		if err != nil {
			return nil, err
		}
	}

	if model == "" && filter.Error == nil {
		return nil, nil
	}
	if model != "" && filter.Error == nil {
		return resp_id_ints(modelIDs), nil
	}
	if model == "" && filter.Error != nil {
		return resp_id_ints(errorIDs), nil
	}

	// model + error: code-level intersection
	modelSet := make(map[sourcedomain.RespId]struct{}, len(modelIDs))
	for _, id := range modelIDs {
		modelSet[id] = struct{}{}
	}
	result := make([]sourcedomain.RespId, 0)
	for _, id := range errorIDs {
		if _, ok := modelSet[id]; ok {
			result = append(result, id)
		}
	}
	return result, nil
}

func resp_id_ints(ids []sourcedomain.RespId) []sourcedomain.RespId {
	return append([]sourcedomain.RespId(nil), ids...)
}

func has_problem_result_response_filter(filter ProblemResultFilter) bool {
	return strings.TrimSpace(filter.Model) != "" || filter.Error != nil
}

func (s *BenchmarkService) problem_result_responses(ctx context.Context, rows []storerepo.BenchmarkRow) (map[sourcedomain.RespId]sourcedomain.LlmResponseRow, error) {
	ids := make([]sourcedomain.RespId, 0, len(rows))
	seen := make(map[sourcedomain.RespId]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.LLMResponseID]; ok {
			continue
		}
		seen[row.LLMResponseID] = struct{}{}
		ids = append(ids, sourcedomain.RespId(row.LLMResponseID))
	}
	return s.store.SourceRepo().GetLlmResponses(ctx, s.store.SourceExecutor(), ids)
}

func (s *BenchmarkService) problem_result_response_ids_map(ctx context.Context, ids []sourcedomain.RespId) (map[sourcedomain.RespId]sourcedomain.LlmResponseRow, error) {
	respIDs := make([]sourcedomain.RespId, 0, len(ids))
	for _, id := range ids {
		respIDs = append(respIDs, sourcedomain.RespId(id))
	}
	return s.store.SourceRepo().GetLlmResponses(ctx, s.store.SourceExecutor(), respIDs)
}

func composeProblemResult(row storerepo.BenchmarkRow, response sourcedomain.LlmResponseRow) ProblemResultRow {
	createdAt := row.CreatedAt
	if !response.CreatedAt.IsZero() {
		createdAt = response.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return ProblemResultRow{
		ID:             row.ID,
		ExamQuestionID: row.ExamQuestionID,
		LLMResponseID:  row.LLMResponseID,
		Model:          response.Model,
		Prompt:         response.Prompt,
		Output:         response.Output,
		Correct:        row.Correct,
		CreatedAt:      createdAt,
	}
}
