# Import dependency linter

`tools/importlint`는 금지 dependency와 adapter 경계를 검사한다.

```text
feature package
    -> external GitHub repo (direct import allowed)
```

예외적으로 `MedicalEntityParser`는 adapter를 통해서만 사용한다.

```text
feature package
    -> internal/external/medicalentityparser
        -> MedicalEntityParser
```

## 실행

```powershell
go run ./tools/importlint --config importlint.json
```

Report only:

```powershell
go run ./tools/importlint --config importlint.json --report-only
```

## 규칙

- `MedicalEntity`는 이 repo에서 직접 import할 수 없다.
  - facade는 `MedicalEntityRepository/repo`를 사용한다.
- `MedicalEntityParser`는 `internal/external/medicalentityparser`를 통해서만 사용한다.
- 아래 외부 모듈은 직접 import할 수 있다.
  - `PipeLineCore`
  - `SourceEntity`
  - `SourceEntityRepository`
  - `SQLCore`
  - `log-sqlite`
  - `llm-client`
- `go-duration`은 `MorbesLLMService` / `MedicalEntityParser` 경유로만 사용한다.
- `MedicalEntityRepository`와 `MorbesLLMService`는 직접 import할 수 있다.
- test import도 검사한다.
- 위반 시 exit code 1.

## Adapter

```text
internal/external/
└── medicalentityparser/
```

`medicalentityparser`는 외부 타입을 alias로 re-export하지 않는다.  
raw 응답을 typed entity로 변환하는 실제 wrapper 함수만 제공한다.

## 설정

`importlint.json`:

```json
{
  "repos": [{ "path": ".", "module": "github.com/G226Medi03/MorbesPipeline", "enforce": true }],
  "forbidden": [
    "github.com/G226Medi03/MedicalEntity"
  ],
  "adapters": [
    {
      "external": "github.com/G226Medi03/MedicalEntityParser",
      "adapter": "github.com/G226Medi03/MorbesPipeline/internal/external/medicalentityparser"
    }
  ]
}
```

`repos`에 하위 repo를 추가하면 같은 설정으로 검사할 수 있다.
