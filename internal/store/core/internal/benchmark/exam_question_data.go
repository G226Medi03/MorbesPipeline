package benchmark

import (
	"context"

	"github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/deps"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"

	source "github.com/G226Medi03/SourceEntity"
)

// ExamQuestionDataCounts summarizes data linked to one exam question.
type ExamQuestionDataCounts = storerepo.ExamQuestionDataCounts

func (s *BenchmarkService) CountExamQuestionData(ctx context.Context, examQuestionID int64) (ExamQuestionDataCounts, error) {
	var result ExamQuestionDataCounts
	executor := s.store.SolveExecutor()
	llmResultCount, err := s.store.BenchmarkRepository().CountByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	responseIDs, err := s.store.BenchmarkRepository().ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	solveCaseCount, err := s.store.SolveRepository().Case.CountByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	return ExamQuestionDataCounts{
		LLMResultCount:   llmResultCount,
		LLMResponseCount: len(responseIDs),
		SolveCaseCount:   solveCaseCount,
	}, nil
}

func (s *BenchmarkService) DeleteExamQuestionData(ctx context.Context, examQuestionID int64) error {
	responseIDs := []source.RespId{}
	err := s.store.RunSolveTx(ctx, func(tx deps.Store) error {
		executor := tx.SolveExecutor()
		caseIDs, err := tx.SolveRepository().Case.ListIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		if len(caseIDs) > 0 {
			if err := tx.SolveRepository().AnswerCheck.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := tx.SolveRepository().Step.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := tx.SolveRepository().Call.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := tx.SolveRepository().Case.DeleteByIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
		}
		responseIDs, err = tx.BenchmarkRepository().ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		return tx.BenchmarkRepository().DeleteByExamQuestionID(ctx, executor, examQuestionID)
	})
	if err != nil {
		return err
	}
	if len(responseIDs) == 0 {
		return nil
	}
	return s.store.SourceRepository().DeleteLlmResponses(ctx, s.store.SourceExecutor(), responseIDs)
}
