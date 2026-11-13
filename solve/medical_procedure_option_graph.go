package csolve

import "github.com/G226Medi03/MorbesPipeline/solve/domain"

// The candidate graph models are owned by the domain package so that the app
// layer can use them without importing the pipeline implementation.
type (
	MedicalProcedureOptionGraph      = domain.MedicalProcedureOptionGraph
	MedicalProcedureOptionGraphItem  = domain.MedicalProcedureOptionGraphItem
	MedicalProcedureOptionOutcome    = domain.MedicalProcedureOptionOutcome
	MedicalProcedureOptionIndication = domain.MedicalProcedureOptionIndication
)
