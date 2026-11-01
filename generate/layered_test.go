package generate_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/grokify/d2vision/convert/mermaid"
	"github.com/grokify/d2vision/generate"
	"github.com/grokify/d2vision/render"
)

// layered is a small left-to-right diagram of the kind a logic model draws:
// one container per stage, nodes in each, edges between stages, and shared
// classes. Its labels hold the characters D2 reads as syntax.
func layered() *generate.DiagramSpec {
	return &generate.DiagramSpec{
		Direction:   "right",
		GridColumns: 2,
		Classes: map[string]generate.StyleSpec{
			"input":  {Fill: "#fff1dc", Stroke: "#e08a1e", BorderRadius: generate.IntPtr(12)},
			"result": {Fill: "#e1f3fd", Stroke: "#2d9cdb", Bold: generate.BoolPtr(true), FontColor: "#111111"},
		},
		Containers: []generate.ContainerSpec{
			{ID: "stage-in", Label: "INPUTS: what goes in", Nodes: []generate.NodeSpec{
				{ID: "ln-staff", Label: "Staff time {2 FTE} #1", Class: "input"},
				{ID: "ln-fund", Label: "It's funding; ${budget}", Class: "input"},
			}},
			{ID: "stage-out", Label: "OUTCOMES", Nodes: []generate.NodeSpec{
				{ID: "ln-skills", Label: `Skills "gained"`, Class: "result"},
			}},
		},
		Edges: []generate.EdgeSpec{
			{From: "stage-in.ln-staff", To: "stage-out.ln-skills", Label: "builds: over time"},
			{From: "stage-in.ln-fund", To: "stage-out.ln-skills"},
		},
	}
}

func TestLayeredSpecCompilesAsD2(t *testing.T) {
	code := generate.NewGenerator().Generate(layered())
	if err := render.Validate(context.Background(), code); err != nil {
		t.Fatalf("generated D2 does not compile: %v\n%s", err, code)
	}
	d, err := render.Compile(context.Background(), code)
	if err != nil {
		t.Fatal(err)
	}
	// An edge between nodes in two containers joins those nodes. It does not
	// make new ones.
	ids := map[string]bool{}
	for _, s := range d.Shapes {
		ids[s.ID] = true
	}
	for _, want := range []string{"stage-in.ln-staff", "stage-in.ln-fund", "stage-out.ln-skills"} {
		if !ids[want] {
			t.Errorf("no shape %q; shapes are %v", want, ids)
		}
	}
	if len(d.Shapes) != 5 {
		t.Errorf("%d shapes, want 5 (2 containers and 3 nodes): %v", len(d.Shapes), ids)
	}
	if len(d.Connections) != 2 {
		t.Errorf("%d connections, want 2", len(d.Connections))
	}
}

// A label is shown as written, whatever characters it holds.
func TestLabelsReachTheDiagramUnchanged(t *testing.T) {
	code := generate.NewGenerator().Generate(layered())
	d, err := render.Compile(context.Background(), code)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, s := range d.Shapes {
		got[s.Label] = true
	}
	for _, c := range d.Connections {
		got[c.Label] = true
	}
	for _, want := range []string{"INPUTS: what goes in", "Staff time {2 FTE} #1", "It's funding; ${budget}", `Skills "gained"`, "builds: over time"} {
		if !got[want] {
			t.Errorf("label %q did not survive; labels are %v", want, got)
		}
	}
}

func TestClassesAreWrittenOnceAndInOrder(t *testing.T) {
	code := generate.NewGenerator().Generate(layered())
	if strings.Count(code, "classes: {") != 1 {
		t.Errorf("want one classes block:\n%s", code)
	}
	if strings.Index(code, "  input: {") > strings.Index(code, "  result: {") {
		t.Errorf("classes are not in name order:\n%s", code)
	}
	if code != generate.NewGenerator().Generate(layered()) {
		t.Error("the same spec gave different D2")
	}
}

func TestMermaidOfALayeredSpec(t *testing.T) {
	out, err := generate.NewGenerator().Mermaid(layered())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"flowchart LR",
		`subgraph n_stage_in["INPUTS: what goes in"]`,
		`n_stage_in_ln_staff["Staff time {2 FTE} #1"]`,
		`n_stage_in_ln_staff -->|"builds: over time"| n_stage_out_ln_skills`,
		"classDef input fill:#fff1dc,stroke:#e08a1e",
		"class n_stage_in_ln_fund input",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	// The output reads back as a flowchart with the same nodes and edges.
	res, err := mermaid.NewConverter().Convert(out)
	if err != nil {
		t.Fatalf("the Mermaid does not parse: %v\n%s", err, out)
	}
	labels := map[string]bool{}
	var edges int
	var walk func(cs []generate.ContainerSpec, ns []generate.NodeSpec, es []generate.EdgeSpec)
	walk = func(cs []generate.ContainerSpec, ns []generate.NodeSpec, es []generate.EdgeSpec) {
		edges += len(es)
		for _, n := range ns {
			// The reader keeps the quotation marks around a quoted label.
			labels[strings.Trim(n.Label, `"`)] = true
		}
		for _, c := range cs {
			walk(c.Containers, c.Nodes, c.Edges)
		}
	}
	walk(res.Spec.Containers, res.Spec.Nodes, res.Spec.Edges)
	for _, want := range []string{"Staff time {2 FTE} #1", "It's funding; ${budget}"} {
		if !labels[want] {
			var got []string
			for l := range labels {
				got = append(got, fmt.Sprintf("%q", l))
			}
			t.Errorf("node %q did not read back; read back %s", want, strings.Join(got, ", "))
		}
	}
	if edges != 2 {
		t.Errorf("read back %d edges, want 2:\n%s", edges, out)
	}
}

func TestMermaidRefusesWhatIsNotAFlowchart(t *testing.T) {
	spec := &generate.DiagramSpec{Sequences: []generate.SequenceSpec{{ID: "s"}}}
	if _, err := generate.NewGenerator().Mermaid(spec); err == nil {
		t.Error("a sequence diagram was written as a flowchart")
	}
	bad := &generate.DiagramSpec{Nodes: []generate.NodeSpec{{ID: "a"}}, Edges: []generate.EdgeSpec{{From: "a", To: "ghost"}}}
	if _, err := generate.NewGenerator().Mermaid(bad); err == nil {
		t.Error("an edge to a node that does not exist was accepted")
	}
}

// Edges written inside a container name nodes relative to it, as in D2.
func TestMermaidResolvesEdgesInsideAContainer(t *testing.T) {
	spec := &generate.DiagramSpec{Containers: []generate.ContainerSpec{{
		ID: "c", Label: "C",
		Nodes: []generate.NodeSpec{{ID: "a"}, {ID: "b"}},
		Edges: []generate.EdgeSpec{{From: "a", To: "b"}},
	}}}
	out, err := generate.NewGenerator().Mermaid(spec)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "n_c_a --> n_c_b") {
		t.Errorf("edge not resolved:\n%s", out)
	}
}
