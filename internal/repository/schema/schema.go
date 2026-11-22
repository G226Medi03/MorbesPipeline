package schema

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

func EnsureSolveIDColumns(ctx context.Context, db sqlutil.Executor) error {
	for _, item := range []struct {
		table     string
		newColumn string
	}{
		{"llm_problem_results", "llm_problem_result_id"},
		{"solve_cases", "case_id"},
		{"solve_steps", "step_id"},
		{"solve_llm_calls", "call_id"},
	} {
		if err := sqlutil.EnsureRenamedColumn(ctx, db, item.table, "id", item.newColumn); err != nil {
			return err
		}
	}
	return nil
}

func EnsureStatusOtherName(ctx context.Context, db sqlutil.Executor) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS status_other_name (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status_type TEXT NOT NULL,
		status_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 1,
		source_id INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_status_other_name_name
		ON status_other_name(name);`); err != nil {
		return err
	}
	return nil
}
