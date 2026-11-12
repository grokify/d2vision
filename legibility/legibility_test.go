package legibility_test

import (
	"strings"
	"testing"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/legibility"
)

// diagram is w x h pixels with one label of each font size given.
func diagram(w, h float64, label string, sizes ...float64) *d2vision.Diagram {
	d := &d2vision.Diagram{ViewBox: d2vision.Bounds{Width: w, Height: h}}
	for i, s := range sizes {
		d.Nodes = append(d.Nodes, d2vision.Node{
			ID: string(rune('a' + i)), Label: label,
			Style: d2vision.NodeStyle{FontSize: s},
		})
	}
	return d
}

func target(t *testing.T, name string) legibility.Target {
	t.Helper()
	tg, ok := legibility.TargetByName(name)
	if !ok {
		t.Fatalf("no target %q", name)
	}
	return tg
}

// A diagram as wide as a portrait page's text, 468 points, shows 16-pixel
// labels at 16 points.
func TestReadableDiagram(t *testing.T) {
	r, err := legibility.Analyze(diagram(468, 300, "Short label", 16, 20), target(t, "docx-portrait"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Readable || r.Scale != 1 || r.LimitedBy != "width" || r.SmallestPt != 16 || len(r.Advice) != 0 {
		t.Errorf("report = %+v", r)
	}
	if r.MaxWidth != 832 || r.MaxHeight != 1152 || r.Shrink != 0.56 {
		t.Errorf("max = %v x %v, shrink %v; want 832 x 1152, 0.56", r.MaxWidth, r.MaxHeight, r.Shrink)
	}
}

// A wide diagram shrinks to fit the page width, and every label with it.
func TestWideDiagramIsTooSmall(t *testing.T) {
	long := strings.Repeat("word ", 20)
	r, err := legibility.Analyze(diagram(4680, 600, long, 16, 22), target(t, "docx-portrait"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Readable || r.LimitedBy != "width" || r.SmallestPt != 1.6 || len(r.TooSmall) != 2 {
		t.Fatalf("report = %+v", r)
	}
	if r.TooSmall[0].FontPx != 16 {
		t.Errorf("smallest first: %+v", r.TooSmall)
	}
	if r.Shrink != 5.63 {
		t.Errorf("shrink = %v, want 5.63", r.Shrink)
	}
	advice := strings.Join(r.Advice, "\n")
	for _, want := range []string{"5.6 times too wide", "fit within 832 x 1152", "Wrap labels to about 28", "fewer columns", "Split the diagram"} {
		if !strings.Contains(advice, want) {
			t.Errorf("advice lacks %q:\n%s", want, advice)
		}
	}
}

// A tall diagram is limited by the page height, and a near miss is told to
// raise the font.
func TestTallDiagramNearMiss(t *testing.T) {
	r, err := legibility.Analyze(diagram(400, 864, "Label", 8), target(t, "docx-portrait"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Readable || r.LimitedBy != "height" || r.SmallestPt != 6 {
		t.Fatalf("report = %+v", r)
	}
	advice := strings.Join(r.Advice, "\n")
	for _, want := range []string{"1.5 times too tall", "fewer items in each column", "Raise the label font from 8 to about 15 pixels"} {
		if !strings.Contains(advice, want) {
			t.Errorf("advice lacks %q:\n%s", want, advice)
		}
	}
}

// The advice names the built-in targets the diagram already fits.
func TestAdviceNamesFittingTargets(t *testing.T) {
	r, err := legibility.Analyze(diagram(1600, 900, "Label", 16), target(t, "docx-portrait"))
	if err != nil {
		t.Fatal(err)
	}
	advice := strings.Join(r.Advice, "\n")
	if !strings.Contains(advice, "readable on: tabloid-landscape") || strings.Contains(advice, "slide-16x9") {
		t.Errorf("advice:\n%s", advice)
	}
}

// A label without a font size, and a diagram without labels, count as
// nothing to read.
func TestNoLabels(t *testing.T) {
	d := diagram(5000, 5000, "Label", 0)
	r, err := legibility.Analyze(d, target(t, "screen"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Readable || r.Labels != 0 {
		t.Errorf("report = %+v", r)
	}
}

func TestBadInput(t *testing.T) {
	if _, err := legibility.Analyze(&d2vision.Diagram{}, target(t, "screen")); err == nil {
		t.Error("a diagram with no size was accepted")
	}
	if _, err := legibility.Analyze(diagram(10, 10, "a", 16), legibility.Target{Name: "x"}); err == nil {
		t.Error("a target with no size was accepted")
	}
	if _, ok := legibility.TargetByName("nope"); ok {
		t.Error("unknown target found")
	}
}

// A rendered SVG is read with its labels' sizes and lines.
func TestAnalyzeRenderedSVG(t *testing.T) {
	d, err := d2vision.ParseFile("../testdata/classed.svg")
	if err != nil {
		t.Fatal(err)
	}
	reports, err := legibility.AnalyzeAll(d, legibility.Targets)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != len(legibility.Targets) {
		t.Fatalf("%d reports", len(reports))
	}
	for _, r := range reports {
		if r.Labels != 3 || r.LongestLine != 15 {
			t.Errorf("%s: %d labels, longest line %d; want 3 and 15", r.Target.Name, r.Labels, r.LongestLine)
		}
	}
	if !reports[0].Readable {
		t.Errorf("a small diagram should be readable on a Word page: %+v", reports[0])
	}
}

// A grid draws an arrow straight over the box between its ends; ELK routes it
// around.
func TestCrossings(t *testing.T) {
	grid, err := d2vision.ParseFile("../testdata/crossing-grid.svg")
	if err != nil {
		t.Fatal(err)
	}
	got := legibility.Crossings(grid)
	if len(got) != 1 || got[0].From != "a" || got[0].To != "c" || got[0].Node != "b" {
		t.Fatalf("grid crossings = %+v, want a -> c over b", got)
	}
	r, err := legibility.Analyze(grid, target(t, "screen"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Readable || r.Clear() || len(r.Advice) != 1 || !strings.Contains(r.Advice[0], "1 arrow is drawn over a box") {
		t.Errorf("report = %+v", r)
	}

	elk, err := d2vision.ParseFile("../testdata/crossing-elk.svg")
	if err != nil {
		t.Fatal(err)
	}
	if got := legibility.Crossings(elk); len(got) != 0 {
		t.Errorf("elk crossings = %+v, want none", got)
	}
}

// A segment that only runs along a box's edge does not cross it.
func TestTouchingIsNotCrossing(t *testing.T) {
	d := &d2vision.Diagram{
		ViewBox: d2vision.Bounds{Width: 300, Height: 100},
		Nodes: []d2vision.Node{
			{ID: "a", Bounds: d2vision.Bounds{X: 0, Y: 0, Width: 50, Height: 50}},
			{ID: "b", Bounds: d2vision.Bounds{X: 100, Y: 0, Width: 50, Height: 50}},
			{ID: "c", Bounds: d2vision.Bounds{X: 200, Y: 0, Width: 50, Height: 50}},
		},
		Edges: []d2vision.Edge{
			{ID: "(a -> c)[0]", Source: "a", Target: "c", Path: []d2vision.Point{{X: 50, Y: 50}, {X: 200, Y: 50}}},
			{ID: "(a -> c)[1]", Source: "a", Target: "c", Path: []d2vision.Point{{X: 50, Y: 25}, {X: 200, Y: 25}}},
		},
	}
	got := legibility.Crossings(d)
	if len(got) != 1 || got[0].Edge != "(a -> c)[1]" || got[0].Node != "b" {
		t.Errorf("crossings = %+v", got)
	}
}
