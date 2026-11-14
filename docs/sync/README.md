# 참조 GitHub repo 동기화

기준일: 2026-11-14

패키지 분리 후 각 영역은 아래 upstream repo에 대응한다.
동기화는 **위에서 아래 순서**로 하나씩 진행한다.

| 순서 | upstream repo | version | 대응 패키지 | 상태 |
| --- | --- | --- | --- | --- |
| 1 | `MedicalEntity` | `v0.13.1-0.20261113140922-8243ad87432a` | `MedicalEntityRepository/repo` facade 경유 | 직접 import 제거 |
| 2 | `MedicalEntityParser` | `v0.0.0-20261113142150-8b1d93272249` | `internal/domain/patient`, `internal/domain/situation`, `internal/pipeline/*` | 동기화 완료 |
| 3 | `SourceEntity` | `v0.2.0` | `internal/support/sourceutil`, `internal/persistence/source.go` | 동기화 완료 |
| 4 | `SourceEntityRepository` | `v0.2.0` | `internal/persistence/source.go` | 동기화 완료 |
| 5 | `SQLCore` | `v0.2.0` | `internal/persistence`, `internal/persistence/*` | 동기화 완료 |
| 6 | `MedicalEntityRepository` | `v0.9.0-2-g355c4b9` | `repo` facade + `internal/persistence/entity*.go`, `internal/persistence/graph`, `internal/persistence/option` | facade 확장 반영 |
| 7 | `PipeLineCore` | `v0.1.0` | `internal/pipeline/pipeline_nodes.go`, `internal/pipeline/service.go` | 동기화 완료 |
| 8 | `llm-client` | `v0.3.0` | `internal/llm/*`, `internal/pipeline/replay_client.go` | 동기화 완료 |
| 9 | `log-sqlite` | `v0.1.1` | `internal/support/logs` | 동기화 완료 |
| 10 | `go-duration` | `v0.1.0` | `internal/domain/label` | 동기화 완료 |

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
  - `parser.ParseStatus`, `parser.NormalizeActionValue`, `parser.ParseActor`, `parser.ParseActorWithKind`
  - `llmparser.ParseStatuses`, `llmparser.ParseActions`, `llmparser.ParseActorItems`, `llmparser.ParseRelations`
- 대응 패키지:
  - `internal/domain/patient`
  - `internal/domain/situation`
  - `internal/pipeline/draft_parse.go`
  - `internal/pipeline/extraction_v2.go`
  - `internal/llm/relation.go`
- 확인:
  - `ParseLLMJSONArray` 시그니처 변경은 미사용
  - `ActionDraft.Sources` 제거는 미사용
  - status 정규화 확장 반영

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
  - `IEStatus`, `IEAction`, `IEActor`, `IMedicalProcedure`, `IESymptom`, `IAspect`
  - concrete status/action/actor/procedure/symptom types
  - `StatusKind`, `ActionKind`, `ActorKind`, `ProcedureKind`
  - `InfuseRoute`
  - `Marshal*` / `Unmarshal*`
- 대응 패키지:
  - `internal/persistence/entity.go`
  - `internal/persistence/entity_id.go`
  - `internal/persistence/entity_read.go`
  - `internal/persistence/graph.go`
  - `internal/persistence/graph/*`
  - `internal/persistence/option/*`
- 확인:
  - v0.9.0의 ID/codec 계층 제거 대응
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
  - `cllmcall.Client`
  - `Generate`, `Model`, `NewFunc`
- 대응 패키지:
  - `internal/llm/llm_ask.go`
  - `internal/llm/relation.go`
  - `internal/pipeline/replay_client.go`
- 확인:
  - replay client `NewFunc` 유지
  - cache mode와 live stage override 유지

## 8. log-sqlite

- 사용 API:
  - `Open`, `OpenFromEnv`, `Config`
  - `Logger`, `Attr`, `Level`
- 대응 패키지:
  - `internal/support/logs`
- 확인:
  - `LOG_DB` 기본값 `./db/logs.db`
  - `LLMError`, `PreprocessChanges` 유지

## 9. go-duration

- 사용 API:
  - `Unmarshal`, `Format`
- 대응 패키지:
  - `internal/domain/label`
- 확인:
  - `post_procedure_period` 라벨 처리 유지

## 동기화 절차

1. upstream repo의 최신 tag/commit 확인
2. `go doc`로 public API diff 확인
3. 해당 패키지만 수정
4. `go build ./...`
5. `go test ./...`
6. `go vet ./...`
7. 이 문서의 version/status 갱신

