package d2vision

import (
	"encoding/base64"
	"html"
	"regexp"
	"strings"
)

// =============================================================================
// D2 ID Encoding/Decoding
// =============================================================================
//
// D2 encodes element IDs as base64 in CSS class names within generated SVGs.
// This allows D2 to use arbitrary characters in IDs (spaces, arrows, etc.)
// while still producing valid CSS class names.
//
// Examples:
//   - "a"                    -> "YQ=="
//   - "(a -> b)[0]"          -> "KGEgLT4gYilbMF0="
//   - "container.inner"      -> "Y29udGFpbmVyLmlubmVy"
//
// IMPORTANT: D2 HTML-encodes special characters BEFORE base64 encoding.
// For example, ">" becomes "&gt;" before encoding. So:
//   - "(a -> b)[0]" is stored as "(a -&gt; b)[0]" then base64 encoded
//   - We must unescape HTML entities after base64 decoding

// edgeIDPattern matches D2 edge ID formats:
//   - Root edge: (source -> target)[index]
//   - Container-scoped edge: container.(source -> target)[index]
//
// The pattern captures: [full, container?, source, direction, target, index?]
var edgeIDPattern = regexp.MustCompile(`^(?:([^(]+)\.)?\((.+?)\s*(<-|->|--)\s*(.+?)\)(?:\[(\d+)\])?$`)

// DecodeBase64ID decodes a base64-encoded D2 element ID.
//
// D2 HTML-encodes special characters (like >) before base64 encoding,
// so we must unescape HTML entities after decoding to get the original ID.
func DecodeBase64ID(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	// Unescape HTML entities (D2 encodes > as &gt; before base64)
	return html.UnescapeString(string(decoded)), nil
}

// EncodeBase64ID encodes a D2 element ID to base64.
// Note: This does NOT HTML-encode first, so it won't produce the same
// result as D2 for IDs containing special characters like >.
func EncodeBase64ID(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

// =============================================================================
// Edge ID Parsing
// =============================================================================
//
// D2 edge IDs follow specific patterns:
//
// Root edges (not inside a container):
//   - "(a -> b)[0]"           - edge from a to b (index 0)
//   - "(source -> target)[1]" - second edge between same nodes
//   - "(a <- b)[0]"           - reverse direction (b to a)
//   - "(a -- b)[0]"           - bidirectional
//
// Container-scoped edges (inside a container):
//   - "container.(a -> b)[0]"       - edge inside "container"
//   - "outer.inner.(x -> y)[0]"     - nested containers
//
// The index [N] disambiguates multiple edges between the same nodes.

// IsEdgeID returns true if the decoded ID represents an edge.
func IsEdgeID(id string) bool {
	return edgeIDPattern.MatchString(id)
}

// EdgeEndpoints represents the parsed components of an edge ID.
type EdgeEndpoints struct {
	Source    string // Source node ID (fully qualified with container prefix)
	Target    string // Target node ID (fully qualified with container prefix)
	Direction string // Arrow direction: "->", "<-", or "--"
	Index     string // Edge index, e.g., "0"
	Container string // Container scope, e.g., "container" for "container.(a -> b)[0]"
}

// ParseEdgeID extracts source and target node IDs from an edge ID.
//
// For container-scoped edges like "container.(a -> b)[0]", the source and
// target are returned with the container prefix: "container.a", "container.b"
func ParseEdgeID(id string) (EdgeEndpoints, bool) {
	matches := edgeIDPattern.FindStringSubmatch(id)
	if matches == nil {
		return EdgeEndpoints{}, false
	}

	// matches: [full, container, source, direction, target, index]
	ep := EdgeEndpoints{
		Container: strings.TrimSpace(matches[1]),
		Source:    strings.TrimSpace(matches[2]),
		Target:    strings.TrimSpace(matches[4]),
		Direction: matches[3],
	}
	if len(matches) > 5 {
		ep.Index = matches[5]
	}

	// Prepend container prefix to make fully-qualified node IDs
	if ep.Container != "" {
		ep.Source = ep.Container + "." + ep.Source
		ep.Target = ep.Container + "." + ep.Target
	}

	// Normalize reverse direction: (a <- b) means b -> a
	if ep.Direction == "<-" {
		ep.Source, ep.Target = ep.Target, ep.Source
	}

	return ep, true
}

// =============================================================================
// Node ID Hierarchy
// =============================================================================
//
// D2 uses dot-separated IDs to represent container hierarchy:
//   - "a"                 - top-level node
//   - "container.a"       - node "a" inside "container"
//   - "outer.inner.leaf"  - deeply nested node

// ExtractParentID extracts the parent container ID from a nested node ID.
// Returns empty string if the node has no parent (is top-level).
//
// Examples:
//   - "a"                 -> ""
//   - "container.node"    -> "container"
//   - "a.b.c"             -> "a.b"
func ExtractParentID(nodeID string) string {
	lastDot := strings.LastIndex(nodeID, ".")
	if lastDot == -1 {
		return ""
	}
	return nodeID[:lastDot]
}

// ExtractBaseName extracts the base name from a node ID (without parent prefix).
//
// Examples:
//   - "a"                 -> "a"
//   - "container.node"    -> "node"
//   - "a.b.c"             -> "c"
func ExtractBaseName(nodeID string) string {
	lastDot := strings.LastIndex(nodeID, ".")
	if lastDot == -1 {
		return nodeID
	}
	return nodeID[lastDot+1:]
}

// NormalizeID cleans up a D2 ID by trimming whitespace.
func NormalizeID(id string) string {
	return strings.TrimSpace(id)
}
