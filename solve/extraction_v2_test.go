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
