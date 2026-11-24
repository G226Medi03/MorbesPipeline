package solve

import (
	"context"

	source "github.com/G226Medi03/SourceEntity"
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
	GetQuestion(ctx context.Context, id int64) (QuestionInput, bool, error)
}
