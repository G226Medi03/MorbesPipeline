package store

import (
	"context"

	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"

	source "github.com/G226Medi03/SourceEntity"
)

// ExamQuestionDataCounts summarizes data linked to one exam question.
type ExamQuestionDataCounts = storerepo.ExamQuestionDataCounts

func (s *StoreContainer) CountExamQuestionData(ctx context.Context, examQuestionID int64) (ExamQuestionDataCounts, error) {
	var result ExamQuestionDataCounts
	executor := s.solveExecutor()
	llmResultCount, err := s.Benchmark.CountByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	responseIDs, err := s.Benchmark.ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	solveCaseCount, err := s.Solve.Case.CountByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	return ExamQuestionDataCounts{
		LLMResultCount:   llmResultCount,
		LLMResponseCount: len(responseIDs),
		SolveCaseCount:   solveCaseCount,
	}, nil
}

func (s *StoreContainer) DeleteExamQuestionData(ctx context.Context, examQuestionID int64) error {
	responseIDs := []source.RespId{}
	err := s.WithSolveTx(ctx, func(tx *StoreContainer) error {
		executor := tx.solveExecutor()
		caseIDs, err := s.Solve.Case.ListIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		if len(caseIDs) > 0 {
			if err := s.Solve.AnswerCheck.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := s.Solve.Step.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := s.Solve.Call.DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := s.Solve.Case.DeleteByIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
		}
		responseIDs, err = s.Benchmark.ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		return s.Benchmark.DeleteByExamQuestionID(ctx, executor, examQuestionID)
	})
	if err != nil {
		return err
	}
	if len(responseIDs) == 0 {
		return nil
	}
	return s.Source.DeleteLlmResponses(ctx, s.sourceExecutor(), responseIDs)
}
