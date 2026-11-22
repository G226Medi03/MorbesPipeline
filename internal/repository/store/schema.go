package store

import (
	responsestore "github.com/G226Medi03/MorbesLLMService/responsestore"
	"github.com/G226Medi03/MorbesPipeline/internal/repository/schema"
)

func (s *Store) InitSchema() error {
	if _, err := s.MedicalEntities(); err != nil {
		return err
	}
	if err := schema.CleanupLegacySchema(s.Context(), s); err != nil {
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
	if err := s.Benchmark.Init(s.Context(), s.SolveExecutor()); err != nil {
		return err
	}
	if err := schema.EnsureSolveIDColumns(s.Context(), s.SolveExecutor()); err != nil {
		return err
	}
	return nil
}
