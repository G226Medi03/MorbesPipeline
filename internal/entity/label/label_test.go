package label

import (
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

func TestStatusLabelIndex(t *testing.T) {
	tests := []struct {
		value repo.IEStatus
		want  string
	}{
		{repo.DecreasedIndex{Index: repo.EIndex{Name: "담즙 배액량"}}, "담즙 배액량 감소"},
		{repo.IncreasedIndex{Index: repo.EIndex{Name: "AST"}}, "AST 증가"},
		{repo.DecreasedIndex{Index: repo.EIndex{Name: "소변량 감소"}}, "소변량 감소"},
	}
	for _, test := range tests {
		if got := StatusLabelValue(test.value); got != test.want {
			t.Fatalf("StatusLabelValue() = %q, want %q", got, test.want)
		}
	}
}

func TestStatusLabelReceivedMedicalProcedure(t *testing.T) {
	value := repo.ReceivedMedicalProcedure{
		MedicalProcedure: repo.Infuse{
			Material: repo.NewUnknownActor("고용량 스테로이드 정맥 투여"),
			Route:    repo.InfuseRouteIV,
		},
	}
	if got, want := StatusLabelValue(value), "고용량 스테로이드 정맥 투여 받음"; got != want {
		t.Fatalf("StatusLabelValue() = %q, want %q", got, want)
	}
}

func TestStatusLabelReceivedMedicalProcedureWithPeriod(t *testing.T) {
	value := repo.ReceivedMedicalProcedure{
		MedicalProcedure:    repo.Transplant{Target: repo.NewUnknownActor("간")},
		PostProcedurePeriod: "수일에서 수주 이내",
	}
	if got, want := StatusLabelValue(value), "간 이식 받음 (수일에서 수주 이내)"; got != want {
		t.Fatalf("StatusLabelValue() = %q, want %q", got, want)
	}
}

func TestStatusLabelPain(t *testing.T) {
	value := repo.HaveSymptom{
		Symptom: repo.EPain{
			Site: repo.NewUnknownActor("우상복부"),
			Aspects: []repo.IPainAspect{
				repo.Quality{Name: "날카로운"},
			},
		},
	}
	if got, want := StatusLabelValue(value), "우상복부 통증 (날카로운)"; got != want {
		t.Fatalf("StatusLabelValue() = %q, want %q", got, want)
	}
}

func TestStatusLabelOtherSymptom(t *testing.T) {
	value := repo.HaveSymptom{Symptom: repo.EOtherSymptom{Name: "육안적 혈뇨"}}
	if got, want := StatusLabelValue(value), "육안적 혈뇨"; got != want {
		t.Fatalf("StatusLabelValue() = %q, want %q", got, want)
	}
}
