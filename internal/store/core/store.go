// Package store is the public SQLite persistence facade.
package store

import (
	"context"
	"database/sql"

	coreservice "github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/service"
	corestore "github.com/G226Medi03/MorbesPipeline/internal/store/core/internal/store"
)

// StoreContainer is the public store facade.
type StoreContainer struct {
	*corestore.Store

	*coreservice.EntityService
	*coreservice.SolveService
	*coreservice.SourceService
	*coreservice.BenchmarkService
	*coreservice.PostprocessService
}

// New keeps the single-file behavior used by tests and command-line tools.
func New(ctx context.Context, db *sql.DB) *StoreContainer {
	return NewWithDBs(ctx, db, db, db)
}

func NewWithDBs(ctx context.Context, medicalDB, sourceDB, solveDB *sql.DB) *StoreContainer {
	return newContainer(corestore.NewWithDBs(ctx, medicalDB, sourceDB, solveDB))
}

func newContainer(st *corestore.Store) *StoreContainer {
	value := &StoreContainer{Store: st}
	value.bindServices()
	return value
}

func (s *StoreContainer) bindServices() {
	s.EntityService = coreservice.NewEntity(s.Store)
	s.SolveService = coreservice.NewSolve(s.Store)
	s.SourceService = coreservice.NewSource(s.Store)
	s.BenchmarkService = coreservice.NewBenchmark(s.Store)
	s.PostprocessService = coreservice.NewPostprocess(s.Store, s.EntityService)
}

func (s *StoreContainer) WithTx(ctx context.Context, fn func(*StoreContainer) error) error {
	return s.Store.WithTx(ctx, func(tx *corestore.Store) error {
		return fn(newContainer(tx))
	})
}

func (s *StoreContainer) WithSourceTx(ctx context.Context, fn func(*StoreContainer) error) error {
	return s.Store.WithSourceTx(ctx, func(tx *corestore.Store) error {
		return fn(newContainer(tx))
	})
}

func (s *StoreContainer) WithSolveTx(ctx context.Context, fn func(*StoreContainer) error) error {
	return s.Store.WithSolveTx(ctx, func(tx *corestore.Store) error {
		return fn(newContainer(tx))
	})
}

func (s *StoreContainer) WithPostprocessTx(ctx context.Context, fn func(*StoreContainer) error) error {
	return s.Store.WithPostprocessTx(ctx, func(tx *corestore.Store) error {
		return fn(newContainer(tx))
	})
}
