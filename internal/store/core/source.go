package store

import (
	"context"

	sourcedomain "github.com/G226Medi03/SourceEntity"
)

func (s *StoreContainer) GetSource(sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error) {
	return s.Source.Get(s.ctx, s.sourceExecutor(), sourceID)
}

func (s *StoreContainer) EnsureSource(kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Source.Ensure(s.ctx, s.sourceExecutor(), kind, idPtr)
}

func (s *StoreContainer) GetLlmResponseRow(respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return s.Source.GetLlmResponse(s.ctx, s.sourceExecutor(), respID)
}

func (s *StoreContainer) ref(ctx context.Context, sourceID sourcedomain.SourceId) sourcedomain.Source {
	return s.Source.Ref(ctx, s.sourceExecutor(), sourceID)
}

func (s *StoreContainer) TraceSource(sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
	return s.Source.Trace(s.ctx, s.sourceExecutor(), sourceID)
}

func (s *StoreContainer) refs(ctx context.Context, ids []sourcedomain.SourceId) []sourcedomain.Source {
	return s.Source.Refs(ctx, s.sourceExecutor(), ids)
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
