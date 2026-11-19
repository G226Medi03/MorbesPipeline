# MorbesPipeline 문서

MorbesPipeline은 Morbes 의학 문제 풀이 파이프라인을 담당하는 Go 모듈이다.
문제에서 임상 소견을 추출하고 `MedicalEntityRepository`로 저장한 뒤,
상태/행동 그래프에서 원인 후보를 계산해 `solve_cases.answer_json`에 답을 남긴다.

## 문서 지도

| 문서 | 내용 |
| --- | --- |
| [packages.md](packages.md) | 현재 패키지 구조와 책임 |
| [pr-guide.md](pr-guide.md) | 브랜치 생성, 검증, PR 생성 절차 |
| [importlint.md](importlint.md) | 금지 dependency와 direct import 규칙 |
| [plans/2026-11-14-json-encode-removal.md](plans/2026-11-14-json-encode-removal.md) | encode/JSON 제거 계획 |
| [sync/README.md](sync/README.md) | 참조 GitHub repo별 동기화 현황 |
| [pipeline.md](pipeline.md) | solve 노드, checkpoint, replay |
| [extraction-llm.md](extraction-llm.md) | 프롬프트, LLM 캐시, 추출 단계 |
| [domain-model.md](domain-model.md) | Case/Step/LLMCall, typed entity, trait |
| [persistence.md](persistence.md) | SQLite 스키마와 그래프 로딩 |

> `architecture.md`, `pipeline.md`, `extraction-llm.md`, `domain-model.md`,
> `persistence.md`는 리팩터링 이전의 세부 설명을 일부 포함한다. 최신 패키지
> 경계는 [packages.md](packages.md)와 [sync/README.md](sync/README.md)를 기준으로 한다.

## 30초 요약

1. 외부 앱은 루트 패키지 `morbes` 하나만 import한다.
2. `morbes.NewStore(ctx, db)`로 저장소를 만들고 `InitSchema()`를 호출한다.
3. `morbes.NewService(ctx, morbes.Dependencies{...})`로 solve 서비스를 만든다.
4. `Service.Run(ctx, caseID)`가 solve 파이프라인을 실행한다.
5. draft는 typed MedicalEntity로 저장되고, graph 후처리 후 원인 후보를 추적한다.
6. 결과는 `solve_cases.answer_json`에 저장된다.

## 패키지 구조

```text
.
├── doc.go, solve.go, store.go   # 공개 facade (package morbes)
├── domain.go, entity.go
├── tracing.go, support.go
│
└── internal/
    ├── pipeline/                # solve 오케스트레이션 (package csolve)
    ├── persistence/             # SQLite 저장소 (package store)
    │   ├── graph/
    │   ├── solve/
    │   ├── option/
    │   └── benchmark/
    ├── tracing/
    │   ├── graph/               # 인메모리 그래프 (package cgraph)
    │   └── cause/               # 원인 추적 (package ccause)
    ├── llm/                     # LLM 캐시/relation/postprocess (package cservice)
    ├── domain/
    │   ├── solve/               # Case/Step/Stage/Question (package domain)
    │   │   └── model/
    │   ├── procedureoption/
    │   ├── patient/
    │   ├── situation/
    │   ├── label/
    │   └── trait/
    ├── support/
    │   ├── ids/
    │   ├── logs/
    │   └── sliceutil/
    └── testdb/

tools/
└── importlint/                # A→B→C dependency linter
```

## 공개 API 사용

```go
import morbes "github.com/G226Medi03/MorbesPipeline"

store := morbes.NewStore(ctx, db)
if err := store.InitSchema(); err != nil {
    return err
}

svc, err := morbes.NewService(ctx, morbes.Dependencies{
    Store: store,
    Solver: solver,
    ExtractLLM: extractor,
})
if err != nil {
    return err
}

err = svc.Run(ctx, caseID)
```

## 검증

```powershell
go build ./...
go test ./...
go vet ./...
```

```powershell
python D:\Coding\MedicalEntity\line_count.py . go
python D:\Coding\MedicalEntity\line_count.py . go --no-blank --no-comment
```
