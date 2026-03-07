package generate

import (
	"strings"
	"testing"
)

func TestGenerateSimple(t *testing.T) {
	spec := &DiagramSpec{
		Nodes: []NodeSpec{
			{ID: "a", Label: "Node A"},
			{ID: "b", Label: "Node B"},
		},
		Edges: []EdgeSpec{
			{From: "a", To: "b"},
		},
	}

	g := NewGenerator()
	got := g.Generate(spec)

	if !strings.Contains(got, "a: Node A") {
		t.Errorf("Generate() missing 'a: Node A': %s", got)
	}
	if !strings.Contains(got, "b: Node B") {
		t.Errorf("Generate() missing 'b: Node B': %s", got)
	}
	if !strings.Contains(got, "a -> b") {
		t.Errorf("Generate() missing 'a -> b': %s", got)
	}
}

func TestGenerateWithShapes(t *testing.T) {
	spec := &DiagramSpec{
		Nodes: []NodeSpec{
			{ID: "db", Label: "Database", Shape: "cylinder"},
			{ID: "server", Label: "Server", Shape: "rectangle"},
		},
	}

	g := NewGenerator()
	got := g.Generate(spec)

	if !strings.Contains(got, "shape: cylinder") {
		t.Errorf("Generate() missing 'shape: cylinder': %s", got)
	}
}

func TestGenerateContainer(t *testing.T) {
	spec := &DiagramSpec{
		Containers: []ContainerSpec{
			{
				ID:        "cluster1",
				Label:     "Cluster 1",
				Direction: "down",
				Nodes: []NodeSpec{
					{ID: "service1", Label: "Service 1"},
					{ID: "db1", Label: "Database 1", Shape: "cylinder"},
				},
				Edges: []EdgeSpec{
					{From: "service1", To: "db1"},
				},
			},
		},
	}

	g := NewGenerator()
	got := g.Generate(spec)

	if !strings.Contains(got, "cluster1: Cluster 1 {") {
		t.Errorf("Generate() missing container declaration: %s", got)
	}
	if !strings.Contains(got, "direction: down") {
		t.Errorf("Generate() missing direction: %s", got)
	}
	if !strings.Contains(got, "service1: Service 1") {
		t.Errorf("Generate() missing service1: %s", got)
	}
	if !strings.Contains(got, "service1 -> db1") {
		t.Errorf("Generate() missing edge: %s", got)
	}
}

func TestGenerateNetworkClusters(t *testing.T) {
	// Recreate the network clusters example
	spec := &DiagramSpec{
		GridColumns: 2,
		Containers: []ContainerSpec{
			{
				ID:        "cluster1",
				Label:     "Cluster 1",
				Direction: "down",
				Containers: []ContainerSpec{
					{
						ID:        "services",
						Label:     "", // Empty label for invisible container
						Direction: "right",
						Style:     &StyleSpec{StrokeWidth: IntPtr(0)},
						Nodes: []NodeSpec{
							{ID: "service1a", Label: "Service 1A"},
							{ID: "service1b", Label: "Service 1B"},
						},
					},
				},
				Nodes: []NodeSpec{
					{ID: "datastore1", Label: "DataStore 1", Shape: "cylinder"},
				},
				Edges: []EdgeSpec{
					{From: "services.service1a", To: "datastore1"},
					{From: "services.service1b", To: "datastore1"},
				},
			},
			{
				ID:        "cluster2",
				Label:     "Cluster 2",
				Direction: "down",
				Nodes: []NodeSpec{
					{ID: "service2", Label: "Service 2"},
					{ID: "datastore2", Label: "DataStore 2", Shape: "cylinder"},
				},
				Edges: []EdgeSpec{
					{From: "service2", To: "datastore2"},
				},
			},
		},
		Edges: []EdgeSpec{
			{From: "cluster2.datastore2", To: "cluster1.datastore1", Label: "replication"},
		},
	}

	g := NewGenerator()
	got := g.Generate(spec)

	// Check key elements
	checks := []string{
		"grid-columns: 2",
		"cluster1: Cluster 1 {",
		"cluster2: Cluster 2 {",
		"direction: down",
		"direction: right",
		"style.stroke-width: 0",
		"shape: cylinder",
		"replication",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("Generate() missing %q:\n%s", check, got)
		}
	}
}

func TestGenerateEdgeWithLabel(t *testing.T) {
	spec := &DiagramSpec{
		Edges: []EdgeSpec{
			{From: "a", To: "b", Label: "connects to"},
		},
	}

	g := NewGenerator()
	got := g.Generate(spec)

	if !strings.Contains(got, "a -> b: connects to") {
		t.Errorf("Generate() missing labeled edge: %s", got)
	}
}

func TestEscapeID(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"with space", `"with space"`},
		{"with-dash", `"with-dash"`},
		{"with.dot", `"with.dot"`},
	}

	for _, tt := range tests {
		got := g.escapeID(tt.input)
		if got != tt.want {
			t.Errorf("escapeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
