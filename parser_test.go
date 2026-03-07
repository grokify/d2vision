package d2vision

import (
	"strings"
	"testing"
)

// =============================================================================
// Basic SVG Parsing Tests
// =============================================================================

func TestParserBasicSVG(t *testing.T) {
	// A minimal D2-like SVG structure
	svg := `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200" data-d2-version="0.6.0">
  <g class="YQ==">
    <rect x="10" y="10" width="80" height="40" fill="#f0f0f0" stroke="#333"/>
    <text x="50" y="35">Node A</text>
  </g>
  <g class="Yg==">
    <rect x="110" y="10" width="80" height="40" fill="#f0f0f0" stroke="#333"/>
    <text x="150" y="35">Node B</text>
  </g>
  <g class="KGEgLT4gYilbMF0=">
    <path d="M90 30 L110 30" stroke="#333" marker-end="url(#arrow)"/>
  </g>
</svg>`

	diagram, err := ParseString(svg)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	// Check version
	if diagram.Version != "0.6.0" {
		t.Errorf("Version = %q, want %q", diagram.Version, "0.6.0")
	}

	// Check viewBox
	if diagram.ViewBox.Width != 200 || diagram.ViewBox.Height != 200 {
		t.Errorf("ViewBox = %+v, want 200x200", diagram.ViewBox)
	}

	// Check nodes
	if len(diagram.Nodes) != 2 {
		t.Errorf("len(Nodes) = %d, want 2", len(diagram.Nodes))
	}

	// Check edges
	if len(diagram.Edges) != 1 {
		t.Errorf("len(Edges) = %d, want 1", len(diagram.Edges))
	}

	// Verify node IDs and labels
	nodeA := diagram.NodeByID("a")
	if nodeA == nil {
		t.Error("NodeByID(\"a\") returned nil")
	} else {
		if nodeA.Label != "Node A" {
			t.Errorf("Node a label = %q, want %q", nodeA.Label, "Node A")
		}
	}

	nodeB := diagram.NodeByID("b")
	if nodeB == nil {
		t.Error("NodeByID(\"b\") returned nil")
	} else {
		if nodeB.Label != "Node B" {
			t.Errorf("Node b label = %q, want %q", nodeB.Label, "Node B")
		}
	}

	// Verify edge
	if len(diagram.Edges) > 0 {
		edge := diagram.Edges[0]
		if edge.Source != "a" {
			t.Errorf("Edge source = %q, want %q", edge.Source, "a")
		}
		if edge.Target != "b" {
			t.Errorf("Edge target = %q, want %q", edge.Target, "b")
		}
	}
}

// =============================================================================
// Nested SVG Structure Tests
// =============================================================================
//
// D2 generates SVGs with nested <svg> elements. The actual diagram content
// is inside an inner <svg>, not directly under the root. This was a bug
// we discovered when bounds were returning (0,0,0,0).

func TestNestedSVGStructure(t *testing.T) {
	// SVG with nested structure like D2 generates
	svg := `<?xml version="1.0" encoding="utf-8"?>
<svg xmlns="http://www.w3.org/2000/svg" data-d2-version="0.7.1" viewBox="0 0 255 600">
  <svg class="d2-inner d2-svg" width="255" height="600" viewBox="-101 -101 255 600">
    <g class="YQ==">
      <g class="shape">
        <rect x="0" y="0" width="53" height="66"/>
      </g>
      <text>a</text>
    </g>
    <g class="Yg==">
      <g class="shape">
        <rect x="0" y="166" width="53" height="66"/>
      </g>
      <text>b</text>
    </g>
  </svg>
</svg>`

	diagram, err := ParseString(svg)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	// Should find nodes inside nested SVG
	if len(diagram.Nodes) != 2 {
		t.Errorf("len(Nodes) = %d, want 2", len(diagram.Nodes))
	}

	// Check that bounds are extracted correctly from nested structure
	nodeA := diagram.NodeByID("a")
	if nodeA == nil {
		t.Fatal("Node 'a' not found")
	}
	if nodeA.Bounds.Width != 53 || nodeA.Bounds.Height != 66 {
		t.Errorf("Node a bounds = %+v, want width=53 height=66", nodeA.Bounds)
	}
	if nodeA.Bounds.X != 0 || nodeA.Bounds.Y != 0 {
		t.Errorf("Node a position = (%v,%v), want (0,0)", nodeA.Bounds.X, nodeA.Bounds.Y)
	}

	nodeB := diagram.NodeByID("b")
	if nodeB == nil {
		t.Fatal("Node 'b' not found")
	}
	if nodeB.Bounds.Y != 166 {
		t.Errorf("Node b Y = %v, want 166 (below node a)", nodeB.Bounds.Y)
	}
}

