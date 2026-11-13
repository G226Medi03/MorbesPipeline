package legacy

import (
	"context"
	"database/sql"
	"errors"

	"github.com/G226Medi03/MedicalEntity/ename/eaction"
	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	eprocedure "github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	"github.com/G226Medi03/MedicalEntity/ename/esymptom"
	"github.com/G226Medi03/MedicalEntity/ename/etrait"
)

// ErrNotFound marks a missing persisted entity.
var ErrNotFound = errors.New("entity not found")

// Executor is the database surface used by repositories.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Typed value aliases keep the domain model compatible with the codec and
// parser packages while the application still uses draft-shaped values.
type (
	Trait             = etrait.Trait
	IEAction          = eaction.IEAction
	IEActor           = eactor.IEActor
	IEStatus          = estatus.IEStatus
	IMedicalProcedure = eprocedure.IMedicalProcedure
	IESymptom         = esymptom.IESymptom
	IAspect           = esymptom.IAspect
)
