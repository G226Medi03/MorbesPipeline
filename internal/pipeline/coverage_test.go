package csolve

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"testing"
)

func coverageTestState(problem, answer, explanation string) *pipelineState {
	return &pipelineState{
		Case:        Case{Problem: problem, Answer: answer},
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
	state := coverageTestState(
		"20 구불결장루가 있는 환자의 장루를 세척한 뒤 배변 자극이 나타났다. 이 처치의 주된 목적은?",
		"",
		"",
	)
	observed := []repo.IEStatus{state.Statuses["s1"]}
	result, added := coverageObservedStatuses(state, observed)
	if len(result) != 2 || len(added) != 1 {
		t.Fatalf("result=%#v added=%#v", result, added)
	}
	if result[1].Kind() != "increased_index" {
		t.Fatalf("added status = %#v", result[1])
	}
	if added[0] != "배변 자극 증가" {
		t.Fatalf("added label = %#v", added)
	}
}

func TestCoverageObservedStatusesIgnoresAnswerAndExplanation(t *testing.T) {
	state := coverageTestState(
		"20 구불결장루가 있는 환자의 장루를 세척하는 주된 목적은?",
		"② 배변 자극",
		"장루 세척은 배변 자극을 목적으로 한다.",
	)
	observed := []repo.IEStatus{state.Statuses["s1"]}
	result, added := coverageObservedStatuses(state, observed)
	if len(result) != 1 || len(added) != 0 {
		t.Fatalf("answer/explanation leaked into findings: result=%#v added=%#v", result, added)
	}
}
