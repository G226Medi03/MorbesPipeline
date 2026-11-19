package postprocess

import (
	"context"
	"fmt"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

// table owns the SQL surface for morbes_postprocesses.
type table struct{}

func (table) init(ctx context.Context, db sqlutil.Executor) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS morbes_postprocesses (
			postprocess_id INTEGER PRIMARY KEY AUTOINCREMENT,
			postprocessor_name TEXT NOT NULL,
			operation TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			target_entity_type TEXT NOT NULL DEFAULT '',
			target_entity_id INTEGER NOT NULL DEFAULT 0,
			finished_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_morbes_postprocesses_entity
			ON morbes_postprocesses(entity_type, entity_id);`,
		`CREATE INDEX IF NOT EXISTS idx_morbes_postprocesses_processor
			ON morbes_postprocesses(postprocessor_name, finished_at);`,
	}
	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("postprocess schema init failed: %w", err)
		}
	}
	return nil
}

func (table) insert(ctx context.Context, db sqlutil.Executor, record Record) (int64, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO morbes_postprocesses (
			postprocessor_name, operation, entity_type, entity_id,
			target_entity_type, target_entity_id, finished_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.PostprocessorName,
		string(record.Operation),
		string(record.EntityType),
		record.EntityID,
		string(record.TargetEntityType),
		record.TargetEntityID,
		record.FinishedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("postprocess record insert failed: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("postprocess record id failed: %w", err)
	}
	return id, nil
}

func (table) listByEntity(
	ctx context.Context,
	db sqlutil.Executor,
	entityType EntityType,
	entityID int64,
) ([]Record, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT postprocess_id, postprocessor_name, operation, entity_type, entity_id,
		       target_entity_type, target_entity_id, finished_at
		FROM morbes_postprocesses
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY postprocess_id`,
		string(entityType), entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("postprocess record list failed: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []Record
	for rows.Next() {
		var record Record
		var operation, entityTypeValue, targetType string
		if err := rows.Scan(
			&record.PostprocessID,
			&record.PostprocessorName,
			&operation,
			&entityTypeValue,
			&record.EntityID,
			&targetType,
			&record.TargetEntityID,
			&record.FinishedAt,
		); err != nil {
			return nil, err
		}
		record.Operation = Operation(operation)
		record.EntityType = EntityType(entityTypeValue)
		record.TargetEntityType = EntityType(targetType)
		result = append(result, record)
	}
	return result, rows.Err()
}
