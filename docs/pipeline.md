# 파이프라인 실행 구조

## 최상위 `SolvePipeline`

`Service.RunWithCache` → `run_v2_from` → `runPipelineNodes`가 아래 노드를 실행한다.
정의는 `solve/pipeline_nodes.go`의 `newSolveNodePipeline`에 있다.

| 순서 | 노드 | stage | kind | 체크포인트 | 주요 파일 |
| --- | --- | --- | --- | --- | --- |
| 1 | `CaseLoadNode` | `prepare` | `KindPrepare` | 없음 | `pipeline_nodes.go` |
| 2 | `CaseExplainNode` | `explain` (10) | `KindLLM` | `explainStepOutput` | `pipeline.go` |
| 3 | `LLMCaseExtractionNode` | `case_extraction` (20) | `KindLLM` | `caseExtractionStepOutput` | `pipeline_v2.go`, `extraction_pipeline.go` |
| 4 | `DraftResolveNode` | `resolve` (30) | `KindResolve` | `resolveStepOutput` | `pipeline_v2.go` |
| 5 | `GraphCauseTraceNode` | `cause_trace` (60) | `KindGraph` | `causeStepOutput` | `pipeline.go` |

`SolvePipeline`은 `PipeLineCore`의 `core.Pipeline`이다. 상태는 `core.State`에
`solve_case_id`, `solve_from_node`, `solve_state` 키로 전달된다.

## 노드별 동작

### 1. `CaseLoadNode`

`caseLoadNode`가 실행된다.

1. `store.Get(caseID)`로 case를 읽는다.
2. `exam_question_id > 0`이면 `QuestionStore.GetQuestion`으로 문제/정답/해설을 덮어쓴다.
3. `load_prompts()`로 DB의 프롬프트와 버전을 읽는다.
4. extraction LLM이 없으면 실패한다.
5. `fromNode == ""`이면 `ResetSteps(caseID)`, 아니면:
   - `domain.IsStage(fromNode)` 검증
   - `MarkDownstreamStale(caseID, fromNode)`
6. `SetRunning(caseID)` 후 `pipelineState{Case, Prompts}`를 state에 넣는다.

### 2. `CaseExplainNode`

1. `fromNode`의 stage 순서가 explain보다 뒤면 `load_explain`으로 체크포인트를 읽는다.
2. `case.Explanation`이 이미 있으면 `record_provided_explanation`:
   - `source_id`는 `case.SourceID`, 없으면 `case.SituationSourceID`
   - explain 단계를 `succeeded`로 남긴다.
3. 없으면 `run_explain_node`:
   - `solver.Generate` 호출
   - JSON이면 `explanation` 필드, 아니면 원문을 해설로 사용
   - `GetOrInsertLlmResponse(model, "solve:<caseID>:explanation", ...)`로 source를 만든다.
   - `SetSource` 후 `explainStepOutput`을 저장한다.

> explain은 `AskLLM` 캐시를 읽지 않고 `GetOrInsertLlmResponse`로 결과만 남긴다.
> 공유 응답 캐시를 사용하는 것은 extraction 단계다.

### 3. `LLMCaseExtractionNode`

