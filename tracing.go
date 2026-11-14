package morbes_pipeline

import (
	cause "github.com/G226Medi03/MorbesPipeline/internal/tracing/cause"
	graph "github.com/G226Medi03/MorbesPipeline/internal/tracing/graph"
)

type (
	Container        = graph.Container
	CauseService     = cause.Service
	Candidate        = cause.Candidate
	TraceNode        = cause.TraceNode
	CauseTraceResult = cause.CauseTraceResult
)

var (
	NewContainer    = graph.New_Container
	NewCauseService = cause.New_ContainerService
	ObservedLabels  = cause.ObservedLabels
)
