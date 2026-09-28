package card

import (
	"math"
	"strings"
	"testing"
)

// stepCard has three nodes in a row, 200 px apart, and edges whose flows are
// wired up per test. Speed 100 px/s makes each ~194 px traversal about 1.94 s.
func stepCard(t *testing.T, edges string, top string) *Card {
	t.Helper()
	return mustParse(t, `{`+top+`"panels":[{"id":"p","height":300,"diagram":{
	  "nodes":[{"id":"a","x":40,"y":100},{"id":"b","x":240,"y":100},{"id":"c","x":440,"y":100}],
	  "edges":[`+edges+`]}}]}`)
}

func flowAt(c *Card, edge string, i int) float64 {
	for _, e := range c.Panels[0].Diagram.Edges {
		if e.ID == edge {
			return e.Flows[i].at
		}
	}
	return math.NaN()
}

func TestStepsRunInOrder(t *testing.T) {
	// Listed out of order on purpose: step numbers, not position, decide.
	c := stepCard(t, `
	  {"from":"b","to":"c","flow":{"step":2,"speed":100}},
	  {"from":"a","to":"b","flow":{"step":1,"speed":100}}`, "")
	c.resolveTiming()
	ab, bc := flowAt(c, "a-b", 0), flowAt(c, "b-c", 0)
	if ab != 0 {
		t.Errorf("step 1 starts at %v, want 0", ab)
	}
	q := edgeQuad(c.Panels[0].Diagram.Nodes[0], c.Panels[0].Diagram.Nodes[1], 0)
	want := q.length()/100 + DefaultStepGap
	if math.Abs(bc-want) > 1e-9 {
		t.Errorf("step 2 starts at %.4f, want %.4f (step 1 duration + gap)", bc, want)
	}
}

func TestEqualStepsRunTogether(t *testing.T) {
	c := stepCard(t, `
	  {"from":"a","to":"b","flow":{"step":1,"speed":100}},
	  {"from":"b","to":"c","flow":{"step":1,"speed":100,"delay":0.5}},
	  {"from":"a","to":"c","bend":30,"flow":{"step":2,"speed":100}}`, "")
	c.resolveTiming()
	if a, b := flowAt(c, "a-b", 0), flowAt(c, "b-c", 0); a != 0 || b != 0.5 {
		t.Errorf("step 1 starts = %v, %v; want 0 and 0.5 (delay is an offset within the step)", a, b)
	}
	// Step 2 waits for the slower member (the delayed one) to finish.
	q := edgeQuad(c.Panels[0].Diagram.Nodes[1], c.Panels[0].Diagram.Nodes[2], 0)
	want := 0.5 + q.length()/100 + DefaultStepGap
	if got := flowAt(c, "a-c", 0); math.Abs(got-want) > 1e-9 {
		t.Errorf("step 2 starts at %.4f, want %.4f", got, want)
	}
}

func TestUnsequencedFlowsKeepDelay(t *testing.T) {
	c := stepCard(t, `{"from":"a","to":"b","flow":{"delay":1.25}}`, "")
	c.resolveTiming()
	if got := flowAt(c, "a-b", 0); got != 1.25 {
		t.Errorf("start = %v, want 1.25", got)
	}
}

func TestPanelsAreIndependentTimelines(t *testing.T) {
	c := mustParse(t, `{"panels":[
	  {"id":"x","height":200,"diagram":{"nodes":[{"id":"a","x":40,"y":60},{"id":"b","x":240,"y":60}],
	    "edges":[{"from":"a","to":"b","flow":{"step":1,"speed":100}}]}},
	  {"id":"y","height":200,"diagram":{"nodes":[{"id":"a","x":40,"y":60},{"id":"b","x":240,"y":60}],
	    "edges":[{"from":"a","to":"b","flow":{"step":1,"speed":100}}]}}]}`)
	c.resolveTiming()
	for i, p := range c.Panels {
		if got := p.Diagram.Edges[0].Flows[0].at; got != 0 {
			t.Errorf("panel %d step 1 starts at %v, want 0", i, got)
		}
	}
}

