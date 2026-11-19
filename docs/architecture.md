# 아키텍처

## 계층

```text
root facade
  -> internal/pipeline/orchestration
  -> internal/pipeline/* helper packages
  -> internal/service
  -> internal/store/facade + store
  -> internal/domain
```

## Dependency direction

```text
internal/support <- internal/domain <- internal/store <- root
internal/support <- internal/domain <- internal/tracing <- root
MorbesLLMService <- internal/pipeline <- root
MedicalEntityRepository <- internal/store
MedicalEntityParser <- internal/pipeline
```

금지:

```text
internal/domain -> store / pipeline / tracing
internal/store -> pipeline
internal/tracing -> store
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
| pipeline/orchestration | case 실행, node, checkpoint, replay |
| pipeline/llm | LLM extraction adapter |
| pipeline/resolve | DTO → Entity helper |
| pipeline/postprocess | deterministic postprocess |
| pipeline/selection | procedure option 선택 |
| store | Store: ctx/db + repositories |
| store/facade | StoreContainer, transaction |
| internal/service | application services |
| store/repository | benchmark/entity/solve/postprocess table과 repository |
| tracing | reverse link cause trace |
| domain | solve/domain/entity key/label/trait |

## 관련 문서

- [project-structure.md](project-structure.md)
- [pipeline.md](pipeline.md)
- [persistence.md](persistence.md)
- [extraction-llm.md](extraction-llm.md)
