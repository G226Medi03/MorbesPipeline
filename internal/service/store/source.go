package store

import (
	"context"

	sourcedomain "github.com/G226Medi03/SourceEntity"
)

func (s *SourceService) GetSource(sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error) {
	return s.store.SourceRepo().Get(s.store.Context(), s.store.SourceExecutor(), sourceID)
}

func (s *SourceService) EnsureSource(kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SourceRepo().Ensure(s.store.Context(), s.store.SourceExecutor(), kind, idPtr)
}

func (s *SourceService) GetLlmResponseRow(respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return s.store.SourceRepo().GetLlmResponse(s.store.Context(), s.store.SourceExecutor(), respID)
}

func (s *SourceService) ref(ctx context.Context, sourceID sourcedomain.SourceId) sourcedomain.Source {
	return s.store.SourceRepo().Ref(ctx, s.store.SourceExecutor(), sourceID)
}

func (s *SourceService) TraceSource(sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
	return s.store.SourceRepo().Trace(s.store.Context(), s.store.SourceExecutor(), sourceID)
}

func (s *SourceService) refs(ctx context.Context, ids []sourcedomain.SourceId) []sourcedomain.Source {
	return s.store.SourceRepo().Refs(ctx, s.store.SourceExecutor(), ids)
}

func firstSourceID(sources []sourcedomain.Source) sourcedomain.SourceId {
	for _, item := range sources {
		if item.SourceID.ToI64() > 0 {
			return item.SourceID
		}
	}
	return sourcedomain.SourceId{}
}

func sourceIDs(sources []sourcedomain.Source) []sourcedomain.SourceId {
	seen := make(map[sourcedomain.SourceId]struct{}, len(sources))
	result := make([]sourcedomain.SourceId, 0, len(sources))
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
