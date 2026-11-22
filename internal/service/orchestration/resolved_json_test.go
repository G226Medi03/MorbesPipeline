package csolve

import (
	"encoding/json"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/patient"
)

func TestResolvedCaseExtractionJSON(t *testing.T) {
	status := repo.HaveDisease{Disease: repo.EDisease("고혈압")}
	value := ResolvedCaseExtraction{
		Statuses: map[string]repo.IEStatus{"s1": status},
		Actions:  map[string]repo.IEAction{"a1": repo.CauseStatus{Status: status}},
		PatientDescriptions: []patient.PatientDescription{{
			Choice:      "②",
			Description: []repo.IEStatus{status},
			Knowledge:   []string{"혈압이 높다는 것을 알고 있음"},
		}},
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var loaded ResolvedCaseExtraction
	if err := json.Unmarshal(encoded, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if loaded.Statuses["s1"].Kind() != "have_disease" {
		t.Fatalf("status = %#v", loaded.Statuses["s1"])
	}
	if loaded.Actions["a1"].Kind() != "cause_status" {
		t.Fatalf("action = %#v", loaded.Actions["a1"])
	}
	if len(loaded.PatientDescriptions) != 1 || len(loaded.PatientDescriptions[0].Description) != 1 {
		t.Fatalf("patient descriptions = %#v", loaded.PatientDescriptions)
	}
}
