// Package facade is the public SQLite persistence facade.
package facade

import (
	"context"
	"database/sql"

	"github.com/G226Medi03/MorbesPipeline/internal/repository/benchmark"
	postprocessrepo "github.com/G226Medi03/MorbesPipeline/internal/repository/postprocess"
	srepo "github.com/G226Medi03/MorbesPipeline/internal/repository/solve"
	coreservice "github.com/G226Medi03/MorbesPipeline/internal/service/store"
	corestore "github.com/G226Medi03/MorbesPipeline/internal/repository/store"
)

type (
	Store = corestore.Store

	ProblemResultFilter    = benchmark.ProblemResultFilter
	ProblemResultRow       = benchmark.ProblemResultRow
	ExamQuestionDataCounts = benchmark.ExamQuestionDataCounts
	ExamQuestionCaseStats  = srepo.ExamQuestionCaseStats
	FailedCase             = srepo.FailedCase

	Operation             = postprocessrepo.Operation
	EntityType            = postprocessrepo.EntityType
	PostprocessOperation  = postprocessrepo.Operation
	PostprocessEntityType = postprocessrepo.EntityType
)

const (
	PostprocessorTransplantRejection  = postprocessrepo.PostprocessorTransplantRejection
	PostprocessorIncreasedIndexAction = postprocessrepo.PostprocessorIncreasedIndexAction
	PostprocessorCauseStatusResult    = postprocessrepo.PostprocessorCauseStatusResult

	OperationObserved    = postprocessrepo.OperationObserved
	OperationCreated     = postprocessrepo.OperationCreated
	OperationUpdated     = postprocessrepo.OperationUpdated
	OperationRemoved     = postprocessrepo.OperationRemoved
	OperationLinkAdded   = postprocessrepo.OperationLinkAdded
	OperationLinkRemoved = postprocessrepo.OperationLinkRemoved

	PostprocessOperationObserved    = postprocessrepo.OperationObserved
	PostprocessOperationCreated     = postprocessrepo.OperationCreated
	PostprocessOperationUpdated     = postprocessrepo.OperationUpdated
	PostprocessOperationRemoved     = postprocessrepo.OperationRemoved
	PostprocessOperationLinkAdded   = postprocessrepo.OperationLinkAdded
	PostprocessOperationLinkRemoved = postprocessrepo.OperationLinkRemoved

	EntityStatus    = postprocessrepo.EntityStatus
	EntityAction    = postprocessrepo.EntityAction
	EntityActor     = postprocessrepo.EntityActor
	EntityProcedure = postprocessrepo.EntityProcedure
	EntityOption    = postprocessrepo.EntityOption

	PostprocessEntityStatus    = postprocessrepo.EntityStatus
	PostprocessEntityAction    = postprocessrepo.EntityAction
	PostprocessEntityActor     = postprocessrepo.EntityActor
	PostprocessEntityProcedure = postprocessrepo.EntityProcedure
	PostprocessEntityOption    = postprocessrepo.EntityOption
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

func (s *StoreContainer) InitSchema() error {
	return s.Store.InitSchema()
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

// ProblemResults returns a consistent result page and model list.
func (s *StoreContainer) ProblemResults(ctx context.Context, filter ProblemResultFilter) ([]ProblemResultRow, int, []string, error) {
	var rows []ProblemResultRow
	var total int
	var models []string
	err := s.WithTx(ctx, func(tx *StoreContainer) error {
		var txErr error
		rows, txErr = tx.ListProblemResults(ctx, filter)
		if txErr != nil {
			return txErr
		}
		total, txErr = tx.CountProblemResults(ctx, filter)
		if txErr != nil {
			return txErr
		}
		models, txErr = tx.ListProblemResultModels(ctx)
		return txErr
	})
	return rows, total, models, err
}

// ProblemResult returns one composed result row.
func (s *StoreContainer) ProblemResult(ctx context.Context, id int64) (ProblemResultRow, bool, error) {
	var row ProblemResultRow
	var found bool
	err := s.WithTx(ctx, func(tx *StoreContainer) error {
		var txErr error
		row, found, txErr = tx.GetProblemResultByID(ctx, id)
		return txErr
	})
	return row, found, err
}
