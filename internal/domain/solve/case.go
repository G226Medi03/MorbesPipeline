package domain

import (
	"time"

	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	"github.com/G226Medi03/SourceEntity"
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
	ID                CaseID             `json:"id"`
	ExamQuestionID    ids.ExamQuestionID `json:"examQuestionId"`
	Processed         bool               `json:"processed"`
	Problem           string             `json:"problem,omitempty"`
	Answer            string             `json:"answer,omitempty"`
	Explanation       string             `json:"explanation,omitempty"`
	SituationJSON     string             `json:"situationJson"`
	AnswerJSON        string             `json:"answerJson"`
	State             string             `json:"state"`
	Stage             string             `json:"stage"`
	QuestionType      string             `json:"questionType"`
	Error             string             `json:"error"`
	SourceID          source.SourceId    `json:"sourceId"`
	SituationSourceID source.SourceId    `json:"situationSourceId"`
	ReplayOf          CaseID             `json:"replayOf"`
	ReplayMode        string             `json:"replayMode"`
	ReplayFromNode    string             `json:"replayFromNode"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`

	Steps []Step `json:"-"`
}
