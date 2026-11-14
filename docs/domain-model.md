# 도메인 모델

## 식별자

식별자는 solve 도메인과 `legacy` 패키지로 나뉜다. `CaseID`는
`solve/domain/case_id.go`의 문자열 타입이고, 나머지는 `legacy/id.go`의 int64
기반 타입이다.

| 타입 | 의미 |
| --- | --- |
| `CaseID` | solve case (`solve/domain`) |
| `StatusID` | 상태 행 |
| `ActionID` | 행동 행 |
| `ActorID` | actor 행 |
| `ProcedureID` | 시술 행 |
| `MedicalProcedureOptionID` | 치료 선택지 행 |
| `DeviceID` | 레거시 장치 ID. 새 장치는 actor(kind=device)로 저장 |
| `ExamQuestionID` | 시험 문제 행 |
| `LLMProblemResultID` | LLM 문제 결과 행 |
| `SolveStepID` | solve 단계 행 |
| `LLMCallID` | solve LLM 호출 행 |
| `StatusEntityID` | disease/symptom/index 행 |

## Case

`solve/domain/case.go`

```go
type Case struct {
    ID                CaseID
    ExamQuestionID    entity.ExamQuestionID
    Processed         bool
    Problem           string
    Answer            string
    Explanation       string
    SituationJSON     string
    AnswerJSON        string
    State             string
    Stage             string
    QuestionType      string
    Error             string
    SourceID          source.SourceId
    SituationSourceID source.SourceId
    ReplayOf          CaseID
    ReplayMode        string
    ReplayFromNode    string
    CreatedAt         time.Time
    UpdatedAt         time.Time
    Steps             []Step
}
```

### 상태 값

| 값 | 의미 |
| --- | --- |
| `pending` | 아직 실행 전 |
| `running` | 실행 중 |
| `done` | 완료 |
| `failed` | 실패 |

### stage

파이프라인 진행 표시는 `stage`에 저장된다. 주요 값:

- `solver_explain`
- `case_extraction`
- `resolve`
- `cause_trace`

`solve_cases.stage`의 기본값은 `queued`다.

## Step

`solve/domain/step.go`

```go
type Step struct {
    ID         entity.SolveStepID
    Node       string
    Seq        int
    State      StepState
    Attempt    int
    Error      string
    StartedAt  time.Time
    FinishedAt *time.Time
    Meta       SolveStepMeta
}
```

`SolveStepMeta`:

```go
type SolveStepMeta struct {
    CaseID        CaseID
    PromptVersion int
    OutputJSON    string
    CallID        entity.LLMCallID
}
```

`StepState`는 문자열 타입이며 실제 값은 `pending`, `running`, `succeeded`,
`failed`, `stale`이다.

주요 node:

| node | stage |
| --- | --- |
| `explain` | explain |
| `case_extraction` | case_extraction |
| `resolve` | resolve |
| `cause_trace` | cause_trace |

## LLM 호출

`solve/model/llmcall.go`

```go
type LLMCall struct {
    ID                 entity.LLMCallID
    CaseID             domain.CaseID
    Seq                int
    Stage              string
    Role               string
    Model              string
    PromptVersion      int
    Prompt             string
    RawOutput          string
    PreprocessedOutput string
    PreprocessChanges  string
    ParsedOutput       string
    SourceID           source.SourceId
    RespID             source.RespId
    State              string
    Error              string
    DurationMS         int64
    CreatedAt          time.Time
}
```

`LLMCallInput`은 ID/CreatedAt을 제외한 같은 필드다.

## 답안 검증

`solve/domain/answer_check.go`

```go
type AnswerCheck struct {
    CaseID          CaseID
    RespID          source.RespId
    QuestionType    string
    ContainsAnswer  bool
    AnswerPresent   bool
    AnswerSatisfied bool
    EffectMatched   bool
    MatchedType     string
    MatchedKey      string
    AnalysisJSON    string
    Error           string
    UpdatedAt       time.Time
}
```

## 프롬프트 버전

`solve/model/prompt_version.go`

```go
type PromptVersion struct {
    Stage     string
    Template  string
    Version   int
    UpdatedAt time.Time
}
```

## Replay

```go
type ReplayMode string

const (
    ReplayModeLLM  ReplayMode = "llm"
    ReplayModeCode ReplayMode = "code"
)
```

## 추출 Draft

`solve/case_extraction.go`

### `CaseExtraction`

한 번의 추출 단계 결과 전체다.

```go
type CaseExtraction struct {
    PatientDescription      *PatientDescriptionDraft
    PatientDescriptions     []PatientDescriptionDraft
    Actors                  []ActorDraft
    Statuses                []StatusDraft
    Actions                 []ActionDraft
    Relations               []RelationDraft
    MedicalProcedureOptions []OptionDraft
}
```

### Draft 타입

