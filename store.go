package morbes_pipeline

import (
	persistence "github.com/G226Medi03/MorbesPipeline/internal/persistence"
	storesolve "github.com/G226Medi03/MorbesPipeline/internal/persistence/solve"
)

type (
	Store               = persistence.Store
	ProblemResultFilter = persistence.ProblemResultFilter
	ProblemResultRow    = persistence.ProblemResultRow

	NamedRecord                = persistence.NamedRecord
	ActorRecord                = persistence.ActorRecord
	ActorActionLink            = persistence.ActorActionLink
	ActionRecord               = persistence.ActionRecord
	ActionResultStatusLink     = persistence.ActionResultStatusLink
	StatusRecord               = persistence.StatusRecord
	StatusActionLink           = persistence.StatusActionLink
	ProcedureRecord            = persistence.ProcedureRecord
	EntityNeighbors            = persistence.EntityNeighbors
	StatusNameRef              = persistence.StatusNameRef
	UnresolvedTransplantTarget = persistence.UnresolvedTransplantTarget
	ExamQuestionDataCounts     = persistence.ExamQuestionDataCounts
	ExamQuestionCaseStats      = storesolve.ExamQuestionCaseStats
)

var ErrUnknownEntityType = persistence.ErrUnknownEntityType

var (
	New             = persistence.New
	NewWithDBs      = persistence.NewWithDBs
	NewStore        = persistence.New
	NewStoreWithDBs = persistence.NewWithDBs
)
