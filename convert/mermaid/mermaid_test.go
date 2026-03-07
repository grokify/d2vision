package mermaid

import (
	"strings"
	"testing"
)

func TestParseFlowchart(t *testing.T) {
	source := `graph LR
    A[Start] --> B{Decision}
    B -->|Yes| C[OK]
    B -->|No| D[Cancel]`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc.Type != DiagramFlowchart {
		t.Errorf("Expected DiagramFlowchart, got %s", doc.Type)
	}

	if doc.Direction != DirectionLR {
		t.Errorf("Expected DirectionLR, got %s", doc.Direction)
	}

	// Check nodes
	if len(doc.Nodes) < 4 {
		t.Errorf("Expected at least 4 nodes, got %d", len(doc.Nodes))
	}

	// Check edges
	if len(doc.Edges) < 3 {
		t.Errorf("Expected at least 3 edges, got %d", len(doc.Edges))
	}
}

func TestParseFlowchartWithSubgraph(t *testing.T) {
	source := `graph TB
    subgraph Frontend
        A[Web App]
        B[Mobile App]
    end
    subgraph Backend
        C[API Server]
    end
    A --> C
    B --> C`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(doc.Subgraphs) < 2 {
		t.Errorf("Expected at least 2 subgraphs, got %d", len(doc.Subgraphs))
	}

	// Check subgraph names
	found := make(map[string]bool)
	for _, sg := range doc.Subgraphs {
		found[sg.ID] = true
	}
	if !found["Frontend"] {
		t.Error("Missing Frontend subgraph")
	}
	if !found["Backend"] {
		t.Error("Missing Backend subgraph")
	}
}

func TestParseSequenceDiagram(t *testing.T) {
	source := `sequenceDiagram
    participant Alice
    participant Bob
    Alice->>Bob: Hello
    Bob-->>Alice: Hi`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc.Type != DiagramSequence {
		t.Errorf("Expected DiagramSequence, got %s", doc.Type)
	}

	if len(doc.Actors) < 2 {
		t.Errorf("Expected at least 2 actors, got %d", len(doc.Actors))
	}

	if len(doc.Messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(doc.Messages))
	}
}

func TestConvertFlowchartToD2(t *testing.T) {
	source := `graph LR
    A[Start] --> B{Decision}
    B -->|Yes| C[OK]
    B -->|No| D[Cancel]`

	converter := NewConverter()
	result, err := converter.Convert(source)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Spec == nil {
		t.Fatal("Expected non-nil spec")
	}

	if result.Spec.Direction != "right" {
		t.Errorf("Expected direction right, got %s", result.Spec.Direction)
	}

	// Should have nodes
	if len(result.Spec.Nodes) == 0 {
		t.Error("Expected nodes in spec")
	}

	// Should have edges
	if len(result.Spec.Edges) == 0 {
		t.Error("Expected edges in spec")
	}
}

func TestLintFlowchart(t *testing.T) {
	source := `graph LR
    A[Start] --> B{Decision}
    click A callback`

	converter := NewConverter()
	lintResult, err := converter.Lint(source)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if !lintResult.Convertible {
		t.Error("Expected diagram to be convertible")
	}

	// Should have unsupported feature for click
	foundClick := false
	for _, u := range lintResult.Unsupported {
		if u.Feature == "click" {
			foundClick = true
			break
		}
	}
	if !foundClick {
		t.Error("Expected unsupported feature for click handler")
	}
}

func TestLintUnsupportedDiagram(t *testing.T) {
	source := `gantt
    title Project Timeline
    section Phase 1
    Task 1: 2024-01-01, 7d`

	converter := NewConverter()
	lintResult, err := converter.Lint(source)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if lintResult.Convertible {
		t.Error("Expected gantt diagram to be non-convertible")
	}
}

func TestNodeShapes(t *testing.T) {
	tests := []struct {
		source    string
		wantShape NodeShape
	}{
		{"A[text]", ShapeRectangle},
		{"A(text)", ShapeRoundedRect},
		{"A{text}", ShapeDiamond},
		{"A((text))", ShapeCircle},
		{"A[(text)]", ShapeCylinder},
		{"A{{text}}", ShapeHexagon},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			source := "graph LR\n    " + tt.source

			doc, err := Parse(source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(doc.Nodes) == 0 {
				t.Fatal("Expected at least one node")
			}

			got := doc.Nodes[0].Shape
			if got != tt.wantShape {
				t.Errorf("Expected shape %s, got %s", tt.wantShape, got)
			}
		})
	}
}

func TestArrowStyles(t *testing.T) {
	tests := []struct {
		arrow    string
		wantDash bool
	}{
		{"-->", false},
		{"-.->", true},
		{"==>", false},
	}

	for _, tt := range tests {
		t.Run(tt.arrow, func(t *testing.T) {
			source := "graph LR\n    A " + tt.arrow + " B"

			doc, err := Parse(source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(doc.Edges) == 0 {
				t.Fatal("Expected at least one edge")
			}

			got := doc.Edges[0].Style.Dashed
			if got != tt.wantDash {
				t.Errorf("Expected dashed=%v, got %v", tt.wantDash, got)
			}
		})
	}
}

func TestDirectionConversion(t *testing.T) {
	tests := []struct {
		direction Direction
		wantD2    string
	}{
		{DirectionLR, "right"},
		{DirectionRL, "left"},
		{DirectionTB, "down"},
		{DirectionTD, "down"},
		{DirectionBT, "up"},
	}

	for _, tt := range tests {
		t.Run(string(tt.direction), func(t *testing.T) {
			got := tt.direction.ToD2Direction()
			if got != tt.wantD2 {
				t.Errorf("Expected %s, got %s", tt.wantD2, got)
			}
		})
	}
}

func TestConvertSequenceDiagram(t *testing.T) {
	source := `sequenceDiagram
    participant Alice
    participant Bob
    Alice->>Bob: Hello
    Bob-->>Alice: Hi`

	converter := NewConverter()
	result, err := converter.Convert(source)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.Spec == nil {
		t.Fatal("Expected non-nil spec")
	}

	// Should have sequence
	if len(result.Spec.Sequences) == 0 {
		t.Error("Expected sequence in spec")
	}

	seq := result.Spec.Sequences[0]

	// Should have actors
	if len(seq.Actors) < 2 {
		t.Errorf("Expected at least 2 actors, got %d", len(seq.Actors))
	}

	// Should have steps
	if len(seq.Steps) < 2 {
		t.Errorf("Expected at least 2 steps, got %d", len(seq.Steps))
	}
}

func TestEdgeLabelParsing(t *testing.T) {
	source := `graph LR
    A -->|Yes| B
    A -->|No| C`

	doc, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check edge labels
	for _, edge := range doc.Edges {
		if edge.Label == "" {
			t.Errorf("Expected edge label, got empty for %s -> %s", edge.From, edge.To)
		}
		if !strings.Contains(edge.Label, "Yes") && !strings.Contains(edge.Label, "No") {
			t.Errorf("Expected label to contain Yes or No, got %s", edge.Label)
		}
	}
}
