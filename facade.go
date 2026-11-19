package morbes_pipeline

import (
	"encoding/json"
	"fmt"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
	clog "github.com/G226Medi03/log-sqlite"

	"github.com/G226Medi03/MedicalEntityParser/llmdto"
	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	label "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/patient"
	solve "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	pipeline "github.com/G226Medi03/MorbesPipeline/internal/pipeline/orchestration"
	persistence "github.com/G226Medi03/MorbesPipeline/internal/store"
	"github.com/G226Medi03/MorbesPipeline/internal/support/logconfig"
	cause "github.com/G226Medi03/MorbesPipeline/internal/tracing"
)

type (
	IEAction          = repo.IEAction
	IEStatus          = repo.IEStatus
	IMedicalProcedure = repo.IMedicalProcedure

	Cleaning                = repo.Cleaning
	DiagnosticTest          = repo.DiagnosticTest
	Education               = repo.Education
	EOtherSymptom           = repo.EOtherSymptom
	HaveSymptom             = repo.HaveSymptom
	Infuse                  = repo.Infuse
	Instruction             = repo.Instruction
	Surgery                 = repo.Surgery
	Transplant              = repo.Transplant
	UnknownActor            = repo.UnknownActor
	UnknownMedicalProcedure = repo.UnknownMedicalProcedure
)

type (
	CaseID        = solve.CaseID
	QuestionType  = solve.QuestionType
	QuestionInput = solve.QuestionInput
	QuestionStore = solve.QuestionStore
	AnswerCheck   = solve.AnswerCheck

	MedicalProcedureOption   = repo.MedicalProcedureOption
	MedicalProcedureOptionID = int64
	PatientDescription       = patient.PatientDescription
	Trait                    = trait.Trait
)

type (
	Service      = pipeline.Service
	Dependencies = pipeline.Dependencies

	Case       = solve.Case
	Step       = solve.Step
	LLMCall    = model.LLMCall
	ReplayMode = solve.ReplayMode

	ResolvedCaseExtraction = pipeline.ResolvedCaseExtraction

	MedicalProcedureOptionGraph      = solve.MedicalProcedureOptionGraph
	MedicalProcedureOptionGraphItem  = solve.MedicalProcedureOptionGraphItem
	MedicalProcedureOptionOutcome    = solve.MedicalProcedureOptionOutcome
	MedicalProcedureOptionIndication = solve.MedicalProcedureOptionIndication
)

type (
	Store          = persistence.StoreContainer
	StoreContainer = persistence.StoreContainer

	ProblemResultFilter = persistence.ProblemResultFilter
	ProblemResultRow    = persistence.ProblemResultRow

	Actor                  = repo.Actor
	Action                 = repo.Action
	Status                 = repo.Status
	Procedure              = repo.Procedure
	ExamQuestionDataCounts = persistence.ExamQuestionDataCounts
	ExamQuestionCaseStats  = persistence.ExamQuestionCaseStats
)

type (
	StatusID           = int64
	ActionID           = int64
	ActorID            = int64
	ProcedureID        = int64
	DeviceID           = int64
	ExamQuestionID     = int64
	LLMProblemResultID = int64
	LLMCallID          = int64
	StatusEntityID     = int64

	Attr    = clog.Attr
	Logger  = clog.Logger
	Options = clog.Config
	Logging = logconfig.Logging

	CauseTraceResult = cause.CauseTraceResult
)

const (
	QuestionTypeCause        = solve.QuestionTypeCause
	QuestionTypeStatus       = solve.QuestionTypeStatus
	QuestionTypeProcedure    = solve.QuestionTypeProcedure
	QuestionTypePurpose      = solve.QuestionTypePurpose
	QuestionTypeMechanism    = solve.QuestionTypeMechanism
	QuestionTypeDeviceAction = solve.QuestionTypeDeviceAction

	IntentEducation       = solve.IntentEducation
	IntentNursing         = solve.IntentNursing
	IntentStatementChoice = solve.IntentStatementChoice
	IntentEmergencyCare   = solve.IntentEmergencyCare

	TraitAcute      = trait.TraitAcute
	TraitDefinitive = trait.TraitDefinitive

	NodeExplain = pipeline.NodeExplain
	NodeCause   = pipeline.NodeCause

	ReplayModeLLM  = solve.ReplayModeLLM
	ReplayModeCode = solve.ReplayModeCode
)

