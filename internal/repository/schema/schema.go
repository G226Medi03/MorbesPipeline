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

// CleanupLegacySchema removes tables that are no longer owned by this repo.
func CleanupLegacySchema(ctx context.Context, db sqlutil.Executor) error {
	for _, table := range []string{"entity_sources", "status_other_name"} {
		if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS `+table); err != nil {
			return err
		}
	}
	return nil
}
