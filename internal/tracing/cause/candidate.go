package ccause

import (
	source_entity "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

type Candidate struct {
	CauseType      string                   `json:"causeType,omitempty"`
	CauseStatusID  ids.StatusID             `json:"causeStatusId"`
	CauseStatusIDs []ids.StatusID           `json:"causeStatusIds,omitempty"`
	CauseActorID   ids.ActorID              `json:"causeActorID,omitempty"`
	CauseActorIDs  []ids.ActorID            `json:"causeActorIDs,omitempty"`
	CauseDeviceID  ids.DeviceID             `json:"causeDeviceId,omitempty"`
	CauseDeviceIDs []ids.DeviceID           `json:"causeDeviceIds,omitempty"`
	CauseKind      string                   `json:"causeKind"`
	CauseKey       string                   `json:"causeKey,omitempty"`
	CauseLabel     string                   `json:"causeLabel"`
	ViaActionID    ids.ActionID             `json:"viaActionID"`
	ViaActionIDs   []ids.ActionID           `json:"viaActionIDs,omitempty"`
	ViaActionKind  string                   `json:"viaActionKind"`
	ViaActionKey   string                   `json:"viaActionKey,omitempty"`
	Depth          int                      `json:"depth"`
	SourceIDs      []source_entity.SourceId `json:"sourceIds"`
	Path           []TraceNode              `json:"path"`
}
