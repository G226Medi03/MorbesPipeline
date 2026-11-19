package csolve

import (
	"encoding/json"
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
)

func TestStatusDraftJSONRoundTrip(t *testing.T) {
	draft := StatusDraft{
		Key:      "s1",
		Value:    repo.HaveSymptom{Symptom: repo.EOtherSymptom{Name: "두통"}},
		SourceID: source.NewSourceId(7),
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var loaded StatusDraft
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if loaded.Key != "s1" || loaded.Value == nil || loaded.Value.Kind() != "have_symptom" {
		t.Fatalf("loaded = %#v", loaded)
	}
	if loaded.SourceID.ToI64() != 7 {
		t.Fatalf("source = %d", loaded.SourceID.ToI64())
	}
}

func TestPatientDescriptionsForService(t *testing.T) {
	status := repo.HaveSymptom{Symptom: repo.EOtherSymptom{Name: "두통"}}
	got := patient_descriptions_for_service([]PatientDescriptionDraft{{
		Choice:    "②",
		Knowledge: []string{"복약"},
		Description: []StatusDraft{{
			Key:   "s1",
			Value: status,
		}},
	}})
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].Choice != "②" || len(got[0].Description) != 1 {
		t.Fatalf("got = %#v", got[0])
	}
	if got[0].Description[0].Kind() != status.Kind() {
		t.Fatalf("status = %s", got[0].Description[0].Kind())
	}
}
