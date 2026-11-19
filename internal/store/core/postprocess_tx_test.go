package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/store/postprocess"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
)

func TestWithPostprocessTxRollback(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "postprocess-tx.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	value := New(ctx, db)
	if err := value.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	want := errors.New("rollback")
	err = value.WithPostprocessTx(ctx, func(tx *StoreContainer) error {
		if _, err := tx.RecordPostprocess(
			"tx_test",
			postprocess.OperationObserved,
			postprocess.EntityStatus,
			1,
			"",
			0,
		); err != nil {
			return err
		}
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("WithPostprocessTx error = %v", err)
	}

	records, err := value.Postprocess.ListByEntity(ctx, value, postprocess.EntityStatus, 1)
	if err != nil {
		t.Fatalf("ListByEntity: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("records = %#v", records)
	}
}
