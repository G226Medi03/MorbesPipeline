// Package store is the single SQLite persistence layer for medical data.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/G226Medi03/MedicalEntityRepository/repo"
	responsestore "github.com/G226Medi03/MorbesLLMService/responsestore"
	storepostprocess "github.com/G226Medi03/MorbesPipeline/internal/store/postprocess"
	storesolve "github.com/G226Medi03/MorbesPipeline/internal/store/solve"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	storesource "github.com/G226Medi03/SourceEntityRepository"
)

type StoreContainer struct {
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
	Solve       storesolve.Repository
	Postprocess storepostprocess.Repository
}

// New keeps the single-file behavior used by tests and command-line tools.
// The server opens the medical/source/solve files separately with NewWithDBs.
func New(ctx context.Context, db *sql.DB) *StoreContainer {
	return NewWithDBs(ctx, db, db, db)
}

func NewWithDBs(ctx context.Context, medicalDB, sourceDB, solveDB *sql.DB) *StoreContainer {
	if sourceDB == nil {
		sourceDB = medicalDB
	}
	if solveDB == nil {
		solveDB = medicalDB
	}
	value := &StoreContainer{
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

func (s *StoreContainer) bind() {
	s.Medical = nil
	s.Response = nil
	s.Postprocess = storepostprocess.NewRepository(s.ctx, s)
}

func (s *StoreContainer) responseStoreWithExecutor() *responsestore.Store {
	if s.Response == nil {
		return nil
	}
	return s.Response.WithExecutor(s.sourceExecutor())
}

func (s *StoreContainer) entities() (*repo.Store, error) {
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

func (s *StoreContainer) withTx(tx *sql.Tx) *StoreContainer {
	clone := *s
	clone.tx = tx
	clone.sourceTx = nil
	clone.solveTx = nil
	clone.bind()
	clone.Response = s.responseStoreWithExecutor()
	return &clone
}

func (s *StoreContainer) withSourceTx(tx *sql.Tx) *StoreContainer {
	clone := *s
	clone.sourceTx = tx
	clone.Response = s.Response.WithExecutor(tx)
	return &clone
}

func (s *StoreContainer) withSolveTx(tx *sql.Tx) *StoreContainer {
	clone := *s
	clone.solveTx = tx
	return &clone
}

func (s *StoreContainer) sourceExecutor() sqlutil.Executor {
	if s.sourceTx != nil {
		return s.sourceTx
	}
	if s.sourceDB != nil && s.sourceDB != s.db {
		return s.sourceDB
	}
	return s
}

func (s *StoreContainer) solveExecutor() sqlutil.Executor {
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
func (s *StoreContainer) lockWrite() func() {
	if s == nil || s.mu == nil || s.tx != nil {
		return func() {}
	}
	s.mu.Lock()
	return s.mu.Unlock
}

func (s *StoreContainer) WithTx(ctx context.Context, fn func(*StoreContainer) error) error {
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

func (s *StoreContainer) WithSolveTx(ctx context.Context, fn func(*StoreContainer) error) error {
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

func (s *StoreContainer) WithSourceTx(ctx context.Context, fn func(*StoreContainer) error) error {
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

func (s *StoreContainer) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.ExecContext(ctx, query, args...)
	}
	return s.db.ExecContext(ctx, query, args...)
}

func (s *StoreContainer) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(ctx, query, args...)
	}
	return s.db.QueryContext(ctx, query, args...)
}

func (s *StoreContainer) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRowContext(ctx, query, args...)
	}
	return s.db.QueryRowContext(ctx, query, args...)
}
