package model

import (
	"time"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

type LLMCall struct {
	ID                 ids.LLMCallID   `json:"id"`
	CaseID             domain.CaseID   `json:"caseId"`
	Seq                int             `json:"seq"`
	Stage              string          `json:"stage"`
	Role               string          `json:"role"`
	Model              string          `json:"model"`
	PromptVersion      int             `json:"promptVersion"`
	Prompt             string          `json:"prompt"`
	RawOutput          string          `json:"rawOutput"`
	PreprocessedOutput string          `json:"preprocessedOutput"`
	PreprocessChanges  string          `json:"preprocessChanges"`
	ParsedOutput       string          `json:"parsedOutput"`
	SourceID           source.SourceId `json:"sourceId"`
	RespID             source.RespId   `json:"respId"`
	State              string          `json:"state"`
	Error              string          `json:"error"`
	DurationMS         int64           `json:"durationMs"`
	CreatedAt          time.Time       `json:"createdAt"`
}
