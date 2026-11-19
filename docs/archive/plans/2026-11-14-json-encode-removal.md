# Encode/JSON 제거 계획

작성일: 2026-11-14
대상 repo: `D:\Coding\MorbesPipeline`
범위: LLM JSON 경계는 제외하고, 나머지 encode/JSON 코드를 제거한다.

## 1. 목표

MorbesPipeline core에서 아래 요소를 제거한다.

```text
encoding/json import
json.Marshal / Unmarshal
json.MarshalIndent
json.RawMessage
json:"..." 태그
MarshalJSON / UnmarshalJSON
entityjson.Encode* / Decode*
encode_json
encodeTraits
encodeIDs
persistence JSON 컬럼
tools/importlint JSON
```

단, LLM 입력/출력 및 draft 변환에 사용하는 JSON은 유지한다.

## 2. LLM JSON 허용 범위

```text
internal/llm/relation.go
internal/pipeline/extraction_v2.go
internal/pipeline/draft_parse.go
internal/pipeline/pipeline_v2.go
internal/pipeline/service.go
internal/pipeline/drafts.go
```

`MedicalEntityParser` 호출 코드도 허용한다.

## 3. 완료된 작업

- root `MarshalAction` / `MarshalStatus` / `MarshalProcedure` 제거
- `repo.Marshal*` / `repo.Unmarshal*` 사용 제거
- `StatusLabel(kind, json)` 제거
- `NodeLabel(kind, json)` 제거
- `IsStatusKind` root export 제거
- label package의 JSON parsing 제거
- `label` 테스트를 typed value 테스트로 교체
- `internal/external/duration` 제거

## 4. 제거 대상

### 4-1. API/domain JSON

- `internal/domain/patient/description.go`
  - `MarshalJSON`
  - `UnmarshalJSON`
- `internal/domain/procedureoption/option.go`
  - `MarshalJSON`
  - `UnmarshalJSON`
  - `json:"-"` 태그
- `internal/pipeline/resolved.go`
  - `MarshalJSON`
  - `UnmarshalJSON`
- `internal/domain/solve/*`
  - `json` 태그
- `internal/tracing/cause/*`
  - `json` 태그

### 4-2. Entity codec

현재:

```text
internal/domain/entityjson/
```

계획:

- LLM 전용 codec은 `internal/llm/entityjson` 등 LLM 경계로 이동
- LLM 외 사용처 제거
- `internal/domain/entityjson` 삭제

### 4-3. Draft / RawMessage

LLM draft는 예외로 유지한다.

```text
internal/pipeline/drafts.go
```

단, LLM 경계 밖에서 `json.RawMessage`를 사용하는 코드는 제거한다.

### 4-4. Persistence JSON

제거 대상 컬럼:

```text
traits_json
context_traits_json
details_json
indication_status_ids_json
situation_json
answer_json
output_json
analysis_json
```

정규화 방향:

```text
procedure_traits(procedure_id, trait)
option_traits(option_id, trait)
option_context_traits(option_id, trait)
status_symptom_aspects(status_id, aspect_kind, value, measure, name, actor_id)
medical_procedure_option_indications(option_id, status_id)
solve_outputs(case_id, node, field, value)
```

제거 함수:

```text
encodeTraits
encodeIDs
```

### 4-5. `encode_json`

제거:

```text
internal/pipeline/trace.go
  encode_json
```

대체:

- step output을 typed column/detail table로 저장
- `output_json` 컬럼 제거

### 4-6. Tools

`tools/importlint`:

현재:

```text
importlint.json
go list -json
json.NewDecoder
```

변경:

- config: key=value 또는 YAML
- `go list -f '{{.ImportPath}}|{{join .Imports ","}}'`

## 5. 단계별 계획

### Phase 1 — API/domain JSON 제거

- `PatientDescription` JSON method 제거
- `ProcedureOption` JSON method 제거
- `ResolvedCaseExtraction` JSON method 제거
- domain/solve, tracing JSON 태그 제거
- 현재 codec 경계를 유지할 부분과 제거할 부분 분리

### Phase 2 — entityjson 분리

- LLM 전용 codec을 LLM 경계 package로 이동
- `internal/domain/entityjson` 삭제
- patient/resolved/persistence에서 entityjson import 제거

### Phase 3 — persistence 정규화

- traits/context traits JSON → join table
- details JSON → aspect table
- indication IDs JSON → join table
- solve output JSON → typed table

### Phase 4 — `encode_json` 제거

- step output typed 저장
- `internal/pipeline/read.go`의 output JSON decode 제거

### Phase 5 — tool JSON 제거

- importlint config/output 변경

### Phase 6 — 검증

```powershell
rg '"encoding/json"' --glob '*.go' .
rg 'json:"' --glob '*.go' .
rg 'json\.RawMessage|MarshalJSON|UnmarshalJSON' --glob '*.go' .
rg '\b(encode|Encode)[A-Za-z0-9_]*\b' --glob '*.go' .
```

LLM 허용 범위 외 결과는 0이어야 한다.

## 6. Acceptance Criteria

```text
go build ./...                                  통과
go test -count=1 ./...                          통과
go vet ./...                                    통과
go run ./tools/importlint --config importlint.json
  importlint: no violations
```

검색 결과:

- LLM 경계 외 `encoding/json`: 0
- `json` 태그: 0
- `json.RawMessage`: LLM draft 외 0
- `MarshalJSON` / `UnmarshalJSON`: 0
- `encode_json`, `encodeTraits`, `encodeIDs`: 0
- `repo.Marshal*` / `repo.Unmarshal*`: 0

## 7. Blocker

- `entityjson`은 LLM과 non-LLM에서 함께 사용 중이므로 분리 필요
- persistence JSON 컬럼 제거는 DB schema migration 필요
- solve output JSON 제거는 typed output table 필요
- LLM draft의 `json.RawMessage`는 LLM 경계이므로 유지
- MorbesGoDB는 root JSON 대신 typed DTO를 받아 자체 JSON으로 변환해야 함
