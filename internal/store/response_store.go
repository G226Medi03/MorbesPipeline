package store

import (
	"fmt"

	source "github.com/G226Medi03/SourceEntity"
)

// FindLlmResponse returns the cached LLM response for the cache key.
func (s *StoreContainer) FindLlmResponse(model, prompt string, referedSourceID source.SourceId) (*source.LlmResponse, bool, error) {
	if s == nil || s.Response == nil {
		return nil, false, fmt.Errorf("llm response store is not initialized")
	}
	return s.Response.FindLlmResponse(model, prompt, referedSourceID)
}

// GetOrInsertLlmResponse stores output and returns the cached response.
func (s *StoreContainer) GetOrInsertLlmResponse(model, prompt string, referedSourceID source.SourceId, output string) (*source.LlmResponse, error) {
	unlock := s.lockWrite()
	defer unlock()
	if s == nil || s.Response == nil {
		return nil, fmt.Errorf("llm response store is not initialized")
	}
	return s.Response.GetOrInsertLlmResponse(model, prompt, referedSourceID, output)
}

// RefreshLlmResponse refreshes the cached output for an existing cache key.
func (s *StoreContainer) RefreshLlmResponse(model, prompt string, referedSourceID source.SourceId, output string) (*source.LlmResponse, error) {
	unlock := s.lockWrite()
	defer unlock()
	if s == nil || s.Response == nil {
		return nil, fmt.Errorf("llm response store is not initialized")
	}
	return s.Response.RefreshLlmResponse(model, prompt, referedSourceID, output)
}
