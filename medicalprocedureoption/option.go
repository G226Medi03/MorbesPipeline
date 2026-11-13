package medicalprocedureoption

import (
	"encoding/json"
	"fmt"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	source "github.com/G226Medi03/SourceEntity"
)

type MedicalProcedureOptionID = entity.MedicalProcedureOptionID

// MedicalProcedureOption is the persistence shape for a procedure candidate.
// Procedure and Indication are hydrated at runtime and are never persisted
// directly.
type MedicalProcedureOption struct {
	id MedicalProcedureOptionID

	ProcedureID             entity.ProcedureID
	Category                string
	Traits                  []entity.Trait
	ContextTraits           []entity.Trait
	IndicationStatusIDs     []entity.StatusID
	ExpectedEffectStatusIDs []entity.StatusID
	Sources                 []source.SourceRef

	Procedure  entity.IMedicalProcedure `json:"-"`
	Indication []entity.IEStatus        `json:"-"`
}

func (value MedicalProcedureOption) Id() MedicalProcedureOptionID { return value.id }

func NewMedicalProcedureOption(
	id MedicalProcedureOptionID,
	procedureID entity.ProcedureID,
	category string,
	traits []entity.Trait,
	contextTraits []entity.Trait,
	indicationIDs []entity.StatusID,
	expectedEffectIDs []entity.StatusID,
	sources []source.SourceRef,
	procedure entity.IMedicalProcedure,
	indication []entity.IEStatus,
) MedicalProcedureOption {
	return MedicalProcedureOption{
		id:                      id,
		ProcedureID:             procedureID,
		Category:                category,
		Traits:                  traits,
		ContextTraits:           contextTraits,
		IndicationStatusIDs:     indicationIDs,
		ExpectedEffectStatusIDs: expectedEffectIDs,
		Sources:                 sources,
		Procedure:               procedure,
		Indication:              indication,
	}
}

func (value MedicalProcedureOption) Validate() error {
	if value.ProcedureID.ToI64() <= 0 {
		return fmt.Errorf("medical procedure option procedure_id is required")
	}
	if len(value.Sources) == 0 {
		return entity.ErrSourceRequired
	}
	return nil
}

func (value MedicalProcedureOption) MarshalJSON() ([]byte, error) {
	type alias MedicalProcedureOption
	return json.Marshal(struct {
		ID MedicalProcedureOptionID `json:"id"`
		*alias
	}{value.id, (*alias)(&value)})
}

func (value *MedicalProcedureOption) UnmarshalJSON(data []byte) error {
	type alias MedicalProcedureOption
	aux := struct {
		ID MedicalProcedureOptionID `json:"id"`
		*alias
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	value.id = aux.ID
	return nil
}
