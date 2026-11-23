package selection

import solve "github.com/G226Medi03/MorbesPipeline/internal/entity/solve"

func ScoreMedicalProcedureOption(option solve.MedicalProcedureOptionGraphItem) int {
	score := 0
	for _, indication := range option.Indications {
		if indication.Observed {
			score += 100
		}
		if indication.Cause {
			score += 30
		}
	}
	for _, effect := range option.ExpectedEffects {
		if effect.Observed {
			score += 60
		}
		if effect.Cause {
			score += 15
		}
	}
	if option.Category == "treatment" {
		score += 5
	}
	return score
}

func SelectMedicalProcedureOption(options []solve.MedicalProcedureOptionGraphItem, problem string) (solve.MedicalProcedureOptionGraphItem, bool) {
	desiredCategory := solve.InferQuestionPolicy(problem).DesiredCategory
	bestIndex := -1
	bestScore := 0
	for index, option := range options {
		score := ScoreMedicalProcedureOption(option)
		if desiredCategory != "" && option.Category == desiredCategory {
			score += 20
		}
		if score <= 0 {
			continue
		}
		if bestIndex < 0 || score > bestScore {
			bestIndex = index
			bestScore = score
		}
	}
	if bestIndex < 0 {
		return solve.MedicalProcedureOptionGraphItem{}, false
	}
	return options[bestIndex], true
}
