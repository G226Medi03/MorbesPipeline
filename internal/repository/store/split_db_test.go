package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	postprocessrepo "github.com/G226Medi03/MorbesPipeline/internal/repository/postprocess"
	testutil "github.com/G226Medi03/MorbesPipeline/internal/testutil"
	source "github.com/G226Medi03/SourceEntity"
)

func TestWithPostprocessTxWithSplitDBs(t *testing.T) {
	ctx := context.Background()
	medicalDB, sourceDB, solveDB := openSplitDBs(t)

	st := NewWithDBs(ctx, medicalDB, sourceDB, solveDB)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	var sourceID source.SourceId
	err := st.WithPostprocessTx(ctx, func(tx *Store) error {
		postprocessID, err := tx.PostprocessRepo().Append(ctx, tx.MainExecutor(), postprocessrepo.Record{
			PostprocessorName: "test",
			Operation:         postprocessrepo.OperationCreated,
			EntityType:        postprocessrepo.EntityAction,
			EntityID:          1,
		})
		if err != nil {
			return err
		}
		sourceID, err = tx.SourceRepo().Ensure(
			ctx,
			tx.SourceExecutor(),
			source.SourceKindMorbesPostprocess,
			postprocessID,
		)
		return err
	})
	if err != nil {
		t.Fatalf("WithPostprocessTx: %v", err)
	}

	if sourceID.ToI64() == 0 {
		t.Fatal("postprocess source was not created")
	}

	var postprocessCount int
	if err := medicalDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM morbes_postprocesses`).Scan(&postprocessCount); err != nil {
		t.Fatalf("count postprocess rows: %v", err)
	}
	if postprocessCount != 1 {
		t.Fatalf("postprocess count = %d, want 1", postprocessCount)
	}

	var sourceCount int
	if err := sourceDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sources`).Scan(&sourceCount); err != nil {
		t.Fatalf("count source rows: %v", err)
	}
	if sourceCount != 1 {
		t.Fatalf("source count = %d, want 1", sourceCount)
	}

	var name string
	err = sourceDB.QueryRowContext(ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'morbes_postprocesses'`).Scan(&name)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("source morbes_postprocesses lookup err = %v, name = %q; want no rows", err, name)
	}
}

func openSplitDBs(t *testing.T) (medicalDB, sourceDB, solveDB *sql.DB) {
	t.Helper()
	dir := t.TempDir()
	open := func(name string) *sql.DB {
		t.Helper()
		db, err := testutil.OpenSQLite(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		t.Cleanup(func() { _ = db.Close() })
		return db
	}
	return open("medical.db"), open("source.db"), open("solve.db")
}