`run_case_extraction_node`가 `case_extraction` step을 시작하고
`extractionPipeline`을 실행한다. 자세한 내부 노드는 아래
[`CaseExtractionPipeline`](#caseextractionpipeline)을 참고한다.

완료 시 체크포인트:

```json
{
  "questionType": "procedure",
  "caseExtraction": {
    "actors": [],
    "statuses": [],
    "actions": [],
    "relations": [],
    "medicalProcedureOptions": [],
    "patientDescriptions": []
  }
}
```

실제 `questionType` 값은 `cause`, `status`, `procedure`, `purpose`, `mechanism`,
`device_action` 중 하나다.

### 4. `DraftResolveNode`

`run_resolve_node`가 `resolve` step을 시작하고 `resolve_case_extraction`을 호출한다.
이 노드는 panic을 recover해 step과 case를 `failed`로 남긴다. status 파싱/저장
실패는 의도적으로 panic으로 올려 resolve 전체를 실패시킨다.

resolve 순서:

1. 문제가 환자 서술 형태가 아니면 patient description draft를 버린다
   (`looks_like_patient_description`).
2. source fallback: `ExplanationSourceID` → `Case.SourceID`.
3. actor draft 저장:
   - `device` → `SaveDevice`
   - 그 외 → `SaveUnknownActor`
   - 이름에 "투여"가 포함된 actor는 infusion option 보강 대상으로 기록
4. status draft 저장: `patient.ParseStatus` → `CanonicalizeStatusEntity`
5. action draft 저장:
   - `parser.NormalizeActionValue`
   - `cause_status`면 `CanonicalizeCauseStatusValue`
   - `SaveAction`
6. medical procedure option 저장:
   - indication/expected effect key → status ID
   - indication이 비면 현재 status 전체를 사용
   - `CanonicalizeProcedure` → `SaveMedicalProcedureOption`
7. infusion actor가 있으면 `ensure_infusion_option`으로 `Infuse` option 생성
8. relation draft를 `LinkStatusAction`, `LinkActionResultStatus`로 연결
9. patient description을 `patient.PatientDescription`으로 조립

완료 시 `state.Statuses`, `state.Actions`, `state.PatientDescriptions`가 채워지고
`resolveStepOutput`이 저장된다.

### 5. `GraphCauseTraceNode`

`run_cause_node`가 `cause_trace` step을 시작한다.

1. `state.Statuses`를 `[]entity.Status`로 변환한다.
2. 비어 있으면 `situation.Situation.ToStatuses()`를 사용한다.
3. `coverageObservedStatuses`로 문제 본문에 명시된 소견을 보강한다
   (정답/해설은 evidence에서 제외).
4. 90초 타임아웃 컨텍스트를 만든다(`cause_trace_timeout`).
5. `service.Postprocess`로 그래프를 로드하고 후처리 규칙을 적용한다.
6. `cause.Service.TraceContext(ctx, statuses, 100)`로 원인 후보를 추적한다.
7. 결과를 `causeStepOutput`으로 만들어 `SetAnswer(caseID, JSON)`에 저장한다.

```json
{
  "observed": ["혈압 증가"],
  "candidates": [
    {
      "causeStatusId": 12,
      "causeKind": "increased_index",
      "causeKey": "increased_index:혈압",
      "causeLabel": "혈압 증가",
      "viaActionID": 3,
      "viaActionKind": "cause_status",
      "depth": 1,
      "sourceIds": [1],
      "path": []
    }
  ],
  "missing": [],
  "coverageAdded": []
}
```

후보가 없으면 `missing`에 그래프에서 원인을 찾지 못했다는 메시지를 넣는다.
`cause_trace`는 LLM 호출이 아니지만 재현 패널을 위해 `solve_llm_calls`에
`stage="cause_trace"`로 기록된다. 같은 stage의 이전 기록은 먼저 삭제된다.

## `CaseExtractionPipeline`

`extractionPipeline`은 별도의 `core.Pipeline`을 실행한다. state에는 바깥
`pipelineState`가 `solve_state` 키로 공유된다.

| 순서 | 노드 | optional | 역할 |
| --- | --- | --- | --- |
| 1 | `LLMQuestionTypeNode` | | 질문 유형 분류 |
| 2 | `LLMParallelExtractionNode` | | actor/status/action 병렬 추출 |
| 3 | `LLMPatientDescriptionNode` | yes | 환자 서술 추출(조건부 no-op) |
| 4 | `LLMRelationNode` | | status/action 관계 추출 |
| 5 | `LLMOptionExtractionNode` | | 치료 선택지 추출 |
| 6 | `CaseExtractionAssembleNode` | | draft를 `CaseExtraction`으로 조립 |

병렬 노드는 `run_parallel`로 actor/status/action LLM 호출을 동시에 실행한다.
status/action 실패는 파이프라인 실패, actor 실패는 경고 후 빈 결과로 처리한다.
patient 추출은 `looks_like_patient_description(problem)`가 참일 때만 실제로
호출되며, 실패해도 경고 후 빈 결과가 된다.

중간 결과는 `core.State`에 다음 키로 저장된다.

| 키 | 값 |
| --- | --- |
| `extract_actors` | `[]ActorDraft` |
| `extract_statuses` | `[]StatusDraft` |
| `extract_actions` | `[]ActionDraft` |
| `extract_patients` | `[]PatientDescriptionDraft` |
| `extract_relations` | `[]RelationDraft` |
| `extract_options` | `[]OptionDraft` |

## Stage와 체크포인트

### stage 순서

`solve/domain/stage.go`:

| stage | sequence | downstream |
| --- | --- | --- |
| `explain` | 10 | case_extraction, resolve, cause_trace |
| `case_extraction` | 20 | resolve, cause_trace |
| `resolve` | 30 | cause_trace |
| `cause_trace` | 60 | 없음 |

> `question_type`, `actor_extract` 등은 LLM 호출 stage이며 최상위 resume 단위가
> 아니다. resume은 위 4개 stage만 지원한다.

### resume

`RunWithCache(ctx, id, ...)`는 항상 `fromNode=""`로 전체 실행한다. `fromNode` 기반
resume은 `RunReplay` 또는 내부 `run_v2_from`에 `fromNode`를 넘길 때 사용한다.

1. `CaseLoadNode`가 해당 stage를 검증하고 `MarkDownstreamStale`을 호출한다.
2. 각 노드는 `StepSeq(fromNode) > StepSeq(자기 stage)`이면 이전 체크포인트를
   `load_*`로 읽고, 아니면 다시 실행한다.
3. `load_step_output`은 step이 `succeeded`이고 `meta.OutputJSON`이 있을 때만
   역직렬화한다. 없으면 오류다.

### step 상태

- `pending`, `running`, `succeeded`, `failed`, `stale`
- `MarkDownstreamStale`은 `seq`가 큰 step을 `stale`로 바꾼다.
- `RecoverInterrupted`(store)는 `running` 상태로 남은 step/case를 실패 처리한다.

## LLM 호출 기록

`Service.record_call`은 `solve_llm_calls`에 호출을 남긴다. seq 상수는
`solve/trace.go`에 있다.

| seq | stage | 사용 여부 |
| --- | --- | --- |
| 10 | `solver_explain` | explain |
| 12 | `question_type` | 질문 유형 |
| 15 | `actor_extract` | actor |
| 17 | `other_names_extract` | 현재 미사용 |
| 20 | `status_extract` | status |
| 20 | `case_extraction` | 호출 기록에는 미사용 |
| 30 | `action_extract` | action |
| 35 | `actor_action_extract` | 현재 미사용 |
| 36 | `patient_description_extract` | 환자 서술 |
| 37 | `medical_procedure_option_extract` | 선택지 |
| 40 | `relation_extract` | 관계 |
| 50 | `situation` | 현재 미사용 |
| 60 | `cause_trace` | 그래프 추적(LLM 아님) |

저장되는 필드: `prompt_version`, `prompt`, `raw_output`, `parsed_output`,
`source_id`, `resp_id`, `state`, `error`, `duration_ms`.

## Replay

### 준비

`PrepareReplay(ctx, sourceID, mode, fromNode)`:

1. mode가 `llm`/`code`인지 확인한다.
2. case가 존재하고 `running`/`pending`이 아닌지 확인한다.
3. `ResetCaseForReplay(sourceID, mode, fromNode, keepCalls)`를 호출한다.
   - code mode는 LLM 호출 기록을 유지한다.
   - llm mode는 호출 기록을 지운다.
4. 같은 case ID를 반환한다. 새 case를 만들지 않는다.

### LLM replay

`RunReplay(..., mode=llm, ...)`:

- `cache_mode = LLMCacheUse` (또는 `useCache=false`면 `LLMCacheRefresh`)
- `run_v2_from`으로 다시 실행한다.

### Code replay

`RunReplay(..., mode=code, ...)`:

1. `ListLLMCalls(sourceID)`로 저장된 호출을 읽는다.
2. `newReplayLLM`이 프롬프트 원문 → raw output, stage → raw output 맵을 만든다.
3. `replay_stage_from_prompt`가 프롬프트 내용으로 stage를 추정한다.
4. `liveStages[stage]`가 참이면 live LLM을 호출하고, 아니면 저장된 출력을 돌려준다.
5. `cache_mode = LLMCacheBypass`로 캐시를 우회한다.
6. case에 source가 없고 situation source가 있으면 source를 복구한다.

## 오류 처리 규칙

- 최상위 노드 실패 시 `store.SetFailed`로 case를 실패 처리한다.
- resolve는 panic을 recover해 `resolve panic: ...`으로 실패 처리한다.
- extraction pipeline 실패 시 `case_extraction` step이 `failed`가 된다.
- actor 추출 실패: 경고 후 actor 없이 계속.
- patient 서술 추출 실패: 경고 후 patient description 없이 계속.
- status/action/relation/option 추출 실패: 파이프라인 실패.
- cause 추적 실패/타임아웃: `cause_trace` step 실패.
