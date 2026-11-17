# Import dependency linter

`tools/importlint`는 금지 dependency를 검사한다.

```text
feature package
    -> external GitHub repo (direct import)
```

adapter 계층은 더 이상 사용하지 않는다.

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
- 아래 외부 모듈은 직접 import할 수 있다.
  - `MedicalEntityParser`
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

## 설정

`importlint.json`:

```json
{
  "repos": [{ "path": ".", "module": "github.com/G226Medi03/MorbesPipeline", "enforce": true }],
  "forbidden": [
    "github.com/G226Medi03/MedicalEntity"
  ],
  "adapters": []
}
```

`repos`에 하위 repo를 추가하면 같은 설정으로 검사할 수 있다.
