# Domain Model

## Solve

```text
Case
Step
Stage
QuestionType
QuestionPolicy
LLMCall
PromptVersion
```

## Entity

Medical entity는 `MedicalEntityRepository/repo` facade를 통해 사용한다.

```text
IEStatus
IEAction
IEActor
IMedicalProcedure
IMedicalProcedureOption
IESymptom
IPainAspect
```

## Identity

```text
status/action/actor/procedure identity -> ename Equal method
medical procedure option identity    -> procedure + category + indications
postprocess audit identity           -> entity DTO payload
case identity                        -> CaseID string
```

Medical entity numeric row identifiers are a SQLite implementation detail.
They are not exposed in public DTOs, repository APIs, or audit records.

## DTO

```text
llmdto.Status
llmdto.Action
llmdto.Procedure
llmdto.Symptom
llmdto.Actor
llmdto.Relation
```

## 관련 문서

- [extraction-llm.md](extraction-llm.md)
- [persistence.md](persistence.md)
