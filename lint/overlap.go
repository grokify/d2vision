package lint

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/grokify/d2vision/render"
)

// minOverlapPx is the intersection each axis must exceed for two boxes to count
// as overlapping. It keeps a 1px touch or rounding jitter from being reported.
const minOverlapPx = 2.0

// box is an axis-aligned rectangle in laid-out diagram coordinates.
type box struct {
	x1, y1, x2, y2 float64
}

func (a box) overlaps(b box) bool {
	dx := min(a.x2, b.x2) - max(a.x1, b.x1)
	dy := min(a.y2, b.y2) - max(a.y1, b.y1)
	return dx > minOverlapPx && dy > minOverlapPx
}

// labelBox is a connection's rendered label rectangle plus the endpoints it
// belongs to (so a label overlapping its own endpoint node is not flagged).
type labelBox struct {
	text     string
	src, dst string
	box
}

// CheckTextOverlap lays out the D2 source with the given engine ("dagre" or
// "elk"; "" = dagre) and reports connection labels whose rendered rectangles
// overlap — either another connection label, or an unrelated node. This is a
// post-layout, engine-dependent check (the same source can overlap under dagre
// yet be clean under elk), so callers pass the engine they actually render with.
//
// Detection is deterministic; the fix (switch engine, add spacing, or shorten
// the label) is guidance, not applied here.
func CheckTextOverlap(d2src, layout string) ([]Finding, error) {
	diagram, err := render.CompileWithLayout(context.Background(), d2src, layout)
	if err != nil {
		return nil, fmt.Errorf("compile+layout: %w", err)
	}

	// Node rectangles (Pos is the top-left of the shape).
	type shapeRect struct {
		id string
		box
	}
	var shapes []shapeRect
	for _, s := range diagram.Shapes {
		shapes = append(shapes, shapeRect{
			id: s.ID,
			box: box{
				x1: float64(s.Pos.X), y1: float64(s.Pos.Y),
				x2: float64(s.Pos.X + s.Width), y2: float64(s.Pos.Y + s.Height),
			},
		})
	}

	// Connection label rectangles.
	var labels []labelBox
	for _, c := range diagram.Connections {
		if strings.TrimSpace(c.Label) == "" || c.LabelWidth <= 0 || c.LabelHeight <= 0 {
			continue
		}
		tl := c.GetLabelTopLeft()
		if tl == nil {
			continue
		}
		labels = append(labels, labelBox{
			text: c.Label, src: c.Src, dst: c.Dst,
			box: box{x1: tl.X, y1: tl.Y, x2: tl.X + float64(c.LabelWidth), y2: tl.Y + float64(c.LabelHeight)},
		})
	}

	rule, _ := Lookup(CodeTextOverlap)
	var findings []Finding

	// label ↔ label
	for i := 0; i < len(labels); i++ {
		for j := i + 1; j < len(labels); j++ {
			if labels[i].overlaps(labels[j].box) {
				findings = append(findings, Finding{
					Code:       CodeTextOverlap,
					Severity:   rule.Severity,
					Object:     oneLine(labels[i].text),
					Message:    fmt.Sprintf("edge labels overlap: %q and %q", oneLine(labels[i].text), oneLine(labels[j].text)),
					Suggestion: rule.Suggestion,
				})
			}
		}
	}

	// label ↔ unrelated node (skip the label's own endpoints)
	for _, l := range labels {
		for _, s := range shapes {
			if s.id == l.src || s.id == l.dst {
				continue
			}
			if l.overlaps(s.box) {
				findings = append(findings, Finding{
					Code:       CodeTextOverlap,
					Severity:   rule.Severity,
					Object:     oneLine(l.text),
					Message:    fmt.Sprintf("edge label %q overlaps node %q", oneLine(l.text), s.id),
					Suggestion: rule.Suggestion,
				})
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool { return findings[i].Message < findings[j].Message })
	return findings, nil
}

// oneLine collapses a multi-line label to a single trimmed line for messages.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
