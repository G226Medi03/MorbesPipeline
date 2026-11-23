package tracing

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source_entity "github.com/G226Medi03/SourceEntity"
)

type Candidate struct {
	CauseType     string                   `json:"causeType,omitempty"`
	CauseKind     string                   `json:"causeKind"`
	CauseLabel    string                   `json:"causeLabel"`
	ViaActionKind string                   `json:"viaActionKind"`
	Depth         int                      `json:"depth"`
	SourceIDs     []source_entity.SourceId `json:"sourceIds"`
	Path          []TraceNode              `json:"path"`

	causeStatus repo.IEStatus
	viaAction   repo.IEAction
	causeActor  repo.IEActor
}
