package csolve

import "github.com/G226Medi03/MorbesPipeline/solve/domain"

// ScoreMedicalProcedureOption ranks a treatment candidate against the case's
// observed findings and traced causes. Larger is better.
func ScoreMedicalProcedureOption(option MedicalProcedureOptionGraphItem) int {
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

// SelectMedicalProcedureOption returns exactly one best candidate. It never
// consults the ground-truth answer; the answer is only used afterwards by the
// evaluator.
func SelectMedicalProcedureOption(options []MedicalProcedureOptionGraphItem, problem string) (MedicalProcedureOptionGraphItem, bool) {
	desired_category := domain.InferQuestionPolicy(problem).DesiredCategory
	bestIndex := -1
	bestScore := 0
	for index, option := range options {
		score := ScoreMedicalProcedureOption(option)
		if desired_category != "" && option.Category == desired_category {
			score += 20
		}
		if score <= 0 {
			continue
		}
		if bestIndex < 0 || score > bestScore ||
			(score == bestScore && optionTieLess(option, options[bestIndex])) {
			bestIndex = index
			bestScore = score
		}
	}
	if bestIndex < 0 {
		return MedicalProcedureOptionGraphItem{}, false
	}
	return options[bestIndex], true
}

func optionTieLess(left, right MedicalProcedureOptionGraphItem) bool {
	if left.TreatOptionPriority > 0 && right.TreatOptionPriority > 0 && left.TreatOptionPriority != right.TreatOptionPriority {
		return left.TreatOptionPriority < right.TreatOptionPriority
	}
	if (left.TreatOptionPriority > 0) != (right.TreatOptionPriority > 0) {
		return left.TreatOptionPriority > 0
	}
	return left.ID < right.ID
}
