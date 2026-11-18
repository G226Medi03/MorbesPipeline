// Package postprocess stores the audit trail of postprocess operations.
package postprocess

import (
	"context"
	"fmt"
	"time"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

// Operation describes one postprocess change.
type Operation string

const (
	OperationObserved    Operation = "observed"
	OperationCreated     Operation = "created"
	OperationUpdated     Operation = "updated"
	OperationRemoved     Operation = "removed"
	OperationLinkAdded   Operation = "link_added"
	OperationLinkRemoved Operation = "link_removed"
)

// EntityType identifies the kind of entity referenced by an operation.
type EntityType string

const (
	EntityStatus    EntityType = "status"
	EntityAction    EntityType = "action"
	EntityActor     EntityType = "actor"
	EntityProcedure EntityType = "procedure"
	EntityOption    EntityType = "option"
)

// Record is one postprocess operation.
type Record struct {
	PostprocessID     int64
	PostprocessorName string
	Operation         Operation
	EntityType        EntityType
	EntityID          int64
	TargetEntityType  EntityType
	TargetEntityID    int64
	FinishedAt        time.Time
}

type Repository struct {
	ctx context.Context
	db  sqlutil.Executor
}

func NewRepository(ctx context.Context, db sqlutil.Executor) Repository {
	return Repository{ctx: ctx, db: db}
}

func (r Repository) Init() error {
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
		if _, err := r.db.ExecContext(r.ctx, query); err != nil {
			return fmt.Errorf("postprocess schema init failed: %w", err)
		}
	}
	return nil
}

// Append inserts one audit record and returns the generated postprocess ID.
func (r Repository) Append(record Record) (int64, error) {
	if record.PostprocessorName == "" {
		return 0, fmt.Errorf("postprocessor name is required")
	}
	if record.Operation == "" {
		return 0, fmt.Errorf("postprocess operation is required")
	}
	if record.EntityType == "" || record.EntityID <= 0 {
		return 0, fmt.Errorf("postprocess entity is required")
	}
	finishedAt := record.FinishedAt
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}
	result, err := r.db.ExecContext(r.ctx, `
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
		finishedAt,
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

// ListByEntity returns the audit records for one entity.
func (r Repository) ListByEntity(entityType EntityType, entityID int64) ([]Record, error) {
	rows, err := r.db.QueryContext(r.ctx, `
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
