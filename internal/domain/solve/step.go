package domain

import (
	"time"
)

type Step struct {
	ID         int64 `json:"id"`
	Node       string          `json:"node"`
	Seq        int             `json:"seq"`
	State      StepState       `json:"state"`
	Attempt    int             `json:"attempt"`
	Error      string          `json:"error,omitempty"`
	StartedAt  time.Time       `json:"startedAt"`
	FinishedAt *time.Time      `json:"finishedAt,omitempty"`
	Meta       SolveStepMeta   `json:"meta,omitempty"`
}

type StepState string

type SolveStepMeta struct {
	CaseID        CaseID        `json:"caseId"`
	PromptVersion int           `json:"promptVersion"`
	OutputJSON    string        `json:"outputJson"`
	CallID        int64 `json:"callId"`
}
