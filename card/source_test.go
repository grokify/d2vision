package card

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPIDL = `{
  "protocol": {"id": "t", "name": "Test"},
  "entities": [
    {"id": "agent", "name": "AI Agent"},
    {"id": "server", "name": "Server"},
    {"id": "hidden", "name": "Hidden"}
  ],
  "flows": [
    {"from": "agent", "to": "server", "action": "ask", "mode": "request"},
    {"from": "server", "to": "server", "action": "think", "mode": "request"},
    {"from": "server", "to": "agent", "action": "answer", "mode": "response"},
    {"from": "agent", "to": "hidden", "action": "leak", "mode": "request"}
  ]
}`

// sourceCard builds a one-panel card whose source shows the given steps.
func sourceCard(t *testing.T, top, steps string) (*Card, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "t.pidl.json"), []byte(testPIDL), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := `{` + top + `
	  "artifacts": [{"id":"req","label":"Request","color":"#f00"},{"id":"resp","label":"Response","color":"#0f0"}],
	  "panels": [{"id":"p","height":300,"source":{
	    "pidl":"t.pidl.json",
	    "actors":[{"entity":"agent","x":50,"y":100},{"entity":"server","label":"Custom","x":250,"y":100},{"id":"human","x":150,"y":200}],
	    "steps":[` + steps + `]}}]}`
	path := filepath.Join(dir, "card.json")
	if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	return ParseFile(path)
}

func TestSourceCompilesFromPIDL(t *testing.T) {
	c, err := sourceCard(t, "", `
	  {"action":"ask","artifact":"req","label":"hi"},
	  {"action":"answer","artifact":"resp"}`)
	if err != nil {
		t.Fatal(err)
	}
	d := c.Panels[0].Diagram
	if d == nil || len(d.Nodes) != 3 || len(d.Edges) != 1 {
		t.Fatalf("want 3 nodes and 1 edge (both steps share the agent/server pair), got %+v", d)
	}
	if got := d.Nodes[0].Label; got != "AI Agent" {
		t.Errorf("label should default to the PIDL entity name, got %q", got)
	}
	if got := d.Nodes[1].Label; got != "Custom" {
		t.Errorf("explicit label should win, got %q", got)
	}
	e := d.Edges[0]
	if e.From != "agent" || e.To != "server" || e.Label != "hi" {
		t.Errorf("edge = %+v", e)
	}
	if len(e.Flows) != 2 {
		t.Fatalf("flows = %+v", e.Flows)
	}
	// Direction comes from the protocol: "answer" is server -> agent, i.e. reverse on this edge.
	if f := e.Flows[0]; f.Mode != FlowForward || f.Step != 1 || f.Color != "#f00" {
		t.Errorf("flow 0 = %+v", f)
	}
	if f := e.Flows[1]; f.Mode != FlowReverse || f.Step != 2 || f.Color != "#0f0" {
		t.Errorf("flow 1 = %+v", f)
	}
}

func TestSourceDerivesLegendFromArtifactsUsed(t *testing.T) {
	c, err := sourceCard(t, "", `{"action":"answer","artifact":"resp"},{"action":"leak","artifact":"req"}`)
	if err == nil {
		t.Fatalf("leak targets an entity with no actor, want an error; got card %+v", c)
	}
	c, err = sourceCard(t, "", `{"action":"answer","artifact":"resp"}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Legend) != 1 || c.Legend[0].Label != "Response" || c.Legend[0].Color != "#0f0" {
		t.Errorf("legend = %+v, want only the artifact actually used", c.Legend)
	}
	// An explicit legend is never overridden.
	c, err = sourceCard(t, `"legend":[{"color":"#123","label":"Mine"}],`, `{"action":"ask","artifact":"req"}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Legend) != 1 || c.Legend[0].Label != "Mine" {
		t.Errorf("legend = %+v", c.Legend)
	}
}

func TestSourceErrors(t *testing.T) {
	tests := []struct {
		name  string
		steps string
		want  string
	}{
		{"unknown action", `{"action":"nope"}`, `action "nope" is not in`},
		{"out of protocol order", `{"action":"answer"},{"action":"ask"}`, "must keep protocol order"},
		{"repeated action", `{"action":"ask"},{"action":"ask"}`, "must keep protocol order"},
		{"internal step", `{"action":"think"}`, "no edge to draw"},
		{"entity without actor", `{"action":"leak"}`, "has no actor"},
		{"unknown artifact", `{"action":"ask","artifact":"zzz"}`, `unknown artifact "zzz"`},
		{"action and from", `{"action":"ask","from":"agent"}`, "not both"},
		{"explicit needs both ends", `{"from":"agent"}`, "needs an action"},
		{"explicit unknown actor", `{"from":"agent","to":"ghost"}`, "unknown actor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sourceCard(t, "", tt.steps)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestSourceExplicitStepsAreWarnings(t *testing.T) {
	c, err := sourceCard(t, "", `{"action":"ask"},{"from":"server","to":"human","artifact":"resp"}`)
	if err != nil {
		t.Fatal(err)
	}
	w := c.Warnings()
	if len(w) != 1 || !strings.Contains(w[0], "not backed by") || !strings.Contains(w[0], "step 2") {
		t.Errorf("warnings = %q", w)
	}
	c, err = sourceCard(t, "", `{"action":"ask"}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Warnings()) != 0 {
		t.Errorf("verified steps must not warn: %q", c.Warnings())
	}
}

func TestSourceRejectsDiagramAndSource(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "t.json"), []byte(testPIDL), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := `{"panels":[{"height":200,"diagram":{"nodes":[]},"source":{"pidl":"t.json","actors":[],"steps":[]}}]}`
	if err := os.WriteFile(filepath.Join(dir, "c.json"), []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseFile(filepath.Join(dir, "c.json")); err == nil || !strings.Contains(err.Error(), "not both") {
		t.Errorf("err = %v", err)
	}
}

func TestSourceEdgeOverrideNeedsMatchingStep(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "t.json"), []byte(testPIDL), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := `{"panels":[{"height":200,"source":{"pidl":"t.json",
	  "actors":[{"entity":"agent","x":50,"y":50},{"entity":"server","x":150,"y":50}],
	  "steps":[{"action":"ask"}],
	  "edges":[{"between":["agent","hidden"],"bend":5}]}}]}`
	if err := os.WriteFile(filepath.Join(dir, "c.json"), []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseFile(filepath.Join(dir, "c.json")); err == nil || !strings.Contains(err.Error(), "matches no step") {
		t.Errorf("err = %v", err)
	}
}
