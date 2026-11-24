package store

import (
	"fmt"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	srepo "github.com/G226Medi03/MorbesPipeline/internal/repository/solve"
	corestore "github.com/G226Medi03/MorbesPipeline/internal/repository/store"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *SolveService) ResetCaseForReplay(id solve.CaseID, mode, fromNode string, keepCalls bool) error {
	unlock := s.store.LockWrite()
	defer unlock()
	seq := 0
	if fromNode != "" {
		seq = solve.Sequence(solve.Stage(fromNode))
		if seq == 0 {
			return fmt.Errorf("알 수 없는 replay node 입니다: %s", fromNode)
		}
	}
	if err := s.store.SolveRepo().Case.ResetForReplay(s.store.Context(), s.store.SolveExecutor(), id, mode, fromNode); err != nil {
		return err
	}
	if err := s.store.SolveRepo().AnswerCheck.Delete(s.store.Context(), s.store.SolveExecutor(), id); err != nil {
		return err
	}
	if err := s.store.SolveRepo().Step.DeleteFromSeq(s.store.Context(), s.store.SolveExecutor(), id, seq); err != nil {
		return err
	}
	if !keepCalls {
		if err := s.store.SolveRepo().Call.DeleteFromSeq(s.store.Context(), s.store.SolveExecutor(), id, seq); err != nil {
			return err
		}
	}
	return nil
}

func (s *SolveService) RecoverInterrupted() error {
	return s.store.WithSolveTx(s.store.Context(), func(tx *corestore.Store) error {
		if err := tx.SolveRepo().Case.MarkInterrupted(tx.Context(), tx.SolveExecutor()); err != nil {
			return err
		}
		return tx.SolveRepo().Step.MarkInterrupted(tx.Context(), tx.SolveExecutor())
	})
}

func (s *SolveService) CreateForExamQuestionWithSource(examQuestionID int64, sourceID source.SourceId) (solve.CaseID, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.CreateForExamQuestion(s.store.Context(), s.store.SolveExecutor(), examQuestionID, sourceID)
}

func (s *SolveService) CreateReplay(examQuestionID int64, replayOf solve.CaseID, mode, fromNode string) (solve.CaseID, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.CreateReplay(s.store.Context(), s.store.SolveExecutor(), examQuestionID, replayOf, mode, fromNode)
}

func (s *SolveService) ListCases(limit int) ([]solve.Case, error) {
	items, err := s.store.SolveRepo().Case.List(s.store.Context(), s.store.SolveExecutor(), limit)
	if err != nil {
		return nil, err
	}
	for index := range items {
		if err := s.ApplyAnswerCheck(&items[index]); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *SolveService) ListExamQuestionIDsWithCases() ([]int64, error) {
	return s.store.SolveRepo().Case.ListExamQuestionIDsWithCases(s.store.Context(), s.store.SolveExecutor())
}

func (s *SolveService) CountFailedCases() (int, error) {
	return s.store.SolveRepo().Case.CountFailed(s.store.Context(), s.store.SolveExecutor())
}

func (s *SolveService) ListFailedCases(limit int) ([]srepo.FailedCase, error) {
	return s.store.SolveRepo().Case.ListFailed(s.store.Context(), s.store.SolveExecutor(), limit)
}

func (s *SolveService) Get(id solve.CaseID) (solve.Case, bool, error) {
	item, found, err := s.store.SolveRepo().Case.Get(s.store.Context(), s.store.SolveExecutor(), id)
	if err != nil || !found {
		return item, found, err
	}
	if err := s.ApplyAnswerCheck(&item); err != nil {
		return solve.Case{}, false, err
	}
	return item, true, nil
}

func (s *SolveService) ApplyAnswerCheck(item *solve.Case) error {
	check, found, err := s.store.SolveRepo().AnswerCheck.Get(s.store.Context(), s.store.SolveExecutor(), item.ID)
	if err != nil {
		return err
	}
	item.Processed = found && check.AnswerSatisfied
	return nil
}

func (s *SolveService) SaveAnswerCheck(value solve.AnswerCheck) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().AnswerCheck.Upsert(s.store.Context(), s.store.SolveExecutor(), value)
}

func (s *SolveService) GetAnswerCheck(caseID solve.CaseID) (solve.AnswerCheck, bool, error) {
	return s.store.SolveRepo().AnswerCheck.Get(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) DeleteAnswerChecksByExamQuestionID(examQuestionID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().AnswerCheck.DeleteByExamQuestionID(s.store.Context(), s.store.SolveExecutor(), examQuestionID)
}

func (s *SolveService) SetRunning(id solve.CaseID) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.SetRunning(s.store.Context(), s.store.SolveExecutor(), id)
}

func (s *SolveService) SetStage(id solve.CaseID, stage string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.SetStage(s.store.Context(), s.store.SolveExecutor(), id, stage)
}

func (s *SolveService) SetQuestionType(id solve.CaseID, questionType string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.SetQuestionType(s.store.Context(), s.store.SolveExecutor(), id, questionType)
}

func (s *SolveService) SetSituation(id solve.CaseID, situation string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.SetSituation(s.store.Context(), s.store.SolveExecutor(), id, situation)
}

func (s *SolveService) SetSource(id solve.CaseID, source_id source.SourceId) error {
	if source_id.ToI64() <= 0 {
		return fmt.Errorf("solve case source가 필요합니다")
	}
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.SetSource(s.store.Context(), s.store.SolveExecutor(), id, source_id)
}

func (s *SolveService) SetSituationSource(id solve.CaseID, source_id source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.SetSituationSource(s.store.Context(), s.store.SolveExecutor(), id, source_id)
}

func (s *SolveService) SetAnswer(id solve.CaseID, answerJSON string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Case.SetAnswer(s.store.Context(), s.store.SolveExecutor(), id, answerJSON)
}

func (s *SolveService) SetFailed(id solve.CaseID, ce error) error {
	unlock := s.store.LockWrite()
	defer unlock()
	message := ""
	if ce != nil {
		message = ce.Error()
	}
	return s.store.SolveRepo().Case.SetFailed(s.store.Context(), s.store.SolveExecutor(), id, message)
}

func (s *SolveService) ExamQuestionCaseStats(ids []int64) (map[int64]srepo.ExamQuestionCaseStats, error) {
	return s.store.SolveRepo().Case.StatsByExamQuestionIDs(s.store.Context(), s.store.SolveExecutor(), ids)
}
