package orchestration

import (
	"testing"

	core "github.com/G226Medi03/PipeLineCore"
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

func TestReplayPipelineNodesKeepsPrepare(t *testing.T) {
	pipeline := newSolveNodePipeline(&Service{})
	tests := []struct {
		name  string
		start string
		want  []string
	}{
		{
			name:  "from extraction",
			start: string(extractionNode),
			want:  []string{string(solveLoadNode), string(extractionNode), string(resolveNode), string(causeNode)},
		},
		{
			name:  "from resolve",
			start: string(resolveNode),
			want:  []string{string(solveLoadNode), string(resolveNode), string(causeNode)},
		},
		{
			name:  "from cause",
			start: string(causeNode),
			want:  []string{string(solveLoadNode), string(causeNode)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			replay, err := replayPipelineNodes(pipeline, core.NodeName(test.start))
			if err != nil {
				t.Fatalf("replayPipelineNodes: %v", err)
			}
			if len(replay.Nodes) != len(test.want) {
				t.Fatalf("nodes = %d, want %d", len(replay.Nodes), len(test.want))
			}
			for index, node := range replay.Nodes {
				if string(node.Name) != test.want[index] {
					t.Fatalf("node[%d] = %q, want %q", index, node.Name, test.want[index])
				}
			}
		})
	}
	if _, err := replayPipelineNodes(pipeline, core.NodeName("missing")); err == nil {
		t.Fatal("expected missing node error")
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
