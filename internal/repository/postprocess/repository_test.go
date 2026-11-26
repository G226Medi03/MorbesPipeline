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
	id, err := repo.Append(ctx, db, Record{
		PostprocessorName: "transplant_rejection",
		Operation:         OperationCreated,
		EntityType:        EntityStatus,
		EntityID:          42,
		FinishedAt:        finished,
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if id <= 0 {
		t.Fatalf("id = %d", id)
	}
	records, err := repo.ListByEntity(ctx, db, EntityStatus, 42)
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
	if records[0].EntityID != 42 {
		t.Fatalf("record = %#v", records[0])
	}
}
