package entity

import (
	"github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/deps"
	source "github.com/G226Medi03/SourceEntity"
)

type EntityService struct {
	store deps.Store
}

func New(store deps.Store) *EntityService {
	return &EntityService{store: store}
}

func sourceIDs(sources []source.Source) []source.SourceId {
	seen := make(map[source.SourceId]struct{}, len(sources))
	result := make([]source.SourceId, 0, len(sources))
	for _, item := range sources {
		if item.SourceID.ToI64() <= 0 {
			continue
		}
		if _, ok := seen[item.SourceID]; ok {
			continue
		}
		seen[item.SourceID] = struct{}{}
		result = append(result, item.SourceID)
	}
	return result
}
