# 아키텍처

## 계층

```text
root facade
  -> internal/service/orchestration
  -> internal/service/* helper packages
  -> internal/service/store
  -> internal/repository/*
  -> internal/entity
```

## Dependency direction

```text
internal/entity <- internal/repository/* <- internal/service/* <- root
MorbesLLMService <- internal/service <- root
MedicalEntityRepository <- internal/service/store
MedicalEntityParser <- internal/service
```

금지:

```text
internal/entity -> internal/repository
internal/entity -> internal/service
internal/repository -> internal/service
internal/service -> root
MedicalEntity direct import
```

`internal/repository/facade`는 제거되었다.
`StoreContainer`는 `internal/service/store`가 소유하고, root facade가 이를 public API로 노출한다.

## 실행 흐름

```text
App
  -> Service.Run
  -> orchestration solve nodes
  -> MorbesLLMService extraction
  -> DTO
  -> resolve DTO -> Entity
  -> MedicalEntityRepository persistence
  -> postprocess
  -> cause trace
  -> answer JSON
```

## 책임

| 영역 | 책임 |
| --- | --- |
| root facade | public API, 전체 조립 |
| entity | solve domain, label, patient, situation, storage key |
| repository/store | Store: ctx/db + repositories |
| repository/* | benchmark/solve/postprocess/schema |
| service/store | StoreContainer + persistence application services + MedicalEntityRepository ename APIs |
| service/orchestration | case 실행, node, checkpoint, replay |
| service/llm | LLM extraction adapter |
| service/resolve | DTO → Entity helper |
| service/postprocess | deterministic postprocess |
| service/tracing | reverse link cause trace |
| support | logging/generic utility |
| testutil | test DB helper |

## 관련 문서

- [project-structure.md](project-structure.md)
- [pipeline.md](pipeline.md)
- [persistence.md](persistence.md)
- [extraction-llm.md](extraction-llm.md)
