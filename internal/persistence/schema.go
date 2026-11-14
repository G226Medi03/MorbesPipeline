package store

import (
	sqlutil "github.com/G226Medi03/MorbesPipeline/internal/external/sqlcore/sqlite"
)

func (s *Store) InitSchema() error {
	if err := s.migrateLegacyMedicalSchema(s.ctx); err != nil {
		return err
	}
	if _, err := s.entities(); err != nil {
		return err
	}
	if err := s.OptionRepository.Init(); err != nil {
		return err
	}
	if err := s.SourceRepository.Init(s.ctx, s.sourceExecutor()); err != nil {
		return err
	}
	if err := s.SolveRepository.Init(s.ctx, s.solveExecutor()); err != nil {
		return err
	}
	if err := s.ensure_llm_problem_results(); err != nil {
		return err
	}
	if err := s.ensure_status_other_name(); err != nil {
		return err
	}
	if err := s.ensure_entity_sources(); err != nil {
		return err
	}
	if err := s.ensure_solve_id_columns(); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensure_solve_id_columns() error {
	executor := s.solveExecutor()
	for _, item := range []struct {
		table     string
		newColumn string
	}{
		{"llm_problem_results", "llm_problem_result_id"},
		{"solve_cases", "case_id"},
		{"solve_steps", "step_id"},
		{"solve_llm_calls", "call_id"},
	} {
		if err := sqlutil.EnsureRenamedColumn(s.ctx, executor, item.table, "id", item.newColumn); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensure_entity_sources() error {
	if _, err := s.ExecContext(s.ctx, `CREATE TABLE IF NOT EXISTS entity_sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_kind TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		source_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(entity_kind, entity_id, source_id)
	);`); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `CREATE INDEX IF NOT EXISTS idx_entity_sources_entity
		ON entity_sources(entity_kind, entity_id);`); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `CREATE INDEX IF NOT EXISTS idx_entity_sources_source
		ON entity_sources(source_id);`); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensure_status_other_name() error {
	if _, err := s.ExecContext(s.ctx, `CREATE TABLE IF NOT EXISTS status_other_name (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status_type TEXT NOT NULL,
		status_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 1,
		source_id INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_status_other_name_name
		ON status_other_name(name);`); err != nil {
		return err
	}
	return nil
}
