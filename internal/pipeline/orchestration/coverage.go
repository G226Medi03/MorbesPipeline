package csolve

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/pipeline/orchestration/coverage"
)

func coverageObservedStatuses(state *pipelineState, observed []repo.IEStatus) ([]repo.IEStatus, []string) {
	return coverage.ObservedStatuses(state, observed)
}
