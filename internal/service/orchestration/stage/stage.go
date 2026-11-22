package stage

import "github.com/G226Medi03/MorbesPipeline/internal/entity/solve"

const (
	NodeExplain        = string(solve.StageExplain)
	NodeCaseExtraction = string(solve.StageCaseExtraction)
	NodeResolve        = string(solve.StageResolve)
	NodeCause          = string(solve.StageCause)
)

type Step = solve.Step

func StepSeq(node string) int { return solve.Sequence(solve.Stage(node)) }

func PipelineStages() []string {
	stages := solve.Stages()
	result := make([]string, 0, len(stages))
	for _, stage := range stages {
		result = append(result, string(stage))
	}
	return result
}

func StageSequence(stage string) int { return solve.Sequence(solve.Stage(stage)) }
