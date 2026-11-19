package csolve

import (
	"context"

	adapter "github.com/G226Medi03/MorbesPipeline/internal/pipeline/orchestration/adapters/postprocess"
	store "github.com/G226Medi03/MorbesPipeline/internal/store"
)

const RejectionStatusName = adapter.RejectionStatusName

// Postprocess applies deterministic postprocess rules to the current case.
func Postprocess(ctx context.Context, st *store.StoreContainer, state *pipelineState) error {
	return adapter.Run(ctx, st, state)
}
