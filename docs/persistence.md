# Persistence

## 구조

```text
internal/repository/
  facade/
    facade.go
  store/
    store.go
    schema.go
  benchmark/
  entity/
  solve/
  postprocess/
  schema/
  migration/
  testdb/
```

## StoreContainer

```text
internal/repository/facade/facade.go
  Store *store.Store
  Entity/Solve/Source/Benchmark/Postprocess service

internal/repository/store/store.go
  ctx
  medical db/tx
  source db/tx
  solve db/tx
  Medical *repo.Store
  Source storesource.Repository
  Entity entity.Repository
  Solve solve.Repository
  Benchmark benchmark.Repository
  Postprocess postprocess.Repository
```

## 주요 구성

```text
internal/repository/facade
  StoreContainer, InitSchema, transaction wrappers
  필요한 public alias만 노출

internal/repository/store
  ctx/db + repositories

internal/service/store
  Entity/Solve/Source/Benchmark/Postprocess services

internal/repository/benchmark
  llm_problem_results
internal/repository/entity
  entity ID/link/read/save
internal/repository/solve
  solve_cases, solve_steps, solve_llm_calls, answer checks
internal/repository/postprocess
  morbes_postprocesses audit
internal/repository/schema
  schema ensure helpers
internal/repository/migration
  legacy schema migration
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
  entity_id
  target_entity_type
  target_entity_id
  finished_at
```

## 관련 문서

- [project-structure.md](project-structure.md)
- [pipeline.md](pipeline.md)
- [domain-model.md](domain-model.md)
