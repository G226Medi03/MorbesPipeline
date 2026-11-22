# MorbesPipeline Workflow

이 문서는 MorbesPipeline에서 작업하는 에이전트와 담당자를 위한 실행 규칙이다.  
버전과 패키지 구조는 항상 실제 `go.mod`와 코드를 기준으로 확인한다.

## 1. Repository ownership

MorbesPipeline owns:

```text
solve pipeline orchestration
pipeline domain DTOs
SQLite persistence and read models
external repository integration
root public facade
postprocess audit store
importlint rules
```

MorbesPipeline does not own:

```text
generic pipeline model          -> PipeLineCore
LLM prompt, call, parser, cache, replay -> MorbesLLMService
LLM response cache              -> MorbesLLMService/responsestore
medical entity types/codec      -> MedicalEntityRepository facade
LLM response parser             -> MedicalEntityParser
source/response row model       -> SourceEntity, SourceEntityRepository
SQL utilities                   -> SQLCore
```

Do not reimplement code that an upstream repository already provides.

## 2. Package boundaries

```text
internal/entity <- internal/repository/* <- internal/service/* <- root
internal/entity <- internal/service/tracing <- root
MorbesLLMService <- internal/service <- root
```

Rules:

```text
internal/entity must not import internal/repository or internal/service.
internal/repository must not import internal/service.
internal/service must not import root.
LLM call, prompt, cache, explain, extraction, replay belong to MorbesLLMService.
llm-client/call is used directly by internal/service.
Only root assembles the full application.
MedicalEntity direct import is forbidden. Use MedicalEntityRepository/repo facade.
go-duration is used only through MorbesLLMService or MedicalEntityParser.
```

`internal/repository/facade` no longer exists.
`StoreContainer` is owned by `internal/service/store`.
`internal/repository/medicalentity` is the medical entity persistence adapter.
`internal/persistence` no longer exists. Use `internal/repository`.

`tools/importlint` checks these boundaries. Fix violations before creating a PR.

## 3. Branch

Create a branch from latest main.

```powershell
git switch main
git pull --ff-only
git switch -c feat/short-description
```

Allowed prefixes:

```text
feat/
fix/
refactor/
chore/
docs/
test/
```

## 4. Local setup

First clone:

```powershell
git clone https://github.com/G226Medi03/MorbesPipeline.git
cd MorbesPipeline
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install-git-hooks.ps1
go mod download
```

Private modules require GitHub authentication. Use Git credential manager.  
Do not put a token directly in a remote URL.

## 5. Verification

Run before every commit/push:

```powershell
go build ./...
go vet ./...
go test ./... -count=1
go run ./tools/importlint --config importlint.json -v
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/lint.ps1
```

Changed-files lint:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/lint.ps1 -ChangedOnly
```

Lint runs locally. There is no GitHub Actions CI for this repository.

## 6. Commit

Use Conventional Commits.

```text
feat:
fix:
refactor:
chore:
docs:
test:
```

Stage only intended files.

```powershell
git add <files>
git commit -m "refactor: short description"
```

## 7. Push and PR

Push:

```powershell
git push -u origin <branch>
```

The pre-push hook runs `scripts/lint.ps1`. Do not use `git push --no-verify` except in an emergency approved by the owner.

Create a PR only. Never merge.

PR body template:

```markdown
## Summary

## Boundary / dependency change

## Verification

## Follow-up
```

PR rules:

```text
base: main
head: working branch
title: state the purpose
merge: owner only
```

Do not call `merge_pull_request`, enable auto-merge, or merge locally into main.

## 8. Dependency version upgrade

When upgrading a package:

```text
1. Check latest upstream tag.
2. Review public API diff with go doc or source.
3. Update go.mod version.
4. go mod tidy
5. go build ./...
6. go test ./... -count=1
7. go vet ./...
8. go run ./tools/importlint --config importlint.json -v
9. scripts/lint.ps1
10. Update docs/sync/README.md versions/status.
11. Create PR.
```

Prefer one dependency stack per PR. Do not mix unrelated dependency upgrades.

## 9. Dependency release/tag procedure

Go modules recognize only tags pushed to the remote.

```powershell
cd <dependency-repo>
git switch main
git pull --ff-only
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
git ls-remote --tags origin
```

Then in MorbesPipeline:

```powershell
go mod tidy
go build ./...
```

Rules:

```text
Do not rewrite an already pushed tag to another commit.
Do not hand-edit go.sum.
If a tag is local-only, push it before go mod tidy.
```

## 10. Common failures

### missing go.sum entry

```text
go mod tidy
go build ./...
```

If the dependency tag is missing remotely, push the tag first.

### unknown revision vX.Y.Z

```powershell
git -C <dependency-repo> tag --list "vX.Y.Z"
git -C <dependency-repo> push origin refs/tags/vX.Y.Z
git -C <dependency-repo> ls-remote --tags origin
```

### HTTP 403 / Permission denied

Check:

```text
Git credential is configured.
PAT has Contents: Read and write.
PAT has Pull requests: Read and write.
Fine-grained PAT includes the target repository.
```

Use credential manager, not a token in the remote URL.

### importlint violation

```powershell
go run ./tools/importlint --config importlint.json -v
```

Replace forbidden imports with the proper upstream facade.

### package boundary violation

Check dependency direction. Move shared types into `internal/entity`. Assemble only at root.

## 11. PR checklist

```text
[ ] go build ./...
[ ] go vet ./...
[ ] go test ./... -count=1
[ ] importlint
[ ] scripts/lint.ps1
[ ] go.sum updated
[ ] dependency tags exist remotely
[ ] no package boundary violation
[ ] public API impact described
[ ] docs updated
[ ] no secrets/tokens/local paths committed
[ ] Verification section included
[ ] merge requested from owner
```

## 12. References

```text
README.md
docs/README.md
docs/packages.md
docs/importlint.md
docs/pipeline.md
docs/extraction-llm.md
docs/domain-model.md
docs/persistence.md
docs/sync/README.md
```

If documentation conflicts with code, `go.mod` and actual code win. Update docs in the same PR.
