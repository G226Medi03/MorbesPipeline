package store

import (
	"context"
	"path/filepath"
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	source "github.com/G226Medi03/SourceEntity"
)

func TestTypedStoreRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "typed-store.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	entityStore, err := st.Entities()
	if err != nil {
		t.Fatalf("Entities: %v", err)
	}
	sources := []source.Source{{SourceID: source.NewSourceId(1)}}
	status := repo.HaveDisease{Disease: repo.EDisease("고혈압")}
	if _, err := entityStore.InsertStatus(status, []source.SourceId{source.NewSourceId(1)}); err != nil {
		t.Fatalf("InsertStatus: %v", err)
	}
	action := repo.CauseStatus{Status: status}
	if _, err := entityStore.InsertAction(action, nil, []source.SourceId{source.NewSourceId(1)}); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}
	if err := entityStore.LinkStatusAction(status, action, source.NewSourceId(1)); err != nil {
		t.Fatalf("LinkStatusAction: %v", err)
	}
	procedure := repo.Surgery{Name: "기관절개술"}
	if _, err := st.SaveProcedure(procedure, nil, sources); err != nil {
		t.Fatalf("SaveProcedure: %v", err)
	}
	if _, err := st.SaveMedicalProcedureOptionByValues(
		procedure, "treatment", nil, nil, []repo.IEStatus{status}, nil, sources,
	); err != nil {
		t.Fatalf("SaveMedicalProcedureOptionByValues: %v", err)
	}
	if _, err := entityStore.GetStatus(status); err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if _, err := entityStore.GetAction(action); err != nil {
		t.Fatalf("GetAction: %v", err)
	}
	links, err := entityStore.ActionsResultingInStatus(status)
	if err != nil {
		t.Fatalf("ActionsResultingInStatus: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("ActionsResultingInStatus = %d, want 0", len(links))
	}
}
