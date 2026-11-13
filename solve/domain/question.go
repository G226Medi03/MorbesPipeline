package domain

import (
	"context"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

// QuestionInput is the exam question payload handed to the solve pipeline.
type QuestionInput struct {
	Problem     string
	Answer      string
	Explanation string
	SourceID    source.SourceId
}

// QuestionStore is the exam question port used by the solve pipeline.
type QuestionStore interface {
	GetQuestion(ctx context.Context, id entity.ExamQuestionID) (QuestionInput, bool, error)
	CreateQuestion(ctx context.Context, problem, answer, explanation string) (entity.ExamQuestionID, error)
}
