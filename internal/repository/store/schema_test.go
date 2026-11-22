package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	testdb "github.com/G226Medi03/MorbesPipeline/internal/repository/testdb"
)

func TestInitSchemaDropsEntitySources(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "schema.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	var name string
	err = db.QueryRowContext(ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'entity_sources'`).Scan(&name)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("entity_sources lookup err = %v, name = %q; want no rows", err, name)
	}

	rows, err := db.QueryContext(ctx, `PRAGMA table_info(morbes_postprocesses)`)
	if err != nil {
		t.Fatalf("postprocess columns: %v", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	found := false
	for rows.Next() {
		var cid, notNull, pk int
		var columnName, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &columnName, &columnType, &notNull, &defaultValue, &pk); err != nil {
			t.Fatalf("scan postprocess column: %v", err)
		}
		if columnName == "entity_key" {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("postprocess columns: %v", err)
	}
	if !found {
		t.Fatal("morbes_postprocesses.entity_key is missing")
	}
}
