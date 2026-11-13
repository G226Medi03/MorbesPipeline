package ccause

type TraceNode struct {
	Type  string `json:"type"`
	ID    int64  `json:"id"`
	Key   string `json:"key,omitempty"`
	Label string `json:"label"`
}
