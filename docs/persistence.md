# Persistence

## 구조

```text
internal/store/
  facade/
    facade.go
  internal/
    store/
      store.go
      schema.go
    service/
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
internal/store/facade/facade.go
  Store *store.Store
  Entity/Solve/Source/Benchmark/Postprocess service

internal/store/internal/store/store.go
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
internal/store/facade
  StoreContainer, InitSchema, transaction wrappers
  필요한 public alias만 노출

internal/store/internal/store
  ctx/db + repositories

internal/store/internal/service
  Entity/Solve/Source/Benchmark/Postprocess services

internal/store/internal/repository/*
  benchmark/entity/solve/postprocess/schema/migration
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
