package d2vision

import (
	"testing"
)

// =============================================================================
// Base64 Encoding/Decoding Tests
// =============================================================================

func TestDecodeBase64ID(t *testing.T) {
	tests := []struct {
		encoded string
		decoded string
		wantErr bool
	}{
		// Simple node IDs
		{"YQ==", "a", false},
		{"Yg==", "b", false},
		{"Yw==", "c", false},
		{"YWJj", "abc", false},

		// Edge IDs
		{"KGEgLT4gYilbMF0=", "(a -> b)[0]", false},

		// Invalid base64
		{"!!!invalid!!!", "", true},
		{"not-base64", "", true},
	}

	for _, tt := range tests {
		got, err := DecodeBase64ID(tt.encoded)
		if (err != nil) != tt.wantErr {
			t.Errorf("DecodeBase64ID(%q) error = %v, wantErr %v", tt.encoded, err, tt.wantErr)
			continue
		}
		if got != tt.decoded {
			t.Errorf("DecodeBase64ID(%q) = %q, want %q", tt.encoded, got, tt.decoded)
		}
	}
}

// TestDecodeBase64IDWithHTMLEntities verifies that HTML entities are unescaped.
// D2 HTML-encodes special characters BEFORE base64 encoding, so:
//   "(a -> b)[0]" is stored as "(a -&gt; b)[0]" then base64 encoded
//
// This was a bug we discovered when edge IDs containing ">" were not
// being recognized as edges because they contained "&gt;" after decoding.
func TestDecodeBase64IDWithHTMLEntities(t *testing.T) {
	// This is how D2 actually encodes "(a -> b)[0]":
	// 1. HTML encode: "(a -&gt; b)[0]"
	// 2. Base64 encode: "KGEgLSZndDsgYilbMF0="
	encoded := "KGEgLSZndDsgYilbMF0="
	want := "(a -> b)[0]"

	got, err := DecodeBase64ID(encoded)
	if err != nil {
		t.Fatalf("DecodeBase64ID(%q) error = %v", encoded, err)
	}
	if got != want {
		t.Errorf("DecodeBase64ID(%q) = %q, want %q", encoded, got, want)
	}

	// Verify it's recognized as an edge
	if !IsEdgeID(got) {
		t.Errorf("IsEdgeID(%q) = false, want true", got)
	}
}

func TestEncodeBase64ID(t *testing.T) {
	tests := []struct {
		id      string
		encoded string
	}{
		{"a", "YQ=="},
		{"b", "Yg=="},
		{"c", "Yw=="},
		{"abc", "YWJj"},
		{"(a -> b)[0]", "KGEgLT4gYilbMF0="},
	}

	for _, tt := range tests {
		got := EncodeBase64ID(tt.id)
		if got != tt.encoded {
			t.Errorf("EncodeBase64ID(%q) = %q, want %q", tt.id, got, tt.encoded)
		}
	}
}

// =============================================================================
// Edge ID Detection Tests
// =============================================================================

func TestIsEdgeID(t *testing.T) {
	tests := []struct {
		id     string
		isEdge bool
	}{
		// Not edges
		{"a", false},
		{"b", false},
		{"container.node", false},
		{"some random text", false},

		// Root edges (various directions)
		{"(a -> b)[0]", true},
		{"(source -> target)[1]", true},
		{"(a <- b)[0]", true},
		{"(a -- b)[0]", true},

		// Edges with complex node IDs
		{"(container.a -> container.b)[0]", true},

		// Container-scoped edges
		{"container.(inner1 -> inner2)[0]", true},
		{"outer.inner.(a -> b)[0]", true},
	}

	for _, tt := range tests {
		got := IsEdgeID(tt.id)
		if got != tt.isEdge {
			t.Errorf("IsEdgeID(%q) = %v, want %v", tt.id, got, tt.isEdge)
		}
	}
}

// =============================================================================
// Edge ID Parsing Tests
// =============================================================================