func TestNestedShapeGroups(t *testing.T) {
	// D2 nests shapes inside <g class="shape"> groups
	svg := `<svg viewBox="0 0 100 100">
    <svg>
      <g class="YQ==">
        <g class="shape">
          <rect x="10" y="20" width="30" height="40"/>
        </g>
      </g>
    </svg>
  </svg>`

	diagram, err := ParseString(svg)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(diagram.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(diagram.Nodes))
	}

	node := diagram.Nodes[0]
	if node.Bounds.X != 10 {
		t.Errorf("Bounds.X = %v, want 10", node.Bounds.X)
	}
	if node.Bounds.Y != 20 {
		t.Errorf("Bounds.Y = %v, want 20", node.Bounds.Y)
	}
	if node.Bounds.Width != 30 {
		t.Errorf("Bounds.Width = %v, want 30", node.Bounds.Width)
	}
	if node.Bounds.Height != 40 {
		t.Errorf("Bounds.Height = %v, want 40", node.Bounds.Height)
	}
}

// =============================================================================
// Container and Hierarchy Tests
// =============================================================================

func TestParserWithContainers(t *testing.T) {
	// SVG with nested containers
	svg := `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 300">
  <svg>
    <g class="Y29udGFpbmVy">
      <rect x="10" y="10" width="200" height="150"/>
      <text>Container</text>
    </g>
    <g class="Y29udGFpbmVyLmlubmVyMQ==">
      <rect x="20" y="30" width="80" height="40"/>
      <text>Inner 1</text>
    </g>
    <g class="Y29udGFpbmVyLmlubmVyMg==">
      <rect x="120" y="30" width="80" height="40"/>
      <text>Inner 2</text>
    </g>
  </svg>
</svg>`

	diagram, err := ParseString(svg)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	// Should have 3 nodes
	if len(diagram.Nodes) != 3 {
		t.Errorf("len(Nodes) = %d, want 3", len(diagram.Nodes))
	}

	// Check container has children
	container := diagram.NodeByID("container")
	if container == nil {
		t.Fatal("container node not found")
	}
	if len(container.Children) != 2 {
		t.Errorf("container.Children = %d, want 2", len(container.Children))
	}

	// Check children have correct parent
	inner1 := diagram.NodeByID("container.inner1")
	if inner1 == nil {
		t.Fatal("container.inner1 node not found")
	}
	if inner1.Parent != "container" {
		t.Errorf("inner1.Parent = %q, want %q", inner1.Parent, "container")
	}

	inner2 := diagram.NodeByID("container.inner2")
	if inner2 == nil {
		t.Fatal("container.inner2 node not found")
	}
	if inner2.Parent != "container" {
		t.Errorf("inner2.Parent = %q, want %q", inner2.Parent, "container")
	}
}

// =============================================================================
// Shape Detection Tests
// =============================================================================

