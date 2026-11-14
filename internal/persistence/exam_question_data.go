package store

import (
	"context"

	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	storesource "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentityrepository"
	benchmarkresult "github.com/G226Medi03/MorbesPipeline/internal/persistence/benchmark"
	storesolve "github.com/G226Medi03/MorbesPipeline/internal/persistence/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

// ExamQuestionDataCounts summarizes data linked to one exam question.
type ExamQuestionDataCounts struct {
	LLMResultCount   int
	LLMResponseCount int
	SolveCaseCount   int
}

func (s *Store) CountExamQuestionData(ctx context.Context, examQuestionID ids.ExamQuestionID) (ExamQuestionDataCounts, error) {
	var result ExamQuestionDataCounts
	executor := s.solveExecutor()
	llmResultCount, err := (benchmarkresult.Table{}).CountByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	responseIDs, err := (benchmarkresult.Table{}).ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	solveCaseCount, err := (storesolve.CaseTable{}).CountByExamQuestionID(ctx, executor, examQuestionID)
	if err != nil {
		return result, err
	}
	return ExamQuestionDataCounts{
		LLMResultCount:   llmResultCount,
		LLMResponseCount: len(responseIDs),
		SolveCaseCount:   solveCaseCount,
	}, nil
}

func (s *Store) DeleteExamQuestionData(ctx context.Context, examQuestionID ids.ExamQuestionID) error {
	responseIDs := []source.RespId{}
	err := s.WithSolveTx(ctx, func(tx *Store) error {
		executor := tx.solveExecutor()
		caseIDs, err := (storesolve.CaseTable{}).ListIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		if len(caseIDs) > 0 {
			if err := (storesolve.AnswerCheckTable{}).DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := (storesolve.StepTable{}).DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := (storesolve.CallTable{}).DeleteByCaseIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
			if err := (storesolve.CaseTable{}).DeleteByIDs(ctx, executor, caseIDs); err != nil {
				return err
			}
		}
		responseIDs, err = (benchmarkresult.Table{}).ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		return (benchmarkresult.Table{}).DeleteByExamQuestionID(ctx, executor, examQuestionID)
	})
	if err != nil {
		return err
	}
	if len(responseIDs) == 0 {
		return nil
	}
	return (storesource.LlmResponseTable{}).DeleteByIDs(ctx, s.sourceExecutor(), responseIDs)
}
