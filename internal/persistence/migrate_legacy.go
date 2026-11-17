package store

import (
	"context"
	"database/sql"
	"fmt"

	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
)

// migrateLegacyMedicalSchema renames the pre-MedicalEntity tables and columns
// to the schema used by entity.*Repository. Renames preserve ids and data.
// It is intentionally idempotent and safe on a fresh database.
func (s *Store) migrateLegacyMedicalSchema(ctx context.Context) error {
	renames := []struct {
		oldTable string
		newTable string
		columns  [][2]string
	}{
		{"e_statuses", "estatus", [][2]string{{"status_id", "estatus_id"}}},
		{"e_actions", "eaction", [][2]string{{"action_id", "eaction_id"}}},
		{"actors", "eactor", [][2]string{{"actor_id", "eactor_id"}}},
		{"eprocedure", "emedical_procedure", [][2]string{{"procedure_id", "emedical_procedure_id"}}},
		{"status_actions", "status_actions", [][2]string{{"status_id", "estatus_id"}, {"action_id", "eaction_id"}}},
		{"action_result_statuses", "action_result_statuses", [][2]string{{"action_id", "eaction_id"}, {"status_id", "estatus_id"}}},
		{"actor_actions", "actor_actions", [][2]string{{"actor_id", "eactor_id"}, {"action_id", "eaction_id"}}},
		{"actor_other_name", "actor_other_name", [][2]string{{"actor_id", "eactor_id"}}},
		{"status_sources", "status_sources", [][2]string{{"status_id", "estatus_id"}}},
		{"action_sources", "action_sources", [][2]string{{"action_id", "eaction_id"}}},
		{"actor_sources", "actor_sources", [][2]string{{"actor_id", "eactor_id"}}},
		{"procedure_sources", "procedure_sources", [][2]string{{"procedure_id", "emedical_procedure_id"}}},
		{"action_cause_status", "action_cause_status", [][2]string{{"status_id", "estatus_id"}}},
		{"action_block", "action_block", [][2]string{{"actor_id", "eactor_id"}}},
		{"action_relax", "action_relax", [][2]string{{"actor_id", "eactor_id"}}},
		{"status_damaged", "status_damaged", [][2]string{{"actor_id", "eactor_id"}}},
		{"status_inflammation", "status_inflammation", [][2]string{{"actor_id", "eactor_id"}}},
		{"status_received_medical_procedure", "status_received_medical_procedure", [][2]string{{"procedure_id", "emedical_procedure_id"}}},
		{"procedure_transplant", "procedure_transplant", [][2]string{{"target_actor_id", "target_eactor_id"}}},
		{"procedure_cleaning", "procedure_cleaning", [][2]string{{"target_actor_id", "target_eactor_id"}}},
		{"actor_procedure", "actor_procedure", [][2]string{{"procedure_id", "emedical_procedure_id"}}},
	}
	for _, item := range renames {
		if err := renameTableIfExists(ctx, s, item.oldTable, item.newTable); err != nil {
			return err
		}
		for _, column := range item.columns {
			if err := renameColumnIfExists(ctx, s, item.newTable, column[0], column[1]); err != nil {
				return err
			}
		}
	}
	// The legacy option table is not compatible with the normalized option
	// repository. Preserve it for audit/manual migration and let InitSchema
	// create the new shape.
	if exists, err := sqlutil.TableExists(ctx, s, "medical_procedure_options"); err != nil {
		return err
	} else if exists {
		hasLegacy, err := columnExists(ctx, s, "medical_procedure_options", "indication_status_ids_json")
		if err != nil {
			return err
		}
		if hasLegacy {
			if _, err := s.ExecContext(ctx,
				`ALTER TABLE medical_procedure_options RENAME TO medical_procedure_options_legacy`); err != nil {
				return fmt.Errorf("legacy medical_procedure_options rename failed: %w", err)
			}
		}
	}
	return nil
}

func renameTableIfExists(ctx context.Context, db sqlutil.Executor, oldTable, newTable string) error {
	oldExists, err := sqlutil.TableExists(ctx, db, oldTable)
	if err != nil || !oldExists {
		return err
	}
	newExists, err := sqlutil.TableExists(ctx, db, newTable)
	if err != nil {
		return err
	}
	if oldTable == newTable || newExists {
		return nil
	}
	if _, err := db.ExecContext(ctx,
		fmt.Sprintf("ALTER TABLE %s RENAME TO %s", oldTable, newTable)); err != nil {
		return fmt.Errorf("legacy table %s rename failed: %w", oldTable, err)
	}
	return nil
}

func renameColumnIfExists(ctx context.Context, db sqlutil.Executor, table, oldColumn, newColumn string) error {
	exists, err := sqlutil.TableExists(ctx, db, table)
	if err != nil || !exists {
		return err
	}
	hasOld, err := columnExists(ctx, db, table, oldColumn)
	if err != nil || !hasOld {
		return err
	}
	hasNew, err := columnExists(ctx, db, table, newColumn)
	if err != nil || hasNew {
		return err
	}
	if _, err := db.ExecContext(ctx,
		fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, oldColumn, newColumn)); err != nil {
		return fmt.Errorf("legacy column %s.%s rename failed: %w", table, oldColumn, err)
	}
	return nil
}

func columnExists(ctx context.Context, db sqlutil.Executor, table, column string) (bool, error) {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return false, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var cid, notnull, pk int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notnull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
