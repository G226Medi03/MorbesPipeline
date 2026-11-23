package orchestration

import (
	"encoding/json"
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/patient"
)

func TestResolvedCaseExtractionJSON(t *testing.T) {
	status := repo.HaveDisease{Disease: repo.EDisease("고혈압")}
	value := ResolvedCaseExtraction{
		Statuses: []repo.IEStatus{status},
		Actions:  []repo.IEAction{repo.CauseStatus{Status: status}},
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
	if loaded.Statuses[0].Kind() != "have_disease" {
		t.Fatalf("status = %#v", loaded.Statuses[0])
	}
	if loaded.Actions[0].Kind() != "cause_status" {
		t.Fatalf("action = %#v", loaded.Actions[0])
	}
	if len(loaded.PatientDescriptions) != 1 || len(loaded.PatientDescriptions[0].Description) != 1 {
		t.Fatalf("patient descriptions = %#v", loaded.PatientDescriptions)
	}
}
