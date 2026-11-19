package store

import (
	"context"

	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
)

type (
	ProblemResultFilter    = storerepo.ProblemResultFilter
	ProblemResultRow       = storerepo.ProblemResultRow
	ExamQuestionDataCounts = storerepo.ExamQuestionDataCounts
)

func (s *StoreContainer) ensure_llm_problem_results() error {
	return s.Benchmark.Init(s.ctx, s.solveExecutor())
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
