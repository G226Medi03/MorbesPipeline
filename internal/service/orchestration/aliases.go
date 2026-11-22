package orchestration

import (
	solve "github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/model"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/selection"
)

// Public aliases owned by the domain package.
type (
	Case          = solve.Case
	LLMCall       = model.LLMCall
	LLMCallInput  = model.LLMCallInput
	QuestionInput = solve.QuestionInput
	QuestionStore = solve.QuestionStore

	MedicalProcedureOptionGraph      = solve.MedicalProcedureOptionGraph
	MedicalProcedureOptionGraphItem  = solve.MedicalProcedureOptionGraphItem
	MedicalProcedureOptionOutcome    = solve.MedicalProcedureOptionOutcome
	MedicalProcedureOptionIndication = solve.MedicalProcedureOptionIndication
)

func ScoreMedicalProcedureOption(option MedicalProcedureOptionGraphItem) int {
	return selection.ScoreMedicalProcedureOption(option)
}

func SelectMedicalProcedureOption(options []MedicalProcedureOptionGraphItem, problem string) (MedicalProcedureOptionGraphItem, bool) {
	return selection.SelectMedicalProcedureOption(options, problem)
}
