package store

import (
	"fmt"
	"github.com/G226Medi03/SourceEntity"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	solve "github.com/G226Medi03/MorbesPipeline/internal/store/solve"
)

func (s *StoreContainer) ResetCaseForReplay(id domain.CaseID, mode, fromNode string, keepCalls bool) error {
	unlock := s.lockWrite()
	defer unlock()
	seq := 0
	if fromNode != "" {
		seq = domain.Sequence(domain.Stage(fromNode))
		if seq == 0 {
			return fmt.Errorf("알 수 없는 replay node 입니다: %s", fromNode)
		}
	}
	if err := s.Solve.Case.ResetForReplay(s.ctx, s.solveExecutor(), id, mode, fromNode); err != nil {
		return err
	}
	if err := s.Solve.AnswerCheck.Delete(s.ctx, s.solveExecutor(), id); err != nil {
		return err
	}
	if err := s.Solve.Step.DeleteFromSeq(s.ctx, s.solveExecutor(), id, seq); err != nil {
		return err
	}
	if !keepCalls {
		if err := s.Solve.Call.DeleteFromSeq(s.ctx, s.solveExecutor(), id, seq); err != nil {
			return err
		}
	}
	return nil
}

func (s *StoreContainer) RecoverInterrupted() error {
	return s.WithSolveTx(s.ctx, func(tx *StoreContainer) error {
		if err := tx.Solve.Case.MarkInterrupted(tx.ctx, tx.solveExecutor()); err != nil {
			return err
		}
		return tx.Solve.Step.MarkInterrupted(tx.ctx, tx.solveExecutor())
	})
}

func (s *StoreContainer) CreateForExamQuestionWithSource(examQuestionID int64, sourceID source.SourceId) (domain.CaseID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.CreateForExamQuestion(s.ctx, s.solveExecutor(), examQuestionID, sourceID)
}

func (s *StoreContainer) CreateReplay(examQuestionID int64, replayOf domain.CaseID, mode, fromNode string) (domain.CaseID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.CreateReplay(s.ctx, s.solveExecutor(), examQuestionID, replayOf, mode, fromNode)
}

func (s *StoreContainer) ListCases(limit int) ([]domain.Case, error) {
	items, err := s.Solve.Case.List(s.ctx, s.solveExecutor(), limit)
	if err != nil {
		return nil, err
	}
	for index := range items {
		if err := s.applyAnswerCheck(&items[index]); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *StoreContainer) ListExamQuestionIDsWithCases() ([]int64, error) {
	return s.Solve.Case.ListExamQuestionIDsWithCases(s.ctx, s.solveExecutor())
}

func (s *StoreContainer) CountFailedCases() (int, error) {
	return s.Solve.Case.CountFailed(s.ctx, s.solveExecutor())
}

func (s *StoreContainer) ListFailedCases(limit int) ([]solve.FailedCase, error) {
	return s.Solve.Case.ListFailed(s.ctx, s.solveExecutor(), limit)
}

func (s *StoreContainer) Get(id domain.CaseID) (domain.Case, bool, error) {
	item, found, err := s.Solve.Case.Get(s.ctx, s.solveExecutor(), id)
	if err != nil || !found {
		return item, found, err
	}
	if err := s.applyAnswerCheck(&item); err != nil {
		return domain.Case{}, false, err
	}
	return item, true, nil
}

func (s *StoreContainer) applyAnswerCheck(item *domain.Case) error {
	check, found, err := s.Solve.AnswerCheck.Get(s.ctx, s.solveExecutor(), item.ID)
	if err != nil {
		return err
	}
	item.Processed = found && check.AnswerSatisfied
	return nil
}

func (s *StoreContainer) SaveAnswerCheck(value domain.AnswerCheck) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.AnswerCheck.Upsert(s.ctx, s.solveExecutor(), value)
}

func (s *StoreContainer) GetAnswerCheck(caseID domain.CaseID) (domain.AnswerCheck, bool, error) {
	return s.Solve.AnswerCheck.Get(s.ctx, s.solveExecutor(), caseID)
}

func (s *StoreContainer) DeleteAnswerChecksByExamQuestionID(examQuestionID int64) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.AnswerCheck.DeleteByExamQuestionID(s.ctx, s.solveExecutor(), examQuestionID)
}

func (s *StoreContainer) SetRunning(id domain.CaseID) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.SetRunning(s.ctx, s.solveExecutor(), id)
}

func (s *StoreContainer) SetStage(id domain.CaseID, stage string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.SetStage(s.ctx, s.solveExecutor(), id, stage)
}

func (s *StoreContainer) SetQuestionType(id domain.CaseID, questionType string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.SetQuestionType(s.ctx, s.solveExecutor(), id, questionType)
}

func (s *StoreContainer) SetSituation(id domain.CaseID, situation string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.SetSituation(s.ctx, s.solveExecutor(), id, situation)
}

func (s *StoreContainer) SetSource(id domain.CaseID, source_id source.SourceId) error {
	if source_id.ToI64() <= 0 {
		return fmt.Errorf("solve case source가 필요합니다")
	}
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.SetSource(s.ctx, s.solveExecutor(), id, source_id)
}

func (s *StoreContainer) SetSituationSource(id domain.CaseID, source_id source.SourceId) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.SetSituationSource(s.ctx, s.solveExecutor(), id, source_id)
}

func (s *StoreContainer) SetAnswer(id domain.CaseID, answerJSON string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Case.SetAnswer(s.ctx, s.solveExecutor(), id, answerJSON)
}

func (s *StoreContainer) SetFailed(id domain.CaseID, ce error) error {
	unlock := s.lockWrite()
	defer unlock()
	message := ""
	if ce != nil {
		message = ce.Error()
	}
	return s.Solve.Case.SetFailed(s.ctx, s.solveExecutor(), id, message)
}

func (s *StoreContainer) ExamQuestionCaseStats(ids []int64) (map[int64]solve.ExamQuestionCaseStats, error) {
	return s.Solve.Case.StatsByExamQuestionIDs(s.ctx, s.solveExecutor(), ids)
}
