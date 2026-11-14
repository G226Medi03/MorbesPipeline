package morbes_pipeline

import (
	solve "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
	llmclient "github.com/G226Medi03/MorbesPipeline/internal/external/llmclient/call"
	pipeline "github.com/G226Medi03/MorbesPipeline/internal/pipeline"
)

type (
	Case                             = solve.Case
	Step                             = solve.Step
	StepState                        = solve.StepState
	Service                          = pipeline.Service
	Pipeline                         = pipeline.Pipeline
	Dependencies                     = pipeline.Dependencies
	LLMClient                        = llmclient.Client
	CaseExtraction                   = pipeline.CaseExtraction
	ResolvedCaseExtraction           = pipeline.ResolvedCaseExtraction
	ActorDraft                       = pipeline.ActorDraft
	StatusDraft                      = pipeline.StatusDraft
	ActionDraft                      = pipeline.ActionDraft
	RelationDraft                    = pipeline.RelationDraft
	OptionDraft                      = pipeline.OptionDraft
	PatientDescriptionDraft          = pipeline.PatientDescriptionDraft
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
	NewService   = pipeline.New_Service
	New_Service  = pipeline.New_Service
	NewLLMClient = llmclient.NewFunc
)

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
