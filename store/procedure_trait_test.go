package store

import (
	"context"
	"github.com/G226Medi03/MorbesPipeline/sourceutil"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/entityconv"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/store/medicalprocedureoption"
	"github.com/G226Medi03/MorbesPipeline/trait"
	"github.com/G226Medi03/SourceEntity"
)

func TestSaveProcedureAndOptionTraits(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "procedure-traits.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	sources := []source.SourceRef{{SourceID: sourceutil.SourceID(1)}}
	procedure_id, err := st.SaveProcedureWithTraits(
		"Surgery",
		`{"name":"기관절개술"}`,
		entityconv.EntityTraits([]trait.Trait{trait.TraitSurgical, trait.TraitInvasive, trait.TraitDefinitive}),
		sources,
	)
	if err != nil {
		t.Fatalf("SaveProcedureWithTraits: %v", err)
	}
	option, err := st.SaveMedicalProcedureOption(medicalprocedureoption.NewMedicalProcedureOption(
		0,
		procedure_id,
		"treatment",
		nil,
		entityconv.EntityTraits([]trait.Trait{trait.TraitAcute}),
		nil,
		nil,
		sources,
		nil,
		nil,
	))
	if err != nil {
		t.Fatalf("SaveMedicalProcedureOption: %v", err)
	}
	if option.Id() == 0 {
		t.Fatal("option id is zero")
	}
	container, err := st.LoadGraph()
	if err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	if len(container.MedicalProcedureOptionsAll) != 1 {
		t.Fatalf("options = %#v", container.MedicalProcedureOptionsAll)
	}
	loaded := container.MedicalProcedureOptionsAll[0]
	if !trait.ContainsTrait(toProcedureTraits(loaded.Traits), trait.TraitSurgical) ||
		!trait.ContainsTrait(toProcedureTraits(loaded.Traits), trait.TraitInvasive) {
		t.Fatalf("procedure traits = %#v", loaded.Traits)
	}
	if !trait.ContainsTrait(toProcedureTraits(loaded.ContextTraits), trait.TraitAcute) {
		t.Fatalf("context traits = %#v", loaded.ContextTraits)
	}
}

func TestMedicalProcedureOptionSeparatesIndicationSets(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "procedure-option-indications.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	sources := []source.SourceRef{{SourceID: sourceutil.SourceID(1)}}
	procedureID, err := st.SaveProcedureWithTraits("Procedure", `{"name":"테스트 주입"}`, nil, sources)
	if err != nil {
		t.Fatalf("SaveProcedureWithTraits: %v", err)
	}
	firstStatus, err := st.SaveStatus(entity.NewStatus("have_disease", map[string]any{"disease": "첫 번째"}, sources))
	if err != nil {
		t.Fatalf("first status: %v", err)
	}
	secondStatus, err := st.SaveStatus(entity.NewStatus("have_disease", map[string]any{"disease": "두 번째"}, sources))
	if err != nil {
		t.Fatalf("second status: %v", err)
	}
	first, err := st.SaveMedicalProcedureOption(medicalprocedureoption.NewMedicalProcedureOption(
		0,
		procedureID,
		"treatment",
		nil,
		nil,
		[]entity.StatusID{firstStatus.Id()},
		nil,
		sources,
		nil,
		nil,
	))
	if err != nil {
		t.Fatalf("first option: %v", err)
	}
	second, err := st.SaveMedicalProcedureOption(medicalprocedureoption.NewMedicalProcedureOption(
		0,
		procedureID,
		"treatment",
		nil,
		nil,
		[]entity.StatusID{secondStatus.Id()},
		nil,
		sources,
		nil,
		nil,
	))
	if err != nil {
		t.Fatalf("second option: %v", err)
	}
	if first.Id() == second.Id() {
		t.Fatalf("options were merged: %d", first.Id())
	}
}

func toProcedureTraits(values []entity.Trait) []trait.Trait {
	if len(values) == 0 {
		return nil
	}
	result := make([]trait.Trait, 0, len(values))
	for _, value := range values {
		result = append(result, trait.Trait(value))
	}
	return result
}
