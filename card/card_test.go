package card

import (
	"encoding/xml"
	"io"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

const testCard = `{
  "loop": 4,
  "panels": [{
    "id": "p", "height": 300, "title": "Title", "body": "Some body text.",
    "diagram": {
      "nodes": [
        {"id": "a", "icon": "A", "label": "Alpha", "x": 50, "y": 60},
        {"id": "b", "label": "Beta", "shape": "box", "x": 250, "y": 60}
      ],
      "edges": [
        {"from": "a", "to": "b", "label": "go", "flow": {"color": "#f00", "delay": 0.5}}
      ]
    }
  }],
  "legend": [{"color": "#f00", "label": "Token"}]
}`

func mustParse(t *testing.T, s string) *Card {
	t.Helper()
	c, err := Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestParseDefaults(t *testing.T) {
	c := mustParse(t, testCard)
	if c.Width != DefaultWidth || c.Height != DefaultHeight {
		t.Errorf("size = %dx%d", c.Width, c.Height)
	}
	e := c.Panels[0].Diagram.Edges[0]
	if e.ID != "a-b" {
		t.Errorf("edge id = %q, want a-b", e.ID)
	}
	if len(e.Flows) != 1 || e.Flow != nil {
		t.Fatalf("Flow should be normalized into Flows, got Flow=%v Flows=%v", e.Flow, e.Flows)
	}
	if f := e.Flows[0]; f.Mode != FlowForward || f.Speed != DefaultFlowSpeed || f.Count != 1 {
		t.Errorf("flow defaults not applied: %+v", f)
	}
	if got := c.Panels[0].Diagram.Nodes[0].Shape; got != ShapeCircle {
		t.Errorf("node shape = %q, want circle", got)
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	if _, err := Parse(strings.NewReader(`{"panels": [], "bogus": 1}`)); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{"no panels", `{"panels": []}`, "no panels"},
		{"unknown node", `{"panels":[{"height":100,"diagram":{"nodes":[{"id":"a","x":10,"y":10}],"edges":[{"from":"a","to":"zz"}]}}]}`, `unknown to node "zz"`},
		{"duplicate node", `{"panels":[{"height":100,"diagram":{"nodes":[{"id":"a"},{"id":"a"}]}}]}`, "duplicate node"},
		{"bad flow mode", `{"panels":[{"height":100,"diagram":{"nodes":[{"id":"a"},{"id":"b"}],"edges":[{"from":"a","to":"b","flow":{"mode":"sideways"}}]}}]}`, "unknown flow mode"},
		{"too tall", `{"panels":[{"height":600},{"height":600}]}`, "exceeding"},
		{"ragged table", `{"panels":[{"kind":"table","height":100,"table":{"columns":["a","b"],"rows":[["x"]]}}]}`, "has 1 cells, want 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.json))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestEdgeQuadClipsToNodeOutlines(t *testing.T) {
	a := Node{ID: "a", Shape: ShapeCircle, X: 0, Y: 0, W: 40, H: 40}
	b := Node{ID: "b", Shape: ShapeBox, X: 200, Y: 0, W: 80, H: 30}
	q := edgeQuad(a, b, 0)
	if math.Abs(q.P0.X-(20+edgeGap)) > 1e-9 || q.P0.Y != 0 {
		t.Errorf("start = %+v", q.P0)
	}
	if math.Abs(q.P2.X-(200-40-edgeGap)) > 1e-9 {
		t.Errorf("end = %+v", q.P2)
	}
	want := q.P2.X - q.P0.X
	if got := q.length(); math.Abs(got-want) > 0.01 {
		t.Errorf("straight length = %.3f, want %.3f", got, want)
	}
	if bent := edgeQuad(a, b, 40); bent.length() <= q.length() {
		t.Error("a bent edge should be longer than a straight one")
	}
}

func TestWrap(t *testing.T) {
	lines := wrap("the quick brown fox jumps over the lazy dog", 16, 120, false)
	if len(lines) < 3 {
		t.Fatalf("expected wrapping, got %q", lines)
	}
	for _, l := range lines {
		if w := textWidth(l, 16, false); w > 120 {
			t.Errorf("line %q is %.0f wide, over 120", l, w)
		}
	}
	if got := wrap("a\nb", 12, 500, false); len(got) != 2 {
		t.Errorf("forced newline: got %q", got)
	}
}

func TestRenderStaticVsAnimated(t *testing.T) {
	c := mustParse(t, testCard)
	static, err := c.Render(Options{})
	if err != nil {
		t.Fatal(err)
	}
	anim, err := c.Render(Options{Animate: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(static), "animateMotion") {
		t.Error("static render must not contain animation")
	}
	if !strings.Contains(string(anim), "animateMotion") {
		t.Error("animated render missing animateMotion")
	}
	for _, id := range []string{`id="panel-p"`, `id="edge-p-a-b"`, `id="path-p-a-b"`, `id="node-p-a"`, `id="legend"`} {
		if !strings.Contains(string(anim), id) {
			t.Errorf("missing stable id %s", id)
		}
	}
	again, _ := c.Render(Options{Animate: true})
	if string(again) != string(anim) {
		t.Error("render is not deterministic")
	}
	wellFormed(t, anim)
}

func wellFormed(t *testing.T, svg []byte) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(string(svg)))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatalf("svg is not well-formed XML: %v", err)
		}
	}
}

func TestRenderEscapesText(t *testing.T) {
	c := mustParse(t, strings.Replace(testCard, "Some body text.", `a <b> & \"c\"`, 1))
	svg, err := c.Render(Options{})
	if err != nil {
		t.Fatal(err)
	}
	wellFormed(t, svg)
	if strings.Contains(string(svg), "<b>") {
		t.Error("text was not escaped")
	}
}

func TestRenderFlowMustFitLoop(t *testing.T) {
	c := mustParse(t, strings.Replace(testCard, `"delay": 0.5`, `"delay": 3.9`, 1))
	_, err := c.Render(Options{Animate: true})
	if err == nil || !strings.Contains(err.Error(), "loop") {
		t.Fatalf("err = %v, want a loop-length error", err)
	}
	// Static rendering has no timing, so the same card still renders.
	if _, err := c.Render(Options{}); err != nil {
		t.Errorf("static render should not depend on flow timing: %v", err)
	}
}

func TestRenderTextMustFitPanel(t *testing.T) {
	long := strings.Repeat("word ", 200)
	c := mustParse(t, strings.Replace(testCard, "Some body text.", long, 1))
	_, err := c.Render(Options{})
	if err == nil || !strings.Contains(err.Error(), "text needs") {
		t.Fatalf("err = %v, want a text overflow error", err)
	}
}

func TestRenderNodeOutsideDiagramArea(t *testing.T) {
	c := mustParse(t, strings.Replace(testCard, `"x": 250`, `"x": 5000`, 1))
	if _, err := c.Render(Options{}); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("err = %v, want an outside-diagram-area error", err)
	}
}

