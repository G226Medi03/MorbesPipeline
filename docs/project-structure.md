# 현재 프로젝트 구조

기준일: 2026-11-22

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
    ├── service/
    ├── support/
    └── testutil/
```

## internal/entity

순수 도메인/value object와 도메인 정책만 둔다.
repository/service를 import하지 않는다.

```text
internal/entity/
  entitykey/   # stable storage/audit key
  label/       # status/action/procedure label
  patient/     # patient description
  situation/   # situation -> status
  solve/       # solve domain DTO/policy
    checkpoint/ # step output / resolved extraction DTO
    input/      # problem stem/choice parsing
    model/      # LLM call model
    selection/  # procedure option scoring policy
```

## internal/repository

SQLite/MER persistence adapter만 둔다.
service를 import하지 않는다.

```text
internal/repository/
  store/          # Store: ctx/db + repositories + transaction
  medicalentity/  # medical entity ename link/read/save
  solve/          # solve tables
  benchmark/      # llm_problem_results
  postprocess/    # postprocess audit
  schema/         # schema ensure helpers
  migration/      # legacy schema migration
```

## internal/service

use case/application service만 둔다.
`StoreContainer`는 `internal/service/store`가 소유한다.

```text
internal/service/
  store/          # StoreContainer + persistence application services
  orchestration/  # solve 실행, node 구성, replay
    state/
    prompt/
    stage/
    coverage/
    trace/
    adapters/postprocess/
  llm/            # LLM extraction adapter + logger
  resolve/        # extraction DTO -> entity resolve helper
  postprocess/    # deterministic postprocess rules
  tracing/        # reverse link cause trace
```

역할:

```text
service/store
  StoreContainer, New/NewWithDBs, InitSchema, transaction wrappers
  Entity/Solve/Source/Benchmark/Postprocess application services

service/orchestration
  solve case load/explain/extract/resolve/cause node 실행
  pipelineState 관리
  replay/checkpoint 재개

service/llm
  MorbesLLMService extraction runner 호출
  question type normalization

service/resolve
  DTO -> Entity helper
  infusion option 보정

service/postprocess
  transplant rejection, increased index canonical action,
  cause_status result link 보정

service/tracing
  ActionsResultingInStatus
  StatusesCausingAction
  ActorsCausingAction
  기반 cause trace
```

## internal/support

```text
internal/support/
  logconfig/
  sliceutil/
```

## internal/testutil

```text
internal/testutil/
  testdb/
```

## Dependency direction

```text
internal/entity <- internal/repository/* <- internal/service/* <- root
MorbesLLMService <- internal/service <- root
MedicalEntityRepository <- internal/repository/medicalentity
MedicalEntityParser <- internal/service
```

금지:

```text
internal/entity -> internal/repository
internal/entity -> internal/service
internal/repository -> internal/service
internal/service -> root
MedicalEntity direct import
```

## 관련 문서

- [architecture.md](architecture.md)
- [persistence.md](persistence.md)
- [pipeline.md](pipeline.md)
- [domain-model.md](domain-model.md)