func TestParserShapeDetection(t *testing.T) {
	tests := []struct {
		name  string
		svg   string
		shape ShapeType
	}{
		{
			name: "rectangle",
			svg: `<svg viewBox="0 0 100 100"><svg>
				<g class="YQ=="><rect x="0" y="0" width="80" height="40"/></g>
			</svg></svg>`,
			shape: ShapeRectangle,
		},
		{
			name: "square",
			svg: `<svg viewBox="0 0 100 100"><svg>
				<g class="YQ=="><rect x="0" y="0" width="50" height="50"/></g>
			</svg></svg>`,
			shape: ShapeSquare,
		},
		{
			name: "circle",
			svg: `<svg viewBox="0 0 100 100"><svg>
				<g class="YQ=="><circle cx="50" cy="50" r="25"/></g>
			</svg></svg>`,
			shape: ShapeCircle,
		},
		{
			name: "oval",
			svg: `<svg viewBox="0 0 100 100"><svg>
				<g class="YQ=="><ellipse cx="50" cy="50" rx="40" ry="20"/></g>
			</svg></svg>`,
			shape: ShapeOval,
		},
		{
			name: "circle from equal ellipse",
			svg: `<svg viewBox="0 0 100 100"><svg>
				<g class="YQ=="><ellipse cx="50" cy="50" rx="25" ry="25"/></g>
			</svg></svg>`,
			shape: ShapeCircle,
		},
		{
			// Cylinder: D2 uses cubic bezier curves (C) for rounded top/bottom
			// and vertical lines (V) for the sides
			name: "cylinder from cubic beziers",
			svg: `<svg viewBox="0 0 100 200"><svg>
				<g class="YQ==">
					<path d="M 0 24 C 0 0 29 0 32 0 C 35 0 64 0 64 24 V 94 C 64 118 35 118 32 118 C 29 118 0 118 0 94 Z" fill="#B5B2FF"/>
				</g>
			</svg></svg>`,
			shape: ShapeCylinder,
		},
		{
			// Diamond: 4 line segments forming a closed rhombus
			name: "diamond from lines",
			svg: `<svg viewBox="0 0 100 100"><svg>
				<g class="YQ==">
					<path d="M 25 0 L 50 25 L 25 50 L 0 25 Z" fill="#FFB5B5"/>
				</g>
			</svg></svg>`,
			shape: ShapeDiamond,
		},
		{
			// Hexagon: 6 line segments forming a closed hexagon
			name: "hexagon from lines",
			svg: `<svg viewBox="0 0 100 100"><svg>
				<g class="YQ==">
					<path d="M 25 0 L 75 0 L 100 50 L 75 100 L 25 100 L 0 50 Z" fill="#B5FFB5"/>
				</g>
			</svg></svg>`,
			shape: ShapeHexagon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram, err := ParseString(tt.svg)
			if err != nil {
				t.Fatalf("ParseString failed: %v", err)
			}
			if len(diagram.Nodes) == 0 {
				t.Fatal("no nodes parsed")
			}
			if diagram.Nodes[0].Shape != tt.shape {
				t.Errorf("Shape = %q, want %q", diagram.Nodes[0].Shape, tt.shape)
			}
		})
	}
}

// =============================================================================
// Bounds Extraction Tests
// =============================================================================

