package domain

import (
	"time"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

type Step struct {
	ID         entity.SolveStepID `json:"id"`
	Node       string             `json:"node"`
	Seq        int                `json:"seq"`
	State      StepState          `json:"state"`
	Attempt    int                `json:"attempt"`
	Error      string             `json:"error,omitempty"`
	StartedAt  time.Time          `json:"startedAt"`
	FinishedAt *time.Time         `json:"finishedAt,omitempty"`
	Meta       SolveStepMeta      `json:"meta,omitempty"`
}
