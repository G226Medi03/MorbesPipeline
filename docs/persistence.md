# Persistence

## 구조

```text
internal/store/
  core/
    internal/
      store/
      service/
      schema/
  repository/
```

## StoreContainer

```text
internal/store/core/store.go
  ctx
  medical db/tx
  source db/tx
  solve db/tx
  Medical *repo.Store
  Source storesource.Repository
  Solve storerepo.SolveRepository
  Benchmark storerepo.BenchmarkRepository
  Postprocess storerepo.PostprocessRepository
```

## 주요 구성

```text
core
  InitSchema
  transaction
  core/internal/store: ctx/db + repositories
  core/internal/service: application services

repository
  Table + Repository
  entity ID/link/read/save
  benchmark/solve/postprocess table과 repository
  ProblemResultRow, ExamQuestionDataCounts
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
