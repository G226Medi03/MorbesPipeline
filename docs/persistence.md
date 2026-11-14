# 영속성 구조

## 두 계층

| 계층 | 패키지 | 역할 |
| --- | --- | --- |
| 포트 | `ports` | 파이프라인이 요구하는 저장소 인터페이스 |
| 어댑터 | `store` | SQLite 구현, 스키마 초기화, 트랜잭션, 확장 조회 |

파이프라인은 `ports.Store`만 알고, 실제 DB는 `store.Store`가 담당한다.
웹/CLI 계층은 `ports.Store`보다 넓은 `store.Store` API를 직접 사용할 수 있다.

## `ports.Store`

`ports/store.go`의 인터페이스는 다음 그룹으로 나눌 수 있다.

### case

- `Get(id) (domain.Case, bool, error)`
- `SetStage`, `SetRunning`, `SetFailed`, `SetAnswer`, `SetQuestionType`, `SetSource`
- `ResetCaseForReplay(id, mode, fromNode, keepCalls)`

### step

- `StartStep(caseID, node, promptVersion) (SolveStepID, error)`
- `FinishStep(id, state, outputJSON, message, callID)`
- `GetStep(caseID, node) (Step, bool, error)`
- `ResetSteps(caseID)`
- `MarkDownstreamStale(caseID, fromNode)`

### LLM

- `AppendLLMCall`, `ListLLMCalls`, `DeleteLLMCallsByStage`
- `FindLlmResponse`
- `GetOrInsertLlmResponse`
- `RefreshLlmResponse`

### 의료 엔티티 저장

- `SaveAction`
- `SaveDevice`
- `SaveUnknownActor`
- `SaveMedicalProcedureOption`
- `LinkStatusAction`
- `LinkActionResultStatus`
- `SaveDisease`, `SaveIndex`, `SaveSymptom`, `SaveStatus`
- `SaveProcedureWithTraits`

### 프롬프트

- `GetPrompt(stage) (template string, version int, error)`

### 그래프

- `LoadGraph() (*cgraph.Container, error)`
- `EnsureStatusIn`, `EnsureActionIn`
- `LinkStatusActionIn`, `LinkActionResultStatusIn`
- `DeleteStatusAction`, `DeleteActionResultStatus`

## `store.Store`

### 구성

```go
type Store struct {
    ctx context.Context
    db  *sql.DB
    tx  *sql.Tx
    mu  *sync.Mutex

    sourceDB *sql.DB
    sourceTx *sql.Tx
    solveDB  *sql.DB
    solveTx  *sql.Tx

    StatusRepository    medicalentity.StatusRepository
    ActionRepository    medicalentity.ActionRepository
    ActorRepository     medicalentity.ActorRepository
    DiseaseRepository   medicalentity.DiseaseRepository
    IndexRepository     medicalentity.IndexRepository
    SymptomRepository   medicalentity.SymptomRepository
    ProcedureRepository medicalentity.ProcedureRepository

    OptionRepository medicalprocedureoption.Repository
    SourceRepository storesource.Repository
    SolveRepository  storesolve.Repository
    GraphRepository  storegraph.Repository
}
```

- `New(ctx, db)`: medical/source/solve를 한 DB로 사용
- `NewWithDBs(ctx, medicalDB, sourceDB, solveDB)`: DB 분리
- `bind()`: 모든 repository를 현재 executor에 연결
- `WithTx`, `WithSourceTx`, `WithSolveTx`: 트랜잭션 복제본 생성
- `lockWrite()`: 애플리케이션 쓰기를 직렬화. LLM 호출처럼 긴 작업은 락 밖에 둔다.

### 스키마 초기화

`InitSchema()` 순서:

1. `migrateLegacyMedicalSchema`: 레거시 테이블/컬럼 이름 변경
2. `medicalentity.InitSchema`: MedicalEntity 엔티티/연결 테이블
3. `OptionRepository.Init`: 선택지 테이블
4. `SourceRepository.Init`: source/LLM 응답 테이블
5. `SolveRepository.Init`: solve 테이블
6. `ensure_llm_problem_results`: 벤치마크 결과 테이블
7. `ensure_status_other_name`: 상태 다른 이름 테이블
8. `ensure_entity_sources`: 일반 엔티티-source 매핑
9. `ensure_solve_id_columns`: solve 테이블의 `id` → `*_id` 컬럼명 변경
10. `validateKindPtr`: kind/ptr 무결성 검사

### 주요 테이블

#### MedicalEntity 계층

`medicalentity.InitSchema`가 관리한다.

- 엔티티: `eactor`, `estatus`, `eaction`, `emedical_procedure`,
  `edisease`, `eindex`, `esymptom`
- 관계: `status_actions`, `action_result_statuses`, `actor_actions`
- source: `status_sources`, `action_sources`, `actor_sources`,
  `procedure_sources`, `disease_sources`, `index_sources`, `symptom_sources`
- 하위 타입: `status_*`, `action_*`, `actor_*`, `procedure_*`
  (kind/ptr이 가리키는 실제 값 테이블)

`validateKindPtr`는 kind와 ptr이 올바른 하위 타입 행을 가리키는지 검사한다.
예: `estatus.kind='increased_index'`이면 `status_increased_index.id = estatus.ptr`.

#### 선택지

`store/medicalprocedureoption`:

- `medical_procedure_options`
  - `medical_procedure_option_id`, `emedical_procedure_id`, `category`,
    `traits_json`, `context_traits_json`, `semantic_key`
  - `semantic_key`는 procedure/category/traits/indications 조합으로 중복 저장을 막는다.
- `medical_procedure_option_indications`
- `medical_procedure_option_expected_effects`
- `medical_procedure_option_sources`

#### solve

`store/solve`:

