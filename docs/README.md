# MorbesPipeline 문서

MorbesPipeline은 Morbes 의학 문제 풀이 파이프라인을 담당하는 Go 모듈이다.

```text
LLM extraction
  -> DTO
  -> Entity
  -> SQLite
  -> postprocess
  -> cause trace
  -> answer JSON
```

## 문서 지도

| 문서 | 내용 |
| --- | --- |
| [project-structure.md](project-structure.md) | 현재 프로젝트 구조와 package 책임 |
| [architecture.md](architecture.md) | 계층과 dependency direction |
| [pipeline.md](pipeline.md) | solve node, checkpoint, replay |
| [persistence.md](persistence.md) | SQLite 저장 구조 |
| [extraction-llm.md](extraction-llm.md) | LLM Output → DTO → Entity |
| [domain-model.md](domain-model.md) | Case, Step, LLMCall, entity |
| [pr-guide.md](pr-guide.md) | branch, 검증, PR 절차 |
| [importlint.md](importlint.md) | dependency / import 규칙 |
| [sync/README.md](sync/README.md) | upstream repo 동기화 |
| [archive/plans/2026-11-14-json-encode-removal.md](archive/plans/2026-11-14-json-encode-removal.md) | 완료된 encode 제거 계획 |

## 30초 요약

1. 외부 앱은 root package `morbes` 하나만 import한다.
2. `morbes.New(ctx, db)`로 store를 만들고 `InitSchema()`를 호출한다.
3. `morbes.NewService`로 pipeline service를 만든다.
4. `Service.Run(ctx, caseID)`가 solve pipeline을 실행한다.
5. LLM extraction 결과는 DTO를 거쳐 Entity로 저장된다.
6. postprocess 후 reverse link로 cause 후보를 추적한다.

## 현재 구조

```text
.
├── facade.go
├── internal/
│   ├── entity/
│   │   ├── entitykey/
│   │   ├── label/
│   │   ├── patient/
│   │   ├── situation/
│   │   └── solve/
│   │       ├── checkpoint/
│   │       ├── input/
│   │       ├── model/
│   │       └── selection/
│   ├── repository/
│   │   ├── benchmark/
│   │   ├── postprocess/
│   │   ├── schema/
│   │   ├── solve/
│   │   └── store/
│   ├── service/
│   │   ├── llm/
│   │   ├── orchestration/
│   │   ├── postprocess/
│   │   ├── resolve/
│   │   ├── store/
│   │   └── tracing/
│   ├── support/
│   │   ├── logconfig/
│   │   └── sliceutil/
│   └── testutil/
│       └── sqlite.go
├── tools/importlint/
├── tools/entityidlint/
└── docs/
```

자세한 내용은 [project-structure.md](project-structure.md)를 참고한다.

## 검증

```powershell
go build ./...
go vet ./...
go test ./... -count=1
go run ./tools/importlint --config importlint.json -v
go run ./tools/entityidlint
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/lint.ps1
```
