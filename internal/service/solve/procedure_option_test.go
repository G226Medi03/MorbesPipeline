package domain

import (
	"encoding/json"
	"strings"
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

func TestMedicalProcedureOptionGraphJSONUsesEnameFields(t *testing.T) {
	graph := MedicalProcedureOptionGraph{
		Options: []MedicalProcedureOptionGraphItem{{
			Key:       "surgery|신장이식|treatment",
			Procedure: repo.Surgery{Name: "신장이식"},
			Category:  "treatment",
			Indications: []MedicalProcedureOptionIndication{{
				Status: repo.HaveDisease{Disease: repo.EDisease("신장")},
				Label:  "신장 질환",
			}},
		}},
		SelectedOptionIndex: 1,
	}
	data, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(data)
	for _, want := range []string{`"status"`, `"procedure"`, `"selectedOptionIndex"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("json %s does not contain %s", body, want)
		}
	}
	for _, unwanted := range []string{`"statusId"`, `"procedureId"`, `"selectedOptionId"`} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("json %s contains %s", body, unwanted)
		}
	}
}
