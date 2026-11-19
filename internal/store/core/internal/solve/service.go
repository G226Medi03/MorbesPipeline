package solve

import "github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/deps"

type SolveService struct {
	store deps.Store
}

func New(store deps.Store) *SolveService {
	return &SolveService{store: store}
}
