package store

import (
	responsestore "github.com/G226Medi03/MorbesLLMService/responsestore"
	coreschema "github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/schema"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
)

func (s *StoreContainer) InitSchema() error {
	if err := storerepo.ApplyMigration(s.ctx, s); err != nil {
		return err
	}
	if _, err := s.entities(); err != nil {
		return err
	}
	if err := s.Postprocess.Init(s.ctx, s); err != nil {
		return err
	}
	if err := s.Source.Init(s.ctx, s.sourceExecutor()); err != nil {
		return err
	}
	responseStore, err := responsestore.NewStore(s.ctx, s.sourceExecutor())
	if err != nil {
		return err
	}
	s.Response = responseStore
	if err := s.Solve.Init(s.ctx, s.solveExecutor()); err != nil {
		return err
	}
	if err := s.ensure_llm_problem_results(); err != nil {
		return err
	}
	if err := coreschema.EnsureStatusOtherName(s.ctx, s); err != nil {
		return err
	}
	if err := coreschema.EnsureEntitySources(s.ctx, s); err != nil {
		return err
	}
	if err := coreschema.EnsureSolveIDColumns(s.ctx, s.solveExecutor()); err != nil {
		return err
	}
	return nil
}
