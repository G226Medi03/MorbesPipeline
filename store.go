package morbes_pipeline

import (
	persistence "github.com/G226Medi03/MorbesPipeline/internal/persistence"
	storesolve "github.com/G226Medi03/MorbesPipeline/internal/persistence/solve"
)

type (
	Store               = persistence.Store
	ProblemResultFilter = persistence.ProblemResultFilter
	ProblemResultRow    = persistence.ProblemResultRow

	ActorRecord                = persistence.ActorRecord
	ActionRecord               = persistence.ActionRecord
	StatusRecord               = persistence.StatusRecord
	ProcedureRecord            = persistence.ProcedureRecord
	EntityNeighbors            = persistence.EntityNeighbors
	StatusNameRef              = persistence.StatusNameRef
	UnresolvedTransplantTarget = persistence.UnresolvedTransplantTarget
	ExamQuestionDataCounts     = persistence.ExamQuestionDataCounts
	ExamQuestionCaseStats      = storesolve.ExamQuestionCaseStats
)

var ErrUnknownEntityType = persistence.ErrUnknownEntityType

var (
	New        = persistence.New
	NewWithDBs = persistence.NewWithDBs
)
