package store

import (
	"errors"

	solvesql "github.com/G226Medi03/MorbesPipeline/internal/persistence/solve"
)

func (s *Store) GetPrompt(stage string) (string, int, error) {
	return s.SolveRepository.Prompt.Get(s.ctx, s.solveExecutor(), stage)
}

// Prompt implements the MorbesLLMService prompt store contract.
func (s *Store) Prompt(stage string) (string, int, bool, error) {
	template, version, err := s.GetPrompt(stage)
	if err != nil {
		if errors.Is(err, solvesql.ErrPromptNotFound) {
			return "", 0, false, nil
		}
		return "", 0, false, err
	}
	return template, version, true, nil
}

func (s *Store) SavePrompt(stage, template string) (int, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Prompt.Save(s.ctx, s.solveExecutor(), stage, template)
}
