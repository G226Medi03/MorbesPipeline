package morbes_pipeline

import (
	solve "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
	pipeline "github.com/G226Medi03/MorbesPipeline/internal/pipeline"
)

type (
	Service      = pipeline.Service
	Pipeline     = pipeline.Pipeline
	Dependencies = pipeline.Dependencies

	Case       = solve.Case
	Step       = solve.Step
	LLMCall    = model.LLMCall
	ReplayMode = solve.ReplayMode

	ResolvedCaseExtraction = pipeline.ResolvedCaseExtraction

	MedicalProcedureOptionGraph      = solve.MedicalProcedureOptionGraph
	MedicalProcedureOptionGraphItem  = solve.MedicalProcedureOptionGraphItem
	MedicalProcedureOptionOutcome    = solve.MedicalProcedureOptionOutcome
	MedicalProcedureOptionIndication = solve.MedicalProcedureOptionIndication
)

const (
	NodeExplain = pipeline.NodeExplain
	NodeCause   = pipeline.NodeCause

	ReplayModeLLM  = solve.ReplayModeLLM
	ReplayModeCode = solve.ReplayModeCode
)

var NewService = pipeline.New_Service

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

func SelectMedicalProcedureOption(options []MedicalProcedureOptionGraphItem, problem string) (MedicalProcedureOptionGraphItem, bool) {
	return pipeline.SelectMedicalProcedureOption(options, problem)
}
