package facade

import (
	responsestore "github.com/G226Medi03/MorbesLLMService/responsestore"
	"github.com/G226Medi03/MorbesPipeline/internal/store/internal/repository/migration"
	"github.com/G226Medi03/MorbesPipeline/internal/store/internal/repository/schema"
)

func (s *StoreContainer) InitSchema() error {
	if err := migration.ApplyMigration(s.Context(), s); err != nil {
		return err
	}
	if _, err := s.Entities(); err != nil {
		return err
	}
	if err := s.Postprocess.Init(s.Context(), s); err != nil {
		return err
	}
	if err := s.Source.Init(s.Context(), s.SourceExecutor()); err != nil {
		return err
	}
	responseStore, err := responsestore.NewStore(s.Context(), s.SourceExecutor())
	if err != nil {
		return err
	}
	s.Response = responseStore
	if err := s.Solve.Init(s.Context(), s.SolveExecutor()); err != nil {
		return err
	}
	if err := s.ensure_llm_problem_results(); err != nil {
		return err
	}
	if err := schema.EnsureStatusOtherName(s.Context(), s); err != nil {
		return err
	}
	if err := schema.EnsureEntitySources(s.Context(), s); err != nil {
		return err
	}
	if err := schema.EnsureSolveIDColumns(s.Context(), s.SolveExecutor()); err != nil {
		return err
	}
	return nil
}
