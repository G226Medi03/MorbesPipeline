# 현재 프로젝트 구조

기준일: 2026-11-19

## 목적

MorbesPipeline은 의학 문제 풀이 case를 실행하고, LLM 추출 결과를 DTO와 Entity로 변환해 SQLite에 저장한 뒤, 저장된 인과 링크에서 원인 후보를 계산한다.

```text
LLM output
  -> MedicalEntityParser DTO
  -> MorbesLLMService extraction DTO
  -> MorbesPipeline resolve
  -> MedicalEntityRepository entity persistence
  -> cause trace / postprocess
  -> solve_cases.answer_json
```

## 최상위 구조

```text
.
├── facade.go                    # root public facade
├── go.mod / go.sum
├── AGENTS.md
├── docs/
├── scripts/
├── tools/
│   └── importlint/
└── internal/
    ├── entity/
    ├── repository/
    └── service/
```

## internal/service

```text
internal/service/
  orchestration/   # solve 실행, node 구성, replay
    state/         # pipeline state, step output
    prompt/        # prompt constants/helpers
    stage/         # stage/node constants
    coverage/      # coverage helper
    trace/         # LLM call trace
    adapters/postprocess/ # deterministic postprocess adapter
  solve/           # solve domain DTO
  store/           # persistence application services
  checkpoint/      # resolve/step checkpoint DTO
  input/           # 문제 stem/choice parsing
  llm/             # LLM extraction adapter + logger
  resolve/         # extraction DTO -> entity resolve helper
  postprocess/     # deterministic postprocess rules
  selection/       # procedure option selection
  tracing/         # reverse link cause trace
  support/         # logconfig, sliceutil
```

역할:

```text
orchestration
  solve case load/explain/extract/resolve/cause node 실행
  pipelineState 관리
  replay/checkpoint 재개

checkpoint
  solve_steps.output_json 형상
  resolved extraction serialization

input
  problem stem, choice parsing

llm
  MorbesLLMService extraction runner 호출
  question type normalization

resolve
  DTO -> Entity helper
  infusion option 보정

postprocess
  transplant rejection, increased index canonical action,
  cause_status result link 보정

selection
  procedure option scoring/selection
```

## internal/repository

```text
internal/repository/
  facade/          # public StoreContainer
  store/           # Store: ctx/db + repositories
  benchmark/       # llm_problem_results
  entity/          # entity ID/link/read/save
  solve/           # solve tables
  postprocess/     # postprocess audit
  schema/          # schema ensure helpers
  migration/       # legacy schema migration
  testdb/          # test db helper
```

역할:

```text
internal/repository/facade
  StoreContainer, New/NewWithDBs, InitSchema
  유일한 public store package

internal/repository/store
  Store: ctx/db + repositories + transaction

internal/service/store
  Entity/Solve/Source/Benchmark/Postprocess application services

internal/repository/benchmark
  llm_problem_results, ProblemResult, ExamQuestionDataCounts

repository/entity
  entity ID/link/read/save

repository/solve
  solve_cases, solve_steps, solve_llm_calls, answer checks

repository/postprocess
  morbes_postprocesses audit

repository/schema
  schema ensure helpers

repository/migration
  legacy schema migration
```

## internal/entity

```text
internal/entity/
  key/       # stable entity identity key
  label/     # status/action/procedure label
  patient/   # patient description
  situation/ # situation -> status
  trait/     # trait policy
```

## internal/service/tracing

```text
internal/service/tracing/
  candidate.go
  cause_trace_result.go
  service.go
  trace.go
  trace_node.go
```

역할:

```text
ActionsResultingInStatus
StatusesCausingAction
ActorsCausingAction
기반 cause trace
```

## internal/service/support

```text
internal/service/support/
  logconfig/   # log DB config
  sliceutil/   # slice helper
```

## tools

```text
tools/importlint/
  package dependency / direct import 검사
```

## dependency direction

```text
internal/entity <- internal/repository/* <- internal/service/* <- root
internal/entity <- internal/service/tracing <- root
MorbesLLMService <- internal/service <- root
MedicalEntityRepository <- internal/repository
MedicalEntityParser <- internal/service
```

예외:

```text
internal/repository/facade -> internal/service/store
internal/repository/solve  -> internal/service/solve
```

금지:

```text
internal/entity -> internal/repository
internal/entity -> internal/service
internal/service -> root
MedicalEntity direct import
```

## 실행 흐름

```text
1. root facade가 StoreContainer와 service orchestration를 조립
2. StoreContainer.InitSchema
3. service.orchestration이 case를 load
4. explain LLM 호출
5. llm adapter가 extraction runner 실행
6. resolve helper가 DTO를 Entity로 변환
7. entity/source store에 저장
8. postprocess rules 적용
9. cause trace reverse link 조회
10. answer_json 저장
```

## 관련 문서

- [architecture.md](architecture.md)
- [pipeline.md](pipeline.md)
- [persistence.md](persistence.md)
- [extraction-llm.md](extraction-llm.md)
- [domain-model.md](domain-model.md)
- [sync/README.md](sync/README.md)
