package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
)

func TestPostprocessRepositoryAppendAndListByEntity(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "postprocess.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	repo := NewPostprocessRepository()
	if err := repo.Init(ctx, db); err != nil {
		t.Fatalf("init: %v", err)
	}
	finished := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	id, err := repo.Append(ctx, db, PostprocessRecord{
		PostprocessorName: "transplant_rejection",
		Operation:         PostprocessOperationCreated,
		EntityType:        PostprocessEntityStatus,
		EntityID:          10,
		FinishedAt:        finished,
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if id <= 0 {
		t.Fatalf("id = %d", id)
	}
	records, err := repo.ListByEntity(ctx, db, PostprocessEntityStatus, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v", records)
	}
	if records[0].PostprocessorName != "transplant_rejection" {
		t.Fatalf("record = %#v", records[0])
	}
	if records[0].Operation != PostprocessOperationCreated {
		t.Fatalf("record = %#v", records[0])
	}
}
