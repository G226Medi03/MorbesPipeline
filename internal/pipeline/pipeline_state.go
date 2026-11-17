package csolve

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/patient"
	csituation "github.com/G226Medi03/MorbesPipeline/internal/domain/situation"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	ccause "github.com/G226Medi03/MorbesPipeline/internal/tracing/cause"
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
	Statuses            map[string]repo.IEStatus
	Actions             map[string]repo.IEAction
	PatientDescriptions []patient.PatientDescription
	Relations           [][]string
	RelationSource      source.SourceRef
	SituationJSON       string
	Situation           csituation.Situation
	SituationSourceID   source.SourceId
	Candidates          []ccause.Candidate
}
