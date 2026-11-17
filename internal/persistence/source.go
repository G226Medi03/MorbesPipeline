package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/support/sourceutil"
	sourcedomain "github.com/G226Medi03/SourceEntity"
)

func (s *Store) GetSource(sourceID sourcedomain.SourceId) (sourcedomain.SourceRow, bool, error) {
	return s.SourceRepository.Source.Get(s.ctx, s.sourceExecutor(), sourceID)
}

// EnsureSource returns a stable source row for a non-file entity such as an
// exam question created without an uploaded source.
func (s *Store) EnsureSource(sourceType string, idPtr int64) (sourcedomain.SourceId, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.SourceRepository.Source.Ensure(s.ctx, s.sourceExecutor(), "", sourceType, idPtr)
}

func (s *Store) GetLlmResponseRow(respID sourcedomain.RespId) (sourcedomain.LlmResponseRow, bool, error) {
	return s.SourceRepository.LlmResponse.Get(s.ctx, s.sourceExecutor(), respID)
}

func (s *Store) FindLlmResponse(model, prompt string, referedSourceID sourcedomain.SourceId) (*sourcedomain.LlmResponse, bool, error) {
	row, found, err := s.SourceRepository.LlmResponse.Find(s.ctx, s.sourceExecutor(), model, prompt, referedSourceID)
	if err != nil || !found {
		return nil, false, err
	}
	sourceRow, found, err := s.SourceRepository.Source.FindByVariant(s.ctx, s.sourceExecutor(), sourcedomain.SourceTypeLlmResponse, row.RespID.ToI64())
	if err != nil || !found {
		return nil, false, err
	}
	response := sourcedomain.NewLlmResponse(row.Model, row.Prompt, row.Output)
	response.SourceID = sourceRow.SourceID
	response.RespID = row.RespID
	response.ReferedSourceID = row.ReferedSourceID
	return &response, true, nil
}

func (s *Store) GetOrInsertLlmResponse(model, prompt string, referedSourceID sourcedomain.SourceId, output string) (*sourcedomain.LlmResponse, error) {
	unlock := s.lockWrite()
	defer unlock()
	respID, err := s.SourceRepository.LlmResponse.EnsureID(s.ctx, s.sourceExecutor(), model, prompt, referedSourceID, output)
	if err != nil {
		return nil, err
	}
	sourceID, err := s.SourceRepository.Source.Ensure(s.ctx, s.sourceExecutor(), "", sourcedomain.SourceTypeLlmResponse, respID.ToI64())
	if err != nil {
		return nil, err
	}
	return s.getLlmResponseBySource(sourceID)
}

// RefreshLlmResponse updates the cached output for an existing prompt/model
// pair, or inserts a new response when no cache row exists. The resp_id remains
// stable so existing solve_llm_calls and answer-check links stay valid.
func (s *Store) RefreshLlmResponse(model, prompt string, referedSourceID sourcedomain.SourceId, output string) (*sourcedomain.LlmResponse, error) {
	unlock := s.lockWrite()
	defer unlock()
	row, found, err := s.SourceRepository.LlmResponse.Find(s.ctx, s.sourceExecutor(), model, prompt, referedSourceID)
	if err != nil {
		return nil, err
	}
	if found {
		if err := s.SourceRepository.LlmResponse.UpdateOutput(s.ctx, s.sourceExecutor(), row.RespID, output); err != nil {
			return nil, err
		}
		sourceRow, sourceFound, sourceErr := s.SourceRepository.Source.FindByVariant(
			s.ctx, s.sourceExecutor(), sourcedomain.SourceTypeLlmResponse, row.RespID.ToI64())
		if sourceErr != nil {
			return nil, sourceErr
		}
		if !sourceFound {
			return nil, fmt.Errorf("llm response source 를 찾지 못했습니다: %d", row.RespID)
		}
		response := sourcedomain.NewLlmResponse(row.Model, row.Prompt, output)
		response.SourceID = sourceRow.SourceID
		response.RespID = row.RespID
		response.ReferedSourceID = row.ReferedSourceID
		return &response, nil
	}
	respID, err := s.SourceRepository.LlmResponse.EnsureID(s.ctx, s.sourceExecutor(), model, prompt, referedSourceID, output)
	if err != nil {
		return nil, err
	}
	sourceID, err := s.SourceRepository.Source.Ensure(s.ctx, s.sourceExecutor(), "", sourcedomain.SourceTypeLlmResponse, respID.ToI64())
	if err != nil {
		return nil, err
	}
	return s.getLlmResponseBySource(sourceID)
}

func (s *Store) getLlmResponseBySource(sourceID sourcedomain.SourceId) (*sourcedomain.LlmResponse, error) {
	sourceRow, found, err := s.SourceRepository.Source.Get(s.ctx, s.sourceExecutor(), sourceID)
	if err != nil || !found {
		return nil, err
	}
	row, found, err := s.SourceRepository.LlmResponse.Get(s.ctx, s.sourceExecutor(), sourceutil.RespID(sourceRow.IDPtr))
	if err != nil || !found {
		return nil, err
	}
	response := sourcedomain.NewLlmResponse(row.Model, row.Prompt, row.Output)
	response.SourceID = sourceID
	response.RespID = row.RespID
	response.ReferedSourceID = row.ReferedSourceID
	return &response, nil
}

func (s *Store) ref(ctx context.Context, sourceID sourcedomain.SourceId) sourcedomain.SourceRef {
	return s.SourceRepository.Source.Ref(ctx, s.sourceExecutor(), sourceID)
}

// TraceSource follows LlmResponse.refered_source_id links from an immediate
// LLM source to the first upstream source that has a path. The immediate
// source is retained separately so callers can keep provenance precise.
func (s *Store) TraceSource(sourceID sourcedomain.SourceId) (sourcedomain.SourceTrace, error) {
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
		row, found, err := s.SourceRepository.Source.Get(s.ctx, s.sourceExecutor(), current)
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
		response, found, err := s.SourceRepository.LlmResponse.Get(s.ctx, s.sourceExecutor(), sourceutil.RespID(row.IDPtr))
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

func (s *Store) refs(ctx context.Context, ids []sourcedomain.SourceId) []sourcedomain.SourceRef {
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
