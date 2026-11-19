package postprocess

import "github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/deps"

type PostprocessService struct {
	store deps.Store
}

func New(store deps.Store) *PostprocessService {
	return &PostprocessService{store: store}
}
