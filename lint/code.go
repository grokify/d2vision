package lint

import "sort"

// Code is a stable identifier for a lint rule. Like net/http status constants,
// the code is the machine contract: it is what appears in CLI/JSON output and
// what CI filters, baselines, and suppressions key off. Human-facing text
// (title, suggestion, remediation) is looked up separately via the registry —
// so the prose can be reworded without breaking anything that depends on a code.
type Code string

// Lint rule codes. Keep these stable; add new ones rather than renaming.
const (
	// CodeCornerNear: an object is pinned to a corner via `near`
	// (top-left/top-right/bottom-left/bottom-right), producing a diagonal
	// split-quadrant layout with ~50% whitespace.
	CodeCornerNear Code = "corner-near"
	// CodeTextOverlap: a connection label's rendered rectangle overlaps another
	// label or an unrelated node (a post-layout, engine-dependent check).
	CodeTextOverlap Code = "text-overlap"
	// CodeCrossContainerEdge: an edge crosses container boundaries and may
	// cause vertical stacking without a grid.
	CodeCrossContainerEdge Code = "cross-container-edge"
	// CodeDuplicateNode: a node id is defined more than once.
	CodeDuplicateNode Code = "duplicate-node"
	// CodeDeepNesting: container nesting is deep enough to affect layout.
	CodeDeepNesting Code = "deep-nesting"
	// CodeMissingGrid: multiple root containers without grid-columns.
	CodeMissingGrid Code = "missing-grid"
	// CodeMixedDirections: inconsistent `direction` settings across the graph.
	CodeMixedDirections Code = "mixed-directions"
	// CodeCompileSkipped: the D2 did not compile, so structural (graph-based)
	// checks were skipped; the text-based checks still ran.
	CodeCompileSkipped Code = "compile-skipped"
)

// Severity levels for a rule.
const (
	// SeverityError marks a finding that should fail a lint run.
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)

