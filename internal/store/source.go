package store

import (
	"context"

	sourcedomain "github.com/G226Medi03/SourceEntity"
)

func (s *StoreContainer) GetSource(sourceID sourcedomain.SourceId) (sourcedomain.Source, bool, error) {
	return s.Source.Source.Get(s.ctx, s.sourceExecutor(), sourceID)
}

// EnsureSource returns a stable source row for a non-file entity such as an
// exam question created without an uploaded source.
func (s *StoreContainer) EnsureSource(kind sourcedomain.SourceKind, idPtr int64) (sourcedomain.SourceId, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Source.Source.Ensure(s.ctx, s.sourceExecutor(), kind, idPtr)
}

func (s *StoreContainer) GetLlmResponseRow(respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return s.Source.LlmResponse.Get(s.ctx, s.sourceExecutor(), respID)
}

func (s *StoreContainer) ref(ctx context.Context, sourceID sourcedomain.SourceId) sourcedomain.Source {
	return s.Source.Source.Ref(ctx, s.sourceExecutor(), sourceID)
}

// TraceSource follows LlmResponse.refered_source_id links from an immediate
// LLM source to the first upstream source.
func (s *StoreContainer) TraceSource(sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
	trace := sourcedomain.SourceTrace{}
	if sourceID.ToI64() <= 0 {
		return trace, nil
	}
	visited := map[sourcedomain.SourceId]struct{}{}
	current := sourceID
	for depth := 0; depth < 64 && current.ToI64() > 0; depth++ {
		if _, seen := visited[current]; seen {
			break
		}
		visited[current] = struct{}{}
		row, found, err := s.Source.Source.Get(s.ctx, s.sourceExecutor(), current)
		if err != nil {
			return trace, err
		}
		if !found {
			break
		}
		trace = append(trace, row)
		if row.SourceKind != sourcedomain.SourceKindLlmResponse {
			break
		}
		response, found, err := s.Source.LlmResponse.Get(s.ctx, s.sourceExecutor(), sourcedomain.NewRespId(row.IDPtr))
		if err != nil {
			return trace, err
		}
		if !found || response.ReferedSourceID.ToI64() <= 0 {
			break
		}
		current = response.ReferedSourceID
	}
	return trace, nil
}

func (s *StoreContainer) refs(ctx context.Context, ids []sourcedomain.SourceId) []sourcedomain.Source {
	result := make([]sourcedomain.Source, 0, len(ids))
	for _, sourceID := range ids {
		result = append(result, s.ref(ctx, sourceID))
	}
	return result
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
