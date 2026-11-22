package entity

import (
	"context"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

func (r Repository) LinkActionResultStatusValues(
	ctx context.Context,
	db sqlutil.Executor,
	action repo.IEAction,
	status repo.IEStatus,
	sourceID source.SourceId,
) error {
	return r.table(ctx, db, nil).LinkActionResultStatusValues(action, status, sourceID)
}

func (r Repository) DeleteStatusActionValues(
	ctx context.Context,
	db sqlutil.Executor,
	status repo.IEStatus,
	action repo.IEAction,
) error {
	return r.table(ctx, db, nil).DeleteStatusActionValues(status, action)
}

func (r Repository) DeleteActionResultStatusValues(
	ctx context.Context,
	db sqlutil.Executor,
	action repo.IEAction,
	status repo.IEStatus,
) error {
	return r.table(ctx, db, nil).DeleteActionResultStatusValues(action, status)
}
