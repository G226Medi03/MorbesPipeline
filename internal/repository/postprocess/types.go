// Package postprocess stores the audit trail of postprocess operations.
package postprocess

import "time"

// Operation describes one postprocess change.
type Operation string

const (
	OperationObserved    Operation = "observed"
	OperationCreated     Operation = "created"
	OperationUpdated     Operation = "updated"
	OperationRemoved     Operation = "removed"
	OperationLinkAdded   Operation = "link_added"
	OperationLinkRemoved Operation = "link_removed"
)

const (
	PostprocessorTransplantRejection  = "transplant_rejection"
	PostprocessorIncreasedIndexAction = "increased_index_action"
	PostprocessorCauseStatusResult    = "cause_status_result"
)

// EntityType identifies the kind of entity referenced by an operation.
type EntityType string

const (
	EntityStatus    EntityType = "status"
	EntityAction    EntityType = "action"
	EntityActor     EntityType = "actor"
	EntityProcedure EntityType = "procedure"
	EntityOption    EntityType = "option"
)

// Record is one postprocess operation.
type Record struct {
	PostprocessID     int64
	PostprocessorName string
	Operation         Operation
	EntityType        EntityType
	EntityJSON        string
	TargetEntityType  EntityType
	TargetEntityJSON  string
	FinishedAt        time.Time
}
