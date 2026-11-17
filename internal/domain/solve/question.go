package domain

import (
	"context"

	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
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
	GetQuestion(ctx context.Context, id ids.ExamQuestionID) (QuestionInput, bool, error)
}