var (
	InferQuestionPolicy   = solve.InferQuestionPolicy
	NormalizeQuestionType = solve.NormalizeQuestionType

	New        = persistence.New
	NewWithDBs = persistence.NewWithDBs

	Open                = logconfig.Open
	OptionsFromEnv      = logconfig.OptionsFromEnv
	AttrOf              = clog.AttrOf
	SetDefault          = clog.SetDefault
	InstallStdlogBridge = clog.InstallStdlogBridge

	SourceID = source.NewSourceId
	RespID   = source.NewRespId

	ProcedureLabel = label.ProcedureLabel

	MergeTraits       = trait.MergeTraits
	PriorityTraits    = trait.PriorityTraits
	DescriptiveTraits = trait.DescriptiveTraits
	ComparePriority   = trait.ComparePriority
	CategoryRank      = trait.CategoryRank
	PriorityReason    = trait.PriorityReason

	NewService = pipeline.New_Service
)

func Info(msg string, attrs ...Attr)  { clog.Info(msg, attrs...) }
func Warn(msg string, attrs ...Attr)  { clog.Warn(msg, attrs...) }
func Error(msg string, attrs ...Attr) { clog.Error(msg, attrs...) }
func Fatal(msg string, attrs ...Attr) { clog.Fatal(msg, attrs...) }

func PipelineStages() []string {
	return pipeline.PipelineStages()
}

func StageSequence(stage string) int {
	return pipeline.StageSequence(stage)
}

func StepSeq(node string) int {
	return pipeline.StepSeq(node)
}

func PromptStages() []string {
	return pipeline.PromptStages()
}

func PromptPlaceholders(stage string) []string {
	return pipeline.PromptPlaceholders(stage)
}

func SelectMedicalProcedureOption(options []MedicalProcedureOptionGraphItem, problem string) (MedicalProcedureOptionGraphItem, bool) {
	return pipeline.SelectMedicalProcedureOption(options, problem)
}

// IsStatusKind reports whether kind is a persisted status variant.
func IsStatusKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "increased_index", "decreased_index", "have_disease", "have_symptom",
		"damaged", "inflammation", "received_medical_procedure":
		return true
	}
	return false
}

// StatusLabel returns a label for a kind-tagged status payload.
func StatusLabel(kind, data string) string {
	raw, err := withJSONKind(kind, data)
	if err != nil {
		return strings.TrimSpace(data)
	}
	var statusDTO llmdto.Status
	if err := json.Unmarshal(raw, &statusDTO); err != nil {
		return strings.TrimSpace(data)
	}
	value, err := llmparser.StatusFromDTO(statusDTO)
	if err != nil {
		return strings.TrimSpace(data)
	}
	return label.StatusLabelValue(value)
}

// NodeLabel returns a label for a kind-tagged entity payload.
func NodeLabel(kind, data string) string {
	if IsStatusKind(kind) {
		return StatusLabel(kind, data)
	}
	raw, err := withJSONKind(kind, data)
	if err == nil {
		var actionDTO llmdto.Action
		if json.Unmarshal(raw, &actionDTO) == nil {
			if value, err := llmparser.ActionFromDTO(actionDTO); err == nil {
				return label.ActionLabelValue(value)
			}
		}
		var procedureDTO llmdto.Procedure
		if json.Unmarshal(raw, &procedureDTO) == nil {
			if value, err := llmparser.ProcedureFromDTO(&procedureDTO); err == nil {
				return label.ProcedureLabel(value)
			}
		}
	}
	return strings.TrimSpace(kind) + ":" + strings.TrimSpace(data)
}

// ExtractJSONElements extracts JSON array elements from an LLM response.
func ExtractJSONElements(rawText string) ([]json.RawMessage, error) {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return nil, fmt.Errorf("응답에서 JSON 을 찾지 못했습니다")
	}
	if start := strings.Index(trimmed, "["); start != -1 {
		if end := strings.LastIndex(trimmed, "]"); end > start {
			var array []json.RawMessage
			if err := json.Unmarshal([]byte(trimmed[start:end+1]), &array); err == nil {
				return array, nil
			}
		}
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &object); err != nil {
		return nil, fmt.Errorf("응답에서 JSON 을 찾지 못했습니다: %w", err)
	}
	for _, key := range []string{"nodes", "chains", "items", "results"} {
		raw, ok := object[key]
		if !ok {
			continue
		}
		var array []json.RawMessage
		if err := json.Unmarshal(raw, &array); err == nil {
			return array, nil
		}
	}
	return []json.RawMessage{json.RawMessage(trimmed)}, nil
}

func withJSONKind(kind, data string) ([]byte, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil, err
	}
	encodedKind, err := json.Marshal(kind)
	if err != nil {
		return nil, err
	}
	payload["kind"] = encodedKind
	return json.Marshal(payload)
}
