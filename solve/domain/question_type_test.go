package domain

import "testing"

func TestNormalizeQuestionType(t *testing.T) {
	cases := map[string]QuestionType{
		"cause":         QuestionTypeCause,
		"STATUS":        QuestionTypeStatus,
		"procedure":     QuestionTypeProcedure,
		"purpose":       QuestionTypePurpose,
		"mechanism":     QuestionTypeMechanism,
		"device_action": QuestionTypeDeviceAction,
		"unknown":       "",
	}
	for input, want := range cases {
		if got := NormalizeQuestionType(input); got != want {
			t.Fatalf("NormalizeQuestionType(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestInferQuestionPolicy(t *testing.T) {
	policy := InferQuestionPolicy("퇴원 교육 후 재교육이 필요한 반응은?")
	if policy.Type != QuestionTypeProcedure || policy.Intent != IntentEducation || policy.Polarity != PolarityNegative {
		t.Fatalf("policy = %#v", policy)
	}
	if policy.DesiredCategory != "education" {
		t.Fatalf("desired category = %q", policy.DesiredCategory)
	}
}

func TestInferQuestionIntentAndPolarity(t *testing.T) {
	if got := InferQuestionIntent("퇴원 교육 후 재교육이 필요한 반응은?"); got != IntentEducation {
		t.Fatalf("intent = %q", got)
	}
	if got := InferQuestionPolarity("퇴원 교육 후 재교육이 필요한 반응은?"); got != "negative" {
		t.Fatalf("polarity = %q", got)
	}
	if got := InferQuestionIntent("수혈 시 간호중재로 옳은 것은?"); got != IntentNursing {
		t.Fatalf("intent = %q", got)
	}
	if got := InferQuestionIntent("항암화학요법을 받는 환자에게 교육할 내용은?"); got != IntentEducation {
		t.Fatalf("intent = %q", got)
	}
	if got := InferQuestionIntent("순음청력검사에 대한 설명으로 옳은 것은?"); got != IntentStatementChoice {
		t.Fatalf("intent = %q", got)
	}
	if got := InferQuestionIntent("피부가 뜨거워지며 갑자기 쓰러진 대상자를 위한 응급처치는?"); got != IntentEmergencyCare {
		t.Fatalf("intent = %q", got)
	}
}

func TestInferQuestionType(t *testing.T) {
	cases := []struct {
		name    string
		problem string
		answer  string
		want    QuestionType
	}{
		{"purpose", "이 중재의 목적은 무엇인가?", "두통 예방", QuestionTypePurpose},
		{"status", "수술 후 나타날 수 있는 것은?", "체온 증가", QuestionTypeStatus},
		{"cause", "이 증상의 원인은?", "감염", QuestionTypeCause},
		{"mechanism", "시메티딘의 약의 작용은?", "위산 분비 억제", QuestionTypeMechanism},
		{"procedure", "가장 우선할 간호중재는?", "침상 안정", QuestionTypeProcedure},
		{"education", "퇴원 교육 후 재교육이 필요한 반응은?", "②", QuestionTypeProcedure},
		{"device", "심장율동전환 시 사용하는 버튼은?", "③", QuestionTypeDeviceAction},
		{"unknown", "다음 중 옳은 것은?", "보기", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := InferQuestionType(tc.problem, tc.answer); got != tc.want {
				t.Fatalf("InferQuestionType() = %q, want %q", got, tc.want)
			}
		})
	}
}
