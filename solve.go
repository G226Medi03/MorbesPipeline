package morbes_pipeline

import (
	"context"

	solve "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
	pipeline "github.com/G226Medi03/MorbesPipeline/internal/pipeline"
)

type (
	Service                          = pipeline.Service
	Pipeline                         = pipeline.Pipeline
	Dependencies                     = pipeline.Dependencies
	CaseExtraction                   = pipeline.CaseExtraction
	ResolvedCaseExtraction           = pipeline.ResolvedCaseExtraction
	ActorDraft                       = pipeline.ActorDraft
	StatusDraft                      = pipeline.StatusDraft
	ActionDraft                      = pipeline.ActionDraft
	RelationDraft                    = pipeline.RelationDraft
	OptionDraft                      = pipeline.OptionDraft
	PatientDescriptionDraft          = pipeline.PatientDescriptionDraft
	ProblemChoice                    = pipeline.ProblemChoice
	ReplayMode                       = solve.ReplayMode
	LLMCall                          = model.LLMCall
	LLMCallInput                     = model.LLMCallInput
	PromptVersion                    = model.PromptVersion
	MedicalProcedureOptionGraph      = solve.MedicalProcedureOptionGraph
	MedicalProcedureOptionGraphItem  = solve.MedicalProcedureOptionGraphItem
	MedicalProcedureOptionOutcome    = solve.MedicalProcedureOptionOutcome
	MedicalProcedureOptionIndication = solve.MedicalProcedureOptionIndication
)

const (
	NodeExplain        = pipeline.NodeExplain
	NodeCaseExtraction = pipeline.NodeCaseExtraction
	NodeResolve        = pipeline.NodeResolve
	NodeCause          = pipeline.NodeCause

	ReplayModeLLM  = solve.ReplayModeLLM
	ReplayModeCode = solve.ReplayModeCode

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

var (
	NewService  = pipeline.New_Service
	New_Service = pipeline.New_Service
)

func Run(ctx context.Context, service *Service, id CaseID) error {
	return service.Run(ctx, id)
}

func PipelineStages() []string {
	return pipeline.PipelineStages()
}

func StageSequence(stage string) int {
	return pipeline.StageSequence(stage)
}

func StepSeq(node string) int {
	return pipeline.StepSeq(node)
}

func PromptStages() []string {
	return pipeline.PromptStages()
}

func PromptPlaceholders(stage string) []string {
	return pipeline.PromptPlaceholders(stage)
}

func RenderPrompt(template string, values map[string]string) string {
	return pipeline.RenderPrompt(template, values)
}

func ScoreMedicalProcedureOption(option MedicalProcedureOptionGraphItem) int {
	return pipeline.ScoreMedicalProcedureOption(option)
}

func SelectMedicalProcedureOption(options []MedicalProcedureOptionGraphItem, problem string) (MedicalProcedureOptionGraphItem, bool) {
	return pipeline.SelectMedicalProcedureOption(options, problem)
}
