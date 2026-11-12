package generate_test

import (
	"context"
	"strings"
	"testing"

	"github.com/grokify/d2vision/generate"
	"github.com/grokify/d2vision/render"
)

// A layout engine is written into the d2-config, and a class's width and
// height size every shape of the class.
func TestLayoutEngineAndClassSize(t *testing.T) {
	spec := &generate.DiagramSpec{
		LayoutEngine: "elk",
		Direction:    "up",
		Classes: map[string]generate.StyleSpec{
			"box": {Width: generate.IntPtr(300), Height: generate.IntPtr(90)},
		},
		Nodes: []generate.NodeSpec{{ID: "a", Label: "Short", Class: "box"}, {ID: "b", Label: "A longer label", Class: "box"}},
		Edges: []generate.EdgeSpec{{From: "a", To: "b"}},
	}
	code := generate.NewGenerator().Generate(spec)
	for _, want := range []string{"vars: {\n  d2-config: {\n    layout-engine: elk\n  }\n}\n", "    width: 300\n", "    height: 90\n"} {
		if !strings.Contains(code, want) {
			t.Errorf("missing %q:\n%s", want, code)
		}
	}
	d, err := render.Compile(context.Background(), code)
	if err != nil {
		t.Fatalf("does not compile: %v\n%s", err, code)
	}
	for _, s := range d.Shapes {
		if s.Width != 300 || s.Height != 90 {
			t.Errorf("%s is %d x %d, want 300 x 90", s.ID, s.Width, s.Height)
		}
	}
}

// Without a layout engine nothing is added.
func TestNoLayoutEngine(t *testing.T) {
	code := generate.NewGenerator().Generate(&generate.DiagramSpec{Nodes: []generate.NodeSpec{{ID: "a"}}})
	if strings.Contains(code, "d2-config") {
		t.Errorf("got:\n%s", code)
	}
}
