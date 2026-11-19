# LLM Extraction

## 흐름

```text
LLM raw output
  -> MedicalEntityParser
  -> llmdto DTO
  -> MorbesLLMService extraction DTO
  -> MorbesPipeline resolve
  -> MedicalEntity Entity
```

## 책임

| repo | 책임 |
| --- | --- |
| MedicalEntityParser | raw JSON → DTO, DTO → Entity |
| MorbesLLMService | prompt/call/cache/replay, extraction DTO 조립 |
| MorbesPipeline | resolve orchestration, Entity persistence |

## Parser API

```text
DecodeStatusDTOs
DecodeActionDTOs
DecodeSymptomDTO
DecodeProcedureDTO
DecodeActorItemDTOs
DecodeRelationDTOs

StatusFromDTO
ActionFromDTO
ProcedureFromDTO
SymptomFromDTO
ActorFromDTO
AspectFromDTO
```

## Extraction DTO

```text
StatusDraft
ActionDraft
OptionDraft
PatientDescriptionDraft
CaseExtraction
```

## Checkpoint

```text
CaseExtraction DTO
ResolvedCaseExtraction
  -> llmdto 기반 serialization
```

## 관련 문서

- [pipeline.md](pipeline.md)
- [domain-model.md](domain-model.md)
- [sync/README.md](sync/README.md)
