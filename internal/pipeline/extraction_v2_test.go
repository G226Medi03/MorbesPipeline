package csolve

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStatusDraftFromRawName(t *testing.T) {
	draft, ok := status_draft_from_raw(json.RawMessage(`{"kind":"have_symptom","name":"두통"}`), "p1_1")
	if !ok {
		t.Fatal("status_draft_from_raw returned false")
	}
	if draft.Key != "p1_1" || draft.Kind != "have_symptom" {
		t.Fatalf("draft = %#v", draft)
	}
	if !strings.Contains(string(draft.Value), "두통") {
		t.Fatalf("value = %s", draft.Value)
	}
}

func TestStatusDraftFromRawValue(t *testing.T) {
	draft, ok := status_draft_from_raw(json.RawMessage(`{"kind":"have_symptom","value":{"name":"두통"}}`), "p1_1")
	if !ok {
		t.Fatal("status_draft_from_raw returned false")
	}
	if !strings.Contains(string(draft.Value), "두통") {
		t.Fatalf("value = %s", draft.Value)
	}
}

func TestStatusDraftsJSONUsesIDField(t *testing.T) {
	raw := status_drafts_json([]StatusDraft{{
		Key:   "s1",
		Kind:  "have_symptom",
		Value: json.RawMessage(`{"name":"두통"}`),
	}})
	if !strings.Contains(raw, `"id":"s1"`) || !strings.Contains(raw, `"kind":"have_symptom"`) {
		t.Fatalf("status json = %s", raw)
	}
}

func TestPatientDescriptionsForService(t *testing.T) {
	got := patient_descriptions_for_service([]PatientDescriptionDraft{{
		Choice:    "②",
		Knowledge: []string{"복약"},
		Description: []StatusDraft{{
			Key:   "s1",
			Kind:  "have_symptom",
			Value: json.RawMessage(`{"name":"두통"}`),
		}},
	}})
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	raw, err := json.Marshal(got[0])
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	for _, want := range []string{
		`"choice":"②"`,
		`"kind":"have_symptom"`,
		`"name":"두통"`,
		`"knowledge":["복약"]`,
	} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("json = %s, want %s", raw, want)
		}
	}
}
