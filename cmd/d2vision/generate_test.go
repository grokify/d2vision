package main

import (
	"strings"
	"testing"

	"github.com/grokify/d2vision/format"
)

const stagesSpec = `{
  "direction": "right",
  "gridColumns": 2,
  "classes": {"input": {"fill": "#fff1dc"}},
  "containers": [
    {"id": "stage-in", "label": "Inputs: what goes in", "nodes": [{"id": "a", "label": "Staff", "class": "input"}]},
    {"id": "stage-out", "label": "Outcomes", "nodes": [{"id": "b", "label": "Skills"}]}
  ],
  "edges": [{"from": "stage-in.a", "to": "stage-out.b"}]
}`

func TestGenerateWritesD2ByDefault(t *testing.T) {
	out, err := generateFromSpec([]byte(stagesSpec), format.JSON, "d2")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"grid-columns: 2", "classes: {", `"stage-in".a -> "stage-out".b`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateWritesAMermaidFlowchart(t *testing.T) {
	out, err := generateFromSpec([]byte(stagesSpec), format.JSON, "mermaid")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"flowchart LR", `subgraph n_stage_in["Inputs: what goes in"]`, "n_stage_in_a --> n_stage_out_b", "classDef input fill:#fff1dc"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateRefusesWhatItCannotWrite(t *testing.T) {
	if _, err := generateFromSpec([]byte(stagesSpec), format.JSON, "svg"); err == nil {
		t.Error("an unknown output format was accepted")
	}
	seq := `{"sequences": [{"id": "s", "actors": [{"id": "a"}], "steps": []}]}`
	if _, err := generateFromSpec([]byte(seq), format.JSON, "mermaid"); err == nil {
		t.Error("a sequence diagram was written as a flowchart")
	}
	if _, err := generateFromSpec([]byte("{not json"), format.JSON, "d2"); err == nil {
		t.Error("an unreadable spec was accepted")
	}
}
