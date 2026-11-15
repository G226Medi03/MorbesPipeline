# 참조 GitHub repo 동기화

기준일: 2026-11-15

패키지 분리 후 각 영역은 아래 upstream repo에 대응한다.
동기화는 **위에서 아래 순서**로 하나씩 진행한다.

| 순서 | upstream repo | version | 대응 패키지 | 상태 |
| --- | --- | --- | --- | --- |
| 1 | `MedicalEntity` | `v0.15.0` | `MedicalEntityRepository/repo` facade 경유 | 직접 import 제거 |
| 2 | `MedicalEntityParser` | `v0.1.0` | `MorbesLLmPipeline` facade 경유 | 직접 import 제거 |
| 3 | `SourceEntity` | `v0.3.0` | `internal/support/sourceutil`, `internal/persistence/source.go` | 동기화 완료 |
| 4 | `SourceEntityRepository` | `v0.2.0` | `internal/persistence/source.go` | 동기화 완료 |
| 5 | `SQLCore` | `v0.2.0` | `internal/persistence`, `internal/persistence/*` | 동기화 완료 |
| 6 | `MedicalEntityRepository` | `v0.10.0` | `repo` facade + `internal/persistence/entity*.go`, `internal/persistence/graph`, `internal/persistence/option` | pain/symptom schema 반영 |
| 7 | `PipeLineCore` | `v0.1.0` | `internal/pipeline/pipeline_nodes.go`, `internal/pipeline/service.go` | 동기화 완료 |
| 8 | `llm-client` | `v0.4.0` | `MorbesLLmPipeline` 경유 | 직접 import 제거 |
| 9 | `log-sqlite` | `v0.1.1` | `internal/support/logs` | 동기화 완료 |
| 10 | `go-duration` | `v0.1.0` | `MorbesLLmPipeline` / parser 경유 | 직접 import 제거 |
| 11 | `MorbesLLmPipeline` | `v0.7.0` | LLM 호출, 캐시, relation, actor 추출, replay, parser facade | 동기화 완료 |

## 1. MedicalEntity

- 사용 경로:
  - `MedicalEntityRepository/repo` facade만 사용
  - `MedicalEntity/*` 직접 import 금지
- facade 제공:
  - ename interface/concrete type
  - kind types/constants
  - `InfuseRoute`
  - codec
  - trait constants/helper
- 확인:
  - aggregate ID 제거 대응
  - `trait` 패키지 경로 변경 대응
  - `emedical_procedure.InfuseRoute` 변경 대응
  - `importlint` forbidden 규칙 적용

## 2. MedicalEntityParser

- 사용 API (via `MorbesLLmPipeline`):
  - `ParseStatusResponse`, `ParseStatusesResponse`
  - `ParseActionsResponse`, `ParseActorItemsResponse`
  - `ParseRelationsResponse`, `ParseActorName`
- 대응 패키지:
  - `internal/domain/patient`
  - `internal/domain/situation`
  - `internal/pipeline/draft_parse.go`
  - `internal/pipeline/extraction_v2.go`
- 확인:
  - parser v0.1.0 decode-only API 반영
  - `MedicalEntityParser` 직접 import 0건
  - typed status/action/actor 결과를 draft value로 변환

## 3. SourceEntity / SourceEntityRepository

- 사용 API:
  - `SourceId`, `RespId`, `SourceRef`, `LlmResponse`
  - source row / llm response row
- 대응 패키지:
  - `internal/support/sourceutil`
  - `internal/persistence/source.go`
- 확인:
  - `SourceId`/`RespId` Scan 기반 복원
  - `GetOrInsertLlmResponse`, `RefreshLlmResponse`, `TraceSource` 유지

## 4. SQLCore

- 사용 API:
  - `sqlite.Executor`
  - `sqlite.TableExists`
  - `sqlite.EnsureRenamedColumn`
