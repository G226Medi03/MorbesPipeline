package orchestration

import "github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/stage"

const (
	NodeExplain        = stage.NodeExplain
	NodeCaseExtraction = stage.NodeCaseExtraction
	NodeResolve        = stage.NodeResolve
	NodeCause          = stage.NodeCause
)

type Step = stage.Step

func StepSeq(node string) int { return stage.StepSeq(node) }

func PipelineStages() []string { return stage.PipelineStages() }

func StageSequence(stageValue string) int { return stage.StageSequence(stageValue) }
