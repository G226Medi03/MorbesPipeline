package repository

import (
	"context"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

type SolveRepository struct {
	Case        SolveCaseTable
	Step        SolveStepTable
	Call        SolveCallTable
	Prompt      SolvePromptTable
	AnswerCheck SolveAnswerCheckTable
}

func NewSolveRepository() SolveRepository { return SolveRepository{} }

func (r SolveRepository) Init(ctx context.Context, db sqlutil.Executor) error {
	for _, init := range []func(context.Context, sqlutil.Executor) error{
		r.Case.init,
		r.Step.init,
		r.Call.init,
		r.Prompt.init,
		r.AnswerCheck.init,
	} {
		if err := init(ctx, db); err != nil {
			return err
		}
	}
	return nil
}
