package source

import "github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/deps"

type SourceService struct {
	store deps.Store
}

func New(store deps.Store) *SourceService {
	return &SourceService{store: store}
}
