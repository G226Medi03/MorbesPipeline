package csolve

import "github.com/G226Medi03/MorbesPipeline/internal/pipeline/selection"

func ScoreMedicalProcedureOption(option MedicalProcedureOptionGraphItem) int {
	return selection.ScoreMedicalProcedureOption(option)
}

func SelectMedicalProcedureOption(options []MedicalProcedureOptionGraphItem, problem string) (MedicalProcedureOptionGraphItem, bool) {
	return selection.SelectMedicalProcedureOption(options, problem)
}