func TestBoundsFromCircle(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100"><svg>
		<g class="YQ=="><circle cx="50" cy="60" r="20"/></g>
	</svg></svg>`

	diagram, err := ParseString(svg)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	node := diagram.Nodes[0]
	// Circle at (50,60) with r=20 should have bounds (30,40,40,40)
	if node.Bounds.X != 30 {
		t.Errorf("Bounds.X = %v, want 30", node.Bounds.X)
	}
	if node.Bounds.Y != 40 {
		t.Errorf("Bounds.Y = %v, want 40", node.Bounds.Y)
	}
	if node.Bounds.Width != 40 {
		t.Errorf("Bounds.Width = %v, want 40", node.Bounds.Width)
	}
	if node.Bounds.Height != 40 {
		t.Errorf("Bounds.Height = %v, want 40", node.Bounds.Height)
	}
}

func TestBoundsFromEllipse(t *testing.T) {
	svg := `<svg viewBox="0 0 100 100"><svg>
		<g class="YQ=="><ellipse cx="50" cy="50" rx="30" ry="20"/></g>
	</svg></svg>`

	diagram, err := ParseString(svg)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	node := diagram.Nodes[0]
	// Ellipse at (50,50) with rx=30,ry=20 should have bounds (20,30,60,40)
	if node.Bounds.X != 20 {
		t.Errorf("Bounds.X = %v, want 20", node.Bounds.X)
	}
	if node.Bounds.Y != 30 {
		t.Errorf("Bounds.Y = %v, want 30", node.Bounds.Y)
	}
	if node.Bounds.Width != 60 {
		t.Errorf("Bounds.Width = %v, want 60", node.Bounds.Width)
	}
	if node.Bounds.Height != 40 {
		t.Errorf("Bounds.Height = %v, want 40", node.Bounds.Height)
	}
}

// TestBoundsFromPath verifies bounds extraction from SVG path elements.
// D2 renders cylinders, diamonds, hexagons, and other shapes using <path>
// elements rather than <rect>, <circle>, or <ellipse>. Without path parsing,
// these shapes would have zero bounds.
func TestBoundsFromPath(t *testing.T) {
	// Simulate a cylinder-like path (simplified): a shape from (0,24) to (64,118)
	// D2 uses paths with curves for cylinder tops/bottoms
	svg := `<svg viewBox="0 0 200 200"><svg>
		<g class="YQ==">
			<path d="M 0 24 C 0 0 29 0 32 0 C 35 0 64 0 64 24 V 94 C 64 118 35 118 32 118 C 29 118 0 118 0 94 Z" fill="#B5B2FF"/>
		</g>
	</svg></svg>`

	diagram, err := ParseString(svg)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(diagram.Nodes) == 0 {
		t.Fatal("no nodes parsed")
	}

	node := diagram.Nodes[0]

	// Path spans X: 0-64, Y: 0-118 → bounds should be (0, 0, 64, 118)
	if node.Bounds.X != 0 {
		t.Errorf("Bounds.X = %v, want 0", node.Bounds.X)
	}
	if node.Bounds.Y != 0 {
		t.Errorf("Bounds.Y = %v, want 0", node.Bounds.Y)
	}
	if node.Bounds.Width != 64 {
		t.Errorf("Bounds.Width = %v, want 64", node.Bounds.Width)
	}
	if node.Bounds.Height != 118 {
		t.Errorf("Bounds.Height = %v, want 118", node.Bounds.Height)
	}
}

// TestExtractBoundsFromPathCommands tests various SVG path command types.
func TestExtractBoundsFromPathCommands(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		wantX  float64
		wantY  float64
		wantW  float64
		wantH  float64
	}{
		{
			name:  "simple moveto/lineto",
			path:  "M 10 20 L 50 60",
			wantX: 10,
			wantY: 20,
			wantW: 40,
			wantH: 40,
		},
		{
			name:  "horizontal and vertical lines",
			path:  "M 0 0 H 100 V 50",
			wantX: 0,
			wantY: 0,
			wantW: 100,
			wantH: 50,
		},
		{
			name:  "cubic bezier",
			path:  "M 0 0 C 10 20 30 40 50 60",
			wantX: 0,
			wantY: 0,
			wantW: 50,
			wantH: 60,
		},
		{
			name:  "closed polygon (diamond-like)",
			path:  "M 25 0 L 50 25 L 25 50 L 0 25 Z",
			wantX: 0,
			wantY: 0,
			wantW: 50,
			wantH: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bounds := extractBoundsFromPath(tt.path)

			if bounds.X != tt.wantX {
				t.Errorf("X = %v, want %v", bounds.X, tt.wantX)
			}
			if bounds.Y != tt.wantY {
				t.Errorf("Y = %v, want %v", bounds.Y, tt.wantY)
			}
			if bounds.Width != tt.wantW {
				t.Errorf("Width = %v, want %v", bounds.Width, tt.wantW)
			}
			if bounds.Height != tt.wantH {
				t.Errorf("Height = %v, want %v", bounds.Height, tt.wantH)
			}
		})
	}
}

// =============================================================================
// Real D2 SVG Tests
// =============================================================================

func TestParseRealD2SVG(t *testing.T) {
	// Test with actual D2-generated SVG files
	diagram, err := ParseFile("testdata/simple.svg")
	if err != nil {
		t.Skipf("testdata/simple.svg not found: %v", err)
	}

	// Check that we get expected structure
	if len(diagram.Nodes) != 3 {
		t.Errorf("len(Nodes) = %d, want 3", len(diagram.Nodes))
	}
	if len(diagram.Edges) != 2 {
		t.Errorf("len(Edges) = %d, want 2", len(diagram.Edges))
	}

	// Verify bounds are extracted (not all zeros)
	for _, node := range diagram.Nodes {
		if node.Bounds.Width == 0 && node.Bounds.Height == 0 {
			t.Errorf("Node %q has zero bounds", node.ID)
		}
	}

	// Verify labels
	nodeA := diagram.NodeByID("a")
	if nodeA != nil && nodeA.Label != "a" {
		t.Errorf("Node a label = %q, want %q", nodeA.Label, "a")
	}
}

func TestParseContainerD2SVG(t *testing.T) {
	diagram, err := ParseFile("testdata/container.svg")
	if err != nil {
		t.Skipf("testdata/container.svg not found: %v", err)
	}

	// Should have container + 2 inner nodes
	if len(diagram.Nodes) != 3 {
		t.Errorf("len(Nodes) = %d, want 3", len(diagram.Nodes))
	}

	// Should have 1 edge (container-scoped)
	if len(diagram.Edges) != 1 {
		t.Errorf("len(Edges) = %d, want 1", len(diagram.Edges))
	}

	// Check container hierarchy
	container := diagram.NodeByID("container")
	if container == nil {
		t.Fatal("container not found")
	}
	if len(container.Children) != 2 {
		t.Errorf("container has %d children, want 2", len(container.Children))
	}

	// Verify edge source/target are fully qualified
	edge := diagram.Edges[0]
	if !strings.HasPrefix(edge.Source, "container.") {
		t.Errorf("Edge source = %q, should start with 'container.'", edge.Source)
	}
	if !strings.HasPrefix(edge.Target, "container.") {
		t.Errorf("Edge target = %q, should start with 'container.'", edge.Target)
	}
}

// =============================================================================
// Diagram Method Tests
// =============================================================================

func TestDiagramMethods(t *testing.T) {
	diagram := &Diagram{
		Nodes: []Node{
			{ID: "a", Label: "Node A", Shape: ShapeRectangle},
			{ID: "b", Label: "Node B", Shape: ShapeCircle},
			{ID: "container", Children: []string{"container.child"}},
			{ID: "container.child", Parent: "container"},
		},
		Edges: []Edge{
			{ID: "(a -> b)[0]", Source: "a", Target: "b"},
		},
	}

	// Test RootNodes
	roots := diagram.RootNodes()
	if len(roots) != 3 {
		t.Errorf("len(RootNodes) = %d, want 3", len(roots))
	}

	// Test ContainerNodes
	containers := diagram.ContainerNodes()
	if len(containers) != 1 {
		t.Errorf("len(ContainerNodes) = %d, want 1", len(containers))
	}

	// Test LeafNodes
	leaves := diagram.LeafNodes()
	if len(leaves) != 3 {
		t.Errorf("len(LeafNodes) = %d, want 3", len(leaves))
	}

	// Test EdgesFrom
	edgesFromA := diagram.EdgesFrom("a")
	if len(edgesFromA) != 1 {
		t.Errorf("len(EdgesFrom(\"a\")) = %d, want 1", len(edgesFromA))
	}

	// Test EdgesTo
	edgesToB := diagram.EdgesTo("b")
	if len(edgesToB) != 1 {
		t.Errorf("len(EdgesTo(\"b\")) = %d, want 1", len(edgesToB))
	}
}

func TestDiagramJSON(t *testing.T) {
	diagram := &Diagram{
		Version: "0.6.0",
		ViewBox: Bounds{X: 0, Y: 0, Width: 200, Height: 200},
		Nodes: []Node{
			{ID: "a", Label: "Node A", Shape: ShapeRectangle},
		},
		Edges: []Edge{},
	}

	// Test JSON output
	jsonBytes, err := diagram.JSON()
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}

	json := string(jsonBytes)
	if !strings.Contains(json, `"version":"0.6.0"`) {
		t.Error("JSON missing version")
	}
	if !strings.Contains(json, `"id":"a"`) {
		t.Error("JSON missing node id")
	}

	// Test indented JSON
	indentedBytes, err := diagram.JSONIndent("", "  ")
	if err != nil {
		t.Fatalf("JSONIndent() failed: %v", err)
	}

	indented := string(indentedBytes)
	if !strings.Contains(indented, "\n") {
		t.Error("JSONIndent should contain newlines")
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestMightBeBase64(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"YQ==", true},
		{"abc123", true},
		{"abc+/=", true},
		{"shape", true},          // Valid base64 chars
		{"fill-N7", false},       // Contains hyphen
		{"stroke-B1", false},     // Contains hyphen
		{"d2-123", false},        // Contains hyphen
		{"class name", false},    // Contains space
		{"", false},              // Empty
	}

	for _, tt := range tests {
		got := mightBeBase64(tt.input)
		if got != tt.want {
			t.Errorf("mightBeBase64(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
