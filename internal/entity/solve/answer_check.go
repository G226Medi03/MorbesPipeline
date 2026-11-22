package solve

import (
	"time"

	source "github.com/G226Medi03/SourceEntity"
)

// AnswerCheck is the latest answer-check decision for one solve case.
// Prompt and raw response remain in llm_responses and are linked by RespID.
type AnswerCheck struct {
	CaseID          CaseID        `json:"caseId"`
	RespID          source.RespId `json:"respId"`
	QuestionType    string        `json:"questionType,omitempty"`
	ContainsAnswer  bool          `json:"containsAnswer"`
	AnswerPresent   bool          `json:"answerPresent"`
	AnswerSatisfied bool          `json:"answerSatisfied"`
	EffectMatched   bool          `json:"effectMatched,omitempty"`
	MatchedType     string        `json:"matchedType"`
	MatchedKey      string        `json:"matchedKey"`
	AnalysisJSON    string        `json:"analysisJson,omitempty"`
	Error           string        `json:"error,omitempty"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}
