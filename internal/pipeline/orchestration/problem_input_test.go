package csolve

import (
	"testing"

	"github.com/G226Medi03/MedicalEntity/ename/edisease"
	"github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
)

func TestProblemStemRemovesChoices(t *testing.T) {
	problem := "38 어지럼, 실신으로 온 환자의 치료는?\n① 디곡신\n② 아트로핀"
	got := problem_stem(problem)
	if got != "38 어지럼, 실신으로 온 환자의 치료는?" {
		t.Fatalf("stem = %q", got)
	}
	choices := problem_choices(problem)
	if len(choices) != 2 || choices[1].Index != 2 || choices[1].Text != "아트로핀" {
		t.Fatalf("choices = %#v", choices)
	}
}

func TestSelectMedicalProcedureOptionIgnoresUnrelatedOptions(t *testing.T) {
	options := []MedicalProcedureOptionGraphItem{
		{Procedure: emedical_procedure.Surgery{Name: "디곡신"}, ProcedureLabel: "디곡신", TreatOptionPriority: 1},
		{
			Procedure:           emedical_procedure.Surgery{Name: "아트로핀"},
			ProcedureLabel:      "아트로핀",
			TreatOptionPriority: 2,
			Indications: []MedicalProcedureOptionIndication{{
				Status:   estatus.HaveDisease{Disease: edisease.EDisease("동서맥")},
				Label:    "동서맥",
				Observed: true,
			}},
		},
		{
			Procedure:           emedical_procedure.Surgery{Name: "아데노신"},
			ProcedureLabel:      "아데노신",
			TreatOptionPriority: 3,
			Indications: []MedicalProcedureOptionIndication{{
				Status: estatus.HaveDisease{Disease: edisease.EDisease("빈맥")},
				Label:  "빈맥",
				Cause:  true,
			}},
		},
	}
	selected, found := SelectMedicalProcedureOption(options, "치료는?")
	if !found || selected.ProcedureLabel != "아트로핀" {
		t.Fatalf("selected = %#v found=%v", selected, found)
	}
}
