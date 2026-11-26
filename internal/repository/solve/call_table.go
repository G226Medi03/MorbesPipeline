package solve

import (
	"context"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/model"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

type CallTable struct{}

func (CallTable) init(ctx context.Context, db sqlutil.Executor) error {
	query := `CREATE TABLE IF NOT EXISTS solve_llm_calls (
		call_id INTEGER PRIMARY KEY AUTOINCREMENT,
		case_id INTEGER NOT NULL,
		seq INTEGER NOT NULL,
		stage TEXT NOT NULL,
		role TEXT NOT NULL,
		prompt_version INTEGER NOT NULL DEFAULT 0,
		parsed_output TEXT NOT NULL DEFAULT '',
		source_id INTEGER NOT NULL CHECK (source_id > 0),
		resp_id INTEGER NOT NULL DEFAULT 0,
		state TEXT NOT NULL DEFAULT 'ok',
		error TEXT NOT NULL DEFAULT '',
		duration_ms INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (case_id) REFERENCES solve_cases(case_id) ON DELETE CASCADE
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("%s: %w", "solve_llm_calls schema init failed", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_solve_llm_calls_case ON solve_llm_calls(case_id, seq);`); err != nil {
		return fmt.Errorf("%s: %w", "solve_llm_calls schema init failed", err)
	}
	for _, column := range []string{"preprocessed_output", "preprocess_changes"} {
		if err := dropColumnIfExists(ctx, db, "solve_llm_calls", column); err != nil {
			return err
		}
	}
	return nil
}

func dropColumnIfExists(ctx context.Context, db sqlutil.Executor, tableName, column string) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+tableName+`)`)
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, notNull, pk int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			rows.Close() //nolint:errcheck // error path
			return err
		}
		if name == column {
			found = true
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if !found {
		return nil
	}
	_, err = db.ExecContext(ctx, `ALTER TABLE `+tableName+` DROP COLUMN `+column)
	return err
}

func (CallTable) DeleteByCaseIDs(ctx context.Context, db sqlutil.Executor, caseIDs []solve.CaseID) error {
	if len(caseIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(caseIDs))
	args := make([]any, 0, len(caseIDs))
	for _, id := range caseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, string(id))
	}
	if _, err := db.ExecContext(ctx, `
		DELETE FROM solve_llm_calls
		WHERE case_id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
		return fmt.Errorf("solve_llm_calls case별 삭제 실패: %w", err)
	}
	return nil
}

func (CallTable) Append(ctx context.Context, db sqlutil.Executor, call model.LLMCallInput) (int64, error) {
	if call.State == "" {
		call.State = "ok"
	}
	result, err := db.ExecContext(ctx, `
		INSERT INTO solve_llm_calls (
			case_id, seq, stage, role, prompt_version, parsed_output,
			source_id, resp_id, state, error, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, 0), (SELECT source_id FROM solve_cases WHERE case_id = ?)), ?, ?, ?, ?)`,
		call.CaseID, call.Seq, call.Stage, call.Role, call.PromptVersion, call.ParsedOutput,
		call.SourceID, call.CaseID, call.RespID, call.State, call.Error, call.DurationMS)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "solve_llm_calls 저장 실패", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "solve_llm_calls LastInsertId 실패", err)
	}
	return int64(id), nil
}

func (CallTable) List(ctx context.Context, db sqlutil.Executor, caseID solve.CaseID) ([]model.LLMCall, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT call_id, case_id, seq, stage, role, prompt_version, parsed_output,
		       source_id, resp_id, state, error, duration_ms, created_at
		FROM solve_llm_calls WHERE case_id = ? ORDER BY seq, call_id`, string(caseID))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "solve_llm_calls list failed", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]model.LLMCall, 0)
	for rows.Next() {
		var call model.LLMCall
		var rawCaseID string
		if err := rows.Scan(&call.ID, &rawCaseID, &call.Seq, &call.Stage, &call.Role,
			&call.PromptVersion, &call.ParsedOutput,
			&call.SourceID, &call.RespID, &call.State, &call.Error, &call.DurationMS,
			&call.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", "solve_llm_calls scan failed", err)
		}
		call.CaseID = solve.CaseID(rawCaseID)
		result = append(result, call)
	}
	return result, rows.Err()
}

func (CallTable) DeleteFromSeq(ctx context.Context, db sqlutil.Executor, caseID solve.CaseID, seq int) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM solve_llm_calls WHERE case_id = ? AND seq >= ?`, string(caseID), seq); err != nil {
		return fmt.Errorf("%s: %w", "solve replay calls delete failed", err)
	}
	return nil
}

func (CallTable) DeleteByStage(ctx context.Context, db sqlutil.Executor, caseID solve.CaseID, stage string) error {
	if _, err := db.ExecContext(ctx,
		`DELETE FROM solve_llm_calls WHERE case_id = ? AND stage = ?`, string(caseID), stage); err != nil {
		return fmt.Errorf("%s: %w", "solve llm call stage delete failed", err)
	}
	return nil
}
