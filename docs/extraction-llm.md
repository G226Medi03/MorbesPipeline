# 추출과 LLM 단계

## 소유권

프롬프트 조회, 렌더링, LLM 호출, 캐시, replay는
`github.com/G226Medi03/MorbesLLMService`가 소유한다.

MorbesPipeline은 다음만 담당한다.

1. case 데이터와 status/action map 구성
2. `MorbesLLMService` typed extraction 메서드 호출
3. 결과를 pipeline draft로 변환
4. 호출 기록 저장

파이프라인은 더 이상 LLM 프롬프트를 직접 렌더링하지 않는다.
`RenderPrompt`/`BuildActorPrompt`/`BuildRelationPrompt`/`AskRole` 경로는 제거됐다.

## 프롬프트 stage

프롬프트는 `solve_prompt_versions` 테이블에 있고,
`persistence.Store.Prompt(stage)`가 `MorbesLLMService`의 `PromptStore`로 전달된다.

| stage | placeholder | 사용 |
| --- | --- | --- |
| `explain` | `problem`, `answer` | 해설 생성 |
| `question_type` | `problem`, `answer`, `explanation` | 질문 유형 분류 |
| `actor` | `problem`, `answer`, `explanation` | actor 추출 |
| `status` | `problem`, `answer`, `explanation` | 상태 추출 |
| `action` | `problem`, `answer`, `explanation` | 행동 추출 |
| `patient_description` | `problem`, `answer`, `explanation`, `status_json` | 환자 서술 추출 |
| `medical_procedure_option` | `problem`, `answer`, `explanation`, `status_json`, `patient_description_json` | 치료 선택지 추출 |
| `relation` | `problem`, `answer`, `explanation`, `status_json`, `action_json` | 관계 추출 |

`other_names`, `actor_action`, `case_extraction`, `situation`은 현재 실행 경로에서
사용하지 않는다.

### 누락 처리

`Service.load_prompts`는 pipeline 상태에 stage별 template/version을 캐시한다.
실제 LLM 호출은 `MorbesLLMService`가 같은 store에서 다시 조회한다.

- `actor`, `patient_description`: template이 비어 있으면 단계 skip
- `relation`: template이 비어 있으면 실패
- `question_type`: template이 비어 있으면 LLM 호출 없이 규칙 기반 추론만 사용
- `status`, `action`, `medical_procedure_option`: 서비스 조회 실패 시 오류

## 문제 본문과 선택지

`solve/problem_input.go`:

- `problem_stem(problem)`: 줄 단위로 선택지를 제거한 본문
  - 정규식: `^\s*(?:([0-9]+)[.)]|([①-⑳]))\s*(.+?)\s*$`
  - `1.`, `1)`, `①`~`⑳` 형태를 선택지로 본다.
- `problem_choices(problem)`: `[]{Index, Text}`로 선택지 파싱 (테스트/유틸)

소견 추출 단계(status/action/relation/patient_description/question_type)는
`problem_stem`을 넘긴다.
선택지 추출 단계(medical_procedure_option)는 원문 문제를 넘긴다.

정답/해설 사용 금지, 선택지 오염 방지 같은 지시문은 이제 파이프라인이 아니라
프롬프트 template이 소유한다.

## 질문 유형 분류

`Service.run_question_type`은 `MorbesLLMService.ExtractQuestionType`을 호출한다.

1. `problem_stem`만 전달하고 answer/explanation은 빈 값으로 둔다.
2. 서비스가 `question_type` template을 조회/렌더링하고 JSON을 파싱한다.
3. `QuestionTypeResult.QuestionType`을 `NormalizeQuestionType`으로 정규화한다.
4. `InferQuestionPolicy(problem)` 결과와 병합한다.
   - 추론된 유형이 `device_action`이면 항상 그것을 사용
   - LLM이 `status`인데 추론이 `procedure`면 `procedure`
   - 추론이 `mechanism`인데 LLM이 mechanism이 아니면 `mechanism`
   - LLM 결과가 비어 있으면 추론값 사용
   - 그래도 비어 있으면 `procedure`
5. `SetQuestionType`으로 case에 저장하고 LLM 호출을 기록한다.

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

LLM 호출과 캐시 정책은 `MorbesLLMService`가 소유한다.
Pipeline은 `pipeLineCore` adapter의 `CacheMode`를 사용한다.

