package source

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	sourcedomain "github.com/G226Medi03/SourceEntity"
	storesource "github.com/G226Medi03/SourceEntityRepository"
)

type Repository struct {
	ctx  context.Context
	db   sqlutil.Executor
	repo storesource.Repository
}

func New(ctx context.Context, db sqlutil.Executor, repo storesource.Repository) Repository {
	return Repository{ctx: ctx, db: db, repo: repo}
}

func (r Repository) Get(sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error) {
	return r.repo.Source.Get(r.ctx, r.db, sourceID)
}

func (r Repository) Ensure(kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error) {
	return r.repo.Source.Ensure(r.ctx, r.db, kind, idPtr)
}

func (r Repository) GetLlmResponse(respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return r.repo.LlmResponse.Get(r.ctx, r.db, respID)
}

func (r Repository) Ref(sourceID sourcedomain.SourceId) sourcedomain.Source {
	return r.repo.Source.Ref(r.ctx, r.db, sourceID)
}

func (r Repository) Refs(ids []sourcedomain.SourceId) []sourcedomain.Source {
	result := make([]sourcedomain.Source, 0, len(ids))
	for _, id := range ids {
		result = append(result, r.Ref(id))
	}
	return result
}

func (r Repository) Trace(sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
	trace := sourcedomain.SourceTrace{}
	if sourceID.ToI64() <= 0 {
		return trace, nil
	}
	visited := map[sourcedomain.SourceId]struct{}{}
	current := sourceID
	for depth := 0; depth < 64 && current.ToI64() > 0; depth++ {
		if _, seen := visited[current]; seen {
			break
		}
		visited[current] = struct{}{}
		row, found, err := r.repo.Source.Get(r.ctx, r.db, current)
		if err != nil {
			return trace, err
		}
		if !found {
			break
		}
		trace = append(trace, row)
		if row.SourceKind != sourcedomain.SourceKindLlmResponse {
			break
		}
		response, found, err := r.repo.LlmResponse.Get(r.ctx, r.db, sourcedomain.NewRespId(row.IDPtr))
		if err != nil {
			return trace, err
		}
		if !found || response.ReferedSourceID.ToI64() <= 0 {
			break
		}
		current = response.ReferedSourceID
	}
	return trace, nil
}
