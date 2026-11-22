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
medical procedure option identity    -> option key
case identity                        -> CaseID string
```

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
