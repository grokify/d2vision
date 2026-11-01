package generate_test

import (
	"context"
	"sort"
	"testing"

	"github.com/grokify/d2vision/generate"
	"github.com/grokify/d2vision/render"
	"oss.terrastruct.com/d2/d2ast"
)

// boardKeywords stay special to D2 even when quoted: they define the boards of a
// multi-board diagram, and an edge cannot be drawn to or from a node inside a
// container with one of these names. The guide says so.
var boardKeywords = map[string]bool{"layers": true, "scenarios": true, "steps": true}

// Every word D2 reserves is quoted when it is used as an ID, in a node, a
// container, and an edge, so the diagram compiles and has the nodes asked for.
func TestReservedWordsAreQuotedAsIDs(t *testing.T) {
	var words []string
	for w := range d2ast.ReservedKeywords {
		words = append(words, w)
	}
	sort.Strings(words)

	for _, w := range words {
		if boardKeywords[w] {
			continue
		}
		spec := &generate.DiagramSpec{
			Containers: []generate.ContainerSpec{{ID: w, Label: "Box", Nodes: []generate.NodeSpec{{ID: "inner"}}}},
			Nodes:      []generate.NodeSpec{{ID: "other"}},
			Edges:      []generate.EdgeSpec{{From: w + ".inner", To: "other"}},
		}
		code := generate.NewGenerator().Generate(spec)
		d, err := render.Compile(context.Background(), code)
		if err != nil {
			t.Errorf("%q as an ID does not compile: %v\n%s", w, err, code)
			continue
		}
		if len(d.Shapes) != 3 || len(d.Connections) != 1 {
			t.Errorf("%q as an ID: %d shapes and %d connections, want 3 and 1\n%s", w, len(d.Shapes), len(d.Connections), code)
		}
	}
}

// A word that only looks like a reserved one is left alone, and the check
// ignores case, as D2 does.
func TestOrdinaryIDsAreNotQuoted(t *testing.T) {
	spec := &generate.DiagramSpec{Nodes: []generate.NodeSpec{{ID: "leftover"}, {ID: "Label"}}}
	code := generate.NewGenerator().Generate(spec)
	if want := "leftover\n\"Label\"\n"; code != want {
		t.Errorf("Generate() = %q, want %q", code, want)
	}
}

// The three board words are the only reserved words that quoting does not rescue.
func TestOnlyTheBoardWordsCannotBeIDs(t *testing.T) {
	for w := range boardKeywords {
		if _, ok := d2ast.ReservedKeywords[w]; !ok {
			t.Errorf("%q is no longer a D2 reserved word; remove it from boardKeywords", w)
		}
	}
}

// What does work with the board words: a plain node, with an edge, and a
// container that has no edges through it.
func TestBoardWordsWorkAsPlainNodes(t *testing.T) {
	for w := range boardKeywords {
		spec := &generate.DiagramSpec{
			Nodes: []generate.NodeSpec{{ID: w}, {ID: "other"}},
			Edges: []generate.EdgeSpec{{From: w, To: "other"}},
		}
		code := generate.NewGenerator().Generate(spec)
		d, err := render.Compile(context.Background(), code)
		if err != nil {
			t.Errorf("a node named %q does not compile: %v\n%s", w, err, code)
			continue
		}
		if len(d.Shapes) != 2 || len(d.Connections) != 1 {
			t.Errorf("a node named %q: %d shapes and %d connections, want 2 and 1", w, len(d.Shapes), len(d.Connections))
		}
	}
}