// Rule is the registry entry for a Code: its severity and the human-facing
// text. Suggestion is the one-line inline fix shown with each finding;
// Remediation is the fuller protocol surfaced on demand (lint --explain).
type Rule struct {
	Code        Code   `json:"code"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Suggestion  string `json:"suggestion"`
	Remediation string `json:"remediation"`
	DocURL      string `json:"docURL,omitempty"`
}

// rules is the registry: the single source of truth mapping each Code to its
// severity and remediation guidance.
var rules = map[Code]Rule{
	CodeCornerNear: {
		Code:       CodeCornerNear,
		Severity:   SeverityError,
		Title:      "Corner-pinned element",
		Suggestion: "use an edge-center near (top-center | bottom-center | center-left | center-right) so it stacks above/below/left/right of the diagram; pick the diagram's short axis",
		Remediation: `A ` + "`near`" + ` set to a corner constant (top-left, top-right, bottom-left,
bottom-right) pins the element to a corner of the diagram's bounding box. The
graph body flows into a different quadrant, so the element sits diagonally from
it and roughly half the canvas is empty.

Fix: use one of the four edge-center constants instead —
  top-center | bottom-center | center-left | center-right
Each stacks the element cleanly above, below, left, or right of the diagram so
the two share an axis. Choose the edge on the diagram's short axis to minimize
whitespace: a tall (top-down) flow -> center-left/center-right; a wide flow ->
top-center/bottom-center.`,
		DocURL: "docs/lint-rules.md#corner-near",
	},
	CodeTextOverlap: {
		Code:       CodeTextOverlap,
		Severity:   SeverityWarning,
		Title:      "Overlapping text",
		Suggestion: "switch layout engine (elk often separates parallel-edge labels), add spacing, or shorten/relocate the label",
		Remediation: `A connection label's rendered rectangle overlaps another label or an
unrelated node, making the text hard to read. This is a post-layout, engine-
dependent condition — the same source can overlap under dagre yet be clean
under elk (commonly the case when several edges share a node pair).

Fix, in order of preference:
  1. Render/lay out with a different engine (elk spaces parallel-edge labels
     better than dagre).
  2. Reduce crowding: fewer parallel edges between the same node pair, or an
     intermediate node.
  3. Shorten the label or move detail into a legend.

Because it requires a full layout pass, this rule is engine-specific and is
opt-in via config (rules.text-overlap.enabled: true, settings.layout: elk).`,
		DocURL: "docs/lint-rules.md#text-overlap",
	},
	CodeCrossContainerEdge: {
		Code:        CodeCrossContainerEdge,
		Severity:    SeverityWarning,
		Title:       "Cross-container edge",
		Suggestion:  "add 'grid-columns: N' at root level to control horizontal layout",
		Remediation: "An edge between different top-level containers can force the layout engine to stack them vertically. Add `grid-columns: N` at the root to control the arrangement, or reconsider whether the edge should cross the boundary.",
		DocURL:      "docs/lint-rules.md#cross-container-edge",
	},
	CodeDuplicateNode: {
		Code:        CodeDuplicateNode,
		Severity:    SeverityWarning,
		Title:       "Duplicate node definition",
		Suggestion:  "consolidate the node's definitions into one",
		Remediation: "A node id is defined more than once. Later definitions merge into the first, which is easy to do by accident. Consolidate the attributes into a single definition.",
		DocURL:      "docs/lint-rules.md#duplicate-node",
	},
	CodeDeepNesting: {
		Code:        CodeDeepNesting,
		Severity:    SeverityInfo,
		Title:       "Deep container nesting",
		Suggestion:  "flatten the structure if possible",
		Remediation: "Container nesting deeper than ~3 levels can slow layout and make the diagram hard to read. Flatten the hierarchy or split the diagram.",
		DocURL:      "docs/lint-rules.md#deep-nesting",
	},
	CodeMissingGrid: {
		Code:        CodeMissingGrid,
		Severity:    SeverityInfo,
		Title:       "Multiple root containers without a grid",
		Suggestion:  "add 'grid-columns: N' to control horizontal arrangement",
		Remediation: "Several root-level containers with no `grid-columns` leave their arrangement to the layout engine, which often stacks them. Add `grid-columns: N` at the root to place them deliberately.",
		DocURL:      "docs/lint-rules.md#missing-grid",
	},
	CodeMixedDirections: {
		Code:        CodeMixedDirections,
		Severity:    SeverityInfo,
		Title:       "Mixed direction settings",
		Suggestion:  "use consistent directions for a cleaner layout",
		Remediation: "Different `direction` values across the graph can produce an inconsistent flow. Prefer one direction unless a sub-graph genuinely needs its own.",
		DocURL:      "docs/lint-rules.md#mixed-directions",
	},
	CodeCompileSkipped: {
		Code:        CodeCompileSkipped,
		Severity:    SeverityInfo,
		Title:       "Structural checks skipped",
		Suggestion:  "fix the D2 compile error so graph-based checks can run",
		Remediation: "The D2 did not compile, so graph-based (structural) checks were skipped; only the text-based checks ran. Fix the reported compile error and re-lint.",
		DocURL:      "docs/lint-rules.md#compile-skipped",
	},
}

// Lookup returns the Rule for a code and whether it is registered.
func Lookup(code Code) (Rule, bool) {
	r, ok := rules[code]
	return r, ok
}

// Text returns the human-readable title for a code, or "" if the code is not
// registered. It is the analogue of net/http.StatusText.
func Text(code Code) string {
	if r, ok := rules[code]; ok {
		return r.Title
	}
	return ""
}

// SeverityOf returns the registered severity for a code, or "" if unknown.
func SeverityOf(code Code) string {
	if r, ok := rules[code]; ok {
		return r.Severity
	}
	return ""
}

// Explain returns the full remediation guidance for a code, or "" if unknown.
func Explain(code Code) string {
	if r, ok := rules[code]; ok {
		return r.Remediation
	}
	return ""
}

// Rules returns every registered rule, sorted by code, for listing and docs.
func Rules() []Rule {
	out := make([]Rule, 0, len(rules))
	for _, r := range rules {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}
