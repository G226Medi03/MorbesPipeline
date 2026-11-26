package solve

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	testutil "github.com/G226Medi03/MorbesPipeline/internal/testutil"
)

func TestCaseGetDerivesAnswerAndQuestionTypeFromSteps(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.OpenSQLite(filepath.Join(t.TempDir(), "solve.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	if err := (CaseTable{}).init(ctx, db); err != nil {
		t.Fatalf("case init: %v", err)
	}
	if err := (StepTable{}).init(ctx, db); err != nil {
		t.Fatalf("step init: %v", err)
	}
	result, err := db.ExecContext(ctx, `
		INSERT INTO solve_cases (exam_question_id, state, stage, source_id)
		VALUES (1, 'done', 'done', 1)`)
	if err != nil {
		t.Fatalf("insert case: %v", err)
	}
	caseID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("case id: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO solve_steps (case_id, node, seq, state, output_json)
		VALUES (?, 'case_extraction', 20, 'succeeded', '{"questionType":"procedure"}'),
		       (?, 'cause_trace', 60, 'succeeded', '{"observed":["test"]}')`,
		caseID, caseID); err != nil {
		t.Fatalf("insert steps: %v", err)
	}

	item, found, err := (CaseTable{}).Get(ctx, db, solve.CaseID(strconv.FormatInt(caseID, 10)))
	if err != nil {
		t.Fatalf("get case: %v", err)
	}
	if !found {
		t.Fatal("case not found")
	}
	if item.QuestionType != "procedure" {
		t.Fatalf("question type = %q", item.QuestionType)
	}
	if item.AnswerJSON != `{"observed":["test"]}` {
		t.Fatalf("answer json = %q", item.AnswerJSON)
	}
}
