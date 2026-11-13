package csolve

import "testing"

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
		{ID: 1, ProcedureLabel: "디곡신", TreatOptionPriority: 1},
		{ID: 2, ProcedureLabel: "아트로핀", TreatOptionPriority: 2, Indications: []MedicalProcedureOptionIndication{{StatusID: 10, Label: "동서맥", Observed: true}}},
		{ID: 3, ProcedureLabel: "아데노신", TreatOptionPriority: 3, Indications: []MedicalProcedureOptionIndication{{StatusID: 11, Label: "빈맥", Cause: true}}},
	}
	selected, found := SelectMedicalProcedureOption(options, "치료는?")
	if !found || selected.ID != 2 {
		t.Fatalf("selected = %#v found=%v", selected, found)
	}
}
