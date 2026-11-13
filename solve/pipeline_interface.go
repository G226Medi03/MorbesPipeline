package csolve

import (
	"context"

	"github.com/G226Medi03/MorbesPipeline/solve/domain"
)

// Pipeline is the application-service-facing contract for running one solve
// case or replaying it.
type Pipeline interface {
	Run(ctx context.Context, id domain.CaseID) error
	RunWithCache(ctx context.Context, id domain.CaseID, useCache bool) error
	PrepareReplay(ctx context.Context, id domain.CaseID, mode ReplayMode, fromNode string) (domain.CaseID, error)
	RunReplay(ctx context.Context, id domain.CaseID, mode ReplayMode, fromNode string, useCache bool) error
}
