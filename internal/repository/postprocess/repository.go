package postprocess

import (
	"context"
	"fmt"
	"time"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

// Repository wraps the postprocess table with validation and defaults.
type Repository struct {
	table table
}

func NewRepository() Repository {
	return Repository{}
}

func (r Repository) Init(ctx context.Context, db sqlutil.Executor) error {
	return r.table.init(ctx, db)
}

// Append inserts one audit record and returns the generated postprocess ID.
func (r Repository) Append(ctx context.Context, db sqlutil.Executor, record Record) (int64, error) {
	if record.PostprocessorName == "" {
		return 0, fmt.Errorf("postprocessor name is required")
	}
	if record.Operation == "" {
		return 0, fmt.Errorf("postprocess operation is required")
	}
	if record.EntityType == "" || record.EntityKey == "" {
		return 0, fmt.Errorf("postprocess entity is required")
	}
	if record.FinishedAt.IsZero() {
		record.FinishedAt = time.Now().UTC()
	}
	return r.table.insert(ctx, db, record)
}

// ListByEntity returns the audit records for one entity key.
func (r Repository) ListByEntity(
	ctx context.Context,
	db sqlutil.Executor,
	entityType EntityType,
	entityKey string,
) ([]Record, error) {
	return r.table.listByEntity(ctx, db, entityType, entityKey)
}
