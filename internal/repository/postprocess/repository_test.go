package postprocess

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	testdb "github.com/G226Medi03/MorbesPipeline/internal/testutil/testdb"
)

func TestRepositoryAppendAndListByEntity(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "postprocess.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	repo := NewRepository()
	if err := repo.Init(ctx, db); err != nil {
		t.Fatalf("init: %v", err)
	}
	finished := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	id, err := repo.Append(ctx, db, Record{
		PostprocessorName: "transplant_rejection",
		Operation:         OperationCreated,
		EntityType:        EntityStatus,
		EntityKey:         "have_disease|신장",
		FinishedAt:        finished,
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if id <= 0 {
		t.Fatalf("id = %d", id)
	}
	records, err := repo.ListByEntity(ctx, db, EntityStatus, "have_disease|신장")
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
	if records[0].EntityKey != "have_disease|신장" {
		t.Fatalf("record = %#v", records[0])
	}
}

func TestRepositoryArchivesLegacyIDTable(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "postprocess-legacy.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	if _, err := db.ExecContext(ctx, `CREATE TABLE morbes_postprocesses (
		postprocess_id INTEGER PRIMARY KEY AUTOINCREMENT,
		postprocessor_name TEXT NOT NULL,
		operation TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		target_entity_type TEXT NOT NULL DEFAULT '',
		target_entity_id INTEGER NOT NULL DEFAULT 0,
		finished_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO morbes_postprocesses (
		postprocessor_name, operation, entity_type, entity_id
	) VALUES ('legacy', 'observed', 'status', 42)`); err != nil {
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
		EntityKey:         "have_disease|신장",
	}); err != nil {
		t.Fatalf("append new row: %v", err)
	}
}
