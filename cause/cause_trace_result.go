package ccause

type CauseTraceResult struct {
	Observed   []string    `json:"observed"`
	Candidates []Candidate `json:"candidates"`
	Missing    []string    `json:"missing"`
}
