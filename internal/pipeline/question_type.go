package csolve

import (
	"context"
	"strings"
	"time"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
)

const defaultQuestionTypePrompt = `주어진 의학 문제의 질문 유형을 분류하라.
반드시 다음 다섯 가지 중 하나만 선택하라.

- cause: 원인, 위험요인, 왜 발생했는지를 묻는 문제
- status: 증상, 징후, 검사 결과, 나타날 수 있는 상태를 묻는 문제
- procedure: 우선 간호중재, 치료, 투여, 시행할 처치를 묻는 문제
- purpose: 어떤 활동/중재의 목적, 이유, 예방 효과를 묻는 문제
- mechanism: 약물/중재의 작용, 기전, 약효, 효과를 묻는 문제
- device_action: 장비의 버튼, 모드, 조작을 묻는 문제

규칙:
- 출력은 JSON 객체 하나만 출력하라.
- questionType은 cause, status, procedure, purpose, mechanism, device_action 중 하나다.
- confidence는 high, medium, low 중 하나다.
- reason은 한 문장으로 작성하라.

[문제]
{{problem}}

[정답]
{{answer}}

[해설]
{{explanation}}

출력 형식:
{"questionType":"mechanism","confidence":"high","reason":"..."}`

func (self *Service) run_question_type(ctx context.Context, id domain.CaseID, state *pipelineState) domain.QuestionType {
	template := state.Prompts.templates[PromptQuestionType]
	if strings.TrimSpace(template) == "" {
		template = defaultQuestionTypePrompt
	}
	problem := problem_stem(state.Case.Problem)
	prompt := RenderPrompt(template, map[string]string{
		"problem":     problem,
		"answer":      "",
		"explanation": "",
	})
	prompt += "\n\n[중요] [정답]과 [해설] 없이 [문제]만 보고 질문 유형을 분류하라." +
		"\n약물/중재의 작용, 기전, 약효, 효과를 묻는 문제는 mechanism으로 분류하라." +
		"\n장비의 버튼, 모드, 조작을 묻는 문제는 device_action으로 분류하라." +
		"\n교육 내용/재교육/간호중재 문제는 procedure로 분류하라."
	started := time.Now()
	raw, response, call_err := self.generate_extraction(ctx, state, prompt)
	question_type := domain.QuestionType("")
	if call_err == nil {
		var parsed struct {
			QuestionType string `json:"questionType"`
			Confidence   string `json:"confidence"`
			Reason       string `json:"reason"`
		}
		if decode_err := decode_json_object(raw, &parsed); decode_err == nil {
			question_type = domain.NormalizeQuestionType(parsed.QuestionType)
		}
	}
	policy := domain.InferQuestionPolicy(problem)
	inferred := policy.Type
	switch {
	case inferred == domain.QuestionTypeDeviceAction:
		question_type = inferred
	case question_type == domain.QuestionTypeStatus && inferred == domain.QuestionTypeProcedure:
		question_type = inferred
	case inferred == domain.QuestionTypeMechanism && question_type != domain.QuestionTypeMechanism:
		question_type = inferred
	case question_type == "" && inferred != "":
		question_type = inferred
	}
	if question_type == "" {
		question_type = domain.QuestionTypeProcedure
	}
	state.QuestionType = question_type
	_ = self.store.SetQuestionType(id, string(question_type))
	self.record_v2_extraction_call(
		id, seqQuestionType, "question_type", "classify",
		state.Prompts.versions[PromptQuestionType], prompt, raw, response,
		map[string]any{
			"questionType":    question_type,
			"intent":          policy.Intent,
			"polarity":        policy.Polarity,
			"desiredCategory": policy.DesiredCategory,
		}, started, call_err,
	)
	return question_type
}
