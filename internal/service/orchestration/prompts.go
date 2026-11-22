package orchestration

import "github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/prompt"

type PromptVersion = prompt.PromptVersion

const (
	PromptExplain                = prompt.PromptExplain
	PromptQuestionType           = prompt.PromptQuestionType
	PromptCaseExtraction         = prompt.PromptCaseExtraction
	PromptActor                  = prompt.PromptActor
	PromptOtherNames             = prompt.PromptOtherNames
	PromptStatus                 = prompt.PromptStatus
	PromptAction                 = prompt.PromptAction
	PromptActorAction            = prompt.PromptActorAction
	PromptPatientDescription     = prompt.PromptPatientDescription
	PromptMedicalProcedureOption = prompt.PromptMedicalProcedureOption
	PromptRelation               = prompt.PromptRelation
	PromptSituation              = prompt.PromptSituation
)

func PromptStages() []string {
	return prompt.PromptStages()
}

func PromptPlaceholders(stage string) []string {
	return prompt.PromptPlaceholders(stage)
}

func RenderPrompt(template string, values map[string]string) string {
	return prompt.RenderPrompt(template, values)
}
