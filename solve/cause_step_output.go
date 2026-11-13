package csolve

import ccause "github.com/G226Medi03/MorbesPipeline/cause"

type causeStepOutput struct {
	Observed      []string           `json:"observed"`
	Candidates    []ccause.Candidate `json:"candidates"`
	Missing       []string           `json:"missing"`
	CoverageAdded []string           `json:"coverageAdded,omitempty"`
}
