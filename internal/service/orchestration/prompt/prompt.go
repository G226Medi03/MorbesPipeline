package prompt

import (
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/model"
)

type PromptVersion = model.PromptVersion

type Set struct {
	Present  map[string]bool
	Versions map[string]int
}

const (
	PromptExplain                = solve.PromptExplain
	PromptQuestionType           = solve.PromptQuestionType
	PromptCaseExtraction         = solve.PromptCaseExtraction
	PromptActor                  = solve.PromptActor
	PromptOtherNames             = solve.PromptOtherNames
	PromptStatus                 = solve.PromptStatus
	PromptAction                 = solve.PromptAction
	PromptActorAction            = solve.PromptActorAction
	PromptPatientDescription     = solve.PromptPatientDescription
	PromptMedicalProcedureOption = solve.PromptMedicalProcedureOption
	PromptRelation               = solve.PromptRelation
	PromptSituation              = solve.PromptSituation
)

func PromptStages() []string {
	return solve.PromptStages()
}

func PromptPlaceholders(stage string) []string {
	return solve.PromptPlaceholders(stage)
}

func RenderPrompt(template string, values map[string]string) string {
	return solve.RenderPrompt(template, values)
}
