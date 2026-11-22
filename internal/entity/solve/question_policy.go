package solve

import "strings"

type QuestionPolarity string

const (
	PolarityPositive QuestionPolarity = "positive"
	PolarityNegative QuestionPolarity = "negative"
)

// QuestionPolicy is the single classification result used by the solve
// pipeline and answer evaluation. It keeps broad type, intent, polarity, and
// the preferred procedure category together instead of recomputing them in
// several packages.
type QuestionPolicy struct {
	Type            QuestionType
	Intent          QuestionIntent
	Polarity        QuestionPolarity
	DesiredCategory string
}

func InferQuestionPolicy(problem string) QuestionPolicy {
	return QuestionPolicy{
		Type:            InferQuestionType(problem, ""),
		Intent:          InferQuestionIntent(problem),
		Polarity:        QuestionPolarity(InferQuestionPolarity(problem)),
		DesiredCategory: InferDesiredProcedureCategory(problem),
	}
}

// InferDesiredProcedureCategory returns the category the question asks for.
// It is a ranking hint only and must not be treated as a hard filter.
func InferDesiredProcedureCategory(problem string) string {
	text := strings.ToLower(strings.TrimSpace(problem))
	switch {
	case strings.Contains(text, "교육"):
		return "education"
	case strings.Contains(text, "간호중재"), strings.Contains(text, "간호"):
		return "nursing_intervention"
	case strings.Contains(text, "모니터링"), strings.Contains(text, "관찰"):
		return "monitoring"
	case strings.Contains(text, "검사"), strings.Contains(text, "초음파"), strings.Contains(text, "촬영"):
		return "diagnostic_test"
	case strings.Contains(text, "투여"), strings.Contains(text, "약물"), strings.Contains(text, "처방"):
		return "treatment"
	default:
		return ""
	}
}
