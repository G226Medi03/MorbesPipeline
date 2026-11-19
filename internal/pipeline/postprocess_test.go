package csolve

import (
	"context"
	"path/filepath"
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/store"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	source "github.com/G226Medi03/SourceEntity"
)

func TestPostprocessCreatesTransplantRejection(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "postprocess.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := store.New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	sourceID := source.NewSourceId(1)
	status := repo.ReceivedMedicalProcedure{
		MedicalProcedure: repo.Transplant{Target: repo.NewUnknownActor("신장")},
	}
	entityStore, err := st.Entities()
	if err != nil {
		t.Fatalf("Entities: %v", err)
	}
	if _, err := entityStore.InsertStatus(status, []source.SourceId{sourceID}); err != nil {
		t.Fatalf("InsertStatus: %v", err)
	}
	state := &pipelineState{
		Statuses: map[string]repo.IEStatus{"s1": status},
		Actions:  map[string]repo.IEAction{},
		Extraction: CaseExtraction{
			Statuses: []StatusDraft{{Key: "s1", Value: status, SourceID: sourceID}},
		},
	}
	if err := Postprocess(ctx, st, state); err != nil {
		t.Fatalf("Postprocess: %v", err)
	}
	effect := repo.HaveSymptom{Symptom: repo.EOtherSymptom{Name: RejectionStatusName}}
	if _, err := entityStore.GetStatus(effect); err != nil {
		t.Fatalf("effect status: %v", err)
	}
	action := repo.CauseStatus{Status: effect}
	results, err := entityStore.ActionsResultingInStatus(effect)
	if err != nil {
		t.Fatalf("ActionsResultingInStatus: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("result links = %d, want 1", len(results))
	}
	causes, err := entityStore.StatusesCausingAction(action)
	if err != nil {
		t.Fatalf("StatusesCausingAction: %v", err)
	}
	if len(causes) != 1 || causes[0].Status.Kind() != status.Kind() {
		t.Fatalf("causes = %#v", causes)
	}
	var auditCount int
	if err := st.QueryRowContext(ctx, `SELECT COUNT(*) FROM morbes_postprocesses`).Scan(&auditCount); err != nil {
		t.Fatalf("audit count: %v", err)
	}
	if auditCount == 0 {
		t.Fatal("postprocess audit records are empty")
	}
}
