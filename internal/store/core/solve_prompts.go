package store

import (
	"errors"

	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
)

func (s *StoreContainer) GetPrompt(stage string) (string, int, error) {
	return s.Solve.Prompt.Get(s.ctx, s.solveExecutor(), stage)
}

// Prompt implements the MorbesLLMService prompt store contract.
func (s *StoreContainer) Prompt(stage string) (string, int, bool, error) {
	template, version, err := s.GetPrompt(stage)
	if err != nil {
		if errors.Is(err, storerepo.ErrPromptNotFound) {
			return "", 0, false, nil
		}
		return "", 0, false, err
	}
	return template, version, true, nil
}

func (s *StoreContainer) SavePrompt(stage, template string) (int, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.Solve.Prompt.Save(s.ctx, s.solveExecutor(), stage, template)
}
