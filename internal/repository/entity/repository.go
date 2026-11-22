package entity

import (
	"context"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

// Repository owns ename-based entity link/delete operations.
type Repository struct{}

func (Repository) table(
	ctx context.Context,
	db sqlutil.Executor,
	entities *repo.Store,
) *entityTable {
	return &entityTable{ctx: ctx, db: db, entities: entities}
}

// EntitySession is a per-call entity repository view.
type entityTable struct {
	ctx      context.Context
	db       sqlutil.Executor
	entities *repo.Store
}

func sourceIDs(sources []source.Source) []source.SourceId {
	seen := make(map[source.SourceId]struct{}, len(sources))
	result := make([]source.SourceId, 0, len(sources))
	for _, item := range sources {
		if item.SourceID.ToI64() <= 0 {
			continue
		}
		if _, ok := seen[item.SourceID]; ok {
			continue
		}
		seen[item.SourceID] = struct{}{}
		result = append(result, item.SourceID)
	}
	return result
}
