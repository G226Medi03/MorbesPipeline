// Package postprocess stores the audit trail of postprocess operations.
package repository

import "time"

// Operation describes one postprocess change.
type PostprocessOperation string

const (
	PostprocessOperationObserved    PostprocessOperation = "observed"
	PostprocessOperationCreated     PostprocessOperation = "created"
	PostprocessOperationUpdated     PostprocessOperation = "updated"
	PostprocessOperationRemoved     PostprocessOperation = "removed"
	PostprocessOperationLinkAdded   PostprocessOperation = "link_added"
	PostprocessOperationLinkRemoved PostprocessOperation = "link_removed"
)

const (
	PostprocessorTransplantRejection  = "transplant_rejection"
	PostprocessorIncreasedIndexAction = "increased_index_action"
	PostprocessorCauseStatusResult    = "cause_status_result"
)

// EntityType identifies the kind of entity referenced by an operation.
type PostprocessEntityType string

const (
	PostprocessEntityStatus    PostprocessEntityType = "status"
	PostprocessEntityAction    PostprocessEntityType = "action"
	PostprocessEntityActor     PostprocessEntityType = "actor"
	PostprocessEntityProcedure PostprocessEntityType = "procedure"
	PostprocessEntityOption    PostprocessEntityType = "option"
)

// Record is one postprocess operation.
type PostprocessRecord struct {
	PostprocessID     int64
	PostprocessorName string
	Operation         PostprocessOperation
	EntityType        PostprocessEntityType
	EntityID          int64
	TargetEntityType  PostprocessEntityType
	TargetEntityID    int64
	FinishedAt        time.Time
}
