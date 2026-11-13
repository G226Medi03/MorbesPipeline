package csolve

import (
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
)

const (
	NodeExplain        = string(domain.StageExplain)
	NodeCaseExtraction = string(domain.StageCaseExtraction)
	NodeResolve        = string(domain.StageResolve)
	NodeCause          = string(domain.StageCause)
)

type Step = domain.Step

func StepSeq(node string) int {
	return domain.Sequence(domain.Stage(node))
}

// PipelineStages returns the canonical pipeline order used by the UI.
func PipelineStages() []string {
	stages := domain.Stages()
	result := make([]string, 0, len(stages))
	for _, stage := range stages {
		result = append(result, string(stage))
	}
	return result
}

func StageSequence(stage string) int {
	return domain.Sequence(domain.Stage(stage))
}
