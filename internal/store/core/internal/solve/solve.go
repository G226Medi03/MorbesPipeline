package solve

import (
	"fmt"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/deps"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
	"github.com/G226Medi03/SourceEntity"
)

func (s *SolveService) ResetCaseForReplay(id domain.CaseID, mode, fromNode string, keepCalls bool) error {
	unlock := s.store.LockWrite()
	defer unlock()
	seq := 0
	if fromNode != "" {
		seq = domain.Sequence(domain.Stage(fromNode))
		if seq == 0 {
			return fmt.Errorf("알 수 없는 replay node 입니다: %s", fromNode)
		}
	}
	if err := s.store.SolveRepository().Case.ResetForReplay(s.store.Context(), s.store.SolveExecutor(), id, mode, fromNode); err != nil {
		return err
	}
	if err := s.store.SolveRepository().AnswerCheck.Delete(s.store.Context(), s.store.SolveExecutor(), id); err != nil {
		return err
	}
	if err := s.store.SolveRepository().Step.DeleteFromSeq(s.store.Context(), s.store.SolveExecutor(), id, seq); err != nil {
		return err
	}
	if !keepCalls {
		if err := s.store.SolveRepository().Call.DeleteFromSeq(s.store.Context(), s.store.SolveExecutor(), id, seq); err != nil {
			return err
		}
	}
	return nil
}

func (s *SolveService) RecoverInterrupted() error {
	return s.store.RunSolveTx(s.store.Context(), func(tx deps.Store) error {
		if err := tx.SolveRepository().Case.MarkInterrupted(tx.Context(), tx.SolveExecutor()); err != nil {
			return err
		}
		return tx.SolveRepository().Step.MarkInterrupted(tx.Context(), tx.SolveExecutor())
	})
}

func (s *SolveService) CreateForExamQuestionWithSource(examQuestionID int64, sourceID source.SourceId) (domain.CaseID, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.CreateForExamQuestion(s.store.Context(), s.store.SolveExecutor(), examQuestionID, sourceID)
}

func (s *SolveService) CreateReplay(examQuestionID int64, replayOf domain.CaseID, mode, fromNode string) (domain.CaseID, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.CreateReplay(s.store.Context(), s.store.SolveExecutor(), examQuestionID, replayOf, mode, fromNode)
}

func (s *SolveService) ListCases(limit int) ([]domain.Case, error) {
	items, err := s.store.SolveRepository().Case.List(s.store.Context(), s.store.SolveExecutor(), limit)
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
	return s.store.SolveRepository().Case.ListExamQuestionIDsWithCases(s.store.Context(), s.store.SolveExecutor())
}

func (s *SolveService) CountFailedCases() (int, error) {
	return s.store.SolveRepository().Case.CountFailed(s.store.Context(), s.store.SolveExecutor())
}

func (s *SolveService) ListFailedCases(limit int) ([]storerepo.FailedCase, error) {
	return s.store.SolveRepository().Case.ListFailed(s.store.Context(), s.store.SolveExecutor(), limit)
}

func (s *SolveService) Get(id domain.CaseID) (domain.Case, bool, error) {
	item, found, err := s.store.SolveRepository().Case.Get(s.store.Context(), s.store.SolveExecutor(), id)
	if err != nil || !found {
		return item, found, err
	}
	if err := s.ApplyAnswerCheck(&item); err != nil {
		return domain.Case{}, false, err
	}
	return item, true, nil
}

func (s *SolveService) ApplyAnswerCheck(item *domain.Case) error {
	check, found, err := s.store.SolveRepository().AnswerCheck.Get(s.store.Context(), s.store.SolveExecutor(), item.ID)
	if err != nil {
		return err
	}
	item.Processed = found && check.AnswerSatisfied
	return nil
}

func (s *SolveService) SaveAnswerCheck(value domain.AnswerCheck) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().AnswerCheck.Upsert(s.store.Context(), s.store.SolveExecutor(), value)
}

func (s *SolveService) GetAnswerCheck(caseID domain.CaseID) (domain.AnswerCheck, bool, error) {
	return s.store.SolveRepository().AnswerCheck.Get(s.store.Context(), s.store.SolveExecutor(), caseID)
}

func (s *SolveService) DeleteAnswerChecksByExamQuestionID(examQuestionID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().AnswerCheck.DeleteByExamQuestionID(s.store.Context(), s.store.SolveExecutor(), examQuestionID)
}

func (s *SolveService) SetRunning(id domain.CaseID) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.SetRunning(s.store.Context(), s.store.SolveExecutor(), id)
}

func (s *SolveService) SetStage(id domain.CaseID, stage string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.SetStage(s.store.Context(), s.store.SolveExecutor(), id, stage)
}

func (s *SolveService) SetQuestionType(id domain.CaseID, questionType string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.SetQuestionType(s.store.Context(), s.store.SolveExecutor(), id, questionType)
}

func (s *SolveService) SetSituation(id domain.CaseID, situation string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.SetSituation(s.store.Context(), s.store.SolveExecutor(), id, situation)
}

func (s *SolveService) SetSource(id domain.CaseID, source_id source.SourceId) error {
	if source_id.ToI64() <= 0 {
		return fmt.Errorf("solve case source가 필요합니다")
	}
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.SetSource(s.store.Context(), s.store.SolveExecutor(), id, source_id)
}

func (s *SolveService) SetSituationSource(id domain.CaseID, source_id source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.SetSituationSource(s.store.Context(), s.store.SolveExecutor(), id, source_id)
}

func (s *SolveService) SetAnswer(id domain.CaseID, answerJSON string) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepository().Case.SetAnswer(s.store.Context(), s.store.SolveExecutor(), id, answerJSON)
}

func (s *SolveService) SetFailed(id domain.CaseID, ce error) error {
	unlock := s.store.LockWrite()
	defer unlock()
	message := ""
	if ce != nil {
		message = ce.Error()
	}
	return s.store.SolveRepository().Case.SetFailed(s.store.Context(), s.store.SolveExecutor(), id, message)
}

func (s *SolveService) ExamQuestionCaseStats(ids []int64) (map[int64]storerepo.ExamQuestionCaseStats, error) {
	return s.store.SolveRepository().Case.StatsByExamQuestionIDs(s.store.Context(), s.store.SolveExecutor(), ids)
}
