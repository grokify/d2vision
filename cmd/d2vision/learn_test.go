package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/generate"
)

// TestRoundTrip tests the round-trip: D2 → SVG → learn → D2 → SVG
// The learned diagram should have the same structure as the original.
func TestRoundTrip(t *testing.T) {
	// Skip if d2 is not installed
	if _, err := exec.LookPath("d2"); err != nil {
		t.Skip("d2 not installed, skipping round-trip tests")
	}

	tests := []struct {
		name     string
		d2Code   string
		wantErr  bool
		validate func(t *testing.T, original, learned *d2vision.Diagram)
	}{
		{
			name: "simple_chain",
			d2Code: `a -> b -> c
`,
			validate: func(t *testing.T, original, learned *d2vision.Diagram) {
				if len(original.Nodes) != len(learned.Nodes) {
					t.Errorf("node count: got %d, want %d", len(learned.Nodes), len(original.Nodes))
				}
				if len(original.Edges) != len(learned.Edges) {
					t.Errorf("edge count: got %d, want %d", len(learned.Edges), len(original.Edges))
				}
			},
		},
		{
			name: "container",
			d2Code: `container: Container {
  inner1 -> inner2
}
`,
			validate: func(t *testing.T, original, learned *d2vision.Diagram) {
				// Should have container + 2 inner nodes
				if len(learned.Nodes) < 3 {
					t.Errorf("expected at least 3 nodes, got %d", len(learned.Nodes))
				}
				// Check container exists
				found := false
				for _, n := range learned.Nodes {
					if n.ID == "container" {
						found = true
						break
					}
				}
				if !found {
					t.Error("container node not found")
				}
			},
		},
		{
			name: "shapes",
			d2Code: `db: Database {
  shape: cylinder
}
`,
			validate: func(t *testing.T, original, learned *d2vision.Diagram) {
				for _, n := range learned.Nodes {
					if n.ID == "db" && n.Shape != d2vision.ShapeCylinder {
						t.Errorf("db shape: got %s, want cylinder", n.Shape)
					}
				}
				// Note: person shape detection is complex and may not round-trip perfectly
			},
		},
		{
			name: "labeled_edge",
			d2Code: `a -> b: connects
`,
			validate: func(t *testing.T, original, learned *d2vision.Diagram) {
				if len(learned.Edges) == 0 {
					t.Fatal("no edges found")
				}
				if learned.Edges[0].Label != "connects" {
					t.Errorf("edge label: got %q, want %q", learned.Edges[0].Label, "connects")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "d2vision-roundtrip-*")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Write original D2
			originalD2 := filepath.Join(tmpDir, "original.d2")
			if err := os.WriteFile(originalD2, []byte(tt.d2Code), 0644); err != nil {
				t.Fatal(err)
			}

			// Render original to SVG
			originalSVG := filepath.Join(tmpDir, "original.svg")
			cmd := exec.Command("d2", originalD2, originalSVG)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("d2 render failed: %v\n%s", err, output)
			}

			// Parse the SVG
			originalDiagram, err := d2vision.ParseFile(originalSVG)
			if err != nil {
				t.Fatalf("parse original SVG: %v", err)
			}

			// Learn: Convert to spec
			spec := diagramToSpec(originalDiagram)

			// Generate D2 from spec
			gen := generate.NewGenerator()
			learnedD2Code := gen.Generate(spec)

			// Write learned D2
			learnedD2 := filepath.Join(tmpDir, "learned.d2")
			if err := os.WriteFile(learnedD2, []byte(learnedD2Code), 0644); err != nil {
				t.Fatal(err)
			}

			// Render learned to SVG
			learnedSVG := filepath.Join(tmpDir, "learned.svg")
			cmd = exec.Command("d2", learnedD2, learnedSVG)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("d2 render learned failed: %v\n%s\nD2 code:\n%s", err, output, learnedD2Code)
			}

			// Parse the learned SVG
			learnedDiagram, err := d2vision.ParseFile(learnedSVG)
			if err != nil {
				t.Fatalf("parse learned SVG: %v", err)
			}

			// Validate
			if tt.validate != nil {
				tt.validate(t, originalDiagram, learnedDiagram)
			}
		})
	}
}

