// Package store is the single SQLite persistence layer for medical data.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/G226Medi03/MedicalEntityRepository/repo"
	storesource "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentityrepository"
	sqlutil "github.com/G226Medi03/MorbesPipeline/internal/external/sqlcore/sqlite"
	storegraph "github.com/G226Medi03/MorbesPipeline/internal/persistence/graph"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence/option"
	storesolve "github.com/G226Medi03/MorbesPipeline/internal/persistence/solve"
)

type Store struct {
	ctx context.Context
	db  *sql.DB
	tx  *sql.Tx
	mu  *sync.Mutex

	sourceDB *sql.DB
	sourceTx *sql.Tx
	solveDB  *sql.DB
	solveTx  *sql.Tx

	entityStore *repo.Store
	entityErr   error
	entityMu    *sync.Mutex

	OptionRepository medicalprocedureoption.Repository
	SourceRepository storesource.Repository
	SolveRepository  storesolve.Repository
	GraphRepository  storegraph.Repository
}

// New keeps the single-file behavior used by tests and command-line tools.
// The server opens the medical/source/solve files separately with NewWithDBs.
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
		entityMu: &sync.Mutex{},
		sourceDB: sourceDB,
		solveDB:  solveDB,
	}
	value.bind()
	return value
}

func (s *Store) bind() {
	s.entityStore = nil
	s.entityErr = nil
	s.OptionRepository = medicalprocedureoption.NewRepository(s.ctx, s)
	s.GraphRepository = storegraph.NewRepository(s.ctx, s)
}

func (s *Store) entities() (*repo.Store, error) {
	s.entityMu.Lock()
	defer s.entityMu.Unlock()
	if s.entityStore != nil {
		return s.entityStore, nil
	}
	if s.entityErr != nil {
		return nil, s.entityErr
	}
	entityStore, err := repo.NewStore(s.ctx, s)
	if err != nil {
		s.entityErr = err
		return nil, err
	}
	s.entityStore = entityStore
	return entityStore, nil
}

func (s *Store) withTx(tx *sql.Tx) *Store {
	clone := *s
	clone.tx = tx
	clone.sourceTx = nil
	clone.solveTx = nil
	clone.bind()
	return &clone
}

func (s *Store) withSourceTx(tx *sql.Tx) *Store {
	clone := *s
	clone.sourceTx = tx
	return &clone
}

func (s *Store) withSolveTx(tx *sql.Tx) *Store {
	clone := *s
	clone.solveTx = tx
	return &clone
}

func (s *Store) sourceExecutor() sqlutil.Executor {
	if s.sourceTx != nil {
		return s.sourceTx
	}
	if s.sourceDB != nil && s.sourceDB != s.db {
		return s.sourceDB
	}
	return s
}

func (s *Store) solveExecutor() sqlutil.Executor {
	if s.solveTx != nil {
		return s.solveTx
	}
	if s.solveDB != nil && s.solveDB != s.db {
		return s.solveDB
	}
	return s
}

// lockWrite serializes application writes while keeping LLM calls outside the
// lock. Calls made through a transaction clone already hold the parent lock.
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
