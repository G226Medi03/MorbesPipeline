package graph

import (
	"context"
	"fmt"

	option "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	sqlutil "github.com/G226Medi03/MorbesPipeline/internal/external/sqlcore/sqlite"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

type Repository struct {
	ctx context.Context
	db  sqlutil.Executor
}

func NewRepository(ctx context.Context, db sqlutil.Executor) Repository {
	return Repository{ctx: ctx, db: db}
}

func (r Repository) Load() (Rows, error) {
	var result Rows
	var err error
	if result.Actors, err = r.ListActors(); err != nil {
		return Rows{}, err
	}
	if result.Statuses, err = r.ListStatuses(); err != nil {
		return Rows{}, err
	}
	if result.Actions, err = r.ListActions(); err != nil {
		return Rows{}, err
	}
	if result.Procedures, err = r.ListProcedures(); err != nil {
		return Rows{}, err
	}
	if result.Options, err = r.ListOptions(); err != nil {
		return Rows{}, err
	}
	if result.StatusActions, err = r.ListStatusActions(); err != nil {
		return Rows{}, err
	}
	if result.ActionResultStatuses, err = r.ListActionResultStatuses(); err != nil {
		return Rows{}, err
	}
	if result.ActorActions, err = r.ListActorActions(); err != nil {
		return Rows{}, err
	}
	if result.StatusSources, err = r.listSources("status_sources", "estatus_id"); err != nil {
		return Rows{}, err
	}
	if result.ActionSources, err = r.listSources("action_sources", "eaction_id"); err != nil {
		return Rows{}, err
	}
	if result.ActorSources, err = r.listSources("actor_sources", "eactor_id"); err != nil {
		return Rows{}, err
	}
	if result.ProcedureSources, err = r.listSources("procedure_sources", "emedical_procedure_id"); err != nil {
		return Rows{}, err
	}
	if result.OptionSources, err = r.listSources("medical_procedure_option_sources", "medical_procedure_option_id"); err != nil {
		return Rows{}, err
	}
	return result, nil
}

func (r Repository) ListActors() ([]ActorRow, error) {
	rows, err := r.db.QueryContext(r.ctx, `SELECT eactor_id, kind, name FROM eactor ORDER BY eactor_id`)
	if err != nil {
		return nil, fmt.Errorf("graph actors load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []ActorRow
	for rows.Next() {
		var row ActorRow
		if err := rows.Scan(&row.ID, &row.Kind, &row.Name); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r Repository) ListStatuses() ([]StatusRow, error) {
	rows, err := r.db.QueryContext(r.ctx, `SELECT estatus_id, kind, ptr FROM estatus ORDER BY estatus_id`)
	if err != nil {
		return nil, fmt.Errorf("graph statuses load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []StatusRow
	for rows.Next() {
		var row StatusRow
		if err := rows.Scan(&row.ID, &row.Kind, &row.Ptr); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r Repository) ListActions() ([]ActionRow, error) {
	rows, err := r.db.QueryContext(r.ctx, `SELECT eaction_id, kind, ptr FROM eaction ORDER BY eaction_id`)
	if err != nil {
		return nil, fmt.Errorf("graph actions load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []ActionRow
	for rows.Next() {
		var row ActionRow
		if err := rows.Scan(&row.ID, &row.Kind, &row.Ptr); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r Repository) ListProcedures() ([]ProcedureRow, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT emedical_procedure_id, kind, ptr, traits_json FROM emedical_procedure ORDER BY emedical_procedure_id`)
	if err != nil {
		return nil, fmt.Errorf("graph procedures load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []ProcedureRow
	for rows.Next() {
		var row ProcedureRow
		if err := rows.Scan(&row.ID, &row.Kind, &row.Ptr, &row.TraitsJSON); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r Repository) ListOptions() ([]OptionRow, error) {
	rows, err := r.db.QueryContext(r.ctx, `
		SELECT medical_procedure_option_id, emedical_procedure_id, category, traits_json, context_traits_json
		FROM medical_procedure_options
		ORDER BY medical_procedure_option_id`)
	if err != nil {
		return nil, fmt.Errorf("graph procedure options load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []OptionRow
	for rows.Next() {
		var rawID int64
		var row OptionRow
		if err := rows.Scan(&rawID, &row.ProcedureID, &row.Category, &row.TraitsJSON, &row.ContextTraitsJSON); err != nil {
			return nil, err
		}
		row.ID = option.MedicalProcedureOptionID(rawID)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range result {
		var err error
		result[index].IndicationStatusIDs, err = r.listStatusIDs("medical_procedure_option_indications", result[index].ID.ToI64())
		if err != nil {
			return nil, err
		}
		result[index].ExpectedEffectStatusIDs, err = r.listStatusIDs("medical_procedure_option_expected_effects", result[index].ID.ToI64())
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r Repository) ListStatusActions() ([]StatusActionRow, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT estatus_id, eaction_id, source_id FROM status_actions ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("graph status actions load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []StatusActionRow
	for rows.Next() {
		var row StatusActionRow
		if err := rows.Scan(&row.StatusID, &row.ActionID, &row.SourceID); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r Repository) ListActionResultStatuses() ([]ActionResultStatusRow, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT eaction_id, estatus_id, source_id FROM action_result_statuses ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("graph action result statuses load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []ActionResultStatusRow
	for rows.Next() {
		var row ActionResultStatusRow
		if err := rows.Scan(&row.ActionID, &row.StatusID, &row.SourceID); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r Repository) ListActorActions() ([]ActorActionRow, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT eactor_id, eaction_id, source_id FROM actor_actions ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("graph actor actions load: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []ActorActionRow
	for rows.Next() {
		var row ActorActionRow
		if err := rows.Scan(&row.ActorID, &row.ActionID, &row.SourceID); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r Repository) listStatusIDs(table string, optionID int64) ([]ids.StatusID, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT estatus_id FROM `+table+` WHERE medical_procedure_option_id = ? ORDER BY estatus_id`,
		optionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []ids.StatusID
	for rows.Next() {
		var statusID ids.StatusID
		if err := rows.Scan(&statusID); err != nil {
			return nil, err
		}
		result = append(result, statusID)
	}
	return result, rows.Err()
}

func (r Repository) listSources(table, idColumn string) ([]SourceRow, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT `+idColumn+`, source_id FROM `+table+` ORDER BY `+idColumn+`, source_id`)
	if err != nil {
		return nil, fmt.Errorf("graph %s load: %w", table, err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []SourceRow
	for rows.Next() {
		var row SourceRow
		var sourceID source.SourceId
		if err := rows.Scan(&row.EntityID, &sourceID); err != nil {
			return nil, err
		}
		row.SourceID = sourceID
		result = append(result, row)
	}
	return result, rows.Err()
}
