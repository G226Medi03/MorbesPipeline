package csolve

import (
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
	"github.com/G226Medi03/MorbesPipeline/solve/model"
)

type PromptVersion = model.PromptVersion

const (
	PromptExplain                = domain.PromptExplain
	PromptQuestionType           = domain.PromptQuestionType
	PromptCaseExtraction         = domain.PromptCaseExtraction
	PromptActor                  = domain.PromptActor
	PromptOtherNames             = domain.PromptOtherNames
	PromptStatus                 = domain.PromptStatus
	PromptAction                 = domain.PromptAction
	PromptActorAction            = domain.PromptActorAction
	PromptPatientDescription     = domain.PromptPatientDescription
	PromptMedicalProcedureOption = domain.PromptMedicalProcedureOption
	PromptRelation               = domain.PromptRelation
	PromptSituation              = domain.PromptSituation
)

func PromptStages() []string {
	return domain.PromptStages()
}

func PromptPlaceholders(stage string) []string {
	return domain.PromptPlaceholders(stage)
}

func RenderPrompt(template string, values map[string]string) string {
	return domain.RenderPrompt(template, values)
}
