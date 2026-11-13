package patient_test

import (
	"encoding/json"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/patient"
)

func TestParseStatus(t *testing.T) {
	status, err := patient.ParseStatus(json.RawMessage(`{"kind":"have_symptom","name":"육안적 혈뇨"}`))
	if err != nil {
		t.Fatalf("ParseStatus: %v", err)
	}
	if status.Variant != "have_symptom" {
		t.Fatalf("variant = %q", status.Variant)
	}
	if status.Value == nil {
		t.Fatal("value is nil")
	}
}

func TestParseStatusProcedureText(t *testing.T) {
	status, err := patient.ParseStatus(json.RawMessage(`{"kind":"received_medical_procedure","medical_procedure":"간 이식"}`))
	if err != nil {
		t.Fatalf("ParseStatus: %v", err)
	}
	if status.Variant != "received_medical_procedure" {
		t.Fatalf("variant = %q", status.Variant)
	}
	payload, ok := status.Value.(map[string]any)
	if !ok {
		t.Fatalf("value = %#v", status.Value)
	}
	procedure, ok := payload["medical_procedure"].(map[string]any)
	if !ok || procedure["kind"] != "transplant" {
		t.Fatalf("medical_procedure = %#v", payload["medical_procedure"])
	}
}

func TestParseStatusScalarShorthand(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		key     string
		field   string
		want    string
		variant string
	}{
		{"actor", `{"kind":"damaged","actor":"도뇨관"}`, "actor", "kind", "unknown_actor", "damaged"},
		{"index", `{"kind":"increased_index","index":"칼륨"}`, "index", "name", "칼륨", "increased_index"},
		{"symptom", `{"kind":"have_symptom","symptom":"복통"}`, "symptom", "name", "복통", "have_symptom"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			status, err := patient.ParseStatus(json.RawMessage(test.raw))
			if err != nil {
				t.Fatalf("ParseStatus: %v", err)
			}
			if status.Variant != test.variant {
				t.Fatalf("variant = %q", status.Variant)
			}
			payload := status.Value.(map[string]any)
			nested, ok := payload[test.key].(map[string]any)
			if !ok || nested[test.field] != test.want {
				t.Fatalf("%s = %#v", test.key, payload[test.key])
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
