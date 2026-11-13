// Package domain contains the solve pipeline's aggregate model and policies.
package domain

type Stage string

const (
	StageExplain        Stage = "explain"
	StageCaseExtraction Stage = "case_extraction"
	StageResolve        Stage = "resolve"
	StageCause          Stage = "cause_trace"
)

var stage_order = []Stage{
	StageExplain,
	StageCaseExtraction,
	StageResolve,
	StageCause,
}

var stage_sequence = map[Stage]int{
	StageExplain:        10,
	StageCaseExtraction: 20,
	StageResolve:        30,
	StageCause:          60,
}

var stage_downstream = map[Stage][]Stage{
	StageExplain:        {StageCaseExtraction, StageResolve, StageCause},
	StageCaseExtraction: {StageResolve, StageCause},
	StageResolve:        {StageCause},
}

func Stages() []Stage {
	result := make([]Stage, len(stage_order))
	copy(result, stage_order)
	return result
}

func Sequence(stage Stage) int {
	return stage_sequence[stage]
}

func Downstream(stage Stage) []Stage {
	values := stage_downstream[stage]
	result := make([]Stage, len(values))
	copy(result, values)
	return result
}

func IsStage(value string) bool {
	_, ok := stage_sequence[Stage(value)]
	return ok
}
