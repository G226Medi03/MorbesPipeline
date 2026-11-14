package store

import (
	"fmt"

	"github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	solve "github.com/G226Medi03/MorbesPipeline/internal/persistence/solve"
)

func (s *Store) ResetCaseForReplay(id domain.CaseID, mode, fromNode string, keepCalls bool) error {
	unlock := s.lockWrite()
	defer unlock()
	seq := 0
	if fromNode != "" {
		seq = domain.Sequence(domain.Stage(fromNode))
		if seq == 0 {
			return fmt.Errorf("알 수 없는 replay node 입니다: %s", fromNode)
		}
	}
	if err := s.SolveRepository.Case.ResetForReplay(s.ctx, s.solveExecutor(), id, mode, fromNode); err != nil {
		return err
	}
	if err := s.SolveRepository.AnswerCheck.Delete(s.ctx, s.solveExecutor(), id); err != nil {
		return err
	}
	if err := s.SolveRepository.Step.DeleteFromSeq(s.ctx, s.solveExecutor(), id, seq); err != nil {
		return err
	}
	if !keepCalls {
		if err := s.SolveRepository.Call.DeleteFromSeq(s.ctx, s.solveExecutor(), id, seq); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) RecoverInterrupted() error {
	return s.WithSolveTx(s.ctx, func(tx *Store) error {
		if err := tx.SolveRepository.Case.MarkInterrupted(tx.ctx, tx.solveExecutor()); err != nil {
			return err
		}
		return tx.SolveRepository.Step.MarkInterrupted(tx.ctx, tx.solveExecutor())
	})
}

func (s *Store) CreateForExamQuestionWithSource(examQuestionID ids.ExamQuestionID, sourceID source.SourceId) (domain.CaseID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.CreateForExamQuestion(s.ctx, s.solveExecutor(), examQuestionID, sourceID)
}

func (s *Store) CreateReplay(examQuestionID ids.ExamQuestionID, replayOf domain.CaseID, mode, fromNode string) (domain.CaseID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.CreateReplay(s.ctx, s.solveExecutor(), examQuestionID, replayOf, mode, fromNode)
}

func (s *Store) ListCases(limit int) ([]domain.Case, error) {
	items, err := s.SolveRepository.Case.List(s.ctx, s.solveExecutor(), limit)
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

func (s *Store) ListExamQuestionIDsWithCases() ([]ids.ExamQuestionID, error) {
	return s.SolveRepository.Case.ListExamQuestionIDsWithCases(s.ctx, s.solveExecutor())
}

func (s *Store) CountFailedCases() (int, error) {
	return s.SolveRepository.Case.CountFailed(s.ctx, s.solveExecutor())
}

func (s *Store) ListFailedCases(limit int) ([]solve.FailedCase, error) {
	return s.SolveRepository.Case.ListFailed(s.ctx, s.solveExecutor(), limit)
}

func (s *Store) Get(id domain.CaseID) (domain.Case, bool, error) {
	item, found, err := s.SolveRepository.Case.Get(s.ctx, s.solveExecutor(), id)
	if err != nil || !found {
		return item, found, err
	}
	if err := s.applyAnswerCheck(&item); err != nil {
		return domain.Case{}, false, err
	}
	return item, true, nil
}

func (s *Store) applyAnswerCheck(item *domain.Case) error {
	check, found, err := s.SolveRepository.AnswerCheck.Get(s.ctx, s.solveExecutor(), item.ID)
	if err != nil {
		return err
	}
	item.Processed = found && check.AnswerSatisfied
	return nil
}

func (s *Store) SaveAnswerCheck(value domain.AnswerCheck) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.AnswerCheck.Upsert(s.ctx, s.solveExecutor(), value)
}

func (s *Store) GetAnswerCheck(caseID domain.CaseID) (domain.AnswerCheck, bool, error) {
	return s.SolveRepository.AnswerCheck.Get(s.ctx, s.solveExecutor(), caseID)
}

func (s *Store) DeleteAnswerChecksByExamQuestionID(examQuestionID ids.ExamQuestionID) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.AnswerCheck.DeleteByExamQuestionID(s.ctx, s.solveExecutor(), examQuestionID)
}

func (s *Store) SetRunning(id domain.CaseID) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.SetRunning(s.ctx, s.solveExecutor(), id)
}

func (s *Store) SetStage(id domain.CaseID, stage string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.SetStage(s.ctx, s.solveExecutor(), id, stage)
}

func (s *Store) SetQuestionType(id domain.CaseID, questionType string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.SetQuestionType(s.ctx, s.solveExecutor(), id, questionType)
}

func (s *Store) SetSituation(id domain.CaseID, situation string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.SetSituation(s.ctx, s.solveExecutor(), id, situation)
}

func (s *Store) SetSource(id domain.CaseID, source_id source.SourceId) error {
	if source_id.ToI64() <= 0 {
		return fmt.Errorf("solve case source가 필요합니다")
	}
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.SetSource(s.ctx, s.solveExecutor(), id, source_id)
}

func (s *Store) SetSituationSource(id domain.CaseID, source_id source.SourceId) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.SetSituationSource(s.ctx, s.solveExecutor(), id, source_id)
}

func (s *Store) SetAnswer(id domain.CaseID, answerJSON string) error {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Case.SetAnswer(s.ctx, s.solveExecutor(), id, answerJSON)
}

func (s *Store) SetFailed(id domain.CaseID, ce error) error {
	unlock := s.lockWrite()
	defer unlock()
	message := ""
	if ce != nil {
		message = ce.Error()
	}
	return s.SolveRepository.Case.SetFailed(s.ctx, s.solveExecutor(), id, message)
}

func (s *Store) ExamQuestionCaseStats(ids []ids.ExamQuestionID) (map[ids.ExamQuestionID]solve.ExamQuestionCaseStats, error) {
	return s.SolveRepository.Case.StatsByExamQuestionIDs(s.ctx, s.solveExecutor(), ids)
}