func TestParseEdgeID(t *testing.T) {
	tests := []struct {
		id        string
		wantOK    bool
		source    string
		target    string
		direction string
		container string
	}{
		// Root edges
		{"(a -> b)[0]", true, "a", "b", "->", ""},
		{"(source -> target)[1]", true, "source", "target", "->", ""},
		{"(a -- b)[0]", true, "a", "b", "--", ""},

		// Reverse direction: (a <- b) means b -> a
		{"(a <- b)[0]", true, "b", "a", "<-", ""},

		// Container-scoped edges
		// "container.(inner1 -> inner2)[0]" should return:
		//   source: "container.inner1", target: "container.inner2"
		{"container.(inner1 -> inner2)[0]", true, "container.inner1", "container.inner2", "->", "container"},
		{"outer.inner.(a -> b)[0]", true, "outer.inner.a", "outer.inner.b", "->", "outer.inner"},

		// Not edges
		{"not an edge", false, "", "", "", ""},
		{"a", false, "", "", "", ""},
		{"container.node", false, "", "", "", ""},
	}

	for _, tt := range tests {
		ep, ok := ParseEdgeID(tt.id)
		if ok != tt.wantOK {
			t.Errorf("ParseEdgeID(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if ep.Source != tt.source {
			t.Errorf("ParseEdgeID(%q) Source = %q, want %q", tt.id, ep.Source, tt.source)
		}
		if ep.Target != tt.target {
			t.Errorf("ParseEdgeID(%q) Target = %q, want %q", tt.id, ep.Target, tt.target)
		}
		if ep.Container != tt.container {
			t.Errorf("ParseEdgeID(%q) Container = %q, want %q", tt.id, ep.Container, tt.container)
		}
		if ep.Direction != tt.direction {
			t.Errorf("ParseEdgeID(%q) Direction = %q, want %q", tt.id, ep.Direction, tt.direction)
		}
	}
}

// TestContainerScopedEdges verifies that edges inside containers are handled.
// D2 uses IDs like "container.(a -> b)[0]" for edges inside containers.
// The source and target should be fully qualified with the container prefix.
func TestContainerScopedEdges(t *testing.T) {
	id := "container.(inner1 -> inner2)[0]"

	if !IsEdgeID(id) {
		t.Errorf("IsEdgeID(%q) = false, want true", id)
	}

	ep, ok := ParseEdgeID(id)
	if !ok {
		t.Fatalf("ParseEdgeID(%q) returned false", id)
	}

	// Source and target should include container prefix
	if ep.Source != "container.inner1" {
		t.Errorf("Source = %q, want %q", ep.Source, "container.inner1")
	}
	if ep.Target != "container.inner2" {
		t.Errorf("Target = %q, want %q", ep.Target, "container.inner2")
	}
	if ep.Container != "container" {
		t.Errorf("Container = %q, want %q", ep.Container, "container")
	}
}

// =============================================================================
// Node Hierarchy Tests
// =============================================================================

func TestExtractParentID(t *testing.T) {
	tests := []struct {
		nodeID string
		parent string
	}{
		// Top-level nodes have no parent
		{"a", ""},
		{"node", ""},

		// Single level of nesting
		{"container.node", "container"},

		// Multiple levels of nesting
		{"a.b.c", "a.b"},
		{"outer.inner.leaf", "outer.inner"},
		{"a.b.c.d.e", "a.b.c.d"},
	}

	for _, tt := range tests {
		got := ExtractParentID(tt.nodeID)
		if got != tt.parent {
			t.Errorf("ExtractParentID(%q) = %q, want %q", tt.nodeID, got, tt.parent)
		}
	}
}

func TestExtractBaseName(t *testing.T) {
	tests := []struct {
		nodeID   string
		baseName string
	}{
		// Top-level nodes
		{"a", "a"},
		{"node", "node"},

		// Nested nodes
		{"container.node", "node"},
		{"a.b.c", "c"},
		{"outer.inner.leaf", "leaf"},
	}

	for _, tt := range tests {
		got := ExtractBaseName(tt.nodeID)
		if got != tt.baseName {
			t.Errorf("ExtractBaseName(%q) = %q, want %q", tt.nodeID, got, tt.baseName)
		}
	}
}

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a", "a"},
		{"  a  ", "a"},
		{"\t\na\n\t", "a"},
		{"(a -> b)[0]", "(a -> b)[0]"},
		{"  (a -> b)[0]  ", "(a -> b)[0]"},
	}

	for _, tt := range tests {
		got := NormalizeID(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
