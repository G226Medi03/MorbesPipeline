# 패키지 구조

## 공개 패키지

외부 앱은 루트 패키지 하나만 import한다.

```go
import morbes "github.com/G226Medi03/MorbesPipeline"
```

| 파일 | 내용 |
| --- | --- |
| `doc.go` | `package morbes` 문서 |
| `solve.go` | `Service`, `Dependencies`, `Run`, replay, prompt, option 선택 |
| `store.go` | `Store`, `NewStore`, `NewStoreWithDBs` |
| `domain.go` | Case/Step/Question/Option/Trait 타입 |
| `entity.go` | MedicalEntity ename 타입 alias |
| `tracing.go` | `Container`, `Candidate`, `TraceNode`, cause service |
| `support.go` | IDs, logs, label, sourceutil, sliceutil |

공개 facade는 내부 구현을 **type alias + 얇은 wrapper**로 노출한다.

## internal 패키지

| 경로 | 패키지 | 역할 |
| --- | --- | --- |
| `internal/pipeline/` | `csolve` | solve 오케스트레이션, extraction, resolve, replay |
| `internal/persistence/` | `store` | SQLite 저장소, schema, transaction |
| `internal/persistence/graph/` | `graph` | 그래프 row loader |
| `internal/persistence/solve/` | `solve` | solve 테이블 |
| `internal/persistence/option/` | `medicalprocedureoption` | option 테이블 repository |
| `internal/persistence/benchmark/` | `benchmarkresult` | benchmark 테이블 |
| `internal/tracing/graph/` | `cgraph` | 인메모리 상태/행동 그래프 |
| `internal/tracing/cause/` | `ccause` | 원인 추적 |
| `internal/llm/` | `cservice` | LLM 캐시, relation, postprocess |
| `internal/domain/solve/` | `domain` | Case/Step/Stage/Question |
| `internal/domain/solve/model/` | `model` | LLMCall/PromptVersion |
| `internal/domain/procedureoption/` | `medicalprocedureoption` | option DTO |
| `internal/domain/patient/` | `patient` | PatientDescription |
| `internal/domain/situation/` | `csituation` | situation → status |
| `internal/domain/label/` | `centitylabel` | label |
| `internal/domain/trait/` | `trait` | trait 규칙 |
| `internal/support/ids/` | `ids` | DB row ID |
| `internal/support/logs/` | `logs` | log-sqlite facade |
| `internal/support/sourceutil/` | `sourceutil` | SourceId 변환 |
| `internal/support/sliceutil/` | `sliceutil` | slice helper |
| `internal/external/*` | adapter | 외부 repo re-export |
| `internal/testdb/` | `testdb` | 테스트 SQLite |
| `tools/importlint/` | `main` | A→B→C dependency linter |

## 의존 방향

```text
internal/support ← internal/domain ← internal/persistence ← root
internal/support ← internal/domain ← internal/tracing     ← root
internal/support ← internal/domain ← internal/llm         ← root
```

- `internal/domain`은 `persistence`/`pipeline`/`tracing`/`llm`을 import하지 않는다.
- `internal/persistence`는 `llm`/`pipeline`을 import하지 않는다.
- `internal/tracing`은 `persistence`를 import하지 않는다.
- 루트 `morbes`만 전체를 조립한다.

## Store 공개 범위

`morbes.Store`는 `internal/persistence.Store`의 alias다.

`persistence.Store`의 repository field는 unexported로 변경했다.

```go
type Store struct {
    // internal fields only
}
```

외부에는 메서드만 노출된다.

## 참조 repo 동기화

[`sync/README.md`](sync/README.md) 참고.
