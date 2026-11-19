// Package store owns the SQLite persistence context, transaction, and repositories.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	responsestore "github.com/G226Medi03/MorbesLLMService/responsestore"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	storesource "github.com/G226Medi03/SourceEntityRepository"
)

// Store owns the database context and repositories.
type Store struct {
	ctx context.Context
	mu  *sync.Mutex

	db *sql.DB
	tx *sql.Tx

	sourceDB *sql.DB
	sourceTx *sql.Tx

	solveDB *sql.DB
	solveTx *sql.Tx

	Medical  *repo.Store
	Response *responsestore.Store

	Source      storesource.Repository
	Solve       storerepo.SolveRepository
	Benchmark   storerepo.BenchmarkRepository
	Postprocess storerepo.PostprocessRepository
	Entity      storerepo.EntityRepository
}

// New keeps the single-file behavior used by tests and command-line tools.
func New(ctx context.Context, db *sql.DB) *Store {
	return NewWithDBs(ctx, db, db, db)
}

func NewWithDBs(ctx context.Context, medicalDB, sourceDB, solveDB *sql.DB) *Store {
	if sourceDB == nil {
		sourceDB = medicalDB
	}
	if solveDB == nil {
		solveDB = medicalDB
	}
	value := &Store{
		ctx:      ctx,
		db:       medicalDB,
		mu:       &sync.Mutex{},
		sourceDB: sourceDB,
		solveDB:  solveDB,
		Source:   storesource.Repository{},
	}
	value.bind()
	return value
}

func (s *Store) bind() {
	s.Medical = nil
	s.Response = nil
	s.Solve = storerepo.NewSolveRepository()
	s.Benchmark = storerepo.BenchmarkRepository{}
	s.Postprocess = storerepo.NewPostprocessRepository()
}

func (s *Store) Context() context.Context {
	return s.ctx
}

func (s *Store) MainExecutor() sqlutil.Executor {
	return s
}

func (s *Store) SourceExecutor() sqlutil.Executor {
	if s.sourceTx != nil {
		return s.sourceTx
	}
	if s.sourceDB != nil && s.sourceDB != s.db {
		return s.sourceDB
	}
	return s
}

func (s *Store) SolveExecutor() sqlutil.Executor {
	if s.solveTx != nil {
		return s.solveTx
	}
	if s.solveDB != nil && s.solveDB != s.db {
		return s.solveDB
	}
	return s
}

func (s *Store) SourceRepo() storesource.Repository {
	return s.Source
}

func (s *Store) SolveRepo() storerepo.SolveRepository {
	return s.Solve
}

func (s *Store) BenchmarkRepo() storerepo.BenchmarkRepository {
	return s.Benchmark
}

func (s *Store) PostprocessRepo() storerepo.PostprocessRepository {
	return s.Postprocess
}

func (s *Store) EntityRepo() storerepo.EntityRepository {
	return s.Entity
}

func (s *Store) ResponseStore() *responsestore.Store {
	return s.Response
}

func (s *Store) Entities() (*repo.Store, error) {
	return s.entities()
}

func (s *Store) LockWrite() func() {
	return s.lockWrite()
}

func (s *Store) responseStoreWithExecutor() *responsestore.Store {
	if s.Response == nil {
		return nil
	}
	return s.Response.WithExecutor(s.SourceExecutor())
}

func (s *Store) entities() (*repo.Store, error) {
	if s.Medical != nil {
		return s.Medical, nil
	}
	medical, err := repo.NewStore(s.ctx, s)
	if err != nil {
		return nil, err
	}
	s.Medical = medical
	return medical, nil
}

func (s *Store) withTx(tx *sql.Tx) *Store {
	clone := *s
	clone.tx = tx
	clone.sourceTx = nil
	clone.solveTx = nil
	clone.bind()
	clone.Response = s.responseStoreWithExecutor()
	return &clone
}

func (s *Store) withSourceTx(tx *sql.Tx) *Store {
	clone := *s
	clone.sourceTx = tx
	clone.Response = s.Response.WithExecutor(tx)
	return &clone
}

func (s *Store) withSolveTx(tx *sql.Tx) *Store {
	clone := *s
	clone.solveTx = tx
	return &clone
}

func (s *Store) lockWrite() func() {
	if s == nil || s.mu == nil || s.tx != nil {
		return func() {}
	}
	s.mu.Lock()
	return s.mu.Unlock
}

func (s *Store) WithTx(ctx context.Context, fn func(*Store) error) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store db 가 없습니다")
	}
	if s.tx != nil {
		return fn(s)
	}
	unlock := s.lockWrite()
	defer unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store transaction 시작 실패: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(s.withTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store transaction 커밋 실패: %w", err)
	}
	return nil
}

func (s *Store) WithSolveTx(ctx context.Context, fn func(*Store) error) error {
	if s == nil || s.solveDB == nil {
		return fmt.Errorf("solve db 가 없습니다")
	}
	if s.solveTx != nil {
		return fn(s)
	}
	unlock := s.lockWrite()
	defer unlock()
	tx, err := s.solveDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("solve transaction 시작 실패: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(s.withSolveTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("solve transaction 커밋 실패: %w", err)
	}
	return nil
}

func (s *Store) WithSourceTx(ctx context.Context, fn func(*Store) error) error {
	if s == nil || s.sourceDB == nil {
		return fmt.Errorf("source db 가 없습니다")
	}
	if s.sourceTx != nil {
		return fn(s)
	}
	unlock := s.lockWrite()
	defer unlock()
	tx, err := s.sourceDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("source transaction 시작 실패: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(s.withSourceTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("source transaction 커밋 실패: %w", err)
	}
	return nil
}

// WithPostprocessTx runs postprocess writes in one main.db transaction.
func (s *Store) WithPostprocessTx(ctx context.Context, fn func(*Store) error) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store db is required")
	}
	if s.sourceDB != s.db || s.solveDB != s.db {
		return fmt.Errorf("postprocess requires a single database")
	}
	return s.WithTx(ctx, fn)
}

func (s *Store) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.ExecContext(ctx, query, args...)
	}
	return s.db.ExecContext(ctx, query, args...)
}

func (s *Store) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(ctx, query, args...)
	}
	return s.db.QueryContext(ctx, query, args...)
}

func (s *Store) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRowContext(ctx, query, args...)
	}
	return s.db.QueryRowContext(ctx, query, args...)
}
