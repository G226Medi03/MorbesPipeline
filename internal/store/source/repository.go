package source

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	sourcedomain "github.com/G226Medi03/SourceEntity"
)

type Repository struct {
	table Table
}

func NewRepository(table Table) Repository { return Repository{table: table} }

func (r Repository) Get(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error) {
	return r.table.Get(ctx, db, sourceID)
}

func (r Repository) Ensure(ctx context.Context, db sqlutil.Executor, kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error) {
	return r.table.Ensure(ctx, db, kind, idPtr)
}

func (r Repository) GetLlmResponse(ctx context.Context, db sqlutil.Executor, respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return r.table.GetLlmResponse(ctx, db, respID)
}

func (r Repository) Ref(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) sourcedomain.Source {
	return r.table.Ref(ctx, db, sourceID)
}

func (r Repository) Refs(ctx context.Context, db sqlutil.Executor, ids []sourcedomain.SourceId) []sourcedomain.Source {
	result := make([]sourcedomain.Source, 0, len(ids))
	for _, id := range ids {
		result = append(result, r.Ref(ctx, db, id))
	}
	return result
}

func (r Repository) Trace(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
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
		row, found, err := r.table.Get(ctx, db, current)
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
		response, found, err := r.table.GetLlmResponse(ctx, db, sourcedomain.NewRespId(row.IDPtr))
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
