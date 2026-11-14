package csolve

import (
	"encoding/json"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
)

// LoadResolvedCaseExtraction returns the persisted resolve-stage output for a
// case. It lets application code read options and patient descriptions without
// depending on the internal step JSON shape.
func (self *Service) LoadResolvedCaseExtraction(id domain.CaseID) (ResolvedCaseExtraction, bool, error) {
	step, found, err := self.store.GetStep(id, NodeResolve)
	if err != nil {
		return ResolvedCaseExtraction{}, false, err
	}
	if !found || step.State != "succeeded" || strings.TrimSpace(step.Meta.OutputJSON) == "" {
		return ResolvedCaseExtraction{}, false, nil
	}
	var output resolveStepOutput
	if err := json.Unmarshal([]byte(step.Meta.OutputJSON), &output); err != nil {
		return ResolvedCaseExtraction{}, false, err
	}
	return output.Resolved, true, nil
}
