package csolve

import (
	domain "github.com/G226Medi03/MorbesPipeline/internal/service/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/service/solve/model"
	"github.com/G226Medi03/MorbesPipeline/internal/service/selection"
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

func ScoreMedicalProcedureOption(option MedicalProcedureOptionGraphItem) int {
	return selection.ScoreMedicalProcedureOption(option)
}

func SelectMedicalProcedureOption(options []MedicalProcedureOptionGraphItem, problem string) (MedicalProcedureOptionGraphItem, bool) {
	return selection.SelectMedicalProcedureOption(options, problem)
}