| 테이블 | 내용 |
| --- | --- |
| `solve_cases` | 문제/정답/해설, 상태, stage, question type, source, replay 메타 |
| `solve_steps` | node별 체크포인트, `output_json`, `prompt_version`, 상태 |
| `solve_llm_calls` | LLM 호출/응답, stage, seq, source/resp, duration |
| `solve_prompt_versions` | stage별 프롬프트 템플릿/버전 |
| `solve_answer_checks` | 답안 검증 결과 |

`SolveRepository`는 `Case`, `Step`, `Call`, `Prompt`, `AnswerCheck` 테이블을 묶는다.

#### 보조

- `llm_problem_results`: 벤치마크 결과. `store/benchmarkresult`
- `entity_sources`: `(entity_kind, entity_id, source_id)` 일반 매핑
- `status_other_name`: 상태 엔티티의 다른 이름/별칭

#### source

`SourceEntityRepository`가 source와 LLM 응답 테이블을 관리한다.
`store/source.go`가 다음을 감싼다.

- `GetSource`, `EnsureSource`
- `GetLlmResponseRow`
- `FindLlmResponse`
- `GetOrInsertLlmResponse`
- `RefreshLlmResponse`
- `TraceSource`

## 쓰기 경로

### status

`SaveStatus`:

1. `Status.Validate()` (variant + source)
2. 레거시 값 → MedicalEntity status 변환
3. `StatusRepository.Insert`
4. `validateKindPtr("estatus")`
5. 저장된 status를 로컬 DTO로 변환

### action

`SaveAction`:

- `cause_status`는 status ID를 분리해 `action_cause_status`와 연결한다.
- 그 외 action은 MedicalEntity action repository에 저장한다.
- 저장 후 source와 결과 상태 링크를 관리한다.

### actor

- `SaveDevice` → `eactor(kind=device)` + device 하위 테이블
- `SaveUnknownActor` → `eactor(kind=unknown_actor)` + unknown 하위 테이블
- `AddActorOtherName`, `LinkActorAction`으로 별칭/행동 연결

### procedure

- `CanonicalizeProcedure`가 만든 payload를 `SaveProcedureWithTraits`로 저장
- intrinsic trait는 `emedical_procedure.traits_json`에 저장
- transplant target이 unknown이면 `unresolved.go`의 조회/해결 API로 보정 가능

### medical procedure option

- `Validate`: procedure ID와 source 필수
- category 정규화
- `semantic_key` 기준 upsert/중복 방지
- indication/expected effect/source를 별도 테이블에 교체 저장

## 그래프 로딩

`LoadGraph()`:

1. `validateKindPtr()`로 무결성 검사
2. `GraphRepository.Load()`로 row 로드
   - actor/status/action/procedure/option
   - status_actions / action_result_statuses / actor_actions
   - 각 source 테이블
3. MedicalEntity repository로 실제 값을 hydrate
4. procedure trait와 option trait를 병합
5. 선택지를 `MedicalProcedureOptions[statusID]`에 인덱싱
6. source/actor/action/status 링크를 `graph.Container`에 채움

이 컨테이너가 `service.Postprocess`와 `cause.TraceContext`의 입력이 된다.

## 트랜잭션과 동시성

- 단일 `Store` 인스턴스는 `sync.Mutex`로 쓰기를 직렬화한다.
- `WithTx`는 이미 트랜잭션 안이면 같은 트랜잭션을 그대로 전달한다.
- `WithSourceTx`, `WithSolveTx`는 분리된 DB를 사용할 때 각각의 executor를 바꾼다.
- `ExecContext`, `QueryContext`, `QueryRowContext`는 tx가 있으면 tx를 사용한다.
- `RecoverInterrupted`는 `running` 상태로 남은 case/step을 실패로 정리한다.

## 레거시 마이그레이션

`migrateLegacyMedicalSchema`는 다음 이름을 새 스키마로 바꾼다.

| 이전 | 이후 |
| --- | --- |
| `e_statuses` | `estatus` |
| `e_actions` | `eaction` |
| `actors` | `eactor` |
| `eprocedure` | `emedical_procedure` |
| `status_id` 컬럼 | `estatus_id` |
| `action_id` 컬럼 | `eaction_id` |
| `actor_id` 컬럼 | `eactor_id` |
| `procedure_id` 컬럼 | `emedical_procedure_id` |

또한 레거시 `medical_procedure_options`에 `indication_status_ids_json` 컬럼이
있으면 `medical_procedure_options_legacy`로 이름을 바꿔 보존하고, 새 정규화
테이블을 만든다.

## 배선 예시

```go
db, err := sql.Open("sqlite", "morbes.db")
if err != nil {
    return err
}
st := store.New(ctx, db)
if err := st.InitSchema(); err != nil {
    return err
}

svc := csolve.New_Service(ctx, csolve.Dependencies{
    Store:         st,
    Canonical:     canonical.New(st),
    Solver:        solverLLM,
    ExtractLLM:    extractLLM,
    QuestionStore: questionStore,
})

if err := svc.Run(ctx, caseID); err != nil {
    return err
}
```

- `store.New`은 테스트/CLI용 단일 DB 구성이다.
- 서버에서 medical/source/solve를 분리하려면 `store.NewWithDBs`를 사용한다.

## 주의 지점

- `store/entity_neighbors.go`의 테이블 매핑은 `actors`, `e_actions`,
  `e_statuses`, `eprocedure` 같은 레거시 이름을 사용한다. 초기 마이그레이션이
  이 이름들을 바꾸므로, 이 API를 사용할 때는 실제 스키마와 매핑이 일치하는지
  확인해야 한다.
- 루트 `doc.go`는 초기 placeholder 문구를 그대로 담고 있다.
