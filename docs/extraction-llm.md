# 추출과 LLM 단계

## 프롬프트 모델

프롬프트는 `solve_prompt_versions` 테이블에서 stage별로 읽는다.
`solve/domain/prompt.go`가 stage 이름과 필수 placeholder를 정의한다.

| stage | placeholder | 파이프라인 사용 |
| --- | --- | --- |
| `explain` | `problem`, `answer` | 해설 생성 |
| `question_type` | `problem`, `answer`, `explanation` | 질문 유형 분류 |
| `actor` | `problem`, `answer`, `explanation` | actor 추출 |
| `other_names` | `problem`, `answer`, `explanation`, `actor_json` | 현재 미사용 |
| `status` | `problem`, `answer`, `explanation` | 상태 추출 |
| `action` | `problem`, `answer`, `explanation` | 행동 추출 |
| `actor_action` | `problem`, `answer`, `explanation`, `actor_json`, `action_json` | 현재 미사용 |
| `patient_description` | `problem`, `answer`, `explanation`, `status_json` | 환자 서술 추출 |
| `medical_procedure_option` | `problem`, `answer`, `explanation`, `status_json`, `patient_description_json` | 치료 선택지 추출 |
| `relation` | `problem`, `answer`, `explanation`, `status_json`, `action_json` | 관계 추출 |
| `situation` | `problem`, `answer`, `explanation` | stage만 정의, 현재 로드/사용 안 함 |
| `case_extraction` | `problem`, `answer`, `explanation`, `patient_description_eligible` | stage만 정의, 현재 로드/사용 안 함 |

`promptStages`에 실제로 로드되는 것은 `explain`, `question_type`, `actor`,
`status`, `action`, `patient_description`, `medical_procedure_option`, `relation`이다.

### 로딩과 누락 허용

`Service.load_prompts`는 각 stage에 대해 `store.GetPrompt(stage)`를 호출한다.

- 누락을 허용하는 stage: `actor`, `other_names`, `patient_description`, `question_type`
- 그 외(`explain`, `status`, `action`, `medical_procedure_option`, `relation`)는
  프롬프트가 없으면 파이프라인이 실패한다.

### 렌더링

`RenderPrompt`는 `{{key}}`를 값으로 단순 치환하고 앞뒤 공백을 제거한다.
`ValidatePromptTemplate`은 stage별 필수 placeholder가 모두 있는지 검사한다.

## 문제 본문과 선택지

`solve/problem_input.go`:

- `problem_stem(problem)`: 줄 단위로 선택지를 제거한 본문
  - 정규식: `^\s*(?:([0-9]+)[.)]|([①-⑳]))\s*(.+?)\s*$`
  - `1.`, `1)`, `①`~`⑳` 형태를 선택지로 본다.
- `problem_choices(problem)`: `[]{Index, Text}`로 선택지 파싱
- `problem_only_instruction(stage)`: 정답/해설 사용 금지 지시문을 덧붙인다.
  - status/action/relation/patient: 선택지의 치료/약물/검사는 소견으로 추출하지 말라는 지시 추가
  - medical_procedure_option: 모든 선택지를 빠짐없이 후보로 추출하고 정답 번호를
    기준으로 고르지 말라는 지시 추가

소견 추출 단계는 `problem_stem`을 사용하고, 선택지 추출 단계는 원문 문제와
선택지 JSON을 함께 사용한다.

## 질문 유형 분류

`Service.run_question_type`은 LLM과 규칙 기반 정책을 결합한다.

1. DB의 `question_type` 템플릿을 사용한다. 없으면 내장
   `defaultQuestionTypePrompt`를 사용한다.
2. `problem_stem`만 렌더링하고 `[정답]`, `[해설]`은 빈 값으로 둔다.
3. 추가 지시문으로 mechanism/device_action/교육·간호중재 규칙을 보강한다.
4. LLM JSON의 `questionType`을 `NormalizeQuestionType`으로 정규화한다.
5. `InferQuestionPolicy(problem)` 결과와 병합한다.
   - 추론된 유형이 `device_action`이면 항상 그것을 사용
   - LLM이 `status`인데 추론이 `procedure`면 `procedure`
   - 추론이 `mechanism`인데 LLM이 mechanism이 아니면 `mechanism`
   - LLM 결과가 비어 있으면 추론값 사용
   - 그래도 비어 있으면 `procedure`
6. `SetQuestionType`으로 case에 저장하고 LLM 호출을 기록한다.

### 질문 유형

| 값 | 의미 |
| --- | --- |
| `cause` | 원인, 위험요인 |
| `status` | 증상, 징후, 검사 결과, 나타날 수 있는 상태 |
| `procedure` | 우선 간호중재, 치료, 투여, 시행할 처치 |
| `purpose` | 목적, 이유, 예방 효과 |
| `mechanism` | 약물/중재의 작용, 기전, 약효 |
| `device_action` | 장비 버튼, 모드, 조작 |

### 질문 정책

`domain.InferQuestionPolicy`는 다음을 한 번에 계산한다.

| 필드 | 값 |
| --- | --- |
| `Type` | `QuestionType` |
| `Intent` | `procedure`, `education`, `nursing`, `device_action`, `purpose`, `mechanism`, `knowledge`, `statement_choice`, `emergency_care` |
| `Polarity` | `positive`, `negative` |
| `DesiredCategory` | `education`, `nursing_intervention`, `monitoring`, `diagnostic_test`, `treatment`, 빈 값 |

