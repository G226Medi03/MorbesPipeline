package store

import (
	"context"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	"github.com/G226Medi03/SourceEntity"
	"path/filepath"
	"testing"
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
	sources := []source.Source{{SourceID: source.NewSourceId(1)}}
	container, err := st.LoadGraph()
	if err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	status := repo.HaveDisease{Disease: repo.EDisease("고혈압")}
	statusID, err := st.EnsureStatusIn(container, status, source.NewSourceId(1))
	if err != nil {
		t.Fatalf("EnsureStatusIn: %v", err)
	}
	action := repo.CauseStatus{Status: status}
	actionID, err := st.EnsureActionIn(container, action, source.NewSourceId(1))
	if err != nil {
		t.Fatalf("EnsureActionIn: %v", err)
	}
	if err := st.LinkStatusActionIn(container, statusID, actionID); err != nil {
		t.Fatalf("LinkStatusActionIn: %v", err)
	}
	procedureID, err := st.SaveProcedure(repo.Surgery{Name: "기관절개술"}, nil, sources)
	if err != nil {
		t.Fatalf("SaveProcedure: %v", err)
	}
	if _, err := st.SaveMedicalProcedureOptionByValues(
		procedureID, "treatment", nil, nil, []repo.IEStatus{status}, nil, sources,
	); err != nil {
		t.Fatalf("SaveMedicalProcedureOptionByValues: %v", err)
	}
	loaded, err := st.LoadGraph()
	if err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	if len(loaded.Statuses) != 1 || len(loaded.Actions) != 1 || len(loaded.MedicalProcedureOptionsAll) != 1 {
		t.Fatalf("graph = statuses:%d actions:%d options:%d", len(loaded.Statuses), len(loaded.Actions), len(loaded.MedicalProcedureOptionsAll))
	}
}
