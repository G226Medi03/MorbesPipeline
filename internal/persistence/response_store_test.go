package store

import (
	"context"
	"path/filepath"
	"testing"

	source "github.com/G226Medi03/SourceEntity"

	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
)

func TestResponseStoreCache(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "response-store.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	refered := source.NewSourceId(1)
	inserted, err := st.GetOrInsertLlmResponse("model", "prompt", refered, "first")
	if err != nil {
		t.Fatalf("GetOrInsertLlmResponse: %v", err)
	}
	if inserted.OutputText() != "first" {
		t.Fatalf("output = %q", inserted.OutputText())
	}

	found, ok, err := st.FindLlmResponse("model", "prompt", refered)
	if err != nil {
		t.Fatalf("FindLlmResponse: %v", err)
	}
	if !ok || found.RespID != inserted.RespID {
		t.Fatalf("found = %#v, ok = %v", found, ok)
	}

	refreshed, err := st.RefreshLlmResponse("model", "prompt", refered, "second")
	if err != nil {
		t.Fatalf("RefreshLlmResponse: %v", err)
	}
	if refreshed.OutputText() != "second" || refreshed.RespID != inserted.RespID {
		t.Fatalf("refreshed = %#v", refreshed)
	}
}