func TestLoopDerivedFromSteps(t *testing.T) {
	c := stepCard(t, `
	  {"from":"a","to":"b","flow":{"step":1,"speed":100}},
	  {"from":"b","to":"c","flow":{"step":2,"speed":100}}`, `"loopRest":1,`)
	loop := c.LoopSeconds()
	q := edgeQuad(c.Panels[0].Diagram.Nodes[0], c.Panels[0].Diagram.Nodes[1], 0)
	end := 2*(q.length()/100) + DefaultStepGap
	if loop < end+1 || loop > end+1.05 {
		t.Errorf("loop = %.3f, want %.3f rounded up by <= 50ms", loop, end+1)
	}
	// A derived loop always fits what is drawn.
	if _, err := c.Render(Options{Animate: true}); err != nil {
		t.Errorf("derived loop should fit all flows: %v", err)
	}
}

func TestExplicitLoopWinsAndIsChecked(t *testing.T) {
	c := stepCard(t, `
	  {"from":"a","to":"b","flow":{"step":1,"speed":100}},
	  {"from":"b","to":"c","flow":{"step":2,"speed":100}}`, `"loop":2,`)
	if got := c.LoopSeconds(); got != 2 {
		t.Errorf("loop = %v, want the explicit 2", got)
	}
	if _, err := c.Render(Options{Animate: true}); err == nil || !strings.Contains(err.Error(), "loop") {
		t.Errorf("err = %v, want a loop-too-short error", err)
	}
}

func TestResolveTimingIsIdempotent(t *testing.T) {
	c := stepCard(t, `
	  {"from":"a","to":"b","flow":{"step":1}},
	  {"from":"b","to":"c","flow":{"step":2}}`, "")
	first, _ := c.Render(Options{Animate: true})
	second, _ := c.Render(Options{Animate: true})
	if string(first) != string(second) {
		t.Error("rendering twice differs")
	}
	if c.Loop != 0 {
		t.Errorf("user-set Loop was mutated to %v", c.Loop)
	}
}

func TestStepBadges(t *testing.T) {
	c := stepCard(t, `
	  {"from":"a","to":"b","flows":[{"step":1},{"step":2,"mode":"reverse"},{"color":"#0f0"}]}`, "")
	svg, err := c.Render(Options{Animate: true})
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(svg), `class="step-badge"`); n != 2 {
		t.Errorf("got %d badges, want 2 (unsequenced flows get none)", n)
	}
	wellFormed(t, svg)
}

func TestFlowSpanMatchesRuns(t *testing.T) {
	for _, mode := range []FlowMode{FlowForward, FlowReverse, FlowAlternate, FlowBoth} {
		f := &Flow{Mode: mode, Count: 3, Gap: 0.3, Pause: 0.2, Speed: 100, at: 2}
		var end float64
		for _, r := range flowRuns(f, 1.5) {
			end = max(end, r.start+r.dur)
		}
		if got := flowSpan(f, 1.5); math.Abs(got-(end-2)) > 1e-9 {
			t.Errorf("%s: span %.3f != last run end - start %.3f", mode, got, end-2)
		}
	}
}

func TestBadgeAvoidsEdgeLabel(t *testing.T) {
	c := stepCard(t, `{"from":"a","to":"b","label":"consent","flow":{"mode":"alternate","step":1}}`, "")
	svg, err := c.Render(Options{})
	if err != nil {
		t.Fatal(err)
	}
	// The label pill is centered on the edge midpoint (140,100); the badge must
	// be moved off it, not drawn on top.
	i := strings.Index(string(svg), `class="step-badge"`)
	if i < 0 {
		t.Fatal("no badge rendered")
	}
	if strings.Contains(string(svg)[i:i+90], `cy="100"`) {
		t.Error("badge is centered on the label pill")
	}
}

func TestBadgesDrawAboveDots(t *testing.T) {
	c := stepCard(t, `{"from":"a","to":"b","flow":{"step":1}}`, "")
	svg, err := c.Render(Options{Animate: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(svg)
	if strings.Index(s, `class="step-badge"`) < strings.LastIndex(s, "<animateMotion") {
		t.Error("badges must be painted after the dots so numbers stay readable")
	}
}
