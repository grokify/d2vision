// Package lint provides deterministic, compiler-based static checks for D2
// diagrams — problems that can be detected structurally, without an LLM. (An
// LLM/agent may still be the right tool to *fix* a finding; detecting one is
// pure analysis.)
package lint

import (
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
)

// Finding is a single static-analysis result. Code is the stable rule
// identifier (see code.go); Message/Suggestion are its human-facing text.
type Finding struct {
	Code       Code   `json:"code"`
	Severity   string `json:"severity"` // "error" | "warning" | "info"
	Object     string `json:"object"`   // the D2 object id (absolute path)
	Near       string `json:"near"`     // the offending near constant, when relevant
	Line       int    `json:"line"`     // 1-indexed source line, 0 if unknown
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// cornerNears are the four `near` constants that pin an element to a corner of
// the diagram's bounding box. A corner pin places the element diagonally from
// the graph body, leaving roughly half the canvas empty and a mismatched
// angled legend-to-diagram relationship.
var cornerNears = map[string]bool{
	"top-left":     true,
	"top-right":    true,
	"bottom-left":  true,
	"bottom-right": true,
}

// CheckCornerNear compiles D2 source and returns an error-severity Finding for
// every object pinned to a corner via a constant `near`
// (top-left/top-right/bottom-left/bottom-right). This is the corner-legend
// anti-pattern: it produces a split-quadrant layout with ~50% whitespace. The
// fix is an edge-center near (top-center/bottom-center/center-left/center-right)
// on the diagram's short axis — chosen by a human or agent, not this check.
//
// Detection is structural: an object literally named "top-left" (a real node,
// not a near constant) is not flagged, because d2graph.IsConstantNear
// distinguishes the two.
func CheckCornerNear(d2src string) ([]Finding, error) {
	g, _, err := d2compiler.Compile("", strings.NewReader(d2src), &d2compiler.CompileOptions{})
	if err != nil {
		return nil, fmt.Errorf("compile d2: %w", err)
	}

	var findings []Finding
	for _, obj := range g.Objects {
		if !obj.IsConstantNear() {
			continue
		}
		key := d2graph.Key(obj.NearKey)
		if len(key) == 0 {
			continue
		}
		near := key[0]
		if !cornerNears[near] {
			continue
		}
		line := 0
		if obj.NearKey != nil {
			// d2ast positions are 0-indexed; report a 1-indexed line.
			line = obj.NearKey.Range.Start.Line + 1
		}
		rule, _ := Lookup(CodeCornerNear)
		findings = append(findings, Finding{
			Code:     CodeCornerNear,
			Severity: rule.Severity,
			Object:   obj.AbsID(),
			Near:     near,
			Line:     line,
			Message: fmt.Sprintf(
				"object %q sets near: %s — a corner pin places it diagonally from the diagram, leaving ~50%% of the canvas empty",
				obj.AbsID(), near),
			Suggestion: rule.Suggestion,
		})
	}
	return findings, nil
}

// HasErrors reports whether any finding is error severity.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}
