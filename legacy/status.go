package legacy

import (
	"encoding/json"

	source "github.com/G226Medi03/SourceEntity"
)

type Status struct {
	id      StatusID
	Variant string `json:"variant"`
	Value   any    `json:"value"`

	// Actions is hydrated one level. Nested Action.ResultStatuses is left empty.
	Actions []struct {
		Action   Action          `json:"action"`
		SourceID source.SourceId `json:"sourceId"`
	} `json:"actions,omitempty"`

	Sources []source.SourceRef `json:"sources"`
}

func (value Status) Id() StatusID { return value.id }

func NewStatus(variant string, data any, sources []source.SourceRef) Status {
	return RestoreStatus(0, variant, data, sources)
}

func RestoreStatus(id StatusID, variant string, data any, sources []source.SourceRef) Status {
	return Status{id: id, Variant: variant, Value: data, Sources: sources}
}

func (value Status) MarshalJSON() ([]byte, error) {
	type alias Status
	return json.Marshal(struct {
		ID StatusID `json:"id"`
		*alias
	}{value.id, (*alias)(&value)})
}

func (value *Status) UnmarshalJSON(data []byte) error {
	type alias Status
	aux := struct {
		ID StatusID `json:"id"`
		*alias
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	value.id = aux.ID
	return nil
}
