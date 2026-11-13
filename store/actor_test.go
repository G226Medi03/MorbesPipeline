package store

import (
	"context"
	"github.com/G226Medi03/MorbesPipeline/sourceutil"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	source "github.com/G226Medi03/SourceEntity"
)

func TestSaveMat(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "actor-kind.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	saved, err := st.SaveMat(eactor.UnknownActor{Name: "간"}, []source.SourceRef{{SourceID: sourceutil.SourceID(1)}})
	if err != nil {
		t.Fatalf("SaveMat: %v", err)
	}
	if saved.ID == 0 || saved.Value.Name != "간" {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestActorOtherNamesResolveToCanonicalActor(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "actor-other-names.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	canonical, err := st.SaveOtherActor(eactor.UnknownActor{Name: "S상결장루"}, []source.SourceRef{{SourceID: sourceutil.SourceID(1)}})
	if err != nil {
		t.Fatalf("SaveOtherActor: %v", err)
	}
	if err := st.AddActorOtherName(canonical.ID, "구불결장루", sourceutil.SourceID(0)); err != nil {
		t.Fatalf("AddActorOtherName: %v", err)
	}
	loaded, found, err := st.GetActorRecordByID(canonical.ID)
	if err != nil || !found {
		t.Fatalf("GetActorRecordByID: found=%v err=%v", found, err)
	}
	if len(loaded.OtherNames) != 1 || loaded.OtherNames[0] != "구불결장루" {
		t.Fatalf("other names = %#v", loaded.OtherNames)
	}
	by_other, found, err := st.GetActorRecordByName("구불결장루")
	if err != nil || !found {
		t.Fatalf("GetActorRecordByName: found=%v err=%v", found, err)
	}
	if by_other.ID != canonical.ID {
		t.Fatalf("other id = %d, canonical = %d", by_other.ID, canonical.ID)
	}
	saved, err := st.SaveOtherActor(eactor.UnknownActor{Name: "구불결장루"}, []source.SourceRef{{SourceID: sourceutil.SourceID(1)}})
	if err != nil {
		t.Fatalf("SaveOtherActor other: %v", err)
	}
	if saved.ID != canonical.ID {
		t.Fatalf("saved id = %d, canonical = %d", saved.ID, canonical.ID)
	}
}

func TestActorNameReusesExistingKind(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "actor-kind-reuse.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	tissue, err := st.SaveTissue(eactor.UnknownActor{Name: "경막"}, []source.SourceRef{{SourceID: sourceutil.SourceID(1)}})
	if err != nil {
		t.Fatalf("SaveTissue: %v", err)
	}
	other, err := st.SaveOtherActor(eactor.UnknownActor{Name: "경막"}, []source.SourceRef{{SourceID: sourceutil.SourceID(2)}})
	if err != nil {
		t.Fatalf("SaveOtherActor should reuse existing tissue: %v", err)
	}
	if other.ID != tissue.ID {
		t.Fatalf("actor id = %d, want tissue id %d", other.ID, tissue.ID)
	}
	record, found, err := st.GetActorRecordByID(other.ID)
	if err != nil || !found {
		t.Fatalf("GetActorRecordByID: found=%v err=%v", found, err)
	}
	if _, ok := record.Value.(eactor.UnknownActor); !ok {
		t.Fatalf("canonical actor value = %T, want tissue", record.Value)
	}
}
