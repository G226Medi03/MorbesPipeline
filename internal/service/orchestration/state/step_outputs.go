package state

import (
	"github.com/G226Medi03/MorbesPipeline/internal/service/checkpoint"
	ccause "github.com/G226Medi03/MorbesPipeline/internal/service/tracing"
)

type CaseExtractionStepOutput = checkpoint.CaseExtractionStepOutput
type ResolveStepOutput = checkpoint.ResolveStepOutput
type ExplainStepOutput = checkpoint.ExplainStepOutput

type CauseStepOutput struct {
	Observed      []string           `json:"observed"`
	Candidates    []ccause.Candidate `json:"candidates"`
	Missing       []string           `json:"missing"`
	CoverageAdded []string           `json:"coverageAdded,omitempty"`
}
