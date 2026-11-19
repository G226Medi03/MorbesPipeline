# Pipeline

## 구조

```text
internal/pipeline/
  orchestration/
    pipeline.go
    pipeline_nodes.go
    pipeline_v2.go
    pipeline_state.go
    replay.go
  checkpoint/
  input/
  llm/
  resolve/
  postprocess/
  selection/
```

## Solve nodes

```text
LLMLoadNode
  -> LLMExplainNode
  -> LLMCaseExtractionNode
  -> LLMResolveNode
  -> GraphCauseTraceNode
```

## Case extraction

```text
MorbesLLMService ExtractionRunner
  -> CaseExtraction DTO
  -> checkpoint.CaseExtractionStepOutput
```

## Resolve

```text
extraction DTO
  -> MedicalEntityParser *FromDTO
  -> MedicalEntityRepository
  -> ResolvedCaseExtraction
```

## Cause

```text
observed statuses
  -> ActionsResultingInStatus
  -> StatusesCausingAction
  -> ActorsCausingAction
  -> candidates
```

## Replay

```text
LLM replay/cache 정책 : MorbesLLMService
case reset/node step  : MorbesPipeline
```

## 관련 문서

- [project-structure.md](project-structure.md)
- [persistence.md](persistence.md)
- [extraction-llm.md](extraction-llm.md)
