package state

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/patient"
	situation "github.com/G226Medi03/MorbesPipeline/internal/entity/situation"
	solve "github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/prompt"
	tracing "github.com/G226Medi03/MorbesPipeline/internal/service/tracing"
	source "github.com/G226Medi03/SourceEntity"
)

type Case = solve.Case

type State struct {
	Case                Case
	Prompts             prompt.Set
	QuestionType        solve.QuestionType
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
	RelationSource      source.Source
	SituationJSON       string
	Situation           situation.Situation
	SituationSourceID   source.SourceId
	Candidates          []tracing.Candidate
}
