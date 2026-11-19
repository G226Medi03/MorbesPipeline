package store

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

func (s *StoreContainer) ensure_llm_problem_results() error {
	return s.Benchmark.Init(s.ctx, s.solveExecutor())
}

func (s *StoreContainer) GetProblemResultByID(ctx context.Context, id int64) (ProblemResultRow, bool, error) {
	row, found, err := s.Benchmark.GetByID(ctx, s.solveExecutor(), id)
	if err != nil || !found {
		return ProblemResultRow{}, found, err
	}
	responses, err := s.Source.GetLlmResponses(
		ctx, s.sourceExecutor(), []sourcedomain.RespId{sourcedomain.RespId(row.LLMResponseID)},
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

func (s *StoreContainer) ListProblemResults(ctx context.Context, filter ProblemResultFilter) ([]ProblemResultRow, error) {
	responseIDs, err := s.problem_result_response_ids(ctx, filter)
	if err != nil {
		return nil, err
	}
	if has_problem_result_response_filter(filter) && len(responseIDs) == 0 {
		return []ProblemResultRow{}, nil
	}
	rows, err := s.Benchmark.List(ctx, s.solveExecutor(), storerepo.BenchmarkFilter{
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

func (s *StoreContainer) CountProblemResults(ctx context.Context, filter ProblemResultFilter) (int, error) {
	responseIDs, err := s.problem_result_response_ids(ctx, filter)
	if err != nil {
		return 0, err
	}
	if has_problem_result_response_filter(filter) && len(responseIDs) == 0 {
		return 0, nil
	}
	return s.Benchmark.Count(ctx, s.solveExecutor(), storerepo.BenchmarkFilter{
		Correct:     filter.Correct,
		ResponseIDs: responseIDs,
	})
}

// ProblemResults returns a consistent result page and model list.
func (s *StoreContainer) ProblemResults(ctx context.Context, filter ProblemResultFilter) ([]ProblemResultRow, int, []string, error) {
	var rows []ProblemResultRow
	var total int
	var models []string
	err := s.WithTx(ctx, func(tx *StoreContainer) error {
		var txErr error
		rows, txErr = tx.ListProblemResults(ctx, filter)
		if txErr != nil {
			return txErr
		}
		total, txErr = tx.CountProblemResults(ctx, filter)
		if txErr != nil {
			return txErr
		}
		models, txErr = tx.ListProblemResultModels(ctx)
		return txErr
	})
	return rows, total, models, err
}

// ProblemResult returns one composed result row.
func (s *StoreContainer) ProblemResult(ctx context.Context, id int64) (ProblemResultRow, bool, error) {
	var row ProblemResultRow
	var found bool
	err := s.WithTx(ctx, func(tx *StoreContainer) error {
		var txErr error
		row, found, txErr = tx.GetProblemResultByID(ctx, id)
		return txErr
	})
	return row, found, err
}

func (s *StoreContainer) ListProblemResultModels(ctx context.Context) ([]string, error) {
	responseIDs, err := s.Benchmark.ListResponseIDs(ctx, s.solveExecutor())
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

func (s *StoreContainer) problem_result_response_ids(ctx context.Context, filter ProblemResultFilter) ([]sourcedomain.RespId, error) {
	model := strings.TrimSpace(filter.Model)
	var modelIDs []sourcedomain.RespId
	if model != "" {
		var err error
		modelIDs, err = s.Source.LlmResponseIDsByModel(ctx, s.sourceExecutor(), model)
		if err != nil {
			return nil, err
		}
	}
	var errorIDs []sourcedomain.RespId
	if filter.Error != nil {
		var err error
		errorIDs, err = s.Source.LlmResponseIDsByError(ctx, s.sourceExecutor(), *filter.Error)
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

func (s *StoreContainer) problem_result_responses(ctx context.Context, rows []storerepo.BenchmarkRow) (map[sourcedomain.RespId]sourcedomain.LlmResponseRow, error) {
	ids := make([]sourcedomain.RespId, 0, len(rows))
	seen := make(map[sourcedomain.RespId]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.LLMResponseID]; ok {
			continue
		}
		seen[row.LLMResponseID] = struct{}{}
		ids = append(ids, sourcedomain.RespId(row.LLMResponseID))
	}
	return s.Source.GetLlmResponses(ctx, s.sourceExecutor(), ids)
}

func (s *StoreContainer) problem_result_response_ids_map(ctx context.Context, ids []sourcedomain.RespId) (map[sourcedomain.RespId]sourcedomain.LlmResponseRow, error) {
	respIDs := make([]sourcedomain.RespId, 0, len(ids))
	for _, id := range ids {
		respIDs = append(respIDs, sourcedomain.RespId(id))
	}
	return s.Source.GetLlmResponses(ctx, s.sourceExecutor(), respIDs)
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
