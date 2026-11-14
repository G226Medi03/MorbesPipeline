package csituation

import (
	"encoding/json"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
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
	if len(statuses) != 3 || statuses[0].Kind() != "have_symptom" {
		t.Fatalf("statuses = %#v", statuses)
	}
	symptom, ok := statuses[0].(repo.HaveSymptom)
	if !ok {
		t.Fatalf("status = %#v", statuses[0])
	}
	if _, ok := symptom.Symptom.(repo.EPain); !ok {
		t.Fatalf("symptom = %#v", symptom.Symptom)
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
	if len(statuses) != 1 || statuses[0].Kind() != "received_medical_procedure" {
		t.Fatalf("statuses = %#v", statuses)
	}
	received := statuses[0].(repo.ReceivedMedicalProcedure)
	if received.MedicalProcedure.Kind() != "infuse" {
		t.Fatalf("procedure = %#v", received.MedicalProcedure)
	}
}
