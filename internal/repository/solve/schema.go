package solve

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

// EnsureLegacyColumns renames legacy solve table columns to their current
// names. It is idempotent and safe on a fresh schema.
func EnsureLegacyColumns(ctx context.Context, db sqlutil.Executor) error {
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
