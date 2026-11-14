package solve

import (
	"context"

	sqlutil "github.com/G226Medi03/MorbesPipeline/internal/external/sqlcore/sqlite"
)

type Repository struct {
	Case        CaseTable
	Step        StepTable
	Call        CallTable
	Prompt      PromptTable
	AnswerCheck AnswerCheckTable
}

func (r Repository) Init(ctx context.Context, db sqlutil.Executor) error {
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