| 타입 | 필드 | 설명 |
| --- | --- | --- |
| `ActorDraft` | `Key`, `Kind`, `Name`, `SourceID` | LLM이 찾은 actor |
| `StatusDraft` | `Key`, `Kind`, `Value`, `SourceID` | `Value`는 JSON 원문 |
| `ActionDraft` | `Key`, `Variant`, `Value`, `SourceID` | 정규화된 action JSON |
| `RelationDraft` | `CauseStatusKey`, `ActionKey`, `ResultStatusKey`, `SourceID` | status/action key 참조 |
| `OptionDraft` | `Key`, `Category`, `Traits`, `MedicalProcedure`, `Indications`, `ExpectedEffects`, `SourceID` | 치료 후보 |
| `PatientDescriptionDraft` | `Choice`, `Description`, `Knowledge`, `SourceID` | 환자 서술 |

draft key는 LLM이 만든 임시 식별자다. `resolve` 단계에서만 실제 ID로 바뀐다.

### `ResolvedCaseExtraction`

```go
type ResolvedCaseExtraction struct {
    Actors                  []ActorDraft
    Statuses                map[string]entity.Status
    Actions                 map[string]entity.Action
    Relations               []RelationDraft
    MedicalProcedureOptions []medicalprocedureoption.MedicalProcedureOption
    PatientDescriptions     []patient.PatientDescription
}
```

## 파이프라인 상태

`solve/pipeline_state.go`

```go
type pipelineState struct {
    Case                Case
    Prompts             promptSet
    QuestionType        domain.QuestionType
    Explanation         string
    Extraction          CaseExtraction
    Resolved            ResolvedCaseExtraction
    Actors              []string
    Devices             []string
    ExplanationSourceID source.SourceId
    Statuses            map[string]entity.Status
    Actions             map[string]entity.Action
    PatientDescriptions []patient.PatientDescription
    Relations           [][]string
    RelationSource      source.SourceRef
    SituationJSON       string
    Situation           csituation.Situation
    SituationSourceID   source.SourceId
    Candidates          []ccause.Candidate
}
```

일부 필드(`Actors`, `Devices`, `Relations`, `RelationSource`, `Situation` 등)는
이전 버전과의 호환을 위해 남아 있으며 현재 v2 흐름에서는 사용하지 않는다.

## 엔티티 래퍼

`legacy` 패키지는 MedicalEntity 값을 파이프라인 내부에서 다루기 쉽게 감싼다.

### Status

```go
type Status struct {
    Variant string
    Value   any
    Actions []struct {
        Action   Action
        SourceID source.SourceId
    }
    Sources []source.SourceRef
}
```

- `id`는 private이며 `Id()`/`NewStatus`/`RestoreStatus`로 접근한다.
- `Actions`는 한 단계만 hydrate된다. 중첩 action의 결과 상태는 비어 있다.

### Action

```go
type Action struct {
    Variant string
    Value   any
    ResultStatuses []struct {
        Status   Status
        SourceID source.SourceId
    }
    Sources []source.SourceRef
}
```

`CauseStatusAction`은 레거시 `status_id` 전용 action이다.

```go
type CauseStatusAction struct {
    StatusID int64
    Status   IEStatus
}
func (CauseStatusAction) Kind() string { return "cause_status" }
```

### Procedure

```go
type Procedure struct {
    Variant string
    Value   any
    Sources []source.SourceRef
}
```

### Actor

| 타입 | 설명 |
| --- | --- |
| `ActorMeta` | `ID`, `Kind`, `Name`, `OtherNames`, `Actions`, `Sources` |
| `Device` | actor meta + `eactor.Device` |
| `UnknownActor` | actor meta + `eactor.UnknownActor` |
| `ActorRecord` | actor meta + `eactor.IEActor` |
| `Organ`/`Tissue`/`Mat`/`TransplantableDevice`/`OtherActor` | `UnknownActor` alias |

## 상태 종류

`label.IsStatusKind`가 인식하는 값:

| kind | 의미 | canonical 저장 |
| --- | --- | --- |
| `increased_index` | 수치 증가 | `index_id` |
| `decreased_index` | 수치 감소 | `index_id` |
| `have_disease` | 질병 보유 | `disease_id` |
| `have_symptom` | 증상 보유 | `symptom_id` (+ pain 상세) |
| `damaged` | 손상 | `actor_id` |
| `inflammation` | 염증 | `actor_id` |
| `received_medical_procedure` | 시술 받음 | `procedure_id` (+ `post_procedure_period`) |

canonicalize 후에는 이름 필드 대신 ID 필드만 남는다. 예:

```json
{"disease_id": 3}
{"index_id": 7}
{"procedure_id": 11, "post_procedure_period": "3일 전"}
```

## 행동 종류

| kind | 의미 | value |
| --- | --- | --- |
| `cause_status` | 어떤 상태가 원인이 되어 결과를 만든다 | `status_id` 또는 `status` |
| `block` | actor 차단 | `actor` |
| `relax` | actor 이완 | `actor` |

`cause_status` action은 `CanonicalizeCauseStatusValue`로 내부 status를 ID로 바꾼다.

## 시술 종류와 canonical payload

