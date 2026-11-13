package legacy

import (
	"encoding/json"

	source "github.com/G226Medi03/SourceEntity"
)

type Procedure struct {
	id      ProcedureID
	Variant string             `json:"variant"`
	Value   any                `json:"value"`
	Sources []source.SourceRef `json:"sources"`
}

func (value Procedure) Id() ProcedureID { return value.id }

func NewProcedure(variant string, data any, sources []source.SourceRef) Procedure {
	return RestoreProcedure(0, variant, data, sources)
}

func RestoreProcedure(id ProcedureID, variant string, data any, sources []source.SourceRef) Procedure {
	return Procedure{id: id, Variant: variant, Value: data, Sources: sources}
}

func (value Procedure) MarshalJSON() ([]byte, error) {
	type alias Procedure
	return json.Marshal(struct {
		ID ProcedureID `json:"id"`
		*alias
	}{value.id, (*alias)(&value)})
}

func (value *Procedure) UnmarshalJSON(data []byte) error {
	type alias Procedure
	aux := struct {
		ID ProcedureID `json:"id"`
		*alias
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	value.id = aux.ID
	return nil
}
