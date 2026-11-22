package store

import (
	"context"

	responsestore "github.com/G226Medi03/MorbesLLMService/responsestore"
	srepo "github.com/G226Medi03/MorbesPipeline/internal/repository/solve"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

func (s *Store) InitSchema() error {
	if _, err := s.MedicalEntities(); err != nil {
		return err
	}
	if err := cleanupLegacySchema(s.Context(), s); err != nil {
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
	if err := srepo.EnsureLegacyColumns(s.Context(), s.SolveExecutor()); err != nil {
		return err
	}
	return nil
}

func cleanupLegacySchema(ctx context.Context, db sqlutil.Executor) error {
	for _, table := range []string{"entity_sources", "status_other_name"} {
		if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS `+table); err != nil {
			return err
		}
	}
	return nil
}
