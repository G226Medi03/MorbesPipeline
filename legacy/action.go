package legacy

import (
	"encoding/json"

	source "github.com/G226Medi03/SourceEntity"
)

// CauseStatusAction preserves the legacy status_id-only action payload until
// old solve rows are migrated to typed MedicalEntity cause-status actions.
type CauseStatusAction struct {
	StatusID int64    `json:"status_id"`
	Status   IEStatus `json:"-"`
}

func (CauseStatusAction) Is_EAction() {}

func (CauseStatusAction) Kind() string { return "cause_status" }

type Action struct {
	id      ActionID
	Variant string `json:"variant"`
	Value   any    `json:"value"`

	// ResultStatuses is hydrated one level. Nested Status.Actions is left empty.
	ResultStatuses []struct {
		Status   Status          `json:"status"`
		SourceID source.SourceId `json:"sourceId"`
	} `json:"resultStatuses,omitempty"`

	Sources []source.SourceRef `json:"sources"`
}

func (value Action) Id() ActionID { return value.id }

func NewAction(variant string, data any, sources []source.SourceRef) Action {
	return RestoreAction(0, variant, data, sources)
}

func RestoreAction(id ActionID, variant string, data any, sources []source.SourceRef) Action {
	return Action{id: id, Variant: variant, Value: data, Sources: sources}
}

func (value Action) MarshalJSON() ([]byte, error) {
	type alias Action
	return json.Marshal(struct {
		ID ActionID `json:"id"`
		*alias
	}{value.id, (*alias)(&value)})
}

func (value *Action) UnmarshalJSON(data []byte) error {
	type alias Action
	aux := struct {
		ID ActionID `json:"id"`
		*alias
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	value.id = aux.ID
	return nil
}
