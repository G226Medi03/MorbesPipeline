package solve

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	testutil "github.com/G226Medi03/MorbesPipeline/internal/testutil"
)

func TestAnswerCheckLegacyMigration(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.OpenSQLite(filepath.Join(t.TempDir(), "solve.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	if _, err := db.ExecContext(ctx, `CREATE TABLE solve_cases (
		case_id INTEGER PRIMARY KEY AUTOINCREMENT
	)`); err != nil {
		t.Fatalf("create solve cases: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO solve_cases (case_id) VALUES (1), (2)`); err != nil {
		t.Fatalf("insert solve cases: %v", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE solve_answer_checks (
		case_id INTEGER PRIMARY KEY,
		resp_id INTEGER NOT NULL DEFAULT 0,
		contains_answer INTEGER NOT NULL DEFAULT 0,
		answer_present INTEGER,
		answer_satisfied INTEGER,
		effect_matched INTEGER,
		matched_type TEXT NOT NULL DEFAULT 'none',
		error TEXT NOT NULL DEFAULT '',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO solve_answer_checks
		(case_id, resp_id, contains_answer, answer_present, answer_satisfied, matched_type)
		VALUES (1, 7, 1, 1, 1, 'choice')`); err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	if err := (AnswerCheckTable{}).init(ctx, db); err != nil {
		t.Fatalf("init: %v", err)
	}

	if exists, err := tableExists(ctx, db, answerCheckTableLegacy); err != nil {
		t.Fatalf("legacy exists: %v", err)
	} else if exists {
		t.Fatal("legacy answer check table still exists")
	}

	migrated, found, err := (AnswerCheckTable{}).Get(ctx, db, "1")
	if err != nil {
		t.Fatalf("get migrated: %v", err)
	}
	if !found || !migrated.AnswerSatisfied || migrated.RespID.ToI64() != 7 {
		t.Fatalf("unexpected migrated result: found=%v value=%+v", found, migrated)
	}

	if err := (AnswerCheckTable{}).Upsert(ctx, db, solve.AnswerCheck{
		CaseID:          "2",
		QuestionType:    "procedure",
		ContainsAnswer:  true,
		AnswerSatisfied: false,
		MatchedType:     "medical_procedure_option",
		AnalysisJSON:    `{"selectionMode":"current"}`,
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	value, found, err := (AnswerCheckTable{}).Get(ctx, db, "2")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !found || value.AnswerSatisfied || value.QuestionType != "procedure" {
		t.Fatalf("unexpected result: found=%v value=%+v", found, value)
	}
}
