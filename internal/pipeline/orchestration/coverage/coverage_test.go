package coverage

import (
	"testing"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/pipeline/orchestration/state"
)

func coverageTestState(problem, answer, explanation string) *state.State {
	return &state.State{
		Case:        state.Case{Problem: problem, Answer: answer},
		Explanation: explanation,
		Statuses: map[string]repo.IEStatus{
			"s1": repo.ReceivedMedicalProcedure{
				MedicalProcedure: repo.Cleaning{Target: repo.NewUnknownActor("S상결장루")},
			},
			"s2": repo.IncreasedIndex{Index: repo.EIndex{Name: "배변 자극"}},
		},
	}
}

func TestCoverageObservedStatusesAddsProblemFinding(t *testing.T) {
	st := coverageTestState(
		"20 구불결장루가 있는 환자의 장루를 세척한 뒤 배변 자극이 나타났다. 이 처치의 주된 목적은?",
		"",
		"",
	)
	observed := []repo.IEStatus{st.Statuses["s1"]}
	result, added := ObservedStatuses(st, observed)
	if len(result) != 2 || len(added) != 1 {
		t.Fatalf("result=%#v added=%#v", result, added)
	}
}

func TestCoverageObservedStatusesIgnoresAnswerAndExplanation(t *testing.T) {
	st := coverageTestState(
		"20 구불결장루가 있는 환자의 장루를 세척하는 주된 목적은?",
		"② 배변 자극",
		"장루 세척은 배변 자극을 목적으로 한다.",
	)
	observed := []repo.IEStatus{st.Statuses["s1"]}
	result, added := ObservedStatuses(st, observed)
	if len(result) != 1 || len(added) != 0 {
		t.Fatalf("answer/explanation leaked into findings: result=%#v added=%#v", result, added)
	}
}