| 모드 | 동작 |
| --- | --- |
| `core.CacheUse` | `FindLlmResponse`로 캐시를 먼저 조회하고, miss면 호출 후 `GetOrInsertLlmResponse` |
| `core.CacheRefresh` | 캐시를 읽지 않고 호출한 뒤 `RefreshLlmResponse`로 기존 행 갱신 |
| `core.CacheBypass` | 캐시를 읽지 않고, 호출한 뒤 `GetOrInsertLlmResponse`를 사용한다(code replay용) |

캐시 키는 `(model, prompt, refered_source_id)`이다. 참조 source는
`ExplanationSourceID`가 있으면 그것, 없으면 `Case.SourceID`다.

## 추출 단계

`extractionPipeline`의 각 LLM 단계는 다음 패턴을 따른다.

1. template 존재 여부 확인(단계별 skip/실패)
2. case 데이터로 typed 파라미터 구성
3. `MorbesLLMService` typed extraction 메서드 호출
4. typed 결과를 pipeline draft로 변환
5. `record_v2_extraction_call`로 호출 기록

### actor

- 함수: `run_v2_actor_extraction`
- 호출: `Service.ExtractActors(ctx, problem, "", "", sourceID, mode)`
- 결과: `[]morbesllm.ActorDraft{Key: "o1", Kind, Name, SourceID}`
- actor kind 보정과 draft 변환은 `MorbesLLMService`가 담당한다.
- 템플릿이 비어 있으면 `nil`을 반환한다(오류 아님).

### status

- 함수: `run_v2_status_extraction`
- 호출: `Service.ExtractStatuses(ctx, problemStem, "", "", sourceID, mode)`
- 결과: `[]StatusDraft{Key, Kind, Value(json.RawMessage), SourceID}`
- typed status를 `entityjson.EncodeStatus`로 직렬화해 draft value로 보관한다.
- key는 LLM이 준 key를 유지하고 사전순으로 정렬한다.

### action

- 함수: `run_v2_action_extraction`
- 호출: `Service.ExtractActions(ctx, problemStem, "", "", sourceID, mode)`
- typed action을 `entityjson.EncodeAction`으로 직렬화한다.
- 결과: `[]ActionDraft{Key, Variant, Value, SourceID}`

### patient description

- 함수: `run_v2_patient_description_extraction`
- `looks_like_patient_description(problem)`가 참일 때만 실행
  - "진술", "서술", "표현", "이해한", "말한", "호소" 또는 인용부호 포함
- 입력: `map[string]repo.IEStatus`
- 호출:
  `Service.ExtractPatientDescription(ctx, problemStem, "", "", statuses, sourceID, mode)`
- 서비스가 `status_json`을 만들고 `patientDescriptions`를 파싱한다.
- 각 description 항목은 `status_draft_from_raw`로 `StatusDraft`가 된다.
  fallback key는 `pN_M`.
- description/knowledge가 모두 비면 버린다.

### medical procedure option

- 함수: `run_v2_medical_procedure_option_extraction`
- 입력: `map[string]repo.IEStatus`, `[]morbesllm.PatientDescription`
- 호출:
  `Service.ExtractMedicalProcedureOption(ctx, problem, "", "", statuses, patients, sourceID, mode)`
- 서비스가 `status_json`, `patient_description_json`을 만들고 옵션을 파싱한다.
- `traits`는 `trait.NormalizeTraits`로 정규화한다.
- `medical_procedure`가 없거나 `null`이면 버린다.
- `indications`, `expected_effects`는 status key 목록이다.

### relation

- 함수: `run_v2_relation_extraction`
- 입력: `map[string]repo.IEStatus`, `map[string]repo.IEAction`
- 호출:
  `Service.ExtractRelations(ctx, problemStem, "", "", statuses, actions, explanationSourceID, mode)`
- 서비스가 relation prompt JSON을 만들고 `DecodeRelations`로 파싱한다.
- parser 결과는 `[causeKey, actionKey, resultKey, actionKey, ...]` 형태의 경로다.
  인접한 세 개를 묶어 `RelationDraft{CauseStatusKey, ActionKey, ResultStatusKey}`로
  변환한다.
- relation 템플릿이 비어 있으면 실패한다.

## 호출 기록

`record_v2_extraction_call`은 다음을 `solve_llm_calls`에 남긴다.

- `Seq`: 단계별 상수(12/15/20/30/36/37/40)
- `Stage`, `Role`
- `PromptVersion`: 서비스 결과의 `PromptVersion`
- `Prompt`: `LlmResponse.Prompt`
- `RawOutput`: `LlmResponse.OutputText()`
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
