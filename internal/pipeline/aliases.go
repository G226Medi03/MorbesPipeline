package csolve

import (
	domain "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
)

// Public aliases owned by the domain package.
type (
	Case          = domain.Case
	LLMCall       = model.LLMCall
	LLMCallInput  = model.LLMCallInput
	QuestionInput = domain.QuestionInput
	QuestionStore = domain.QuestionStore

	MedicalProcedureOptionGraph      = domain.MedicalProcedureOptionGraph
	MedicalProcedureOptionGraphItem  = domain.MedicalProcedureOptionGraphItem
	MedicalProcedureOptionOutcome    = domain.MedicalProcedureOptionOutcome
	MedicalProcedureOptionIndication = domain.MedicalProcedureOptionIndication
)
