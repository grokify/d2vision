package card

import "math"

// pt is a 2D point.
type pt struct{ X, Y float64 }

func (a pt) add(b pt) pt        { return pt{a.X + b.X, a.Y + b.Y} }
func (a pt) sub(b pt) pt        { return pt{a.X - b.X, a.Y - b.Y} }
func (a pt) scale(k float64) pt { return pt{a.X * k, a.Y * k} }
func (a pt) len() float64       { return math.Hypot(a.X, a.Y) }
func (a pt) unit() pt {
	l := a.len()
	if l == 0 {
		return pt{}
	}
	return a.scale(1 / l)
}

// quad is a quadratic Bezier segment. A straight edge has C on the chord.
type quad struct{ P0, C, P2 pt }

func (q quad) at(t float64) pt {
	u := 1 - t
	return q.P0.scale(u * u).add(q.C.scale(2 * u * t)).add(q.P2.scale(t * t))
}

// length approximates the arc length by sampling.
func (q quad) length() float64 {
	const n = 48
	var sum float64
	prev := q.P0
	for i := 1; i <= n; i++ {
		p := q.at(float64(i) / n)
		sum += p.sub(prev).len()
		prev = p
	}
	return sum
}

// edgeGap is the space left between an edge end and a node outline.
const edgeGap = 3.0

// edgeQuad computes the clipped curve for an edge between two nodes. Ends stop
// at the node outlines so dots and arrowheads don't sit under the node fill.
func edgeQuad(from, to Node, bend float64) quad {
	a, b := pt{from.X, from.Y}, pt{to.X, to.Y}
	mid := a.add(b).scale(0.5)
	dir := b.sub(a).unit()
	normal := pt{-dir.Y, dir.X}
	c := mid.add(normal.scale(bend))

	return quad{
		P0: boundary(from, c),
		C:  c,
		P2: boundary(to, c),
	}
}

// boundary returns the point on n's outline in the direction of toward,
// pushed out by edgeGap.
func boundary(n Node, toward pt) pt {
	center := pt{n.X, n.Y}
	d := toward.sub(center).unit()
	if d == (pt{}) {
		return center
	}
	switch n.Shape {
	case ShapeCircle:
		return center.add(d.scale(n.W/2 + edgeGap))
	default:
		hw, hh := n.W/2+edgeGap, n.H/2+edgeGap
		// Scale d until it touches the rectangle.
		tx, ty := math.Inf(1), math.Inf(1)
		if d.X != 0 {
			tx = hw / math.Abs(d.X)
		}
		if d.Y != 0 {
			ty = hh / math.Abs(d.Y)
		}
		return center.add(d.scale(math.Min(tx, ty)))
	}
}
