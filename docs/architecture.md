# 아키텍처

## 계층

```text
root facade
  -> internal/service/orchestration
  -> internal/service/* helper packages
  -> internal/repository/facade
  -> internal/entity
```

## Dependency direction

```text
internal/entity <- internal/repository/* <- internal/service/* <- root
internal/entity <- internal/service/tracing <- root
MorbesLLMService <- internal/service <- root
MedicalEntityRepository <- internal/repository
MedicalEntityParser <- internal/service
```

예외:

```text
internal/repository/facade -> internal/service/store
internal/repository/solve  -> internal/service/solve
```

금지:

```text
internal/entity -> internal/repository
internal/entity -> internal/service
internal/service -> root
MedicalEntity direct import
```

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
| service/orchestration | case 실행, node, checkpoint, replay |
| service/llm | LLM extraction adapter |
| service/resolve | DTO → Entity helper |
| service/postprocess | deterministic postprocess |
| service/selection | procedure option 선택 |
| service/store | persistence application services |
| repository/facade | public store package |
| repository/store | Store: ctx/db + repositories |
| repository/* | benchmark/entity/solve/postprocess table과 repository |
| service/tracing | reverse link cause trace |
| entity | solve domain/entity key/label |

## 관련 문서

- [project-structure.md](project-structure.md)
- [pipeline.md](pipeline.md)
- [persistence.md](persistence.md)
- [extraction-llm.md](extraction-llm.md)
