# Persistence

## 구조

```text
internal/store/
  core/
  entity/
  read/
  solve/
  benchmark/
  postprocess/
  migration/
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
  Solve storesolve.Repository
  Postprocess storepostprocess.Repository
```

## 주요 구성

```text
core
  InitSchema
  transaction
  entity ID/link SQL
  entity read는 MedicalEntityRepository facade 위임
  solve/source access

entity
  actor/trait helper

read
  benchmark/result read model types

solve
  solve_cases, solve_steps, solve_llm_calls, answer checks

benchmark
  llm_problem_results

postprocess
  table: morbes_postprocesses schema/CRUD
  repository: validation and record mapping

migration
  legacy schema rename
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
