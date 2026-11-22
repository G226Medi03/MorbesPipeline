package solve

import "strings"

type QuestionType string

const (
	QuestionTypeCause        QuestionType = "cause"
	QuestionTypeStatus       QuestionType = "status"
	QuestionTypeProcedure    QuestionType = "procedure"
	QuestionTypePurpose      QuestionType = "purpose"
	QuestionTypeMechanism    QuestionType = "mechanism"
	QuestionTypeDeviceAction QuestionType = "device_action"
)

func NormalizeQuestionType(value string) QuestionType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "cause", "causes", "why":
		return QuestionTypeCause
	case "status", "symptom", "finding", "result":
		return QuestionTypeStatus
	case "procedure", "intervention", "treatment", "nursing":
		return QuestionTypeProcedure
	case "purpose", "goal", "prevention", "prevent":
		return QuestionTypePurpose
	case "mechanism", "action", "drug_action", "effect", "mode_of_action":
		return QuestionTypeMechanism
	case "device_action", "device", "button", "equipment_action":
		return QuestionTypeDeviceAction
	default:
		return ""
	}
}

func InferQuestionType(problem, answer string) QuestionType {
	// Solve-time classification uses the problem. The answer parameter is kept
	// for compatibility with older callers, but answer text must not influence
	// type selection.
	_ = answer
	text := strings.ToLower(strings.TrimSpace(problem))
	switch {
	case containsAny(text, "버튼", "장비", "모드", "sync", "charge", "pacer", "defibrillator"):
		return QuestionTypeDeviceAction
	case containsAny(text, "교육 내용", "교육내용", "교육해야", "재교육", "퇴원 교육", "교육 후", "교육으로"):
		return QuestionTypeProcedure
	case containsAny(text, "간호중재", "간호"):
		return QuestionTypeProcedure
	case containsAny(text, "약의 작용", "작용을", "기전", "어떻게 작용", "약효", "효과를"):
		return QuestionTypeMechanism
	case containsAny(text, "목적", "위함", "무엇을 위해", "왜 ", "이유", "예방"):
		return QuestionTypePurpose
	case containsAny(text, "원인", "위험요인", "왜 발생"):
		return QuestionTypeCause
	case containsAny(text, "나타날 수 있는", "나타나는", "증상", "징후", "소견", "사정자료"):
		return QuestionTypeStatus
	case containsAny(text, "중재", "치료", "투여", "시행", "우선", "검사"):
		return QuestionTypeProcedure
	default:
		return ""
	}
}

type QuestionIntent string

const (
	IntentProcedure       QuestionIntent = "procedure"
	IntentEducation       QuestionIntent = "education"
	IntentNursing         QuestionIntent = "nursing"
	IntentDeviceAction    QuestionIntent = "device_action"
	IntentPurpose         QuestionIntent = "purpose"
	IntentMechanism       QuestionIntent = "mechanism"
	IntentKnowledge       QuestionIntent = "knowledge"
	IntentStatementChoice QuestionIntent = "statement_choice"
	IntentEmergencyCare   QuestionIntent = "emergency_care"
)

func InferQuestionIntent(problem string) QuestionIntent {
	text := strings.ToLower(strings.TrimSpace(problem))
	switch {
	case containsAny(text, "버튼", "장비", "모드", "sync", "charge", "pacer"):
		return IntentDeviceAction
	case containsAny(text, "응급처치", "응급 처치", "즉시 처치", "초기 처치", "응급간호", "응급 중재"):
		return IntentEmergencyCare
	case containsAny(text, "교육 내용", "교육내용", "교육해야", "교육할", "교육하는", "교육이 필요한", "교육이 필요", "재교육", "퇴원 교육", "교육 후", "교육으로"):
		return IntentEducation
	case containsAny(text, "설명으로", "설명에 대한", "대한 설명", "설명한 내용", "설명이 옳은"):
		return IntentStatementChoice
	case containsAny(text, "간호중재", "간호"):
		return IntentNursing
	case containsAny(text, "약의 작용", "작용을", "기전", "약효", "효과를"):
		return IntentMechanism
	case containsAny(text, "목적", "이유", "위함", "예방"):
		return IntentPurpose
	default:
		return IntentProcedure
	}
}

func InferQuestionPolarity(problem string) string {
	text := strings.ToLower(strings.TrimSpace(problem))
	if containsAny(text, "옳지 않은", "아닌", "잘못된", "재교육이 필요한", "필요 없는") {
		return "negative"
	}
	return "positive"
}

func containsAny(value string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(value, strings.ToLower(token)) {
			return true
		}
	}
	return false
}
