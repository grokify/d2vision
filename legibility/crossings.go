package legibility

import (
	"sort"

	"github.com/grokify/d2vision"
)

// Crossing is an arrow drawn over a box that is neither of its ends, which
// hides part of the box and makes the arrow hard to follow.
type Crossing struct {
	// Edge is the arrow's ID, and From and To the boxes it joins.
	Edge string `json:"edge"`
	From string `json:"from"`
	To   string `json:"to"`
	// Node is the box it is drawn over.
	Node string `json:"node"`
}

// crossingInset keeps an arrow that only touches a box's edge, as one leaving
// a box beside it may, from counting as drawn over it.
const crossingInset = 3

// Crossings lists the arrows drawn over boxes they do not join, in edge
// order. Containers are not counted: an arrow may cross into one. A grid
// layout draws arrows straight, over whatever is in the way; a layout engine
// that routes them, such as ELK, draws them around boxes.
func Crossings(d *d2vision.Diagram) []Crossing {
	var boxes []d2vision.Node
	for _, n := range d.Nodes {
		if !n.HasChildren() && n.Bounds.Width > 2*crossingInset && n.Bounds.Height > 2*crossingInset {
			boxes = append(boxes, n)
		}
	}
	sort.SliceStable(boxes, func(i, j int) bool { return boxes[i].ID < boxes[j].ID })
	var out []Crossing
	for _, e := range d.Edges {
		for _, n := range boxes {
			if n.ID == e.Source || n.ID == e.Target {
				continue
			}
			if pathCrosses(e.Path, n.Bounds) {
				out = append(out, Crossing{Edge: e.ID, From: e.Source, To: e.Target, Node: n.ID})
			}
		}
	}
	return out
}

// pathCrosses reports whether any segment of a path enters a box, inset by
// crossingInset.
func pathCrosses(path []d2vision.Point, b d2vision.Bounds) bool {
	x0, y0 := b.X+crossingInset, b.Y+crossingInset
	x1, y1 := b.X+b.Width-crossingInset, b.Y+b.Height-crossingInset
	for i := 1; i < len(path); i++ {
		if segmentEnters(path[i-1], path[i], x0, y0, x1, y1) {
			return true
		}
	}
	return false
}

// segmentEnters reports whether the segment p-q passes through the inside of
// the rectangle, by Liang-Barsky clipping.
func segmentEnters(p, q d2vision.Point, x0, y0, x1, y1 float64) bool {
	dx, dy := q.X-p.X, q.Y-p.Y
	t0, t1 := 0.0, 1.0
	for _, c := range [4][2]float64{
		{-dx, p.X - x0}, {dx, x1 - p.X}, {-dy, p.Y - y0}, {dy, y1 - p.Y},
	} {
		pp, qq := c[0], c[1]
		if pp == 0 {
			if qq <= 0 {
				return false
			}
			continue
		}
		r := qq / pp
		if pp < 0 {
			t0 = max(t0, r)
		} else {
			t1 = min(t1, r)
		}
		if t0 >= t1 {
			return false
		}
	}
	return true
}
