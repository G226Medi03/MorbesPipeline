package postprocess

import (
	"context"
	"fmt"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

// table owns the SQL surface for morbes_postprocesses.
type table struct{}

func (table) init(ctx context.Context, db sqlutil.Executor) error {
	if err := (table{}).archiveLegacy(ctx, db); err != nil {
		return err
	}
	queries := []string{
		`CREATE TABLE IF NOT EXISTS morbes_postprocesses (
			postprocess_id INTEGER PRIMARY KEY AUTOINCREMENT,
			postprocessor_name TEXT NOT NULL,
			operation TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_key TEXT NOT NULL,
			target_entity_type TEXT NOT NULL DEFAULT '',
			target_entity_key TEXT NOT NULL DEFAULT '',
			finished_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_morbes_postprocesses_entity
			ON morbes_postprocesses(entity_type, entity_key);`,
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

// archiveLegacy preserves a pre-key audit table under an immutable legacy name.
func (table) archiveLegacy(ctx context.Context, db sqlutil.Executor) error {
	exists, err := sqlutil.TableExists(ctx, db, "morbes_postprocesses")
	if err != nil || !exists {
		return err
	}
	hasEntityID, err := columnExists(ctx, db, "morbes_postprocesses", "entity_id")
	if err != nil || !hasEntityID {
		return err
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS morbes_postprocesses_legacy`); err != nil {
		return fmt.Errorf("postprocess legacy archive cleanup failed: %w", err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE morbes_postprocesses RENAME TO morbes_postprocesses_legacy`); err != nil {
		return fmt.Errorf("postprocess legacy archive rename failed: %w", err)
	}
	for _, index := range []string{
		"idx_morbes_postprocesses_entity",
		"idx_morbes_postprocesses_processor",
	} {
		if _, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS `+index); err != nil {
			return fmt.Errorf("postprocess legacy index cleanup failed: %w", err)
		}
	}
	return nil
}

func columnExists(ctx context.Context, db sqlutil.Executor, table, column string) (bool, error) {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return false, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var cid, notNull, pk int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (table) insert(ctx context.Context, db sqlutil.Executor, record Record) (int64, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO morbes_postprocesses (
			postprocessor_name, operation, entity_type, entity_key,
			target_entity_type, target_entity_key, finished_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.PostprocessorName,
		string(record.Operation),
		string(record.EntityType),
		record.EntityKey,
		string(record.TargetEntityType),
		record.TargetEntityKey,
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
	entityKey string,
) ([]Record, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT postprocess_id, postprocessor_name, operation, entity_type, entity_key,
		       target_entity_type, target_entity_key, finished_at
		FROM morbes_postprocesses
		WHERE entity_type = ? AND entity_key = ?
		ORDER BY postprocess_id`,
		string(entityType), entityKey,
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
			&record.EntityKey,
			&targetType,
			&record.TargetEntityKey,
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
