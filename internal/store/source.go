package store

import (
	"context"
	"strings"

	sourcedomain "github.com/G226Medi03/SourceEntity"
)

func (s *StoreContainer) GetSource(sourceID sourcedomain.SourceId) (sourcedomain.SourceRow, bool, error) {
	return s.Source.Source.Get(s.ctx, s.sourceExecutor(), sourceID)
}

// EnsureSource returns a stable source row for a non-file entity such as an
// exam question created without an uploaded source.
func (s *StoreContainer) EnsureSource(sourceType string, idPtr int64) (sourcedomain.SourceId, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Source.Source.Ensure(s.ctx, s.sourceExecutor(), "", sourceType, idPtr)
}

func (s *StoreContainer) GetLlmResponseRow(respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return s.Source.LlmResponse.Get(s.ctx, s.sourceExecutor(), respID)
}

func (s *StoreContainer) ref(ctx context.Context, sourceID sourcedomain.SourceId) sourcedomain.SourceRef {
	return s.Source.Source.Ref(ctx, s.sourceExecutor(), sourceID)
}

// TraceSource follows LlmResponse.refered_source_id links from an immediate
// LLM source to the first upstream source that has a path. The immediate
// source is retained separately so callers can keep provenance precise.
func (s *StoreContainer) TraceSource(sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
	trace := sourcedomain.SourceTrace{Chain: []sourcedomain.SourceRef{}}
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
		ref := sourcedomain.SourceRef{SourceID: row.SourceID, SourceType: row.SourceType, Path: row.Path}
		trace.Chain = append(trace.Chain, ref)
		if depth == 0 {
			trace.Immediate = ref
		}
		if row.SourceType != sourcedomain.SourceTypeLlmResponse {
			if strings.TrimSpace(row.Path) != "" {
				origin := ref
				trace.Origin = &origin
			}
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

func (s *StoreContainer) refs(ctx context.Context, ids []sourcedomain.SourceId) []sourcedomain.SourceRef {
	result := make([]sourcedomain.SourceRef, 0, len(ids))
	for _, sourceID := range ids {
		result = append(result, s.ref(ctx, sourceID))
	}
	return result
}

func firstSourceID(sources []sourcedomain.SourceRef) sourcedomain.SourceId {
	for _, item := range sources {
		if item.SourceID.ToI64() > 0 {
			return item.SourceID
		}
	}
	return sourcedomain.SourceId{}
}

func sourceIDs(sources []sourcedomain.SourceRef) []sourcedomain.SourceId {
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
