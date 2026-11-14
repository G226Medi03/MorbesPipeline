package morbes_pipeline

import (
	"context"
	"database/sql"

	persistence "github.com/G226Medi03/MorbesPipeline/internal/persistence"
)

type (
	Store               = persistence.Store
	ProblemResultFilter = persistence.ProblemResultFilter
	ProblemResultRow    = persistence.ProblemResultRow
)

var (
	NewStore        = persistence.New
	New             = persistence.New
	NewStoreWithDBs = persistence.NewWithDBs
	NewWithDBs      = persistence.NewWithDBs
)

func NewStoreContext(ctx context.Context, db *sql.DB) *Store {
	return persistence.New(ctx, db)
}

func InitSchema(store *Store) error {
	return store.InitSchema()
}
