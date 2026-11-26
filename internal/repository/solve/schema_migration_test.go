package solve

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	testutil "github.com/G226Medi03/MorbesPipeline/internal/testutil"
)

func TestInitDropsRedundantColumns(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.OpenSQLite(filepath.Join(t.TempDir(), "solve.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	statements := []string{
		`CREATE TABLE solve_cases (
			case_id INTEGER PRIMARY KEY AUTOINCREMENT,
			exam_question_id INTEGER NOT NULL,
			situation_json TEXT NOT NULL DEFAULT '',
			answer_json TEXT NOT NULL DEFAULT '',
			state TEXT NOT NULL DEFAULT 'pending',
			stage TEXT NOT NULL DEFAULT 'queued',
			question_type TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			source_id INTEGER NOT NULL,
			situation_source_id INTEGER NOT NULL DEFAULT 0,
			replay_of INTEGER NOT NULL DEFAULT 0,
			replay_mode TEXT NOT NULL DEFAULT '',
			replay_from_node TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE solve_steps (
			step_id INTEGER PRIMARY KEY AUTOINCREMENT,
			case_id INTEGER NOT NULL,
			node TEXT NOT NULL,
			seq INTEGER NOT NULL,
			state TEXT NOT NULL DEFAULT 'pending',
			attempt INTEGER NOT NULL DEFAULT 1,
			prompt_version INTEGER NOT NULL DEFAULT 0,
			output_json TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			call_id INTEGER NOT NULL DEFAULT 0,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME,
			UNIQUE(case_id, node)
		)`,
		`CREATE TABLE solve_llm_calls (
			call_id INTEGER PRIMARY KEY AUTOINCREMENT,
			case_id INTEGER NOT NULL,
			seq INTEGER NOT NULL,
			stage TEXT NOT NULL,
			role TEXT NOT NULL,
			model TEXT NOT NULL DEFAULT '',
			prompt_version INTEGER NOT NULL DEFAULT 0,
			prompt TEXT NOT NULL DEFAULT '',
			raw_output TEXT NOT NULL DEFAULT '',
			preprocessed_output TEXT NOT NULL DEFAULT '',
			preprocess_changes TEXT NOT NULL DEFAULT '',
			parsed_output TEXT NOT NULL DEFAULT '',
			source_id INTEGER NOT NULL,
			resp_id INTEGER NOT NULL DEFAULT 0,
			state TEXT NOT NULL DEFAULT 'ok',
			error TEXT NOT NULL DEFAULT '',
			duration_ms INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("create legacy table: %v", err)
		}
	}

	if err := (CaseTable{}).init(ctx, db); err != nil {
		t.Fatalf("case init: %v", err)
	}
	if err := (StepTable{}).init(ctx, db); err != nil {
		t.Fatalf("step init: %v", err)
	}
	if err := (CallTable{}).init(ctx, db); err != nil {
		t.Fatalf("call init: %v", err)
	}

	assertColumnMissing(t, ctx, db, "solve_cases", "situation_json")
	assertColumnMissing(t, ctx, db, "solve_steps", "call_id")
	assertColumnMissing(t, ctx, db, "solve_llm_calls", "preprocessed_output")
	assertColumnMissing(t, ctx, db, "solve_llm_calls", "preprocess_changes")
}

func assertColumnMissing(t *testing.T, ctx context.Context, db interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, table, column string) {
	t.Helper()
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		t.Fatalf("table info %s: %v", table, err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var cid, notNull, pk int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			t.Fatalf("scan %s: %v", table, err)
		}
		if name == column {
			t.Fatalf("%s.%s still exists", table, column)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("table info %s: %v", table, err)
	}
}
