package read

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
)

type ActorRecord struct {
	Key        string
	Value      repo.IEActor
	OtherNames []string
	Sources    []source.Source
	Actions    []ActorActionLink
}

type ActorActionLink struct {
	Action   ActionRecord
	SourceID source.SourceId
}

type ActionRecord struct {
	Key            string
	Value          repo.IEAction
	Sources        []source.Source
	ResultStatuses []ActionResultStatusLink
}

type ActionResultStatusLink struct {
	Status   StatusRecord
	SourceID source.SourceId
}

type StatusRecord struct {
	Key     string
	Value   repo.IEStatus
	Sources []source.Source
	Actions []StatusActionLink
}

type StatusActionLink struct {
	Action   ActionRecord
	SourceID source.SourceId
}

type ProcedureRecord struct {
	Key        string
	Value      repo.IMedicalProcedure
	OtherNames []string
	Traits     []repo.Trait
	Sources    []source.Source
}

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
