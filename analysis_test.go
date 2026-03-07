package d2vision

import (
	"strings"
	"testing"
)

func TestAnalyzeLayout(t *testing.T) {
	// Create a simple diagram with containers
	diagram := &Diagram{
		ViewBox: Bounds{X: 0, Y: 0, Width: 800, Height: 600},
		Nodes: []Node{
			{ID: "cluster1", Label: "Cluster 1", Shape: ShapeRectangle, Bounds: Bounds{X: 0, Y: 0, Width: 350, Height: 500}, Children: []string{"cluster1.service1", "cluster1.db1"}},
			{ID: "cluster1.service1", Label: "Service 1", Shape: ShapeRectangle, Bounds: Bounds{X: 50, Y: 50, Width: 100, Height: 50}, Parent: "cluster1"},
			{ID: "cluster1.db1", Label: "Database", Shape: ShapeCylinder, Bounds: Bounds{X: 50, Y: 150, Width: 100, Height: 50}, Parent: "cluster1"},
			{ID: "cluster2", Label: "Cluster 2", Shape: ShapeRectangle, Bounds: Bounds{X: 400, Y: 0, Width: 350, Height: 500}, Children: []string{"cluster2.service2"}},
			{ID: "cluster2.service2", Label: "Service 2", Shape: ShapeRectangle, Bounds: Bounds{X: 450, Y: 50, Width: 100, Height: 50}, Parent: "cluster2"},
		},
		Edges: []Edge{
			{ID: "e1", Source: "cluster1.service1", Target: "cluster1.db1"},
			{ID: "e2", Source: "cluster2.service2", Target: "cluster1.db1"},
		},
	}

	analysis := diagram.AnalyzeLayout()

	// Check basic properties
	if !analysis.HasContainers {
		t.Error("Expected HasContainers to be true")
	}

	if analysis.ContainerCount != 2 {
		t.Errorf("Expected ContainerCount 2, got %d", analysis.ContainerCount)
	}

	// Check for cross-container edges
	if analysis.CrossContainerEdges != 1 {
		t.Errorf("Expected 1 cross-container edge, got %d", analysis.CrossContainerEdges)
	}

	// Check that insights are generated
	if len(analysis.Insights) == 0 {
		t.Error("Expected non-empty Insights")
	}

	// Check that hints are generated
	if len(analysis.GenerationHints) == 0 {
		t.Error("Expected non-empty GenerationHints")
	}
}

func TestDescribeForGeneration(t *testing.T) {
	diagram := &Diagram{
		ViewBox: Bounds{X: 0, Y: 0, Width: 800, Height: 600},
		Nodes: []Node{
			{ID: "a", Label: "Node A", Shape: ShapeRectangle, Bounds: Bounds{X: 0, Y: 0, Width: 100, Height: 50}},
			{ID: "b", Label: "Node B", Shape: ShapeCylinder, Bounds: Bounds{X: 200, Y: 0, Width: 100, Height: 50}},
		},
		Edges: []Edge{
			{ID: "e1", Source: "a", Target: "b", Label: "connects"},
		},
	}

	output := diagram.DescribeForGeneration()

	// Check that output contains expected sections
	expectedSections := []string{
		"# D2 Diagram Analysis for Recreation",
		"## Overview",
		"## Structure",
		"### Nodes",
		"### Edges",
		"## Suggested D2 Code Skeleton",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected output to contain %q", section)
		}
	}

	// Check that nodes and edges are mentioned
	if !strings.Contains(output, "Node A") && !strings.Contains(output, "a") {
		t.Error("Expected output to mention node a")
	}
	if !strings.Contains(output, "cylinder") {
		t.Error("Expected output to mention cylinder shape")
	}
}

func TestCalculateVariance(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 0},
		{"identical", []float64{3, 3, 3}, 0},
		{"simple", []float64{1, 2, 3}, 0.6666666666666666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateVariance(tt.values)
			if got != tt.want {
				t.Errorf("calculateVariance(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestNodeDepth(t *testing.T) {
	diagram := &Diagram{
		Nodes: []Node{
			{ID: "root", Children: []string{"level1"}},
			{ID: "level1", Parent: "root", Children: []string{"level2"}},
			{ID: "level2", Parent: "level1", Children: []string{"level3"}},
			{ID: "level3", Parent: "level2"},
		},
	}

	tests := []struct {
		nodeID string
		want   int
	}{
		{"root", 0},
		{"level1", 1},
		{"level2", 2},
		{"level3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.nodeID, func(t *testing.T) {
			got := diagram.nodeDepth(tt.nodeID)
			if got != tt.want {
				t.Errorf("nodeDepth(%q) = %d, want %d", tt.nodeID, got, tt.want)
			}
		})
	}
}

func TestDetectFlowDirection(t *testing.T) {
	// Horizontal flow (right)
	horizontalDiagram := &Diagram{
		Nodes: []Node{
			{ID: "a", Bounds: Bounds{X: 0, Y: 100, Width: 50, Height: 50}},
			{ID: "b", Bounds: Bounds{X: 100, Y: 100, Width: 50, Height: 50}},
			{ID: "c", Bounds: Bounds{X: 200, Y: 100, Width: 50, Height: 50}},
		},
		Edges: []Edge{
			{Source: "a", Target: "b"},
			{Source: "b", Target: "c"},
		},
	}

	direction := horizontalDiagram.detectFlowDirection()
	if direction != "right" {
		t.Errorf("Expected direction 'right', got %q", direction)
	}

	// Vertical flow (down)
	verticalDiagram := &Diagram{
		Nodes: []Node{
			{ID: "a", Bounds: Bounds{X: 100, Y: 0, Width: 50, Height: 50}},
			{ID: "b", Bounds: Bounds{X: 100, Y: 100, Width: 50, Height: 50}},
			{ID: "c", Bounds: Bounds{X: 100, Y: 200, Width: 50, Height: 50}},
		},
		Edges: []Edge{
			{Source: "a", Target: "b"},
			{Source: "b", Target: "c"},
		},
	}

	direction = verticalDiagram.detectFlowDirection()
	if direction != "down" {
		t.Errorf("Expected direction 'down', got %q", direction)
	}
}
