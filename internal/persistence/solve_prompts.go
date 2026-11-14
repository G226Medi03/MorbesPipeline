package store

func (s *Store) GetPrompt(stage string) (string, int, error) {
	return s.SolveRepository.Prompt.Get(s.ctx, s.solveExecutor(), stage)
}

func (s *Store) SavePrompt(stage, template string) (int, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.SolveRepository.Prompt.Save(s.ctx, s.solveExecutor(), stage, template)
}
