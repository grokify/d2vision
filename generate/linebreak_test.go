package generate_test

import (
	"context"
	"strings"
	"testing"

	"github.com/grokify/d2vision/generate"
	"github.com/grokify/d2vision/render"
)

// A label with line breaks keeps them, and characters D2 would otherwise read
// as the end of the string or a variable are shown as written.
func TestLabelLineBreaksAreKept(t *testing.T) {
	label := "Costs $5 and ${x},\n  a \"quoted\" word \\ and it's\n\nthe last  line"
	spec := &generate.DiagramSpec{Nodes: []generate.NodeSpec{{ID: "a", Label: label}}}
	code := generate.NewGenerator().Generate(spec)
	d, err := render.Compile(context.Background(), code)
	if err != nil {
		t.Fatalf("does not compile: %v\n%s", err, code)
	}
	want := "Costs $5 and ${x},\na \"quoted\" word \\ and it's\nthe last line"
	if got := d.Shapes[0].Label; got != want {
		t.Errorf("label = %q, want %q\n%s", got, want, code)
	}
}

// A label on one line is written as before.
func TestOneLineLabelIsUnchanged(t *testing.T) {
	spec := &generate.DiagramSpec{Nodes: []generate.NodeSpec{{ID: "a", Label: "  Plain   text "}}}
	code := generate.NewGenerator().Generate(spec)
	if !strings.Contains(code, "a: Plain text\n") {
		t.Errorf("got:\n%s", code)
	}
}

// Mermaid shows a line break as <br/>.
func TestMermaidLabelLineBreaks(t *testing.T) {
	spec := &generate.DiagramSpec{Nodes: []generate.NodeSpec{{ID: "a", Label: "one <b>\ntwo"}}}
	code, err := generate.NewGenerator().Mermaid(spec)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, `"one #lt;b#gt;<br/>two"`) {
		t.Errorf("got:\n%s", code)
	}
}
