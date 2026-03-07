package d2vision

import (
	"strings"
	"testing"
)

func TestDescribeSummary(t *testing.T) {
	tests := []struct {
		name     string
		diagram  *Diagram
		contains []string
	}{
		{
			name: "empty diagram",
			diagram: &Diagram{
				Nodes: []Node{},
				Edges: []Edge{},
			},
			contains: []string{"0 nodes", "0 connections"},
		},
		{
			name: "single node",
			diagram: &Diagram{
				Nodes: []Node{{ID: "a"}},
				Edges: []Edge{},
			},
			contains: []string{"1 node", "0 connections"},
		},
		{
			name: "single edge",
			diagram: &Diagram{
				Nodes: []Node{{ID: "a"}, {ID: "b"}},
				Edges: []Edge{{Source: "a", Target: "b"}},
			},
			contains: []string{"2 nodes", "1 connection"},
		},
		{
			name: "with container",
			diagram: &Diagram{
				Nodes: []Node{
					{ID: "container", Children: []string{"child"}},
					{ID: "child", Parent: "container"},
				},
				Edges: []Edge{},
			},
			contains: []string{"2 nodes", "1 container"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.diagram.DescribeSummary()
			for _, want := range tt.contains {
				if !strings.Contains(summary, want) {
					t.Errorf("DescribeSummary() = %q, should contain %q", summary, want)
				}
			}
		})
	}
}

func TestDescribeDetailed(t *testing.T) {
	diagram := &Diagram{
		Nodes: []Node{
			{ID: "a", Label: "Node A", Shape: ShapeRectangle},
			{ID: "b", Label: "Node B", Shape: ShapeCircle},
		},
		Edges: []Edge{
			{ID: "(a -> b)[0]", Source: "a", Target: "b", Label: "connects"},
		},
	}

	detailed := diagram.DescribeDetailed()

	// Should contain node descriptions
	if !strings.Contains(detailed, "Node A") {
		t.Error("Should contain 'Node A'")
	}
	if !strings.Contains(detailed, "Node B") {
		t.Error("Should contain 'Node B'")
	}
	if !strings.Contains(detailed, "circle") {
		t.Error("Should mention circle shape")
	}

	// Should contain edge description
	if !strings.Contains(detailed, "connects to") {
		t.Error("Should contain connection description")
	}
}

func TestDescribeForLLM(t *testing.T) {
	diagram := &Diagram{
		Title: "Test Diagram",
		Nodes: []Node{
			{ID: "container", Children: []string{"container.child"}},
			{ID: "container.child", Parent: "container", Shape: ShapeRectangle},
			{ID: "standalone", Shape: ShapeCircle},
		},
		Edges: []Edge{
			{ID: "(a -> b)[0]", Source: "a", Target: "b"},
		},
	}

	llm := diagram.DescribeForLLM()

	// Should have markdown headers
	if !strings.Contains(llm, "# D2 Diagram Structure") {
		t.Error("Should have main header")
	}
	if !strings.Contains(llm, "## Containers") {
		t.Error("Should have containers section")
	}
	if !strings.Contains(llm, "## Nodes") {
		t.Error("Should have nodes section")
	}
	if !strings.Contains(llm, "## Edges") {
		t.Error("Should have edges section")
	}

	// Should mention title
	if !strings.Contains(llm, "Test Diagram") {
		t.Error("Should contain diagram title")
	}
}

func TestFormatLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", `"with space"`},
		{"with.dot", `"with.dot"`},
		{"normal_underscore", "normal_underscore"},
	}

	for _, tt := range tests {
		got := formatLabel(tt.input)
		if got != tt.expected {
			t.Errorf("formatLabel(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNodeDescriptions(t *testing.T) {
	tests := []struct {
		node     Node
		contains string
	}{
		{
			node:     Node{ID: "a", Shape: ShapeRectangle},
			contains: "rectangle",
		},
		{
			node:     Node{ID: "person", Shape: ShapePerson},
			contains: "person",
		},
		{
			node:     Node{ID: "db", Shape: ShapeCylinder},
			contains: "cylinder",
		},
		{
			node:     Node{ID: "container", Shape: ShapeRectangle, Children: []string{"child"}},
			contains: "container",
		},
	}

	for _, tt := range tests {
		desc := describeNode(tt.node)
		if !strings.Contains(desc, tt.contains) {
			t.Errorf("describeNode(%+v) = %q, should contain %q", tt.node, desc, tt.contains)
		}
	}
}

func TestEdgeDescriptions(t *testing.T) {
	diagram := &Diagram{
		Nodes: []Node{
			{ID: "a", Label: "Source"},
			{ID: "b", Label: "Target"},
		},
	}

	tests := []struct {
		edge     Edge
		contains string
	}{
		{
			edge:     Edge{Source: "a", Target: "b"},
			contains: "connects to",
		},
		{
			edge:     Edge{Source: "a", Target: "b", SourceArrow: ArrowTriangle, TargetArrow: ArrowTriangle},
			contains: "bidirectionally",
		},
		{
			edge:     Edge{Source: "a", Target: "b", Label: "relationship"},
			contains: "relationship",
		},
	}

	for _, tt := range tests {
		desc := describeEdge(tt.edge, diagram)
		if !strings.Contains(desc, tt.contains) {
			t.Errorf("describeEdge(%+v) = %q, should contain %q", tt.edge, desc, tt.contains)
		}
	}
}
