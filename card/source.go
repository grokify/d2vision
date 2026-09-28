package card

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Source derives a panel's diagram from a protocol definition (PIDL) instead of
// hand-drawing it. The card author chooses which actors and messages to show,
// and how; the protocol file decides what each message is, who sends it to
// whom, and in what order. That keeps the picture consistent with the spec:
// a step that does not exist, or that contradicts the protocol's order, is a
// compile error rather than a wrong arrow.
type Source struct {
	PIDL   string         `json:"pidl"`            // path to a PIDL JSON file, relative to the card file
	Actors []Actor        `json:"actors"`          // nodes to draw
	Steps  []SourceStep   `json:"steps"`           // messages to show, in display order
	Edges  []EdgeOverride `json:"edges,omitempty"` // presentation tweaks for the edge between two actors
}

// Actor places a protocol entity on the diagram. Node.ID defaults to Entity.
// An actor with no Entity is not part of the protocol file (for example a human
// approving out of band) and can only be used by explicit steps.
type Actor struct {
	Entity string `json:"entity,omitempty"`
	Node
}

// SourceStep is one message to draw. Set Action to reference a PIDL flow, which
// makes it verified: From, To and order come from the protocol file. Or set
// From and To directly for something the protocol file does not model; those
// steps are drawn but reported by Card.Warnings as unverified.
type SourceStep struct {
	Action   string   `json:"action,omitempty"`   // PIDL flow action id
	From     string   `json:"from,omitempty"`     // actor id, explicit steps only
	To       string   `json:"to,omitempty"`       // actor id, explicit steps only
	Artifact string   `json:"artifact,omitempty"` // key into Card.Artifacts: dot color and legend entry
	Color    string   `json:"color,omitempty"`    // overrides the artifact color
	Label    string   `json:"label,omitempty"`    // pill label for the edge this step travels on
	Mode     FlowMode `json:"mode,omitempty"`     // e.g. alternate; default from the step direction
	Delay    float64  `json:"delay,omitempty"`
}

// EdgeOverride adjusts the edge joining two actors, whichever way it is drawn.
type EdgeOverride struct {
	Between [2]string `json:"between"`
	Bend    float64   `json:"bend,omitempty"`
	Style   EdgeStyle `json:"style,omitempty"`
	Arrow   bool      `json:"arrow,omitempty"`
	Label   string    `json:"label,omitempty"`
}

// Artifact names a kind of message (a token, a request) with the color its dots
// use. When Card.Legend is empty, the legend lists the artifacts used.
type Artifact struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"`
}

// Warnings lists things worth a human look after parsing, such as steps that
// are drawn but not backed by a protocol file.
func (c *Card) Warnings() []string { return c.warnings }

// pidlDoc is the subset of a PIDL file the compiler reads. Other fields are
// ignored so richer PIDL files keep working.
type pidlDoc struct {
	Protocol struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"protocol"`
	Entities []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"entities"`
	Flows []struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Action string `json:"action"`
		Mode   string `json:"mode"`
	} `json:"flows"`
}

func readPIDL(path string) (*pidlDoc, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path comes from the card definition
	if err != nil {
		return nil, err
	}
	var d pidlDoc
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &d, nil
}

// compileSources turns every panel Source into a Diagram. baseDir is where
// relative PIDL paths resolve from. It runs before defaults and validation, so
// the generated diagram gets the same treatment as a hand-written one.
func (c *Card) compileSources(baseDir string) error {
	art := map[string]Artifact{}
	for _, a := range c.Artifacts {
		if a.ID == "" {
			return errors.New("artifact with empty id")
		}
		if _, dup := art[a.ID]; dup {
			return fmt.Errorf("duplicate artifact id %q", a.ID)
		}
		art[a.ID] = a
	}

	var used []string
	seen := map[string]bool{}
	c.warnings = nil
	for i := range c.Panels {
		p := &c.Panels[i]
		if p.Source == nil {
			continue
		}
		if p.Diagram != nil {
			return fmt.Errorf("panel %d (%s): set either diagram or source, not both", i, p.ID)
		}
		d, uses, err := p.Source.compile(baseDir, art, p.ID, &c.warnings)
		if err != nil {
			return fmt.Errorf("panel %d (%s): source: %w", i, p.ID, err)
		}
		p.Diagram = d
		for _, id := range uses {
			if !seen[id] {
				seen[id] = true
				used = append(used, id)
			}
		}
	}

	if len(c.Legend) == 0 {
		for _, id := range used {
			c.Legend = append(c.Legend, LegendItem{Color: art[id].Color, Label: art[id].Label})
		}
	}
	return nil
}

