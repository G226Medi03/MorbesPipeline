# Persistence

## 구조

```text
internal/store/
  store.go
  facade/
    store.go
    schema.go
    benchmark_tx.go
  repository/
    benchmark/
    entity/
    solve/
    postprocess/
    schema/
    migration/
```

## StoreContainer

```text
internal/store/facade/store.go
  Store *store.Store
  Entity/Solve/Source/Benchmark/Postprocess service

internal/store/store.go
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
internal/service
  Entity/Solve/Source/Benchmark/Postprocess services

store/facade
  InitSchema
  transaction wrappers
  StoreContainer facade

store
  ctx/db + repositories

repository/benchmark
  llm_problem_results
repository/entity
  entity ID/link/read/save
repository/solve
  solve_cases, solve_steps, solve_llm_calls, answer checks
repository/postprocess
  morbes_postprocesses
repository/schema
  schema ensure helpers
repository/migration
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
