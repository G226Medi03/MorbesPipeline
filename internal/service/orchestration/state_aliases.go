package csolve

import (
	"github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/state"
)

type pipelineState = state.State

type (
	CaseExtraction          = state.CaseExtraction
	ActorDraft              = state.ActorDraft
	StatusDraft             = state.StatusDraft
	ActionDraft             = state.ActionDraft
	RelationDraft           = state.RelationDraft
	OptionDraft             = state.OptionDraft
	PatientDescriptionDraft = state.PatientDescriptionDraft

	ResolvedCaseExtraction = state.ResolvedCaseExtraction
)

type (
	caseExtractionStepOutput = state.CaseExtractionStepOutput
	resolveStepOutput        = state.ResolveStepOutput
	explainStepOutput        = state.ExplainStepOutput
	causeStepOutput          = state.CauseStepOutput
)
