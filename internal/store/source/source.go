package source

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	sourcedomain "github.com/G226Medi03/SourceEntity"
)

// Table is the source persistence primitive surface.
type Table interface {
	Get(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error)
	Ensure(ctx context.Context, db sqlutil.Executor, kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error)
	GetLlmResponse(ctx context.Context, db sqlutil.Executor, respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error)
	Ref(ctx context.Context, db sqlutil.Executor, sourceID sourcedomain.SourceId) sourcedomain.Source
}
