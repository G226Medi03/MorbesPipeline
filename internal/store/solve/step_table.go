package solve

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

type StepTable struct{}

func (StepTable) init(ctx context.Context, db sqlutil.Executor) error {
	query := `CREATE TABLE IF NOT EXISTS solve_steps (
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
		UNIQUE(case_id, node),
		FOREIGN KEY (case_id) REFERENCES solve_cases(case_id) ON DELETE CASCADE
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("%s: %w", "solve_steps schema init failed", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_solve_steps_case ON solve_steps(case_id, seq);`); err != nil {
		return fmt.Errorf("%s: %w", "solve_steps schema init failed", err)
	}
	return nil
}

func (StepTable) Reset(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM solve_steps WHERE case_id = ?`, string(caseID)); err != nil {
		return fmt.Errorf("%s: %w", "solve_steps legacy cleanup failed", err)
	}
	return nil
}

func (StepTable) DeleteByCaseIDs(ctx context.Context, db sqlutil.Executor, caseIDs []domain.CaseID) error {
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
		DELETE FROM solve_steps
		WHERE case_id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
		return fmt.Errorf("solve_steps case별 삭제 실패: %w", err)
	}
	return nil
}

func (StepTable) Start(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID, node string, promptVersion int) (int64, error) {
	seq := domain.Sequence(domain.Stage(node))
	if seq == 0 {
		return 0, fmt.Errorf("unknown solve node: %s", node)
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO solve_steps (case_id, node, seq, state, attempt, prompt_version, output_json, error, call_id, started_at, finished_at)
		VALUES (?, ?, ?, 'running', 1, ?, '', '', 0, CURRENT_TIMESTAMP, NULL)
		ON CONFLICT(case_id, node) DO UPDATE SET
			state = 'running',
			attempt = solve_steps.attempt + 1,
			prompt_version = excluded.prompt_version,
			output_json = '',
			error = '',
			call_id = 0,
			started_at = CURRENT_TIMESTAMP,
			finished_at = NULL`,
		caseID, node, seq, promptVersion)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "solve_steps query failed", err)
	}
	var id int64
	if err := db.QueryRowContext(ctx, `SELECT step_id FROM solve_steps WHERE case_id = ? AND node = ?`, string(caseID), node).Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", "solve_steps last insert id failed", err)
	}
	return id, nil
}

func (StepTable) Finish(ctx context.Context, db sqlutil.Executor, id int64, state, outputJSON, message string, callID int64) error {
	_, err := db.ExecContext(ctx, `
		UPDATE solve_steps
		SET state = ?, output_json = ?, error = ?, call_id = ?, finished_at = CURRENT_TIMESTAMP
		WHERE step_id = ?`, state, outputJSON, message, callID, id)
	if err != nil {
		return fmt.Errorf("%s: %w", "solve_steps finish failed", err)
	}
	return nil
}

func (StepTable) List(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID) ([]domain.Step, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT step_id, case_id, node, seq, state, attempt, prompt_version,
		       output_json, error, call_id, started_at, finished_at
		FROM solve_steps WHERE case_id = ? ORDER BY seq`, string(caseID))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "solve_steps query failed", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]domain.Step, 0)
	for rows.Next() {
		var (
			step          domain.Step
			rawStepID     int64
			rawCaseID     string
			node          string
			state         string
			promptVersion int
			outputJSON    string
			rawcallID     int64
		)
		if err := rows.Scan(&rawStepID, &rawCaseID, &node, &step.Seq, &state,
			&step.Attempt, &promptVersion, &outputJSON, &step.Error,
			&rawcallID, &step.StartedAt, &step.FinishedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", "solve_steps scan failed", err)
		}
		step.ID = int64(rawStepID)
		step.Node = node
		step.State = domain.StepState(state)
		step.Meta = domain.SolveStepMeta{CaseID: domain.CaseID(rawCaseID), PromptVersion: promptVersion, OutputJSON: outputJSON, CallID: int64(rawcallID)}
		result = append(result, step)
	}
	return result, rows.Err()
}

func (StepTable) Get(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID, node string) (domain.Step, bool, error) {
	var (
		step          domain.Step
		rawStepID     int64
		rawCaseID     string
		rowNode       string
		state         string
		promptVersion int
		outputJSON    string
		rawcallID     int64
	)
	err := db.QueryRowContext(ctx, `
		SELECT step_id, case_id, node, seq, state, attempt, prompt_version,
		       output_json, error, call_id, started_at, finished_at
		FROM solve_steps WHERE case_id = ? AND node = ?`, string(caseID), node).
		Scan(&rawStepID, &rawCaseID, &rowNode, &step.Seq, &state,
			&step.Attempt, &promptVersion, &outputJSON, &step.Error,
			&rawcallID, &step.StartedAt, &step.FinishedAt)
	if err == sql.ErrNoRows {
		return domain.Step{}, false, nil
	}
	if err != nil {
		return domain.Step{}, false, fmt.Errorf("%s: %w", "solve_steps query failed", err)
	}
	step.ID = int64(rawStepID)
	step.Node = rowNode
	step.State = domain.StepState(state)
	step.Meta = domain.SolveStepMeta{CaseID: domain.CaseID(rawCaseID), PromptVersion: promptVersion, OutputJSON: outputJSON, CallID: int64(rawcallID)}
	return step, true, nil
}

func (StepTable) DeleteFromSeq(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID, seq int) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM solve_steps WHERE case_id = ? AND seq >= ?`, string(caseID), seq); err != nil {
		return fmt.Errorf("%s: %w", "solve replay step reset failed", err)
	}
	return nil
}

func (StepTable) MarkInterrupted(ctx context.Context, db sqlutil.Executor) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE solve_steps
		SET state = 'failed',
		    error = '??? ????? ???????.',
		    finished_at = CURRENT_TIMESTAMP
		WHERE state = 'running'`); err != nil {
		return fmt.Errorf("%s: %w", "interrupted step update failed", err)
	}
	return nil
}

func (StepTable) MarkStaleFrom(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID, fromSeq int) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE solve_steps SET state = 'stale', finished_at = NULL
		WHERE case_id = ? AND seq >= ? AND node <> ?`,
		caseID, fromSeq, string(domain.StageExplain)); err != nil {
		return fmt.Errorf("%s: %w", "solve_steps stale update failed", err)
	}
	return nil
}

func (StepTable) MarkDownstreamStale(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID, fromNode string) error {
	if !domain.IsStage(fromNode) {
		return fmt.Errorf("unknown solve node: %s", fromNode)
	}
	for _, target := range domain.Downstream(domain.Stage(fromNode)) {
		if _, err := db.ExecContext(ctx, `
			UPDATE solve_steps SET state = 'stale', finished_at = NULL
			WHERE case_id = ? AND node = ?`, caseID, string(target)); err != nil {
			return fmt.Errorf("%s: %w", "solve_steps stale update failed", err)
		}
	}
	return nil
}
