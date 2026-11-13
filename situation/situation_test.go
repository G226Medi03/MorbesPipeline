package csituation

import (
	"encoding/json"
	"testing"
)

func TestToStatuses(t *testing.T) {
	raw := `{
		"findings":{"status":[
			{"kind":"have_symptom","symptom":{"kind":"symptom_code","code":"chest_pain"}},
			{"kind":"increased_index","index":{"name":"blood_pressure"}},
			{"kind":"have_disease","disease":"angina"}
		]}
	}`
	var situation Situation
	if err := json.Unmarshal([]byte(raw), &situation); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	statuses, err := situation.ToStatuses()
	if err != nil {
		t.Fatalf("ToStatuses: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("statuses = %d", len(statuses))
	}
	if statuses[0].Variant != "have_symptom" {
		t.Fatalf("variant = %s", statuses[0].Variant)
	}
	payload, ok := statuses[0].Value.(map[string]any)
	if !ok {
		t.Fatalf("value = %#v", statuses[0].Value)
	}
	symptom, ok := payload["symptom"].(map[string]any)
	if !ok || symptom["kind"] != "pain" {
		t.Fatalf("symptom = %#v", payload["symptom"])
	}
}

func TestToStatusesProcedureText(t *testing.T) {
	raw := `{
		"findings":{"status":[
			{"kind":"received_medical_procedure","medical_procedure":"정맥 수액 투여"}
		]}
	}`
	var situation Situation
	if err := json.Unmarshal([]byte(raw), &situation); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	statuses, err := situation.ToStatuses()
	if err != nil {
		t.Fatalf("ToStatuses: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("statuses = %d", len(statuses))
	}
	payload := statuses[0].Value.(map[string]any)
	procedure, ok := payload["medical_procedure"].(map[string]any)
	if !ok || procedure["kind"] != "infuse" {
		t.Fatalf("medical_procedure = %#v", payload["medical_procedure"])
	}
}
