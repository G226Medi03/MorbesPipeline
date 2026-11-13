package csolve

import (
	"encoding/json"
	"testing"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

func coverageTestState(problem, answer, explanation string) *pipelineState {
	return &pipelineState{
		Case: Case{
			Problem: problem,
			Answer:  answer,
		},
		Explanation: explanation,
		Statuses: map[string]entity.Status{
			"s1": entity.NewStatus(
				"received_medical_procedure",
				json.RawMessage(`{"medical_procedure":{"kind":"Cleaning","cleaning":{"device":"S상결장루"}}}`),
				nil,
			),
			"s2": entity.NewStatus(
				"increased_index",
				json.RawMessage(`{"index":{"name":"배변 자극"}}`),
				nil,
			),
		},
	}
}

func TestCoverageObservedStatusesAddsProblemFinding(t *testing.T) {
	state := coverageTestState(
		"20 구불결장루가 있는 환자의 장루를 세척한 뒤 배변 자극이 나타났다. 이 처치의 주된 목적은?",
		"",
		"",
	)
	observed := []entity.Status{state.Statuses["s1"]}
	result, added := coverageObservedStatuses(state, observed)
	if len(result) != 2 || len(added) != 1 {
		t.Fatalf("result=%#v added=%#v", result, added)
	}
	if result[1].Variant != "increased_index" {
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
	observed := []entity.Status{state.Statuses["s1"]}
	result, added := coverageObservedStatuses(state, observed)
	if len(result) != 1 || len(added) != 0 {
		t.Fatalf("answer/explanation leaked into findings: result=%#v added=%#v", result, added)
	}
}
