package service

import (
	"fmt"

	source "github.com/G226Medi03/SourceEntity"
)

// FindLlmResponse returns the cached LLM response for the cache key.
func (s *SourceService) FindLlmResponse(model, prompt string, referedSourceID source.SourceId) (*source.LlmResponse, bool, error) {
	if s == nil || s.store.ResponseStore() == nil {
		return nil, false, fmt.Errorf("llm response store is not initialized")
	}
	return s.store.ResponseStore().FindLlmResponse(model, prompt, referedSourceID)
}

// GetOrInsertLlmResponse stores output and returns the cached response.
func (s *SourceService) GetOrInsertLlmResponse(model, prompt string, referedSourceID source.SourceId, output string) (*source.LlmResponse, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	if s == nil || s.store.ResponseStore() == nil {
		return nil, fmt.Errorf("llm response store is not initialized")
	}
	return s.store.ResponseStore().GetOrInsertLlmResponse(model, prompt, referedSourceID, output)
}

// RefreshLlmResponse refreshes the cached output for an existing cache key.
func (s *SourceService) RefreshLlmResponse(model, prompt string, referedSourceID source.SourceId, output string) (*source.LlmResponse, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	if s == nil || s.store.ResponseStore() == nil {
		return nil, fmt.Errorf("llm response store is not initialized")
	}
	return s.store.ResponseStore().RefreshLlmResponse(model, prompt, referedSourceID, output)
}
