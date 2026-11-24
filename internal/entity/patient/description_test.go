package patient_test

import (
	"encoding/json"
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/patient"
)

func TestParseStatus(t *testing.T) {
	status, err := patient.ParseStatus(json.RawMessage(`{"kind":"have_symptom","name":"육안적 혈뇨"}`))
	if err != nil {
		t.Fatalf("ParseStatus: %v", err)
	}
	if status.Kind() != "have_symptom" {
		t.Fatalf("kind = %q", status.Kind())
	}
}

func TestParseStatusRejectsMultiple(t *testing.T) {
	raw := json.RawMessage(`[
		{"kind":"have_disease","disease":"고혈압"},
		{"kind":"have_disease","disease":"저혈압"}
	]`)
	if _, err := patient.ParseStatus(raw); err == nil {
		t.Fatal("expected multiple status error")
	}
}

func TestParseStatusProcedureText(t *testing.T) {
	status, err := patient.ParseStatus(json.RawMessage(`{"kind":"received_medical_procedure","medical_procedure":"간 이식"}`))
	if err != nil {
		t.Fatalf("ParseStatus: %v", err)
	}
	received, ok := status.(repo.ReceivedMedicalProcedure)
	if !ok || received.MedicalProcedure.Kind() != "transplant" {
		t.Fatalf("status = %#v", status)
	}
}

func TestParseStatusScalarShorthand(t *testing.T) {
	cases := map[string]struct {
		raw  string
		want string
	}{
		"index":   {`{"kind":"increased_index","index":"칼륨"}`, "칼륨"},
		"symptom": {`{"kind":"have_symptom","symptom":"복통"}`, "복통"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			status, err := patient.ParseStatus(json.RawMessage(test.raw))
			if err != nil {
				t.Fatalf("ParseStatus: %v", err)
			}
			switch value := status.(type) {
			case repo.IncreasedIndex:
				if value.Index.Name != test.want {
					t.Fatalf("index = %q", value.Index.Name)
				}
			case repo.HaveSymptom:
				symptom, ok := value.Symptom.(repo.EOtherSymptom)
				if !ok || symptom.Name != test.want {
					t.Fatalf("symptom = %#v", value.Symptom)
				}
			default:
				t.Fatalf("status = %#v", status)
			}
		})
	}
}

func TestPatientDescriptionJSON(t *testing.T) {
	description := patient.PatientDescription{
		Choice:    "②",
		Knowledge: []string{"방광경 검사 후 소변에 피가 섞일 수 있다고 알고 있음"},
	}
	encoded, err := json.Marshal(description)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var loaded patient.PatientDescription
	if err := json.Unmarshal(encoded, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if loaded.Choice != "②" || len(loaded.Knowledge) != 1 {
		t.Fatalf("loaded = %#v", loaded)
	}
}
