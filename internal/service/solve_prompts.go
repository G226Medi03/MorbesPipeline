package service

import (
	"errors"

	srepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository/solve"
)

func (s *SolveService) GetPrompt(stage string) (string, int, error) {
	return s.store.SolveRepo().Prompt.Get(s.store.Context(), s.store.SolveExecutor(), stage)
}

// Prompt implements the MorbesLLMService prompt store contract.
func (s *SolveService) Prompt(stage string) (string, int, bool, error) {
	template, version, err := s.GetPrompt(stage)
	if err != nil {
		if errors.Is(err, srepo.ErrPromptNotFound) {
			return "", 0, false, nil
		}
		return "", 0, false, err
	}
	return template, version, true, nil
}

func (s *SolveService) SavePrompt(stage, template string) (int, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.SolveRepo().Prompt.Save(s.store.Context(), s.store.SolveExecutor(), stage, template)
}
