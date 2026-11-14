package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
)

func newAnswerCheckStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "answer-check.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO solve_cases (case_id, exam_question_id, source_id, state, stage)
		VALUES (6, 8, 1, 'done', 'done')`); err != nil {
		t.Fatalf("insert case: %v", err)
	}
	return st
}

func TestApplyAnswerCheckUsesAnswerSatisfied(t *testing.T) {
	st := newAnswerCheckStore(t)
	if _, err := st.ExecContext(context.Background(), `
		INSERT INTO solve_answer_checks
			(case_id, question_type, contains_answer, answer_present, answer_satisfied, matched_type, matched_key, analysis_json)
		VALUES (6, 'status', 0, 1, 1, 'observed', '체온 증가', '{"verdict":"correct"}')`); err != nil {
		t.Fatalf("insert check: %v", err)
	}
	item := domain.Case{ID: "6"}
	if err := st.applyAnswerCheck(&item); err != nil {
		t.Fatalf("applyAnswerCheck: %v", err)
	}
	if !item.Processed {
		t.Fatal("processed must use answer_satisfied, not contains_answer")
	}
}

func TestLegacyAnswerSatisfiedNullFallsBackToContainsAnswer(t *testing.T) {
	st := newAnswerCheckStore(t)
	if _, err := st.ExecContext(context.Background(), `
		INSERT INTO solve_answer_checks
			(case_id, question_type, contains_answer, answer_present, answer_satisfied, matched_type, matched_key, analysis_json)
		VALUES (6, '', 1, NULL, NULL, 'medical_procedure_option', '1', '{"verdict":"correct"}')`); err != nil {
		t.Fatalf("insert check: %v", err)
	}
	check, found, err := st.SolveRepository.AnswerCheck.Get(context.Background(), st.solveExecutor(), domain.CaseID("6"))
	if err != nil || !found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}
	if !check.AnswerSatisfied {
		t.Fatal("legacy NULL answer_satisfied must fall back to contains_answer")
	}
}
