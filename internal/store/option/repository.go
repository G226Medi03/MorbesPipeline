package medicalprocedureoption

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G226Medi03/MedicalEntityRepository/repo"
	mp "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

type Repository struct {
	ctx context.Context
	db  sqlutil.Executor
}

func NewRepository(ctx context.Context, db sqlutil.Executor) Repository {
	return Repository{ctx: ctx, db: db}
}

func (r Repository) Init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS medical_procedure_options (
			medical_procedure_option_id INTEGER PRIMARY KEY AUTOINCREMENT,
			emedical_procedure_id INTEGER NOT NULL,
			category TEXT NOT NULL DEFAULT 'treatment',
			traits_json TEXT NOT NULL DEFAULT '[]',
			context_traits_json TEXT NOT NULL DEFAULT '[]',
			semantic_key TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(semantic_key),
			FOREIGN KEY (emedical_procedure_id) REFERENCES emedical_procedure(emedical_procedure_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS medical_procedure_option_indications (
			medical_procedure_option_id INTEGER NOT NULL,
			estatus_id INTEGER NOT NULL,
			PRIMARY KEY (medical_procedure_option_id, estatus_id),
			FOREIGN KEY (medical_procedure_option_id) REFERENCES medical_procedure_options(medical_procedure_option_id) ON DELETE CASCADE,
			FOREIGN KEY (estatus_id) REFERENCES estatus(estatus_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS medical_procedure_option_expected_effects (
			medical_procedure_option_id INTEGER NOT NULL,
			estatus_id INTEGER NOT NULL,
			PRIMARY KEY (medical_procedure_option_id, estatus_id),
			FOREIGN KEY (medical_procedure_option_id) REFERENCES medical_procedure_options(medical_procedure_option_id) ON DELETE CASCADE,
			FOREIGN KEY (estatus_id) REFERENCES estatus(estatus_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS medical_procedure_option_sources (
			medical_procedure_option_id INTEGER NOT NULL,
			source_id INTEGER NOT NULL,
			PRIMARY KEY (medical_procedure_option_id, source_id),
			FOREIGN KEY (medical_procedure_option_id) REFERENCES medical_procedure_options(medical_procedure_option_id) ON DELETE CASCADE
		)`,
	}
	for _, query := range queries {
		if _, err := r.db.ExecContext(r.ctx, query); err != nil {
			return fmt.Errorf("medical procedure option schema init failed: %w", err)
		}
	}
	for _, query := range []string{
		"CREATE INDEX IF NOT EXISTS idx_medical_procedure_options_procedure ON medical_procedure_options(emedical_procedure_id);",
		"CREATE INDEX IF NOT EXISTS idx_medical_procedure_option_indications_status ON medical_procedure_option_indications(estatus_id);",
		"CREATE INDEX IF NOT EXISTS idx_medical_procedure_option_effects_status ON medical_procedure_option_expected_effects(estatus_id);",
	} {
		if _, err := r.db.ExecContext(r.ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (r Repository) Save(option mp.MedicalProcedureOption) (mp.MedicalProcedureOption, error) {
	if err := option.Validate(); err != nil {
		return option, err
	}
	option.Category = normalize_category(option.Category)
	traits, err := encodeTraits(option.Traits)
	if err != nil {
		return option, err
	}
	contextTraits, err := encodeTraits(option.ContextTraits)
	if err != nil {
		return option, err
	}
	indications, err := encodeIDs(option.IndicationStatusIDs)
	if err != nil {
		return option, err
	}
	semanticKey := fmt.Sprintf(
		"%d|%s|%s|%s|%s",
		option.ProcedureID.ToI64(),
		option.Category,
		string(traits),
		string(contextTraits),
		string(indications),
	)
	if _, err := r.db.ExecContext(r.ctx, `
		INSERT INTO medical_procedure_options (
			emedical_procedure_id, category, traits_json, context_traits_json, semantic_key
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(semantic_key) DO UPDATE SET
			category = excluded.category,
			traits_json = excluded.traits_json,
			context_traits_json = excluded.context_traits_json`,
		option.ProcedureID.ToI64(),
		option.Category,
		string(traits),
		string(contextTraits),
		semanticKey,
	); err != nil {
		return option, fmt.Errorf("medical procedure option save failed: %w", err)
	}
	var id int64
	if err := r.db.QueryRowContext(r.ctx,
		`SELECT medical_procedure_option_id FROM medical_procedure_options WHERE semantic_key = ?`,
		semanticKey,
	).Scan(&id); err != nil {
		return option, fmt.Errorf("medical procedure option lookup failed: %w", err)
	}
	option = option.WithID(mp.MedicalProcedureOptionID(id))
	if err := r.replaceStatusIDs("medical_procedure_option_indications", option.Id(), option.IndicationStatusIDs); err != nil {
		return option, err
	}
	if err := r.replaceStatusIDs("medical_procedure_option_expected_effects", option.Id(), option.ExpectedEffectStatusIDs); err != nil {
		return option, err
	}
	if err := r.replaceSources(option.Id(), option.Sources); err != nil {
		return option, err
	}
	return option, nil
}

func (r Repository) Get(id mp.MedicalProcedureOptionID) (mp.MedicalProcedureOption, error) {
	var procedureID ids.ProcedureID
	var category, traitsJSON, contextTraitsJSON string
	err := r.db.QueryRowContext(r.ctx, `
		SELECT emedical_procedure_id, category, traits_json, context_traits_json
		FROM medical_procedure_options
		WHERE medical_procedure_option_id = ?`,
		id.ToI64(),
	).Scan(&procedureID, &category, &traitsJSON, &contextTraitsJSON)
	if err == sql.ErrNoRows {
		return mp.MedicalProcedureOption{}, repo.ErrNotFound
	}
	if err != nil {
		return mp.MedicalProcedureOption{}, fmt.Errorf("medical procedure option get failed: %w", err)
	}
	var traits []trait.Trait
	if err := json.Unmarshal([]byte(traitsJSON), &traits); err != nil {
		return mp.MedicalProcedureOption{}, fmt.Errorf("medical procedure option traits decode failed: %w", err)
	}
	var contextTraits []trait.Trait
	if err := json.Unmarshal([]byte(contextTraitsJSON), &contextTraits); err != nil {
		return mp.MedicalProcedureOption{}, fmt.Errorf("medical procedure option context traits decode failed: %w", err)
	}
	indications, err := r.listStatusIDs("medical_procedure_option_indications", id)
	if err != nil {
		return mp.MedicalProcedureOption{}, err
	}
	expectedEffects, err := r.listStatusIDs("medical_procedure_option_expected_effects", id)
	if err != nil {
		return mp.MedicalProcedureOption{}, err
	}
	sources, err := r.listSources(id)
	if err != nil {
		return mp.MedicalProcedureOption{}, err
	}
	return mp.NewMedicalProcedureOption(
		id,
		procedureID,
		category,
		traits,
		contextTraits,
		indications,
		expectedEffects,
		sources,
		nil,
		nil,
	), nil
}

func (r Repository) replaceStatusIDs(table string, id mp.MedicalProcedureOptionID, ids []ids.StatusID) error {
	if _, err := r.db.ExecContext(r.ctx,
		`DELETE FROM `+table+` WHERE medical_procedure_option_id = ?`, id.ToI64()); err != nil {
		return err
	}
	return r.mergeStatusIDs(table, id, ids)
}

func (r Repository) mergeStatusIDs(table string, id mp.MedicalProcedureOptionID, ids []ids.StatusID) error {
	for _, statusID := range ids {
		if statusID.ToI64() <= 0 {
			continue
		}
		if _, err := r.db.ExecContext(r.ctx,
			`INSERT OR IGNORE INTO `+table+` (medical_procedure_option_id, estatus_id) VALUES (?, ?)`,
			id.ToI64(), statusID.ToI64()); err != nil {
			return fmt.Errorf("%s insert failed: %w", table, err)
		}
	}
	return nil
}

func (r Repository) listStatusIDs(table string, id mp.MedicalProcedureOptionID) ([]ids.StatusID, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT estatus_id FROM `+table+` WHERE medical_procedure_option_id = ? ORDER BY estatus_id`,
		id.ToI64())
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

func (r Repository) replaceSources(id mp.MedicalProcedureOptionID, sources []source.SourceRef) error {
	if _, err := r.db.ExecContext(r.ctx,
		`DELETE FROM medical_procedure_option_sources WHERE medical_procedure_option_id = ?`, id.ToI64()); err != nil {
		return err
	}
	for _, ref := range sources {
		if ref.SourceID.ToI64() <= 0 {
			continue
		}
		if _, err := r.db.ExecContext(r.ctx,
			`INSERT OR IGNORE INTO medical_procedure_option_sources (medical_procedure_option_id, source_id) VALUES (?, ?)`,
			id.ToI64(), ref.SourceID.ToI64()); err != nil {
			return err
		}
	}
	return nil
}

func (r Repository) listSources(id mp.MedicalProcedureOptionID) ([]source.SourceRef, error) {
	rows, err := r.db.QueryContext(r.ctx,
		`SELECT source_id FROM medical_procedure_option_sources WHERE medical_procedure_option_id = ? ORDER BY source_id`,
		id.ToI64())
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var result []source.SourceRef
	for rows.Next() {
		var sourceID source.SourceId
		if err := rows.Scan(&sourceID); err != nil {
			return nil, err
		}
		result = append(result, source.SourceRef{SourceID: sourceID})
	}
	return result, rows.Err()
}

func normalize_category(value string) string {
	switch strings.TrimSpace(value) {
	case "diagnostic_test", "test", "diagnostic":
		return "diagnostic_test"
	case "monitoring":
		return "monitoring"
	case "education":
		return "education"
	case "nursing_intervention":
		return "nursing_intervention"
	default:
		return "treatment"
	}
}

func encodeTraits(values []trait.Trait) ([]byte, error) {
	if values == nil {
		values = []trait.Trait{}
	}
	return json.Marshal(values)
}

func encodeIDs(values []ids.StatusID) ([]byte, error) {
	if values == nil {
		values = []ids.StatusID{}
	}
	return json.Marshal(values)
}
