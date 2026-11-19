package csolve

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/pipeline/orchestration"
)

const (
	NodeExplain        = orchestration.NodeExplain
	NodeCaseExtraction = orchestration.NodeCaseExtraction
	NodeResolve        = orchestration.NodeResolve
	NodeCause          = orchestration.NodeCause
)

type Step = domain.Step

func StepSeq(node string) int { return orchestration.StepSeq(node) }

func PipelineStages() []string { return orchestration.PipelineStages() }

func StageSequence(stage string) int { return orchestration.StageSequence(stage) }
