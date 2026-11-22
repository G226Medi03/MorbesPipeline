# MorbesPipeline PR 작성 설명서

이 문서는 MorbesPipeline에서 브랜치를 만들고, 검증하고, PR을 생성하는 절차를 정리한다.

- 대상: MorbesPipeline 저장소 담당자
- 기준: `main` (2026-11-19, PR #15 merge 이후)
- 원칙: **PR 생성까지만 하고 merge는 저장소 소유자가 직접 한다.**

---

## 1. 빠른 절차

```powershell
# 1. 최신 main에서 작업 브랜치 생성
git switch main
git pull --ff-only
git switch -c feat/짧은-설명

# 2. 코드 수정 후 검증
go build ./...
go vet ./...
go test ./... -count=1
go run ./tools/importlint --config importlint.json
go run ./tools/entityidlint
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/lint.ps1

# 3. 커밋 및 push
git add <변경 파일>
git commit -m "feat: 짧은 설명"
git -c credential.helper=wincred push -u origin feat/짧은-설명

# 4. GitHub에서 PR 생성
# base: main
# head: feat/짧은-설명
```

push 후 GitHub가 출력하는 아래 주소로 바로 PR을 만들 수 있다.

```text
https://github.com/G226Medi03/MorbesPipeline/pull/new/<branch>
```

또는 직접 비교 페이지를 연다.

```text
https://github.com/G226Medi03/MorbesPipeline/compare/main...<branch>?expand=1
```

---

## 2. 최초 1회 설정

### Go

```powershell
go version
```

이 저장소는 Go 1.26을 사용한다.

### Git hook

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install-git-hooks.ps1
```

pre-push hook이 설치된다. hook은 push 전에 `scripts/lint.ps1`을 실행한다.

### 의존성 다운로드

```powershell
go mod download
```

private GitHub module이 포함되어 있으므로 Git 인증이 필요하다.

---

## 3. 브랜치 규칙

브랜치 이름은 변경 목적이 드러나게 만든다.

- `feat/기능명`
- `fix/버그명`
- `refactor/리팩터링명`
- `docs/문서명`
- `chore/작업명`

`main`에 직접 commit하거나 push하지 않는다.

---

## 4. 코드 수정 시 확인할 경계

- 외부 앱은 루트 패키지 `morbes`만 import한다.
- `internal/entity`은 `store`, `pipeline`, `tracing`을 import하지 않는다.
- `internal/repository`는 `pipeline`을 import하지 않는다.
- `internal/service/tracing`은 `store`를 import하지 않는다.
- LLM 프롬프트, 호출, 캐시, replay는 `MorbesLLMService`가 소유한다.
- `MedicalEntity` 직접 import는 금지한다. `MedicalEntityRepository/repo` facade를 사용한다.
- `go-duration`은 `MorbesLLMService` 또는 `MedicalEntityParser` 경유로만 사용한다.

상세 규칙은 `docs/project-structure.md`, `docs/importlint.md`를 따른다.

---

## 5. 검증 명령

PR을 올리기 전에 아래 명령을 모두 통과해야 한다.

```powershell
go build ./...
go vet ./...
go test ./... -count=1
go run ./tools/importlint --config importlint.json
go run ./tools/entityidlint
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/lint.ps1
```

변경 파일만 lint하려면:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/lint.ps1 -ChangedOnly
```

---

## 6. 커밋과 push

```powershell
git add <변경 파일>
git commit -m "refactor: 짧은 설명"
git -c credential.helper=wincred push -u origin <branch>
```

권장 커밋 prefix:

- `feat:`
- `fix:`
- `refactor:`
- `docs:`
- `chore:`
- `test:`

---

## 7. PR 본문 작성

PR 본문에는 최소한 아래 항목을 넣는다.

```markdown
## Summary

- 변경 목적
- 주요 변경 사항

## Boundary / dependency changes

- 패키지 경계 변경 여부
- 의존 모듈 버전 변경 여부
- 공개 API 변경 여부

## Verification

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./... -count=1`
- [ ] `go run ./tools/importlint --config importlint.json`
- [ ] `go run ./tools/entityidlint`
- [ ] `scripts/lint.ps1`

## Follow-up

- 후속 PR 또는 소비자 변경 사항
```

`.github/pull_request_template.md`가 있으므로 GitHub에서 PR을 만들면 본문이 자동으로 채워진다.

---

## 8. PR 생성이 안 될 때

### 8.1 `Permission ... denied` / HTTP 403

증상:

```text
remote: Permission to G226Medi03/MorbesPipeline.git denied to G226Medi03.
fatal: ... The requested URL returned error: 403
```

원인:

- Git credential이 없음
- PAT에 `Contents: Read and write`가 없음
- fine-grained PAT의 Repository access에 `MorbesPipeline`이 없음

해결:

```powershell
git -c credential.helper=wincred push -u origin <branch>
```

Windows Credential Manager에 저장된 GitHub 자격증명을 사용한다.

PAT를 새로 만드는 경우:

- Repository access: `MorbesPipeline` 포함
- Contents: **Read and write**
- Pull requests: **Read and write**
- Metadata: **Read**

토큰을 remote URL에 직접 넣지 않는다.

### 8.2 `missing go.sum entry`

증상:

```text
missing go.sum entry for module providing package ...
```

해결:

```powershell
go mod tidy
go build ./...
```

`go.sum`을 직접 편집하지 않는다.

### 8.3 `unknown revision vX.Y.Z`

증상:

```text
reading ... at revision vX.Y.Z: unknown revision vX.Y.Z
```

원인:

- 의존 모듈의 tag가 원격에 없음
- tag가 로컬에만 있음

해결:

```powershell
git -C <dependency-repo> tag --list "vX.Y.Z"
git -C <dependency-repo> push origin refs/tags/vX.Y.Z
git -C <dependency-repo> ls-remote --tags origin
```

Go module은 **원격에 push된 tag**만 인식한다.

실제 사례:

- `SourceEntityRepository v0.3.1`이 `go.mod`에는 있었지만 원격 tag가 없어서 `go mod tidy`가 실패했다.
- 로컬 release commit의 `v0.3.1` tag를 원격에 push한 뒤 해결했다.

### 8.4 pre-push lint 실패

증상:

```text
0 issues.
error: failed to push some refs
```

또는 lint 오류가 출력된다.

해결:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/lint.ps1
```

lint 오류를 수정한 뒤 다시 push한다. `git push --no-verify`로 우회하지 않는다.

### 8.5 `A pull request already exists`

같은 head 브랜치로 이미 열린 PR이 있을 때 발생한다.

해결:

- 기존 PR에 commit을 추가해서 update한다.
- 또는 새 브랜치를 만들어 새 PR을 생성한다.

### 8.6 `No commits between main and <branch>`

브랜치에 main과 비교할 새 commit이 없다는 뜻이다.

해결:

```powershell
git status
git log --oneline main..HEAD
```

commit이 없으면 변경 사항을 commit한 뒤 다시 push한다.

### 8.7 protected branch push 거부

`main`에 직접 push하지 않는다. 작업 브랜치에서 PR을 만든다.

---

## 9. Merge 정책

- PR 생성은 담당자가 한다.
- 리뷰와 merge는 저장소 소유자가 한다.
- 자동 merge를 켜지 않는다.
- 에이전트가 `merge_pull_request`를 호출하지 않는다.

---

## 10. PR 전 최종 체크리스트

- [ ] `main` 최신화 후 작업 브랜치 생성
- [ ] `go build ./...` 통과
- [ ] `go vet ./...` 통과
- [ ] `go test ./... -count=1` 통과
- [ ] importlint 통과
- [ ] `scripts/lint.ps1` 통과
- [ ] `go.sum` 갱신 완료
- [ ] 의존 모듈 tag가 원격에 존재
- [ ] 패키지 경계 위반 없음
- [ ] 공개 API 변경 시 PR 본문에 명시
- [ ] 문서 갱신
- [ ] 비밀번호·토큰·로컬 경로 커밋 없음
- [ ] PR 본문에 Verification 포함
- [ ] merge는 소유자에게 요청
