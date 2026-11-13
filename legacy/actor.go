package legacy

import (
	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	"github.com/G226Medi03/SourceEntity"
)

type ActorActionLink struct {
	Action   Action          `json:"action"`
	SourceID source.SourceId `json:"sourceId"`
}

type ActorMeta struct {
	ID         ActorID            `json:"id"`
	Kind       string             `json:"kind"`
	Name       string             `json:"name"`
	OtherNames []string           `json:"otherNames,omitempty"`
	Actions    []ActorActionLink  `json:"actions,omitempty"`
	Sources    []source.SourceRef `json:"sources"`
}

func (value ActorMeta) Validate() error {
	return ValidateSources(value.Sources)
}

type Device struct {
	ActorMeta
	Value eactor.Device `json:"value"`
}

type UnknownActor struct {
	ActorMeta
	Value eactor.UnknownActor `json:"value"`
}

// Deprecated aliases for old code paths. They all resolve to UnknownActor.
type (
	Organ                = UnknownActor
	Tissue               = UnknownActor
	Mat                  = UnknownActor
	TransplantableDevice = UnknownActor
	OtherActor           = UnknownActor
)

type Transplantable struct {
	ActorMeta
	Value eactor.IEActor `json:"value"`
}

type Cleanable struct {
	ActorMeta
	Value eactor.IEActor `json:"value"`
}

type ActorRecord struct {
	ActorMeta
	Value eactor.IEActor `json:"value"`
}
