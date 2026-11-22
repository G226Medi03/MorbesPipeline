package orchestration

import (
	"testing"
)

func TestSolveNodePipelineOrder(t *testing.T) {
	pipeline := newSolveNodePipeline(&Service{})
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	want := []string{
		string(solveLoadNode),
		string(explainNode),
		string(extractionNode),
		string(resolveNode),
		string(causeNode),
	}
	if len(pipeline.Nodes) != len(want) {
		t.Fatalf("nodes = %d, want %d", len(pipeline.Nodes), len(want))
	}
	for index, node := range pipeline.Nodes {
		if string(node.Name) != want[index] {
			t.Fatalf("node[%d] = %q, want %q", index, node.Name, want[index])
		}
	}
}

func TestNodeForStage(t *testing.T) {
	tests := map[string]string{
		"":                 "",
		NodeExplain:        string(explainNode),
		NodeCaseExtraction: string(extractionNode),
		NodeResolve:        string(resolveNode),
		NodeCause:          string(causeNode),
	}
	for stage, want := range tests {
		if got := string(nodeForStage(stage)); got != want {
			t.Fatalf("nodeForStage(%q) = %q, want %q", stage, got, want)
		}
	}
}
