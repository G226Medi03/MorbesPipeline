package medicalprocedureoption

import (
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	mp "github.com/G226Medi03/MorbesPipeline/medicalprocedureoption"
	source "github.com/G226Medi03/SourceEntity"
)

type MedicalProcedureOptionID = mp.MedicalProcedureOptionID
type MedicalProcedureOption = mp.MedicalProcedureOption

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
	return mp.NewMedicalProcedureOption(
		id,
		procedureID,
		category,
		traits,
		contextTraits,
		indicationIDs,
		expectedEffectIDs,
		sources,
		procedure,
		indication,
	)
}
