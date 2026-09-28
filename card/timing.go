package card

import (
	"math"
	"sort"
)

// LoopSeconds returns the effective animation loop length: Loop if set,
// otherwise derived from the flows. Use it for frame capture.
func (c *Card) LoopSeconds() float64 {
	c.ApplyDefaults()
	c.resolveTiming()
	return c.loop
}

// resolveTiming decides when every flow starts and how long the loop is.
//
// Within a panel, flows with a Step run in ascending step order: a step starts
// when the previous one has fully finished (plus StepGap), and flows sharing a
// step start together. Flow.Delay is then an offset within the step. Flows
// with no Step start at their Delay from the top of the loop. Each panel is
// its own timeline starting at zero, so panels animate side by side.
//
// When Loop is 0 the loop is the latest finish time plus LoopRest, rounded up
// to 50 ms. It is idempotent and never modifies user-set fields.
func (c *Card) resolveTiming() {
	var end float64
	for pi := range c.Panels {
		d := c.Panels[pi].Diagram
		if d == nil {
			continue
		}
		nodes := make(map[string]Node, len(d.Nodes))
		for _, n := range d.Nodes {
			nodes[n.ID] = n
		}

		type ref struct {
			f    *Flow
			span float64
		}
		steps := map[int][]ref{}
		for ei := range d.Edges {
			e := &d.Edges[ei]
			length := edgeQuad(nodes[e.From], nodes[e.To], e.Bend).length()
			for fi := range e.Flows {
				f := &e.Flows[fi]
				span := flowSpan(f, length/f.Speed)
				if f.Step > 0 {
					steps[f.Step] = append(steps[f.Step], ref{f, span})
					continue
				}
				f.at = f.Delay
				end = max(end, f.at+span)
			}
		}

		order := make([]int, 0, len(steps))
		for s := range steps {
			order = append(order, s)
		}
		sort.Ints(order)
		var t float64
		for _, s := range order {
			var dur float64
			for _, r := range steps[s] {
				r.f.at = t + r.f.Delay
				dur = max(dur, r.f.Delay+r.span)
			}
			t += dur
			end = max(end, t)
			t += c.StepGap
		}
	}

	if c.Loop > 0 {
		c.loop = c.Loop
		return
	}
	c.loop = math.Ceil((end+c.LoopRest)*20) / 20
}
