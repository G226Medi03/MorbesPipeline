package morbes_pipeline

import (
	label "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	"github.com/G226Medi03/MorbesPipeline/internal/support/logs"
	"github.com/G226Medi03/MorbesPipeline/internal/support/sliceutil"
	"github.com/G226Medi03/MorbesPipeline/internal/support/sourceutil"
)

type (
	StatusID           = ids.StatusID
	ActionID           = ids.ActionID
	ActorID            = ids.ActorID
	ProcedureID        = ids.ProcedureID
	DeviceID           = ids.DeviceID
	ExamQuestionID     = ids.ExamQuestionID
	LLMProblemResultID = ids.LLMProblemResultID
	SolveStepID        = ids.SolveStepID
	LLMCallID          = ids.LLMCallID
	StatusEntityID     = ids.StatusEntityID

	Attr             = logs.Attr
	Entry            = logs.Entry
	ISink            = logs.ISink
	Level            = logs.Level
	Logger           = logs.Logger
	Options          = logs.Options
	Logging          = logs.Logging
	PreprocessChange = logs.PreprocessChange
)

const (
	LevelDebug = logs.LevelDebug
	LevelInfo  = logs.LevelInfo
	LevelWarn  = logs.LevelWarn
	LevelError = logs.LevelError
)

var (
	Open              = logs.Open
	OptionsFromEnv    = logs.OptionsFromEnv
	AttrOf            = logs.AttrOf
	SetDefault        = logs.SetDefault
	LLMError          = logs.LLMError
	PreprocessChanges = logs.PreprocessChanges

	SourceID = sourceutil.SourceID
	RespID   = sourceutil.RespID

	StatusLabel    = label.StatusLabel
	NodeLabel      = label.NodeLabel
	ProcedureLabel = label.ProcedureLabel
	IsStatusKind   = label.IsStatusKind
)

func Debug(msg string, attrs ...Attr) { logs.Debug(msg, attrs...) }
func Info(msg string, attrs ...Attr)  { logs.Info(msg, attrs...) }
func Warn(msg string, attrs ...Attr)  { logs.Warn(msg, attrs...) }
func Error(msg string, attrs ...Attr) { logs.Error(msg, attrs...) }
func Fatal(msg string, attrs ...Attr) { logs.Fatal(msg, attrs...) }

func AppendUnique[T comparable](values []T, add ...T) []T {
	return sliceutil.AppendUnique(values, add...)
}
