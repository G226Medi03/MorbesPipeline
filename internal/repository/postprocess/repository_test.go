package postprocess

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	testutil "github.com/G226Medi03/MorbesPipeline/internal/testutil"
)

func TestRepositoryAppendAndListByEntity(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.OpenSQLite(filepath.Join(t.TempDir(), "postprocess.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	repo := NewRepository()
	if err := repo.Init(ctx, db); err != nil {
		t.Fatalf("init: %v", err)
	}
	finished := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	const entityJSON = `{"kind":"have_disease","disease":"신장"}`
	id, err := repo.Append(ctx, db, Record{
		PostprocessorName: "transplant_rejection",
		Operation:         OperationCreated,
		EntityType:        EntityStatus,
		EntityJSON:        entityJSON,
		FinishedAt:        finished,
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if id <= 0 {
		t.Fatalf("id = %d", id)
	}
	records, err := repo.ListByEntity(ctx, db, EntityStatus, entityJSON)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v", records)
	}
	if records[0].PostprocessorName != "transplant_rejection" {
		t.Fatalf("record = %#v", records[0])
	}
	if records[0].Operation != OperationCreated {
		t.Fatalf("record = %#v", records[0])
	}
	if records[0].EntityJSON != entityJSON {
		t.Fatalf("record = %#v", records[0])
	}
}

func TestRepositoryArchivesKeyTable(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.OpenSQLite(filepath.Join(t.TempDir(), "postprocess-key.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	if _, err := db.ExecContext(ctx, `CREATE TABLE morbes_postprocesses (
		postprocess_id INTEGER PRIMARY KEY AUTOINCREMENT,
		postprocessor_name TEXT NOT NULL,
		operation TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_key TEXT NOT NULL,
		target_entity_type TEXT NOT NULL DEFAULT '',
		target_entity_key TEXT NOT NULL DEFAULT '',
		finished_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create key table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO morbes_postprocesses (
		postprocessor_name, operation, entity_type, entity_key
	) VALUES ('legacy', 'observed', 'status', 'have_disease|신장')`); err != nil {
		t.Fatalf("insert key row: %v", err)
	}

	repo := NewRepository()
	if err := repo.Init(ctx, db); err != nil {
		t.Fatalf("init: %v", err)
	}
	var legacyCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM morbes_postprocesses_legacy`).Scan(&legacyCount); err != nil {
		t.Fatalf("legacy count: %v", err)
	}
	if legacyCount != 1 {
		t.Fatalf("legacy count = %d, want 1", legacyCount)
	}
	columns, err := tableColumns(ctx, db, "morbes_postprocesses")
	if err != nil {
		t.Fatalf("columns: %v", err)
	}
	if columns["entity_key"] || columns["target_entity_key"] {
		t.Fatalf("key columns remain: %+v", columns)
	}
	if !columns["entity_json"] {
		t.Fatalf("entity_json is missing: %+v", columns)
	}
	if _, err := repo.Append(ctx, db, Record{
		PostprocessorName: "new",
		Operation:         OperationObserved,
		EntityType:        EntityStatus,
		EntityJSON:        `{"kind":"have_disease","disease":"신장"}`,
	}); err != nil {
		t.Fatalf("append new row: %v", err)
	}
}

func TestRepositoryArchivesLegacyTable(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.OpenSQLite(filepath.Join(t.TempDir(), "postprocess-legacy.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	if _, err := db.ExecContext(ctx, `CREATE TABLE morbes_postprocesses (
		postprocess_id INTEGER PRIMARY KEY AUTOINCREMENT,
		postprocessor_name TEXT NOT NULL,
		operation TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		legacy_ref TEXT NOT NULL,
		target_entity_type TEXT NOT NULL DEFAULT '',
		target_ref TEXT NOT NULL DEFAULT '',
		finished_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO morbes_postprocesses (
		postprocessor_name, operation, entity_type, legacy_ref
	) VALUES ('legacy', 'observed', 'status', 'legacy-ref')`); err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	repo := NewRepository()
	if err := repo.Init(ctx, db); err != nil {
		t.Fatalf("init: %v", err)
	}
	var legacyCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM morbes_postprocesses_legacy`).Scan(&legacyCount); err != nil {
		t.Fatalf("legacy count: %v", err)
	}
	if legacyCount != 1 {
		t.Fatalf("legacy count = %d, want 1", legacyCount)
	}
	if _, err := repo.Append(ctx, db, Record{
		PostprocessorName: "new",
		Operation:         OperationObserved,
		EntityType:        EntityStatus,
		EntityJSON:        `{"kind":"have_disease","disease":"신장"}`,
	}); err != nil {
		t.Fatalf("append new row: %v", err)
	}
}
