package entity

import (
	"context"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/trait"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

func (r Repository) SaveMedicalProcedureOptionByValues(
	ctx context.Context,
	db sqlutil.Executor,
	entities *repo.Store,
	procedure repo.IMedicalProcedure,
	category string,
	traits []trait.Trait,
	_ []trait.Trait,
	indications []repo.IEStatus,
	effects []repo.IEStatus,
	sources []source.Source,
) (repo.MedicalProcedureOption, error) {
	return r.table(ctx, db, entities).SaveMedicalProcedureOptionByValues(
		procedure, category, traits, nil, indications, effects, sources,
	)
}

func (r Repository) LinkActionResultStatusValues(
	ctx context.Context,
	db sqlutil.Executor,
	action repo.IEAction,
	status repo.IEStatus,
	sourceID source.SourceId,
) error {
	return r.table(ctx, db, nil).LinkActionResultStatusValues(action, status, sourceID)
}

func (r Repository) LinkStatusAction(ctx context.Context, db sqlutil.Executor, statusID, actionID int64, sourceID source.SourceId) error {
	return r.table(ctx, db, nil).LinkStatusAction(statusID, actionID, sourceID)
}

func (r Repository) LinkActionResultStatus(ctx context.Context, db sqlutil.Executor, actionID, statusID int64, sourceID source.SourceId) error {
	return r.table(ctx, db, nil).LinkActionResultStatus(actionID, statusID, sourceID)
}

func (r Repository) DeleteStatusAction(ctx context.Context, db sqlutil.Executor, statusID, actionID int64) error {
	return r.table(ctx, db, nil).DeleteStatusAction(statusID, actionID)
}

func (r Repository) DeleteStatusActionValues(
	ctx context.Context,
	db sqlutil.Executor,
	status repo.IEStatus,
	action repo.IEAction,
) error {
	return r.table(ctx, db, nil).DeleteStatusActionValues(status, action)
}

func (r Repository) DeleteActionResultStatus(ctx context.Context, db sqlutil.Executor, actionID, statusID int64) error {
	return r.table(ctx, db, nil).DeleteActionResultStatus(actionID, statusID)
}

func (r Repository) DeleteActionResultStatusValues(
	ctx context.Context,
	db sqlutil.Executor,
	action repo.IEAction,
	status repo.IEStatus,
) error {
	return r.table(ctx, db, nil).DeleteActionResultStatusValues(action, status)
}

func (r Repository) StatusIDForValue(ctx context.Context, db sqlutil.Executor, value repo.IEStatus) (int64, error) {
	return r.table(ctx, db, nil).statusIDForValue(value)
}

func (r Repository) ActionIDForValue(ctx context.Context, db sqlutil.Executor, value repo.IEAction) (int64, error) {
	return r.table(ctx, db, nil).actionIDForValue(value)
}
