package model

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/SourceEntity"
)

type LLMCallInput struct {
	CaseID             domain.CaseID
	Seq                int
	Stage              string
	Role               string
	Model              string
	PromptVersion      int
	Prompt             string
	RawOutput          string
	PreprocessedOutput string
	PreprocessChanges  string
	ParsedOutput       string
	SourceID           source.SourceId
	RespID             source.RespId
	State              string
	Error              string
	DurationMS         int64
}
