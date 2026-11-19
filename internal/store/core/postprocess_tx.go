package store

import (
	"context"
	"fmt"
)

// WithPostprocessTx runs postprocess writes in one main.db transaction.
func (s *StoreContainer) WithPostprocessTx(ctx context.Context, fn func(*StoreContainer) error) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store db is required")
	}
	if s.sourceDB != s.db || s.solveDB != s.db {
		return fmt.Errorf("postprocess requires a single database")
	}
	return s.WithTx(ctx, fn)
}
