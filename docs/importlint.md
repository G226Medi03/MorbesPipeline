# Import dependency linter

`tools/importlint` enforces the A → B → C dependency rule.

```text
feature package (A)
    -> internal/external/<repo> adapter (B)
        -> external GitHub repo (C)
```

Direct `A -> C` import is forbidden.

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
- `MedicalEntityParser`는 이 repo에서 직접 import할 수 없다.
  - parser facade는 `MorbesLLmPipeline`을 사용한다.
- 아래 외부 모듈은 `internal/external/*` adapter를 통해서만 import할 수 있다.
  - `PipeLineCore`
  - `SourceEntity`
  - `SourceEntityRepository`
  - `SQLCore`
  - `llm-client`
  - `log-sqlite`
- `MedicalEntityRepository`와 `MorbesLLmPipeline`은 직접 import할 수 있다.
- `MedicalEntityParser`는 `MorbesLLmPipeline` 내부에서만 사용한다.
- adapter 패키지 자신만 해당 외부 모듈을 직접 import할 수 있다.
- test import도 검사한다.
- 위반 시 exit code 1.

## Adapter

```text
internal/external/
├── pipelinecore/
├── sourceentity/
├── sourceentityrepository/
├── sqlcore/
└── logsqlite/
```

각 adapter는 외부 repo의 타입/함수를 alias 또는 var로 re-export한다.
기능 패키지는 외부 repo 대신 adapter를 import한다.

## 설정

`importlint.json`:

```json
{
  "repos": [{ "path": ".", "module": "github.com/G226Medi03/MorbesPipeline", "enforce": true }],
  "forbidden": [
    "github.com/G226Medi03/MedicalEntity",
    "github.com/G226Medi03/MedicalEntityParser"
  ],
  "adapters": [
    {
      "external": "github.com/G226Medi03/PipeLineCore",
      "adapter": "github.com/G226Medi03/MorbesPipeline/internal/external/pipelinecore"
    }
  ]
}
```

`repos`에 하위 repo를 추가하면 같은 설정으로 검사할 수 있다.
하위 repo는 자체 adapter 경로가 필요하므로 별도 `importlint.json`을 두는 것을 권장한다.
