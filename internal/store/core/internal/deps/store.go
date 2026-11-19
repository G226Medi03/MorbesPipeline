package deps

import (
	"context"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	responsestore "github.com/G226Medi03/MorbesLLMService/responsestore"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
	storesource "github.com/G226Medi03/SourceEntityRepository"
)

// Store is the dependency surface used by core internal services.
type Store interface {
	Context() context.Context
	MainExecutor() sqlutil.Executor
	SourceExecutor() sqlutil.Executor
	SolveExecutor() sqlutil.Executor

	Entities() (*repo.Store, error)
	SourceRepository() storesource.Repository
	SolveRepository() storerepo.SolveRepository
	BenchmarkRepository() storerepo.BenchmarkRepository
	PostprocessRepository() storerepo.PostprocessRepository
	ResponseStore() *responsestore.Store

	LockWrite() func()

	RunTx(ctx context.Context, fn func(Store) error) error
	RunSolveTx(ctx context.Context, fn func(Store) error) error

	StatusIDForValue(value repo.IEStatus) (int64, error)
	ActionIDForValue(value repo.IEAction) (int64, error)
	EnsureSource(kind source.SourceKind, idPtr int64) (source.SourceId, error)
}
