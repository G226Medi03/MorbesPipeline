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
    ├── pipeline/
    ├── store/
    ├── domain/
    ├── tracing/
    ├── support/
    └── testdb/
```

## internal/pipeline

```text
internal/pipeline/
  orchestration/   # solve case 실행, node 구성, 상태, replay
  checkpoint/      # resolve/step checkpoint DTO
  input/           # 문제 stem/choice parsing
  llm/             # LLM extraction adapter + logger
  resolve/         # extraction DTO -> entity resolve helper
  postprocess/     # deterministic postprocess rules
  selection/       # procedure option selection
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

## internal/store

```text
internal/store/
  core/            # StoreContainer, schema, transaction, core SQL
  entity/          # entity helper
  read/            # read model types
  source/          # source/LLM response trace
  solve/           # solve table repository
  benchmark/       # benchmark table repository
  postprocess/     # postprocess audit repository
  migration/       # legacy schema migration
```

역할:

```text
core
  StoreContainer, New/NewWithDBs, InitSchema
  transaction executor, entity/solve/source 조립
  entity ID lookup, link/delete, read API

entity
  actor value helper, MedicalEntity trait conversion

read
  ActorRecord, ActionRecord, StatusRecord, ProcedureRecord,
  ProblemResultRow, ExamQuestionDataCounts

source
  SourceEntityRepository 기반 source/response trace

solve
  solve_cases, solve_steps, solve_llm_calls, answer checks

benchmark
  llm_problem_results

postprocess
  morbes_postprocesses audit

migration
  legacy table/column rename
```

## internal/domain

```text
internal/domain/
  solve/           # Case, Step, Stage, Question, replay
  solve/model/     # LLMCall, PromptVersion
  entitykey/       # stable entity identity key
  label/           # status/action/procedure label
  patient/         # patient description
  situation/       # situation -> status
  procedureoption/ # option DTO
  trait/           # trait policy
```

## internal/tracing

```text
internal/tracing/
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

## internal/support

```text
internal/support/
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
internal/support <- internal/domain <- internal/store <- root
internal/support <- internal/domain <- internal/tracing <- root
MorbesLLMService <- internal/pipeline <- root
MedicalEntityRepository <- internal/store
MedicalEntityParser <- internal/pipeline
```

금지:

```text
internal/domain -> store / pipeline / tracing
internal/store -> pipeline
internal/tracing -> store
MedicalEntity direct import
```

## 실행 흐름

```text
1. root facade가 StoreContainer와 pipeline Service를 조립
2. StoreContainer.InitSchema
3. pipeline.orchestration이 case를 load
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
