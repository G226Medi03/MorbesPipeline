package morbes_pipeline

import (
	"context"
	"path/filepath"
	"testing"

	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
)

func TestPublicFacade(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "public-facade.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	store := NewStore(ctx, db)
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	container := NewContainer()
	status := HaveDisease{Disease: EDisease("고혈압")}
	if _, err := store.EnsureStatusIn(container, status, SourceID(1)); err != nil {
		t.Fatalf("EnsureStatusIn: %v", err)
	}
	if _, err := store.LoadGraph(); err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	service := NewService(ctx, Dependencies{Store: store})
	if service == nil {
		t.Fatal("NewService returned nil")
	}
}
