package repository

import (
	"context"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

func (r EntityRepository) SaveProcedure(
	ctx context.Context,
	db sqlutil.Executor,
	entities *repo.Store,
	value repo.IMedicalProcedure,
	traits []trait.Trait,
	sources []source.Source,
) (int64, error) {
	return r.table(ctx, db, entities).SaveProcedure(value, traits, sources)
}

func (r EntityRepository) SaveMedicalProcedureOption(
	ctx context.Context,
	db sqlutil.Executor,
	entities *repo.Store,
	option repo.MedicalProcedureOption,
) (repo.MedicalProcedureOption, error) {
	return r.table(ctx, db, entities).SaveMedicalProcedureOption(option)
}

func (r EntityRepository) SaveMedicalProcedureOptionByValues(
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

func (r EntityRepository) LinkActionResultStatusValues(
	ctx context.Context,
	db sqlutil.Executor,
	action repo.IEAction,
	status repo.IEStatus,
	sourceID source.SourceId,
) error {
	return r.table(ctx, db, nil).LinkActionResultStatusValues(action, status, sourceID)
}

func (r EntityRepository) LinkStatusAction(ctx context.Context, db sqlutil.Executor, statusID, actionID int64, sourceID source.SourceId) error {
	return r.table(ctx, db, nil).LinkStatusAction(statusID, actionID, sourceID)
}

func (r EntityRepository) LinkStatusActionValues(
	ctx context.Context,
	db sqlutil.Executor,
	status repo.IEStatus,
	action repo.IEAction,
	sourceID source.SourceId,
) error {
	return r.table(ctx, db, nil).LinkStatusActionValues(status, action, sourceID)
}

func (r EntityRepository) LinkActionResultStatus(ctx context.Context, db sqlutil.Executor, actionID, statusID int64, sourceID source.SourceId) error {
	return r.table(ctx, db, nil).LinkActionResultStatus(actionID, statusID, sourceID)
}

func (r EntityRepository) DeleteStatusAction(ctx context.Context, db sqlutil.Executor, statusID, actionID int64) error {
	return r.table(ctx, db, nil).DeleteStatusAction(statusID, actionID)
}

func (r EntityRepository) DeleteStatusActionValues(
	ctx context.Context,
	db sqlutil.Executor,
	status repo.IEStatus,
	action repo.IEAction,
) error {
	return r.table(ctx, db, nil).DeleteStatusActionValues(status, action)
}

func (r EntityRepository) DeleteActionResultStatus(ctx context.Context, db sqlutil.Executor, actionID, statusID int64) error {
	return r.table(ctx, db, nil).DeleteActionResultStatus(actionID, statusID)
}

func (r EntityRepository) DeleteActionResultStatusValues(
	ctx context.Context,
	db sqlutil.Executor,
	action repo.IEAction,
	status repo.IEStatus,
) error {
	return r.table(ctx, db, nil).DeleteActionResultStatusValues(action, status)
}

func (r EntityRepository) GetActorByName(ctx context.Context, db sqlutil.Executor, entities *repo.Store, name string) (repo.Actor, bool, error) {
	return r.table(ctx, db, entities).GetActorByName(name)
}

func (r EntityRepository) GetDeviceByName(ctx context.Context, db sqlutil.Executor, entities *repo.Store, name string) (repo.Actor, bool, error) {
	return r.table(ctx, db, entities).GetDeviceByName(name)
}

func (r EntityRepository) GetAction(ctx context.Context, db sqlutil.Executor, entities *repo.Store, value repo.IEAction) (repo.Action, bool, error) {
	return r.table(ctx, db, entities).GetAction(value)
}

func (r EntityRepository) GetStatus(ctx context.Context, db sqlutil.Executor, entities *repo.Store, value repo.IEStatus) (repo.Status, bool, error) {
	return r.table(ctx, db, entities).GetStatus(value)
}

func (r EntityRepository) GetProcedure(ctx context.Context, db sqlutil.Executor, entities *repo.Store, value repo.IMedicalProcedure) (repo.Procedure, bool, error) {
	return r.table(ctx, db, entities).GetProcedure(value)
}

func (r EntityRepository) StatusIDForValue(ctx context.Context, db sqlutil.Executor, value repo.IEStatus) (int64, error) {
	return r.table(ctx, db, nil).statusIDForValue(value)
}

func (r EntityRepository) ActionIDForValue(ctx context.Context, db sqlutil.Executor, value repo.IEAction) (int64, error) {
	return r.table(ctx, db, nil).actionIDForValue(value)
}