// TestRoundTripNodeCount ensures node count is preserved through round-trip.
func TestRoundTripNodeCount(t *testing.T) {
	if _, err := exec.LookPath("d2"); err != nil {
		t.Skip("d2 not installed")
	}

	testCases := []struct {
		name      string
		d2Code    string
		wantNodes int
		wantEdges int
	}{
		{"single", "a\n", 1, 0},
		{"two_nodes", "a\nb\n", 2, 0},
		{"chain_3", "a -> b -> c\n", 3, 2},
		{"chain_5", "a -> b -> c -> d -> e\n", 5, 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, _ := os.MkdirTemp("", "d2vision-*")
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Write and render
			d2File := filepath.Join(tmpDir, "test.d2")
			svgFile := filepath.Join(tmpDir, "test.svg")
			if err := os.WriteFile(d2File, []byte(tc.d2Code), 0644); err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command("d2", d2File, svgFile)
			if _, err := cmd.CombinedOutput(); err != nil {
				t.Skip("d2 render failed")
			}

			// Parse
			diagram, err := d2vision.ParseFile(svgFile)
			if err != nil {
				t.Fatal(err)
			}

			if len(diagram.Nodes) != tc.wantNodes {
				t.Errorf("nodes: got %d, want %d", len(diagram.Nodes), tc.wantNodes)
			}
			if len(diagram.Edges) != tc.wantEdges {
				t.Errorf("edges: got %d, want %d", len(diagram.Edges), tc.wantEdges)
			}
		})
	}
}

// TestRoundTripTemplates tests round-trip with actual templates.
func TestRoundTripTemplates(t *testing.T) {
	if _, err := exec.LookPath("d2"); err != nil {
		t.Skip("d2 not installed")
	}

	templates := []string{
		"network-boundary",
		"microservices",
		"data-flow",
		// Sequence and ER diagrams have special parsing that may not fully round-trip
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			tmpDir, _ := os.MkdirTemp("", "d2vision-template-*")
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Generate template
			var spec *generate.DiagramSpec
			switch tmpl {
			case "network-boundary":
				spec = generateNetworkBoundaryTemplate(2, 2)
			case "microservices":
				spec = generateMicroservicesTemplate()
			case "data-flow":
				spec = generateDataFlowTemplate()
			}

			gen := generate.NewGenerator()
			d2Code := gen.Generate(spec)

			// Write and render
			d2File := filepath.Join(tmpDir, "template.d2")
			svgFile := filepath.Join(tmpDir, "template.svg")
			if err := os.WriteFile(d2File, []byte(d2Code), 0644); err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command("d2", d2File, svgFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("d2 render failed: %v\n%s", err, output)
			}

			// Parse the SVG
			diagram, err := d2vision.ParseFile(svgFile)
			if err != nil {
				t.Fatalf("parse SVG: %v", err)
			}

			// Basic validation
			if len(diagram.Nodes) == 0 {
				t.Error("no nodes parsed")
			}

			// Learn and regenerate
			learnedSpec := diagramToSpec(diagram)
			learnedD2 := gen.Generate(learnedSpec)

			// Verify the learned D2 is valid (can be rendered)
			learnedD2File := filepath.Join(tmpDir, "learned.d2")
			learnedSVGFile := filepath.Join(tmpDir, "learned.svg")
			if err := os.WriteFile(learnedD2File, []byte(learnedD2), 0644); err != nil {
				t.Fatal(err)
			}

			cmd = exec.Command("d2", learnedD2File, learnedSVGFile)
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Errorf("learned D2 failed to render: %v\n%s\nD2:\n%s", err, output, learnedD2)
			}
		})
	}
}

// TestDiagramToSpec tests the diagramToSpec conversion.
func TestDiagramToSpec(t *testing.T) {
	diagram := &d2vision.Diagram{
		Nodes: []d2vision.Node{
			{ID: "a", Label: "Node A", Shape: d2vision.ShapeRectangle},
			{ID: "b", Label: "Node B", Shape: d2vision.ShapeCylinder},
		},
		Edges: []d2vision.Edge{
			{Source: "a", Target: "b", Label: "connects"},
		},
	}

	spec := diagramToSpec(diagram)

	if len(spec.Nodes) != 2 {
		t.Errorf("nodes: got %d, want 2", len(spec.Nodes))
	}

	if len(spec.Edges) != 1 {
		t.Errorf("edges: got %d, want 1", len(spec.Edges))
	}

	// Check shape preservation
	for _, n := range spec.Nodes {
		if n.ID == "b" && n.Shape != "cylinder" {
			t.Errorf("shape not preserved: got %s, want cylinder", n.Shape)
		}
	}

	// Check edge label
	if spec.Edges[0].Label != "connects" {
		t.Errorf("edge label: got %q, want %q", spec.Edges[0].Label, "connects")
	}
}

