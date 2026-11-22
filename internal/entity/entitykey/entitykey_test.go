package entitykey

import (
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

func TestStatusKey(t *testing.T) {
	status := repo.HaveDisease{Disease: repo.EDisease(" 신장 ")}
	if got, want := Status(status), "have_disease|신장"; got != want {
		t.Fatalf("Status() = %q, want %q", got, want)
	}
}

func TestActionAndProcedureKeys(t *testing.T) {
	procedure := repo.Surgery{Name: " 신장이식 "}
	if got, want := Procedure(procedure), "surgery|신장이식"; got != want {
		t.Fatalf("Procedure() = %q, want %q", got, want)
	}
	action := repo.CauseStatus{Status: repo.HaveDisease{Disease: repo.EDisease("신장")}}
	if got, want := Action(action), "cause_status|have_disease|신장"; got != want {
		t.Fatalf("Action() = %q, want %q", got, want)
	}
}

func TestOptionKeyFormat(t *testing.T) {
	procedure := repo.Surgery{Name: "신장이식"}
	indications := []repo.IEStatus{
		repo.HaveDisease{Disease: repo.EDisease("신장")},
		repo.HaveSymptom{Symptom: repo.EOtherSymptom{Name: "통증"}},
	}
	if got, want := OptionKey(procedure, "treatment", indications),
		"surgery|신장이식|treatment|have_disease|신장|have_symptom|other_symptom|통증"; got != want {
		t.Fatalf("OptionKey() = %q, want %q", got, want)
	}
}