- 대응 패키지:
  - `internal/persistence`
  - `internal/persistence/graph`
  - `internal/persistence/solve`
  - `internal/persistence/option`
  - `internal/persistence/benchmark`
- 확인:
  - `*sql.DB`/`*sql.Tx` executor 호환
  - `v0.2.0` 사용

## 5. MedicalEntityRepository

- 사용 API:
  - `repo.NewStore`
  - `Store.InsertStatus`, `Store.GetStatus`
  - `Store.InsertAction`, `Store.GetAction`
  - `Store.LinkStatusAction`, `Store.LinkActorAction`
  - `Store.InsertProcedure`, `Store.GetProcedure`
  - `Store.InsertActor`, `Store.GetActor`
  - `Store.GetStatus/GetAction`
  - facade types/constants/codec
- facade:
  - `IEStatus`, `IEAction`, `IEActor`, `IMedicalProcedure`, `IESymptom`, `IPainAspect`
  - concrete status/action/actor/procedure/symptom types
  - `StatusKind`, `ActionKind`, `ActorKind`, `ProcedureKind`
  - `InfuseRoute`
- 대응 패키지:
  - `internal/persistence/entity.go`
  - `internal/persistence/entity_id.go`
  - `internal/persistence/entity_read.go`
  - `internal/persistence/graph.go`
  - `internal/persistence/graph/*`
  - `internal/persistence/option/*`
- 확인:
  - v0.10.0 pain schema 반영 (`symptom_pain`, `symptom_other`)
  - `EPain.Site` + `[]IPainAspect`
  - `repo.ErrNotFound` 사용
  - numerical ID는 `internal/persistence/entity_id.go`에서만 SQL로 조회

## 6. PipeLineCore

- 사용 API:
  - `Pipeline`, `Node`, `State`
  - `Pipeline.Run`, `Pipeline.RunFrom`
- 대응 패키지:
  - `internal/pipeline/pipeline_nodes.go`
  - `internal/pipeline/extraction_pipeline.go`
  - `internal/pipeline/service.go`
- 확인:
  - `core` re-export 패키지 삭제
  - `PipeLineCore` 직접 import

## 7. llm-client

- 사용 API:
  - `llmclient.LlmResponse`
  - `Generator.Generate` / `Model`
- 대응 패키지:
  - `MorbesLLmPipeline`
- 확인:
  - `MorbesPipeline` 직접 import 0건
  - replay generator와 cache mode는 `MorbesLLmPipeline`이 소유
  - live stage override는 `SetReplayLive`로 전달

## 8. log-sqlite

- 사용 API:
  - `Open`, `OpenFromEnv`, `Config`
  - `Logger`, `Attr`, `Level`
- 대응 패키지:
  - `internal/support/logs`
- 확인:
  - `LOG_DB` 기본값 `./db/logs.db`
  - `LLMError` 유지

## 9. go-duration

- 사용 API:
  - parser 내부에서만 사용
- 대응 패키지:
  - `MorbesLLmPipeline` / `MedicalEntityParser` 경유
- 확인:
  - `MorbesPipeline` 직접 import 0건

## 10. MorbesLLmPipeline

- 사용 API:
  - `Service`, `Clients`, `CacheMode`
  - `AskRole`, `ExplainResult`, `ExtractRelations`
  - `BuildActorPrompt`, `ExtractActors`, `ActorDraft`
  - `WithReplay`, `WithReplayLive`
  - parser facade
- 대응 패키지:
  - `internal/pipeline/*`
  - `internal/domain/patient`
  - `internal/domain/situation`
- 확인:
  - `MorbesLLmPipeline v0.7.0`
  - LLM 호출/캐시/relation/actor 추출/replay 소유
  - `Generator`는 `llmclient.LlmResponse` 반환

## 동기화 절차

1. upstream repo의 최신 tag/commit 확인
2. `go doc`로 public API diff 확인
3. 해당 패키지만 수정
4. `go build ./...`
5. `go test ./...`
6. `go vet ./...`
7. 이 문서의 version/status 갱신