`canonical.CanonicalizeProcedure`가 처리하는 종류:

| kind | payload |
| --- | --- |
| `transplant` | `target_kind`, `target_id` (또는 `target_name`) |
| `infuse` | `material_id`, `route` |
| `surgery` | `name` (없으면 "수술") |
| `insertion` | `name`, `tube`, `site` |
| `cleaning` | `target_kind`, `target_id` |
| `diagnostic_test` | `name` |
| `education` | `name`, `topic` |
| `instruction` | `name`, `action`, `duration`, `position` |
| `unknown` | `name` |

intrinsic trait가 없으면 `trait.DefaultIntrinsicTraits(kind)`를 사용한다.

## Trait

`trait.Trait`은 시술/선택지의 의미 속성이다.

### 우선순위 trait

높은 순서:

1. `resuscitative`
2. `acute`
3. `definitive`
4. `first_line`
5. `rescue`

### 문맥 trait

`acute`, `chronic`, `first_line`, `rescue`, `maintenance`, `prophylactic`

- `IntrinsicTraits`: 시술 자체 속성. procedure에 저장
- `ContextTraits`: case 문맥 속성. medical procedure option에 저장
- `PriorityTraits`: 우선순위 순서로 정렬
- `DescriptiveTraits`: 우선순위가 아닌 trait을 입력 순서로 반환

### 카테고리 순위

`CategoryRank`:

1. `treatment`
2. `diagnostic_test`
3. `monitoring`
4. `nursing_intervention`
5. `education`

## 치료 선택지

`medicalprocedureoption.MedicalProcedureOption`

```go
type MedicalProcedureOption struct {
    ProcedureID             entity.ProcedureID
    Category                string
    Traits                  []entity.Trait
    ContextTraits           []entity.Trait
    IndicationStatusIDs     []entity.StatusID
    ExpectedEffectStatusIDs []entity.StatusID
    Sources                 []source.SourceRef
    Procedure               entity.IMedicalProcedure // runtime only
    Indication              []entity.IEStatus        // runtime only
}
```

`Procedure`/`Indication`은 `json:"-"`이며 영속화하지 않는다.
`Validate`는 procedure ID와 source를 요구한다.

### 선택지 그래프

`solve/domain/procedure_option.go`

```go
type MedicalProcedureOptionGraph struct {
    Options                         []MedicalProcedureOptionGraphItem
    SelectedOptionID                entity.MedicalProcedureOptionID
    Selected                        *MedicalProcedureOptionGraphItem
    AnswerInMedicalProcedureOptions bool
}
```

`MedicalProcedureOptionGraphItem`은 다음을 갖는다.

- `ID`, `ProcedureID`, `Category`
- `ProcedureTraits`, `ContextTraits`, `PriorityTraits`, `DescriptiveTraits`
- `TreatOptionPriority`, `PriorityReason`
- `ProcedureLabel`, `Action`
- `Indications`, `ExpectedEffects`, `Outcomes`
- `Sources`

점수화(`ScoreMedicalProcedureOption`):

| 조건 | 점수 |
| --- | --- |
| indication observed | +100 |
| indication cause | +30 |
| expected effect observed | +60 |
| expected effect cause | +15 |
| category == treatment | +5 |
| 희망 카테고리 일치 | +20 |

`SelectMedicalProcedureOption`은 점수가 0 이하인 후보를 버리고, 동점이면
`TreatOptionPriority`, 그다음 ID로 결정한다. 정답은 참조하지 않는다.

## 원인 후보

`cause/candidate.go`

```go
type Candidate struct {
    CauseType      string
    CauseStatusID  entity.StatusID
    CauseStatusIDs []entity.StatusID
    CauseActorID   entity.ActorID
    CauseActorIDs  []entity.ActorID
    CauseDeviceID  entity.DeviceID
    CauseDeviceIDs []entity.DeviceID
    CauseKind      string
    CauseKey       string
    CauseLabel     string
    ViaActionID    entity.ActionID
    ViaActionIDs   []entity.ActionID
    ViaActionKind  string
    ViaActionKey   string
    Depth          int
    SourceIDs      []source.SourceId
    Path           []TraceNode
}
```

`TraceNode`:

```go
type TraceNode struct {
    Type  string
    ID    int64
    Key   string
    Label string
}
```

후보는 `(CauseKey, ViaActionKey)` 기준으로 병합되고, 같은 키면 더 짧은 path를
선호한다. 최종 정렬은 `Depth` 오름차순, `CauseKey` 오름차순이다.

## 라벨

`label` 패키지는 저장된 ID/JSON에서 사람이 읽는 라벨을 만든다.

- `StatusLabel`: kind prefix 없는 상태 라벨
- `NodeLabel`: `kind:label`
- `ProcedureLabel`: 시술 라벨
- 수치: `증가`/`감소` suffix
- 시술: `받음` suffix와 시술 후 기간
- 통증: 부위 + 강도/기간/양상/방사
- 기간: 문자열 또는 `go-duration` 포맷
