package benchmark

import "github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/deps"

type BenchmarkService struct {
	store deps.Store
}

func New(store deps.Store) *BenchmarkService {
	return &BenchmarkService{store: store}
}
