# 참조 GitHub repo 동기화

기준일: 2026-11-19

패키지 분리 후 각 영역은 아래 upstream repo에 대응한다.
동기화는 **위에서 아래 순서**로 하나씩 진행한다.

| 순서 | upstream repo | version | 대응 패키지 | 상태 |
| --- | --- | --- | --- | --- |
| 1 | `MedicalEntity` | `v0.17.0` | `MedicalEntityRepository/repo` facade 경유 | option ename 추가 |
| 2 | `MedicalEntityParser` | `v0.6.1` | `internal/service/llm`, `internal/service/resolve` | DTO decode + Entity 변환 |
| 3 | `SourceEntity` | `v0.4.1` | `internal/repository/facade` | 동기화 완료 |
| 4 | `SourceEntityRepository` | `v0.4.0` | `internal/repository/facade` | Repository-only API |
| 5 | `SQLCore` | `v0.2.0` | `internal/repository/facade`, `internal/repository/*` | 동기화 완료 |
| 6 | `MedicalEntityRepository` | `v0.14.0` | `repo` facade + `internal/repository/facade`, `internal/repository/internal/repository/*` | entity read alias |
| 7 | `PipeLineCore` | `v0.1.1-0.20261116143831-62093f9a17d1` | `internal/service/orchestration` | 동기화 완료 |
| 8 | `llm-client` | `v0.5.0` | `internal/service/orchestration` | `*call.Client` 직접 사용 |
| 9 | `log-sqlite` | `v0.1.1` | `internal/service/support/logconfig` + direct import | 동기화 완료 |
| 10 | `go-duration` | `v0.1.0` | `MorbesLLMService` / parser 경유 | 직접 import 제거 |
| 11 | `MorbesLLMService` | `v0.14.1` | extraction DTO, prompt, cache, replay | 동기화 완료 |

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
  - `DecodeStatusDTOs`
  - `DecodeActionDTOs`
  - `DecodeSymptomDTO`
  - `DecodeProcedureDTO`
  - `DecodeActorItemDTOs`
  - `DecodeRelationDTOs`
  - `StatusFromDTO`, `ActionFromDTO`, `ProcedureFromDTO`
- 대응 패키지:
  - `internal/service/llm`
  - `internal/service/resolve`
- 확인:
  - LLM output -> DTO -> Entity 흐름
  - parser direct import는 pipeline helper에 격리
  - local entityjson codec 제거

## 3. SourceEntity / SourceEntityRepository

- 사용 API:
  - `SourceId`, `RespId`, `SourceRef`, `LlmResponse`
  - source row / llm response row
- 대응 패키지:
  - `internal/repository/facade`
- 확인:
  - `SourceId`/`RespId`는 `SourceEntity.NewSourceId` / `NewRespId` 직접 사용
  - `SourceEntityRepository.Repository` facade만 사용
  - LLM response cache는 `MorbesLLMService/responsestore`에 위임
  - `TraceSource`는 MorbesPipeline에 유지

## 4. SQLCore

- 사용 API:
  - `sqlite.Executor`
  - `sqlite.TableExists`
  - `sqlite.EnsureRenamedColumn`
- 대응 패키지:
  - `internal/repository/facade`
  - `internal/repository/internal/repository/*`
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
  - `Store.InsertActor`, `Store.GetActor`, `Store.GetActorByName`
  - facade types/constants/codec
- facade:
  - `IEStatus`, `IEAction`, `IEActor`, `IMedicalProcedure`, `IESymptom`, `IPainAspect`
  - `Actor`, `Action`, `Status`, `Procedure` entity alias
  - `StatusKind`, `ActionKind`, `ActorKind`, `ProcedureKind`
  - `InfuseRoute`
- 대응 패키지:
  - `internal/repository/facade`
  - `internal/repository/internal/repository/*`
- 확인:
  - actor/action/status/procedure read는 facade entity 반환
  - link 타입은 facade에서 export하지 않음
  - `repo.ErrNotFound` 사용
  - `ActionsResultingInStatus`, `StatusesCausingAction`, `ActorsCausingAction` 기반 cause tracing
  - numerical ID는 `internal/repository/entity/table_id.go`에서만 SQL로 조회

## 6. PipeLineCore

- 사용 API:
  - `Pipeline`, `Node`, `State`
  - `Pipeline.Run`, `Pipeline.RunFrom`
- 대응 패키지:
  - `internal/service/pipeline_nodes.go`
  - `internal/service/extraction_pipeline.go`
  - `internal/service/service.go`
- 확인:
  - `core` re-export 패키지 삭제
  - `PipeLineCore` 직접 import

## 7. llm-client

- 사용 API:
  - `llmclient.LlmResponse`
  - `*cllmcall.Client`
- 대응 패키지:
  - `MorbesLLMService`
  - `internal/service/service.go`
- 확인:
  - `MorbesPipeline`은 `*call.Client`를 직접 사용
  - replay는 `MorbesLLMService`, `CacheMode` 타입은 `PipeLineCore`가 소유
  - live stage override는 `SetReplayLive`로 전달

## 8. log-sqlite

- 사용 API:
  - `Open`, `OpenFromEnv`, `Config`
  - `Logger`, `Attr`, `Level`
- 대응 패키지:
  - `internal/service/support/logconfig`
  - `internal/service` direct import
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
  - `internal/service/*`
  - `internal/repository` (`PromptStore`, `LlmResponseStore`)
- 확인:
  - `MorbesLLMService v0.11.1`
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

