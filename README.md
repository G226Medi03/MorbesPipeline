# MorbesPipeline

Morbes medical pipeline module.

This repository owns:

- pipeline solve orchestration
- pipeline domain DTOs
- SQLite persistence and read models
- external repository adapters
- root public facade for consumer applications

The generic pipeline model lives in `github.com/G226Medi03/PipeLineCore`.

## Node construction

Use `core.NewNode` through the `internal/external/pipelinecore` adapter.
Direct `core.Node` composite literals are rejected by the
`nodeconstructor` golangci-lint module plugin.

## Lint

Lint runs locally, not in GitHub Actions. Run it before every push:

```powershell
scripts/lint.ps1
```

Enable the repository pre-push hook once per clone:

```powershell
scripts/install-git-hooks.ps1
```

The hook runs the same lint script, so a failing lint blocks the push.

See [docs/README.md](docs/README.md) for the code structure and pipeline documentation.
