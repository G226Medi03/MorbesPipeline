package source

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	sourcedomain "github.com/G226Medi03/SourceEntity"
)

type Store struct {
	Repository Repository
}

func NewStore(table Table) Store {
	return Store{Repository: NewRepository(table)}
}

func (s Store) Get(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error) {
	return s.Repository.Get(ctx, db, sourceID)
}

func (s Store) Ensure(ctx context.Context, db sqlutil.Executor, kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error) {
	return s.Repository.Ensure(ctx, db, kind, idPtr)
}

func (s Store) GetLlmResponse(ctx context.Context, db sqlutil.Executor, respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return s.Repository.GetLlmResponse(ctx, db, respID)
}

func (s Store) Ref(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) sourcedomain.Source {
	return s.Repository.Ref(ctx, db, sourceID)
}

func (s Store) Refs(ctx context.Context, db sqlutil.Executor, ids []sourcedomain.SourceId) []sourcedomain.Source {
	return s.Repository.Refs(ctx, db, ids)
}

func (s Store) Trace(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
	return s.Repository.Trace(ctx, db, sourceID)
}
