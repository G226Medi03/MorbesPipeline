package csolve

import (
	ccause "github.com/G226Medi03/MorbesPipeline/cause"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/patient"
	csituation "github.com/G226Medi03/MorbesPipeline/situation"
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
	source "github.com/G226Medi03/SourceEntity"
)

type pipelineState struct {
	Case                Case
	Prompts             promptSet
	QuestionType        domain.QuestionType
	Explanation         string
	Extraction          CaseExtraction
	Resolved            ResolvedCaseExtraction
	Actors              []string
	Devices             []string
	ExplanationSourceID source.SourceId
	Statuses            map[string]entity.Status
	Actions             map[string]entity.Action
	PatientDescriptions []patient.PatientDescription
	Relations           [][]string
	RelationSource      source.SourceRef
	SituationJSON       string
	Situation           csituation.Situation
	SituationSourceID   source.SourceId
	Candidates          []ccause.Candidate
}
