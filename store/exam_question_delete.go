package store

import (
	"context"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/store/benchmarkresult"
	storesolve "github.com/G226Medi03/MorbesPipeline/store/solve"
	sourcedomain "github.com/G226Medi03/SourceEntity"
	storesource "github.com/G226Medi03/SourceEntityRepository"
)

type ExamQuestionDataCounts struct {
	LLMResultCount   int
	LLMResponseCount int
	SolveCaseCount   int
}

func (s *Store) CountExamQuestionData(ctx context.Context, examQuestionID entity.ExamQuestionID) (ExamQuestionDataCounts, error) {
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
	result = ExamQuestionDataCounts{
		LLMResultCount:   llmResultCount,
		LLMResponseCount: len(responseIDs),
		SolveCaseCount:   solveCaseCount,
	}
	return result, nil
}

func (s *Store) DeleteExamQuestionData(ctx context.Context, examQuestionID entity.ExamQuestionID) error {
	responseIDs := []sourcedomain.RespId{}
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

		ids, err := (benchmarkresult.Table{}).ListResponseIDsByExamQuestionID(ctx, executor, examQuestionID)
		if err != nil {
			return err
		}
		responseIDs = ids
		if err := (benchmarkresult.Table{}).DeleteByExamQuestionID(ctx, executor, examQuestionID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(responseIDs) == 0 {
		return nil
	}
	return (storesource.LlmResponseTable{}).DeleteByIDs(ctx, s.sourceExecutor(), append([]sourcedomain.RespId(nil), responseIDs...))
}
