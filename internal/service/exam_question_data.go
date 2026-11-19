package service

import (
	"context"

	corestore "github.com/G226Medi03/MorbesPipeline/internal/store/core/store"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"

	source "github.com/G226Medi03/SourceEntity"
)

// ExamQuestionDataCounts summarizes data linked to one exam question.
type ExamQuestionDataCounts = storerepo.ExamQuestionDataCounts

func (s *BenchmarkService) CountExamQuestionData(ctx context.Context, examQuestionID int64) (ExamQuestionDataCounts, error) {
	var result ExamQuestionDataCounts
	executor := s.store.SolveExecutor()
	llmResultCount, err := s.store.BenchmarkRepo().CountByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	responseIDs, err := s.store.BenchmarkRepo().ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	solveCaseCount, err := s.store.SolveRepo().Case.CountByExamQuestionID(ctx, executor, examQuestionID)
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
	err := s.store.WithSolveTx(ctx, func(tx *corestore.Store) error {
		executor := tx.SolveExecutor()
		caseIDs, err := tx.SolveRepo().Case.ListIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		if len(caseIDs) > 0 {
			if err := tx.SolveRepo().AnswerCheck.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := tx.SolveRepo().Step.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := tx.SolveRepo().Call.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := tx.SolveRepo().Case.DeleteByIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
		}
		responseIDs, err = tx.BenchmarkRepo().ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		return tx.BenchmarkRepo().DeleteByExamQuestionID(ctx, executor, examQuestionID)
	})
	if err != nil {
		return err
	}
	if len(responseIDs) == 0 {
		return nil
	}
	return s.store.SourceRepo().DeleteLlmResponses(ctx, s.store.SourceExecutor(), responseIDs)
}
