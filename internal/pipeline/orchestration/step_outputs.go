package csolve

import (
	"github.com/G226Medi03/MorbesPipeline/internal/pipeline/checkpoint"
	ccause "github.com/G226Medi03/MorbesPipeline/internal/tracing"
)

type caseExtractionStepOutput = checkpoint.CaseExtractionStepOutput
type resolveStepOutput = checkpoint.ResolveStepOutput
type explainStepOutput = checkpoint.ExplainStepOutput

type causeStepOutput struct {
	Observed      []string           `json:"observed"`
	Candidates    []ccause.Candidate `json:"candidates"`
	Missing       []string           `json:"missing"`
	CoverageAdded []string           `json:"coverageAdded,omitempty"`
}
