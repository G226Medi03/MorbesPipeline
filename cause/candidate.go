package ccause

import (
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	source_entity "github.com/G226Medi03/SourceEntity"
)

type Candidate struct {
	CauseType      string                   `json:"causeType,omitempty"`
	CauseStatusID  entity.StatusID          `json:"causeStatusId"`
	CauseStatusIDs []entity.StatusID        `json:"causeStatusIds,omitempty"`
	CauseActorID   entity.ActorID           `json:"causeActorID,omitempty"`
	CauseActorIDs  []entity.ActorID         `json:"causeActorIDs,omitempty"`
	CauseDeviceID  entity.DeviceID          `json:"causeDeviceId,omitempty"`
	CauseDeviceIDs []entity.DeviceID        `json:"causeDeviceIds,omitempty"`
	CauseKind      string                   `json:"causeKind"`
	CauseKey       string                   `json:"causeKey,omitempty"`
	CauseLabel     string                   `json:"causeLabel"`
	ViaActionID    entity.ActionID          `json:"viaActionID"`
	ViaActionIDs   []entity.ActionID        `json:"viaActionIDs,omitempty"`
	ViaActionKind  string                   `json:"viaActionKind"`
	ViaActionKey   string                   `json:"viaActionKey,omitempty"`
	Depth          int                      `json:"depth"`
	SourceIDs      []source_entity.SourceId `json:"sourceIds"`
	Path           []TraceNode              `json:"path"`
}
