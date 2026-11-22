package solve

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

// ErrPromptNotFound reports a missing prompt stage.
var ErrPromptNotFound = errors.New("prompt not found in DB")

type PromptTable struct{}

func (PromptTable) init(ctx context.Context, db sqlutil.Executor) error {
	query := `CREATE TABLE IF NOT EXISTS solve_prompt_versions (
		stage TEXT PRIMARY KEY,
		template TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("%s: %w", "solve_prompt_versions schema init failed", err)
	}
	return nil
}

func (PromptTable) Get(ctx context.Context, db sqlutil.Executor, stage string) (string, int, error) {
	var template string
	var version int
	err := db.QueryRowContext(ctx, `SELECT template, version FROM solve_prompt_versions WHERE stage = ?`, stage).
		Scan(&template, &version)
	if err == sql.ErrNoRows {
		return "", 0, fmt.Errorf("%w: %s", ErrPromptNotFound, stage)
	}
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", "solve prompt get failed", err)
	}
	return template, version, nil
}

func (PromptTable) Save(ctx context.Context, db sqlutil.Executor, stage, template string) (int, error) {
	if err := solve.ValidatePromptTemplate(stage, template); err != nil {
		return 0, err
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO solve_prompt_versions (stage, template, version, updated_at)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(stage) DO UPDATE SET
			template = excluded.template,
			version = solve_prompt_versions.version + 1,
			updated_at = CURRENT_TIMESTAMP`, stage, template); err != nil {
		return 0, fmt.Errorf("%s: %w", "solve prompt get failed", err)
	}
	var version int
	if err := db.QueryRowContext(ctx, `SELECT version FROM solve_prompt_versions WHERE stage = ?`, stage).Scan(&version); err != nil {
		return 0, fmt.Errorf("%s: %w", "solve prompt version save failed", err)
	}
	return version, nil
}
