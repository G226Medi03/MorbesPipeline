package model

import (
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	source "github.com/G226Medi03/SourceEntity"
)

type LLMCallInput struct {
	CaseID        solve.CaseID
	Seq           int
	Stage         string
	Role          string
	Model         string
	PromptVersion int
	Prompt        string
	RawOutput     string
	ParsedOutput  string
	SourceID      source.SourceId
	RespID        source.RespId
	State         string
	Error         string
	DurationMS    int64
}
