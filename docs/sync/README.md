# 참조 GitHub repo 동기화

기준일: 2026-11-17

패키지 분리 후 각 영역은 아래 upstream repo에 대응한다.
동기화는 **위에서 아래 순서**로 하나씩 진행한다.

| 순서 | upstream repo | version | 대응 패키지 | 상태 |
| --- | --- | --- | --- | --- |
| 1 | `MedicalEntity` | `v0.15.0` | `MedicalEntityRepository/repo` facade 경유 | 직접 import 제거 |
| 2 | `MedicalEntityParser` | `v0.4.0` | `internal/domain/entityjson` | parser 경계 격리 |
| 3 | `SourceEntity` | `v0.3.0` | `internal/store/source.go` | 동기화 완료 |
| 4 | `SourceEntityRepository` | `v0.2.0` | `internal/store/source.go` | 동기화 완료 |
| 5 | `SQLCore` | `v0.2.0` | `internal/store`, `internal/store/*` | 동기화 완료 |
| 6 | `MedicalEntityRepository` | `v0.10.1` | `repo` facade + `internal/store/entity*.go`, `internal/store/graph`, `internal/store/option` | pain/symptom schema 반영 |
| 7 | `PipeLineCore` | `v0.1.1-0.20261116143831-62093f9a17d1` | `internal/pipeline/pipeline_nodes.go`, `internal/pipeline/service.go` | 동기화 완료 |
| 8 | `llm-client` | `v0.5.0` | `internal/pipeline` | `*call.Client` 직접 사용 |
| 9 | `log-sqlite` | `v0.1.1` | `internal/support/logconfig` + direct import | 동기화 완료 |
| 10 | `go-duration` | `v0.1.0` | `MorbesLLMService` / parser 경유 | 직접 import 제거 |
| 11 | `MorbesLLMService` | `v0.10.1` | prompt 소유, typed extraction, replay | 동기화 완료 |

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

- 사용 API:
  - `DecodeStatuses`
  - `DecodeActions`
- 대응 패키지:
  - `internal/domain/entityjson`
- 확인:
  - parser direct import는 `entityjson`에만 격리
  - `internal/pipeline` parser direct import 0건
  - typed status/action 결과를 domain 값으로 변환

## 3. SourceEntity / SourceEntityRepository

- 사용 API:
  - `SourceId`, `RespId`, `SourceRef`, `LlmResponse`
  - source row / llm response row
- 대응 패키지:
  - `internal/store/source.go`
  - `internal/store/response_store.go`
- 확인:
  - `SourceId`/`RespId`는 `SourceEntity.NewSourceId` / `NewRespId` 직접 사용
  - LLM response cache는 `MorbesLLMService/responsestore`에 위임
  - `TraceSource`는 MorbesPipeline에 유지

## 4. SQLCore

- 사용 API:
  - `sqlite.Executor`
  - `sqlite.TableExists`
  - `sqlite.EnsureRenamedColumn`
- 대응 패키지:
  - `internal/store`
  - `internal/store/graph`
  - `internal/store/solve`
  - `internal/store/option`
  - `internal/store/benchmark`
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
  - `internal/store/entity.go`
  - `internal/store/entity_id.go`
  - `internal/store/entity_read.go`
  - `internal/store/graph.go`
  - `internal/store/graph/*`
  - `internal/store/option/*`
- 확인:
  - v0.10.1 pain schema 반영 (`symptom_pain`, `symptom_other`)
  - `EPain.Site` + `[]IPainAspect`
  - `repo.ErrNotFound` 사용
  - numerical ID는 `internal/store/entity_id.go`에서만 SQL로 조회

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
  - `*cllmcall.Client`
- 대응 패키지:
  - `MorbesLLMService`
  - `internal/pipeline/service.go`
- 확인:
  - `MorbesPipeline`은 `*call.Client`를 직접 사용
  - replay는 `MorbesLLMService`, `CacheMode` 타입은 `PipeLineCore`가 소유
  - live stage override는 `SetReplayLive`로 전달

## 8. log-sqlite

- 사용 API:
  - `Open`, `OpenFromEnv`, `Config`
  - `Logger`, `Attr`, `Level`
- 대응 패키지:
  - `internal/support/logconfig`
  - `internal/pipeline` direct import
- 확인:
  - `LOG_DB` 기본값 `./db/logs.db`
  - logging re-export wrapper 제거
  - 앱 전용 log DB 설정만 `logconfig`에 유지

## 9. go-duration

- 사용 API:
  - parser 내부에서만 사용
- 대응 패키지:
  - `MorbesLLMService` / `MedicalEntityParser` 경유
- 확인:
  - `MorbesPipeline` 직접 import 0건

## 10. MorbesLLMService

- 사용 API:
  - `New`, `Service`, `PromptStore`, `LlmResponseStore`
  - `ExplainResult`
  - `ExtractQuestionType`, `ExtractPatientDescription`, `ExtractMedicalProcedureOption`
  - `ExtractStatuses`, `ExtractActions`, `ExtractActors`, `ExtractRelations`
  - `QuestionTypeResult`, `PatientDescriptionResult`, `MedicalProcedureOptionResult`
  - `ActorDraft`, `WithReplayLive`
- 대응 패키지:
  - `internal/pipeline/*`
  - `internal/store` (`PromptStore`, `LlmResponseStore`)
- 확인:
  - `MorbesLLMService v0.10.1`
  - prompt 조회/렌더링, LLM 호출, 파싱, replay 소유
  - `CacheMode`/`ReplayCall`은 `PipeLineCore` 소유

## 동기화 절차

1. upstream repo의 최신 tag/commit 확인
2. `go doc`로 public API diff 확인
3. 해당 패키지만 수정
4. `go build ./...`
5. `go test ./...`
6. `go vet ./...`
7. 이 문서의 version/status 갱신

