package domain

// ReplayMode selects which stages a replay runs.
type ReplayMode string

const (
	ReplayModeLLM  ReplayMode = "llm"
	ReplayModeCode ReplayMode = "code"
)
