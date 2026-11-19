package source

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	sourcedomain "github.com/G226Medi03/SourceEntity"
	storesource "github.com/G226Medi03/SourceEntityRepository"
)

// Table owns the source and LLM response SQL primitives.
type Table struct {
	source      storesource.SourceTable
	llmResponse storesource.LlmResponseTable
}

func NewTable() Table { return Table{} }

func (t Table) Get(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error) {
	return t.source.Get(ctx, db, sourceID)
}

func (t Table) Ensure(ctx context.Context, db sqlutil.Executor, kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error) {
	return t.source.Ensure(ctx, db, kind, idPtr)
}

func (t Table) GetLlmResponse(ctx context.Context, db sqlutil.Executor, respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return t.llmResponse.Get(ctx, db, respID)
}

func (t Table) Ref(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) sourcedomain.Source {
	return t.source.Ref(ctx, db, sourceID)
}