func (s *Source) compile(baseDir string, art map[string]Artifact, panelID string, warnings *[]string) (*Diagram, []string, error) {
	if s.PIDL == "" {
		return nil, nil, errors.New("pidl path is required")
	}
	path := s.PIDL
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}
	doc, err := readPIDL(path)
	if err != nil {
		return nil, nil, err
	}

	entityName := map[string]string{}
	for _, e := range doc.Entities {
		entityName[e.ID] = e.Name
	}
	flowIndex := map[string]int{}
	for i, f := range doc.Flows {
		flowIndex[f.Action] = i
	}

	// Actors become nodes; entity ids map to node ids.
	var d Diagram
	nodeOf := map[string]string{} // PIDL entity -> node id
	nodes := map[string]bool{}
	for _, a := range s.Actors {
		n := a.Node
		if a.Entity != "" {
			name, ok := entityName[a.Entity]
			if !ok {
				return nil, nil, fmt.Errorf("actor entity %q is not in %s", a.Entity, s.PIDL)
			}
			if n.ID == "" {
				n.ID = a.Entity
			}
			if n.Label == "" {
				n.Label = name
			}
			nodeOf[a.Entity] = n.ID
		}
		if n.ID == "" {
			return nil, nil, errors.New("actor needs an entity or an id")
		}
		if nodes[n.ID] {
			return nil, nil, fmt.Errorf("duplicate actor %q", n.ID)
		}
		nodes[n.ID] = true
		d.Nodes = append(d.Nodes, n)
	}

	// Steps become flows grouped onto one edge per actor pair.
	edgeAt := map[[2]string]int{} // sorted pair -> index in d.Edges
	lastIdx := -1
	var uses []string
	for si, st := range s.Steps {
		from, to := st.From, st.To
		if st.Action != "" {
			if from != "" || to != "" {
				return nil, nil, fmt.Errorf("step %d: set action or from/to, not both", si+1)
			}
			idx, ok := flowIndex[st.Action]
			if !ok {
				return nil, nil, fmt.Errorf("step %d: action %q is not in %s", si+1, st.Action, s.PIDL)
			}
			if idx <= lastIdx {
				return nil, nil, fmt.Errorf("step %d: action %q is #%d in the protocol but follows #%d; steps must keep protocol order",
					si+1, st.Action, idx+1, lastIdx+1)
			}
			lastIdx = idx
			pf := doc.Flows[idx]
			if pf.From == pf.To {
				return nil, nil, fmt.Errorf("step %d: action %q happens inside %q and has no edge to draw; pick a message between two actors", si+1, st.Action, pf.From)
			}
			var okF, okT bool
			from, okF = nodeOf[pf.From]
			to, okT = nodeOf[pf.To]
			if !okF || !okT {
				return nil, nil, fmt.Errorf("step %d: action %q is %s -> %s but that entity has no actor; add it to actors", si+1, st.Action, pf.From, pf.To)
			}
		} else {
			if from == "" || to == "" {
				return nil, nil, fmt.Errorf("step %d: needs an action, or both from and to", si+1)
			}
			if !nodes[from] || !nodes[to] {
				return nil, nil, fmt.Errorf("step %d: explicit step uses unknown actor %q or %q", si+1, from, to)
			}
			*warnings = append(*warnings, fmt.Sprintf("panel %s step %d (%s -> %s) is not backed by %s; add it to the protocol file or accept it as editorial",
				panelID, si+1, from, to, s.PIDL))
		}

		color := st.Color
		if st.Artifact != "" {
			a, ok := art[st.Artifact]
			if !ok {
				return nil, nil, fmt.Errorf("step %d: unknown artifact %q", si+1, st.Artifact)
			}
			if color == "" {
				color = a.Color
			}
			uses = append(uses, st.Artifact)
		}

		key := [2]string{from, to}
		if key[0] > key[1] {
			key = [2]string{to, from}
		}
		ei, ok := edgeAt[key]
		if !ok {
			ei = len(d.Edges)
			edgeAt[key] = ei
			d.Edges = append(d.Edges, Edge{From: from, To: to})
		}
		e := &d.Edges[ei]
		if e.Label == "" {
			e.Label = st.Label
		}
		mode := st.Mode
		if mode == "" {
			mode = FlowForward
			if e.From != from {
				mode = FlowReverse
			}
		}
		e.Flows = append(e.Flows, Flow{Mode: mode, Color: color, Delay: st.Delay, Step: si + 1})
	}

	for _, o := range s.Edges {
		key := o.Between
		if key[0] > key[1] {
			key = [2]string{key[1], key[0]}
		}
		ei, ok := edgeAt[key]
		if !ok {
			return nil, nil, fmt.Errorf("edge override between %q and %q matches no step", o.Between[0], o.Between[1])
		}
		e := &d.Edges[ei]
		e.Bend, e.Arrow = o.Bend, o.Arrow
		if o.Style != "" {
			e.Style = o.Style
		}
		if o.Label != "" {
			e.Label = o.Label
		}
	}
	return &d, uses, nil
}