`DesiredCategory`는 선택지 정렬 힌트로만 사용되며 hard filter가 아니다.

## LLM 캐시

`Service.generate_extraction`이 추출 LLM 호출을 담당한다.

| 모드 | 동작 |
| --- | --- |
| `LLMCacheUse` | `FindLlmResponse`로 캐시를 먼저 조회하고, miss면 호출 후 `GetOrInsertLlmResponse` |
| `LLMCacheRefresh` | 캐시를 읽지 않고 호출한 뒤 `RefreshLlmResponse`로 기존 행 갱신 |
| `LLMCacheBypass` | 캐시를 읽지 않고, 호출한 뒤 `GetOrInsertLlmResponse`를 사용한다(code replay용) |

`service.AskLLMWithMode`는 relation 추출에서 같은 모드를 사용한다.
`explain`은 별도로 `GetOrInsertLlmResponse`만 사용한다.

캐시 키는 `(model, prompt, refered_source_id)`이다. 참조 source는
`ExplanationSourceID`가 있으면 그것, 없으면 `Case.SourceID`다.

## 추출 단계

`extractionPipeline`의 각 LLM 단계는 다음 패턴을 따른다.

1. 프롬프트 템플릿 확인(없으면 단계별로 skip 또는 실패)
2. `problem_stem` 기준으로 렌더링
3. `generate_extraction` 호출
4. parser로 draft 생성
5. `record_v2_extraction_call`로 호출 기록

### actor

- 함수: `run_v2_actor_extraction`
- 파서: `MedicalEntityParser/llmparser.ParseActorItems`
- 결과: `[]ActorDraft{Key: "o1", Kind, Name, SourceID}`
- kind가 비면 `parser.ParseActor`로 보정하고, 그래도 비면 `unknown_actor`
- 템플릿이 비어 있으면 `nil`을 반환한다(오류 아님).

### status

- 함수: `run_v2_status_extraction`
- 파서: `llmparser.ParseStatuses`
- 결과: `[]StatusDraft{Key, Kind, Value(json.RawMessage), SourceID}`
- key는 LLM이 준 key를 유지하고 사전순으로 정렬한다.

### action

- 함수: `run_v2_action_extraction`
- 파서: `parser.ParseActions`
- 각 action을 `parser.NormalizeActionValue(kind, value)`로 정규화한 뒤 JSON으로 저장
- 결과: `[]ActionDraft{Key, Variant, Value, SourceID}`

### patient description

- 함수: `run_v2_patient_description_extraction`
- `looks_like_patient_description(problem)`가 참일 때만 실행
  - "진술", "서술", "표현", "이해한", "말한", "호소" 또는 인용부호 포함
- 입력: `status_json`
- 기대 출력:

```json
{
  "patientDescriptions": [
    {
      "choice": "②",
      "description": [
        {"kind": "have_symptom", "name": "육안적 혈뇨"}
      ],
      "knowledge": ["방광경 검사 후 소변에 피가 섞일 수 있음을 알고 있음"]
    }
  ]
}
```

- 각 description 항목은 `status_draft_from_raw`로 `StatusDraft`가 된다.
  fallback key는 `pN_M`.
- description/knowledge가 모두 비면 버린다.

### medical procedure option

- 함수: `run_v2_medical_procedure_option_extraction`
- 입력: `status_json`, `patient_description_json`, 선택지 JSON이 덧붙은 문제
- 기대 출력:

```json
{
  "medical_procedure_options": [
    {
      "id": "m1",
      "category": "treatment",
      "traits": ["first_line", "pharmacologic"],
      "medical_procedure": {"kind": "infuse", "infuse": {"material": {"kind": "unknown_actor", "name": "약물"}, "route": "iv"}},
      "indications": ["s1"],
      "expected_effects": ["s2"]
    }
  ]
}
```

- `traits`는 `trait.NormalizeTraits`로 정규화한다.
- `medical_procedure`가 없거나 `null`이면 버린다.
- `indications`, `expected_effects`는 status key 목록이다.

### relation

- 함수: `run_v2_relation_extraction`
- `Build_RelationPromptFromMaps`로 status/action map을 JSON에 넣어 렌더링
- `service.ExtractRelationsWithPrompt`가 `llmparser.ParseRelations`를 호출
- parser 결과는 `[causeKey, actionKey, resultKey, actionKey, ...]` 형태의 경로다.
  인접한 세 개를 묶어 `RelationDraft{CauseStatusKey, ActionKey, ResultStatusKey}`로
  변환한다.
- relation 템플릿이 비어 있으면 실패한다.

## 호출 기록

`record_v2_extraction_call`은 다음을 `solve_llm_calls`에 남긴다.

- `Seq`: 단계별 상수(12/15/20/30/36/37/40)
- `Stage`, `Role`
- `PromptVersion`
- `Prompt`
- `RawOutput`
- `ParsedOutput` (draft 목록 JSON)
- `SourceID`, `RespID`
- `State`, `Error`, `DurationMS`

`record_v2_relation_call`은 relation 경로를 `ParsedOutput`으로 저장하고,
응답이 있으면 `raw_output`, `source_id`, `resp_id`를 채운다.

## source 전파

- `extraction_refered_source`: explain source → case source
- `extraction_source_id`: LLM 응답 source → 위 참조 source
- draft의 `SourceID`는 resolve에서 entity source로 사용된다.
- source가 없는 draft/entity는 저장 시 검증에서 거부될 수 있다.
