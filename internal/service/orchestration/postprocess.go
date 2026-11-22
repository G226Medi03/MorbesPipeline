package orchestration

import (
	"context"

	adapter "github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/adapters/postprocess"
	store "github.com/G226Medi03/MorbesPipeline/internal/service/store"
)

const RejectionStatusName = adapter.RejectionStatusName

// Postprocess applies deterministic postprocess rules to the current case.
func Postprocess(ctx context.Context, st *store.StoreContainer, state *pipelineState) error {
	return adapter.Run(ctx, st, state)
}
