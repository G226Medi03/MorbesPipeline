package medicalprocedureoption

import (
	"encoding/json"
	"errors"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	source "github.com/G226Medi03/SourceEntity"
)

var ErrSourceRequired = errors.New("source가 필요합니다")

type MedicalProcedureOptionID = ids.MedicalProcedureOptionID

// MedicalProcedureOption is the persistence shape for a procedure candidate.
// Procedure and Indication are hydrated at runtime and are never persisted
// directly.
type MedicalProcedureOption struct {
	id MedicalProcedureOptionID

	ProcedureID             ids.ProcedureID
	Category                string
	Traits                  []trait.Trait
	ContextTraits           []trait.Trait
	IndicationStatusIDs     []ids.StatusID
	ExpectedEffectStatusIDs []ids.StatusID
	Sources                 []source.SourceRef

	Procedure  repo.IMedicalProcedure `json:"-"`
	Indication []repo.IEStatus        `json:"-"`
}

func (value MedicalProcedureOption) Id() MedicalProcedureOptionID { return value.id }

func (value MedicalProcedureOption) WithID(id MedicalProcedureOptionID) MedicalProcedureOption {
	value.id = id
	return value
}

func NewMedicalProcedureOption(
	id MedicalProcedureOptionID,
	procedureID ids.ProcedureID,
	category string,
	traits []trait.Trait,
	contextTraits []trait.Trait,
	indicationIDs []ids.StatusID,
	expectedEffectIDs []ids.StatusID,
	sources []source.SourceRef,
	procedure repo.IMedicalProcedure,
	indication []repo.IEStatus,
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
		return ErrSourceRequired
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
