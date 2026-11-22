# Persistence

## 구조

```text
internal/repository/
  store/
    store.go
    schema.go
  benchmark/
  solve/
  postprocess/

internal/service/store/
  container.go
  medical_entity.go
  solve.go
  ...
```

## StoreContainer

```text
internal/service/store/container.go
  Store *repository/store.Store
  MedicalEntity/Solve/Source/Benchmark/Postprocess service

internal/repository/store/store.go
  ctx
  medical db/tx
  source db/tx
  solve db/tx
  Medical *repo.Store
  Source storesource.Repository
  Solve solve.Repository
  Benchmark benchmark.Repository
  Postprocess postprocess.Repository
```

## 주요 구성

```text
internal/service/store
  StoreContainer, InitSchema, transaction wrappers
  MedicalEntity/Solve/Source/Benchmark/Postprocess application services
  MedicalEntityRepository ename-only link/unlink APIs

internal/repository/store
  ctx/db + repositories

internal/repository/benchmark
  llm_problem_results

internal/repository/solve
  solve_cases, solve_steps, solve_llm_calls, answer checks

internal/repository/postprocess
  morbes_postprocesses audit
```

## MedicalEntity ID

MorbesPipeline은 MedicalEntity numeric row ID를 사용하지 않는다.

```text
status/action/actor/procedure/option write
  -> MedicalEntityRepository ename API

numeric row ID
  -> MedicalEntityRepository 내부 구현
```

## 인과 링크

```text
ActionsResultingInStatus
StatusesCausingAction
ActorsCausingAction
```

## Postprocess audit

```text
morbes_postprocesses
  postprocessor_name
  operation
  entity_type
  entity_key
  target_entity_type
  target_entity_key
  finished_at
```

기존 numeric ID 기반 audit table은 migration 시
`morbes_postprocesses_legacy`로 보존된다.

`entity_sources`, `status_other_name` legacy table은 schema 초기화 시 제거된다.

## 관련 문서

- [project-structure.md](project-structure.md)
- [pipeline.md](pipeline.md)
- [domain-model.md](domain-model.md)
