package ccause

import source_entity "github.com/G226Medi03/SourceEntity"

type Candidate struct {
	CauseType     string                   `json:"causeType,omitempty"`
	CauseKey      string                   `json:"causeKey,omitempty"`
	CauseKind     string                   `json:"causeKind"`
	CauseLabel    string                   `json:"causeLabel"`
	ViaActionKey  string                   `json:"viaActionKey,omitempty"`
	ViaActionKind string                   `json:"viaActionKind"`
	Depth         int                      `json:"depth"`
	SourceIDs     []source_entity.SourceId `json:"sourceIds"`
	Path          []TraceNode              `json:"path"`
}
