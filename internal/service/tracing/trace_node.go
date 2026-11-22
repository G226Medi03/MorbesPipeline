package tracing

type TraceNode struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Label string `json:"label"`
}
