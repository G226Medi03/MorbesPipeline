package solve

import (
	"time"

	source "github.com/G226Medi03/SourceEntity"
)

const (
	CasePending CaseState = "pending"
	CaseRunning CaseState = "running"
	CaseDone    CaseState = "done"
	CaseFailed  CaseState = "failed"
)

type CaseID string

type CaseState string

type Case struct {
	ID                CaseID          `json:"id"`
	ExamQuestionID    int64           `json:"examQuestionId"`
	Processed         bool            `json:"processed"`
	Problem           string          `json:"problem,omitempty"`
	Answer            string          `json:"answer,omitempty"`
	Explanation       string          `json:"explanation,omitempty"`
	AnswerJSON        string          `json:"answerJson"`
	State             string          `json:"state"`
	Stage             string          `json:"stage"`
	QuestionType      string          `json:"questionType"`
	Error             string          `json:"error"`
	SourceID          source.SourceId `json:"sourceId"`
	SituationSourceID source.SourceId `json:"situationSourceId"`
	ReplayOf          CaseID          `json:"replayOf"`
	ReplayMode        string          `json:"replayMode"`
	ReplayFromNode    string          `json:"replayFromNode"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`

	Steps []Step `json:"-"`
}
