package csolve

import (
	"encoding/json"
	"strings"

	solve "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/pipeline/checkpoint"
)

type ResolvedCaseExtraction = checkpoint.ResolvedCaseExtraction

func (self *Service) LoadResolvedCaseExtraction(id solve.CaseID) (ResolvedCaseExtraction, bool, error) {
	step, found, err := self.store.GetStep(id, NodeResolve)
	if err != nil {
		return ResolvedCaseExtraction{}, false, err
	}
	if !found || step.State != "succeeded" || strings.TrimSpace(step.Meta.OutputJSON) == "" {
		return ResolvedCaseExtraction{}, false, nil
	}
	var output checkpoint.ResolveStepOutput
	if err := json.Unmarshal([]byte(step.Meta.OutputJSON), &output); err != nil {
		return ResolvedCaseExtraction{}, false, err
	}
	return output.Resolved, true, nil
}
