# 패키지 구조

## 공개 패키지

외부 앱은 루트 패키지 하나만 import한다.

```go
import morbes "github.com/G226Medi03/MorbesPipeline"
```

| 파일 | 내용 |
| --- | --- |
| `facade.go` | entity/domain 타입 alias, trait helper |
| `solve.go` | `Service`, `Dependencies`, `Run`, replay, prompt, option 선택 |
| `store.go` | `Store`, `New`, `NewWithDBs`, read model |
| `support.go` | IDs, logs, label, sliceutil |
| `tracing.go` | `Container`, `Candidate`, `TraceNode`, cause service |

공개 facade는 내부 구현을 **type alias + 얇은 wrapper**로 노출한다.
root package 선언은 `package morbes_pipeline`이며, 외부 앱은 `morbes` alias를 권장한다.

## internal 패키지

| 경로 | 패키지 | 역할 |
| --- | --- | --- |
| `internal/pipeline/` | `csolve` | solve 오케스트레이션, extraction, resolve, replay |
| `internal/store/` | `store` | SQLite 저장소, schema, transaction |
| `internal/store/graph/` | `graph` | 그래프 row loader |
| `internal/store/solve/` | `solve` | solve 테이블 |
| `internal/store/option/` | `medicalprocedureoption` | option 테이블 repository |
| `internal/store/benchmark/` | `benchmarkresult` | benchmark 테이블 |
| `internal/tracing/graph/` | `cgraph` | 인메모리 상태/행동 그래프 |
| `internal/tracing/cause/` | `ccause` | 원인 추적 |
| `internal/domain/entityjson/` | `entityjson` | typed medical entity JSON codec |
| `internal/domain/entitykey/` | `entitykey` | typed entity identity key |
| `internal/domain/solve/` | `domain` | Case/Step/Stage/Question |
| `internal/domain/solve/model/` | `model` | LLMCall/PromptVersion |
| `internal/domain/procedureoption/` | `medicalprocedureoption` | option DTO |
| `internal/domain/patient/` | `patient` | PatientDescription |
| `internal/domain/situation/` | `csituation` | situation → status |
| `internal/domain/label/` | `centitylabel` | label |
| `internal/domain/trait/` | `trait` | trait 규칙 |
| `internal/support/ids/` | `ids` | DB row ID |
| `internal/support/logconfig/` | `logconfig` | log DB 설정 |
| `internal/support/sliceutil/` | `sliceutil` | slice helper |

| `internal/testdb/` | `testdb` | 테스트 SQLite |
| `tools/importlint/` | `main` | A→B→C dependency linter |

## 의존 방향

```text
internal/support ← internal/domain ← internal/store ← root
internal/support ← internal/domain ← internal/tracing     ← root
MorbesLLMService ← internal/pipeline ← root
```

- `internal/domain`은 `store`/`pipeline`/`tracing`을 import하지 않는다.
- `internal/store`는 `pipeline`을 import하지 않는다.
- `internal/tracing`은 `store`를 import하지 않는다.
- LLM 호출, 프롬프트, 캐시, explain, actor/status/action/relation/patient/option 추출, replay는 `MorbesLLMService`가 소유한다.
- `llm-client/call`은 `MorbesLLMService`의 client 타입을 위해 `internal/pipeline`에서 직접 사용한다.
- 루트 `morbes`만 전체를 조립한다.

## Store 공개 범위

`morbes.Store`는 `internal/store.Store`의 alias다.  
`internal/store.Store`는 `StoreContainer`의 alias다.

```go
type Store = StoreContainer
```

`StoreContainer`는 store들을 sibling field로 가진다.

```go
type StoreContainer struct {
    Medical  *repo.Store
    Response *responsestore.Store
    Option   option.Repository
    Source   storesource.Repository
    Solve    storesolve.Repository
    Graph    graph.Repository
}
```

`morbes.StoreContainer`도 public facade에 노출한다.

외부에는 메서드만 노출된다.

## 참조 repo 동기화

[`sync/README.md`](sync/README.md) 참고.