func TestFlowModes(t *testing.T) {
	q := quad{P0: pt{0, 0}, C: pt{50, 0}, P2: pt{100, 0}}
	count := func(mode FlowMode, n int) (circles int, svg string) {
		f := &Flow{Mode: mode, Color: "#fff", Size: 5, Speed: 100, Count: n, Gap: 0.25, Pause: 0.1}
		out, err := flowDots("p", Edge{ID: "e"}, f, q, 10)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Count(out, "<circle"), out
	}
	if n, _ := count(FlowForward, 3); n != 3 {
		t.Errorf("forward count=3 gave %d dots", n)
	}
	if n, _ := count(FlowBoth, 2); n != 4 {
		t.Errorf("both count=2 should be 4 dots (2 each way), got %d", n)
	}
	_, alt := count(FlowAlternate, 1)
	if !strings.Contains(alt, `keyPoints="0;0;1;1;0;0"`) {
		t.Errorf("alternate should go out and back: %s", alt)
	}
	_, rev := count(FlowReverse, 1)
	if !strings.Contains(rev, `keyPoints="1;1;0;0"`) {
		t.Errorf("reverse should run 1 -> 0: %s", rev)
	}
}

// TestExamplesRender keeps the documented examples working: every card under
// examples/card must parse, render statically and animated, and have a loop.
func TestExamplesRender(t *testing.T) {
	files, err := filepath.Glob("../examples/card/*.card.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no example cards found")
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			c, err := ParseFile(f)
			if err != nil {
				t.Fatal(err)
			}
			for _, w := range c.Warnings() {
				t.Errorf("examples must be fully protocol-backed, got warning: %s", w)
			}
			for _, animate := range []bool{false, true} {
				svg, err := c.Render(Options{Animate: animate})
				if err != nil {
					t.Fatalf("render (animate=%v): %v", animate, err)
				}
				wellFormed(t, svg)
			}
			if c.LoopSeconds() <= 0 {
				t.Error("loop should be positive")
			}
		})
	}
}