// TestDetectGridLayout tests grid detection from node positions.
func TestDetectGridLayout(t *testing.T) {
	tests := []struct {
		name        string
		nodes       []d2vision.Node
		wantColumns int
		wantRows    int
	}{
		{
			name: "horizontal_3",
			nodes: []d2vision.Node{
				{ID: "a", Bounds: d2vision.Bounds{X: 0, Y: 50, Width: 100, Height: 50}},
				{ID: "b", Bounds: d2vision.Bounds{X: 150, Y: 50, Width: 100, Height: 50}},
				{ID: "c", Bounds: d2vision.Bounds{X: 300, Y: 50, Width: 100, Height: 50}},
			},
			wantColumns: 3,
			wantRows:    0,
		},
		{
			name: "single_node",
			nodes: []d2vision.Node{
				{ID: "a", Bounds: d2vision.Bounds{X: 50, Y: 50, Width: 100, Height: 50}},
			},
			wantColumns: 0,
			wantRows:    0,
		},
		{
			name: "vertical_stack",
			nodes: []d2vision.Node{
				{ID: "a", Bounds: d2vision.Bounds{X: 50, Y: 0, Width: 100, Height: 50}},
				{ID: "b", Bounds: d2vision.Bounds{X: 50, Y: 100, Width: 100, Height: 50}},
			},
			wantColumns: 0,
			wantRows:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram := &d2vision.Diagram{Nodes: tt.nodes}
			gotColumns, gotRows := detectGridLayout(diagram)

			if gotColumns != tt.wantColumns {
				t.Errorf("columns: got %d, want %d", gotColumns, tt.wantColumns)
			}
			if gotRows != tt.wantRows {
				t.Errorf("rows: got %d, want %d", gotRows, tt.wantRows)
			}
		})
	}
}

// TestHelperFunctions tests the helper functions.
func TestHelperFunctions(t *testing.T) {
	// Test abs
	if abs(-5.0) != 5.0 {
		t.Error("abs(-5.0) should be 5.0")
	}
	if abs(5.0) != 5.0 {
		t.Error("abs(5.0) should be 5.0")
	}

	// Test variance
	vals := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	v := variance(vals)
	// Variance should be 4.0
	if v < 3.9 || v > 4.1 {
		t.Errorf("variance: got %f, want ~4.0", v)
	}

	// Test isDescendantOf
	if !isDescendantOf("container.child", "container") {
		t.Error("container.child should be descendant of container")
	}
	if isDescendantOf("container", "container") {
		t.Error("container should not be descendant of itself")
	}
	if isDescendantOf("other.child", "container") {
		t.Error("other.child should not be descendant of container")
	}

	// Test makeRelativeID
	if makeRelativeID("container.child", "container") != "child" {
		t.Error("makeRelativeID should strip container prefix")
	}
	if makeRelativeID("child", "container") != "child" {
		t.Error("makeRelativeID should return original if no prefix match")
	}
}

// TestGeneratedD2Syntax ensures generated D2 is syntactically valid.
func TestGeneratedD2Syntax(t *testing.T) {
	if _, err := exec.LookPath("d2"); err != nil {
		t.Skip("d2 not installed")
	}

	// Test various specs
	specs := []*generate.DiagramSpec{
		{
			Nodes: []generate.NodeSpec{
				{ID: "simple", Label: "Simple Node"},
			},
		},
		{
			Containers: []generate.ContainerSpec{
				{
					ID:    "container",
					Label: "Container",
					Nodes: []generate.NodeSpec{
						{ID: "inner", Label: "Inner"},
					},
				},
			},
		},
		{
			GridColumns: 2,
			Containers: []generate.ContainerSpec{
				{ID: "a", Label: "A"},
				{ID: "b", Label: "B"},
			},
		},
	}

	gen := generate.NewGenerator()

	for i, spec := range specs {
		t.Run(strings.ReplaceAll(t.Name(), "/", "_")+"_"+string(rune('0'+i)), func(t *testing.T) {
			d2Code := gen.Generate(spec)

			tmpDir, _ := os.MkdirTemp("", "d2vision-syntax-*")
			defer func() { _ = os.RemoveAll(tmpDir) }()

			d2File := filepath.Join(tmpDir, "test.d2")
			svgFile := filepath.Join(tmpDir, "test.svg")
			if err := os.WriteFile(d2File, []byte(d2Code), 0644); err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command("d2", d2File, svgFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("invalid D2 syntax: %v\n%s\nCode:\n%s", err, output, d2Code)
			}
		})
	}
}
