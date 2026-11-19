package read

import (
	source "github.com/G226Medi03/SourceEntity"
)

type ProblemResultFilter struct {
	Model   string
	Correct *bool
	Error   *bool
	Order   string
	Limit   int
	Offset  int
}

type ProblemResultRow struct {
	ID             int64
	ExamQuestionID int64
	LLMResponseID  source.RespId
	Model          string
	Prompt         string
	Output         string
	Correct        bool
	CreatedAt      string
}

type ExamQuestionDataCounts struct {
	LLMResultCount   int
	LLMResponseCount int
	SolveCaseCount   int
}
