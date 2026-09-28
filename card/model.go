// Package card renders "datasheet" style infographic cards: a fixed-size poster
// of stacked panels, each pairing a short text block with a simple node/edge
// diagram whose edges can carry animated "flow" dots.
//
// A Card is plain data (JSON) so many cards can be produced by the same
// pipeline: Card -> SVG (optionally animated with SMIL) -> PNG / GIF.
//
// These cards are intentionally simpler than documentation diagrams: a handful
// of nodes per panel, explicit coordinates, one message per panel. For
// auto-laid-out diagrams use D2 and animate the result with the flow layer.
package card

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Defaults applied by Card.ApplyDefaults.
const (
	DefaultWidth      = 800
	DefaultHeight     = 1000
	DefaultStepGap    = 0.2 // seconds between consecutive steps
	DefaultLoopRest   = 1.0 // seconds to hold the finished state before the loop restarts
	DefaultFlowSpeed  = 220 // px per second
	DefaultDotSize    = 5.0
	DefaultTextWidth  = 350
	DefaultMargin     = 24
	DefaultTitleSize  = 30
	DefaultBodySize   = 17
	DefaultFontFamily = `Inter, -apple-system, "Helvetica Neue", Helvetica, Arial, sans-serif`
)

// Card is a complete datasheet image.
type Card struct {
	Title     string       `json:"title,omitempty"`     // document title (SVG <title>), not drawn
	Width     int          `json:"width,omitempty"`     // default 800
	Height    int          `json:"height,omitempty"`    // default 1000
	Loop      float64      `json:"loop,omitempty"`      // loop length in seconds; 0 derives it from the flows
	StepGap   float64      `json:"stepGap,omitempty"`   // seconds between steps, default 0.2
	LoopRest  float64      `json:"loopRest,omitempty"`  // seconds held after the last step when Loop is derived, default 1
	Theme     Theme        `json:"theme,omitzero"`      // colors and fonts
	Panels    []Panel      `json:"panels"`              // stacked top to bottom
	Legend    []LegendItem `json:"legend,omitempty"`    // dot color key, drawn in a footer row
	Artifacts []Artifact   `json:"artifacts,omitempty"` // named dot colors used by sources; the legend is derived from them when Legend is empty
	Footer    string       `json:"footer,omitempty"`    // small attribution text

	loop     float64  // effective loop length, set by resolveTiming
	warnings []string // unverified steps etc., see Warnings
}

// Theme holds colors and typography. Zero values fall back to a dark theme.
type Theme struct {
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
	Muted      string `json:"muted,omitempty"`  // secondary text
	Stroke     string `json:"stroke,omitempty"` // node outlines and edges
	Fill       string `json:"fill,omitempty"`   // node fill
	Rule       string `json:"rule,omitempty"`   // panel separators
	FontFamily string `json:"fontFamily,omitempty"`
}

// PanelKind selects how a panel is laid out.
type PanelKind string

const (
	// PanelDiagram is a text block beside a node/edge diagram (default).
	PanelDiagram PanelKind = "diagram"
	// PanelTable is a full-width comparison table.
	PanelTable PanelKind = "table"
)

// Panel is one horizontal band of the card.
type Panel struct {
	ID        string    `json:"id,omitempty"`
	Kind      PanelKind `json:"kind,omitempty"`
	Height    int       `json:"height"`              // px
	Title     string    `json:"title,omitempty"`     // bold heading
	TitleSize float64   `json:"titleSize,omitempty"` // default 30
	Body      string    `json:"body,omitempty"`      // paragraph; "\n" starts a new line
	BodySize  float64   `json:"bodySize,omitempty"`  // default 17
	TextSide  string    `json:"textSide,omitempty"`  // "left" (default) or "right"
	TextWidth int       `json:"textWidth,omitempty"` // default 350
	Center    bool      `json:"center,omitempty"`    // center-align the text block (hero panels)
	RuleBelow bool      `json:"ruleBelow,omitempty"` // draw a thick separator under the panel
	Diagram   *Diagram  `json:"diagram,omitempty"`
	Source    *Source   `json:"source,omitempty"` // derive the diagram from a protocol (PIDL) file instead
	Table     *Table    `json:"table,omitempty"`
}

// Diagram is a set of nodes and edges. Coordinates are in pixels relative to
// the top-left of the panel's diagram area (the panel minus its text column).
type Diagram struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// NodeShape is the visual form of a node.
type NodeShape string

const (
	ShapeCircle NodeShape = "circle" // icon in a circle, label underneath (default)
	ShapeBox    NodeShape = "box"    // rounded rectangle with label inside
	ShapePill   NodeShape = "pill"   // fully rounded box, for small step labels
)

// Node is a diagram vertex. X, Y are the node center.
type Node struct {
	ID    string    `json:"id"`
	Label string    `json:"label,omitempty"` // "\n" starts a new line
	Icon  string    `json:"icon,omitempty"`  // emoji or short text drawn inside the node
	Shape NodeShape `json:"shape,omitempty"`
	X     float64   `json:"x"`
	Y     float64   `json:"y"`
	W     float64   `json:"w,omitempty"`     // box/pill width; circle diameter (default 44)
	H     float64   `json:"h,omitempty"`     // box/pill height
	Color string    `json:"color,omitempty"` // outline override
	Size  float64   `json:"size,omitempty"`  // label font size (default 13)
}

// EdgeStyle is the line style of an edge.
type EdgeStyle string

const (
	StyleSolid  EdgeStyle = "solid"
	StyleDashed EdgeStyle = "dashed"
)

// Edge connects two nodes with a straight or curved line.
type Edge struct {
	ID    string    `json:"id,omitempty"` // stable SVG id: "edge-<id>"; defaults to "<from>-<to>"
	From  string    `json:"from"`
	To    string    `json:"to"`
	Label string    `json:"label,omitempty"` // pill label at the edge midpoint
	Style EdgeStyle `json:"style,omitempty"`
	Bend  float64   `json:"bend,omitempty"`  // px offset of the curve's control point; sign picks the side
	Arrow bool      `json:"arrow,omitempty"` // draw an arrowhead at To
	Flow  *Flow     `json:"flow,omitempty"`  // animated dots (shorthand for a single entry in Flows)
	Flows []Flow    `json:"flows,omitempty"` // several flows, e.g. a different artifact in each direction
}

// FlowMode is the direction semantics of a flow.
type FlowMode string

const (
	// FlowForward moves dots From -> To.
	FlowForward FlowMode = "forward"
	// FlowReverse moves dots To -> From.
	FlowReverse FlowMode = "reverse"
	// FlowAlternate sends a dot From -> To, then back To -> From: request/response.
	FlowAlternate FlowMode = "alternate"
	// FlowBoth runs independent dots in both directions at once: full duplex.
	FlowBoth FlowMode = "both"
)

// Flow describes dots travelling along an edge. All flows share the card's
// Loop so the animation repeats seamlessly.
type Flow struct {
	Mode  FlowMode `json:"mode,omitempty"`  // default forward
	Color string   `json:"color,omitempty"` // default theme foreground
	Size  float64  `json:"size,omitempty"`  // dot radius, default 5
	Speed float64  `json:"speed,omitempty"` // px per second, default 220
	Count int      `json:"count,omitempty"` // dots per traversal, default 1
	Gap   float64  `json:"gap,omitempty"`   // seconds between dots when Count > 1, default 0.25
	Delay float64  `json:"delay,omitempty"` // seconds before the first dot leaves: into the loop, or into the step when Step is set
	Step  int      `json:"step,omitempty"`  // 1-based order within the panel; steps run in sequence, equal steps run together. 0 = unsequenced
	Pause float64  `json:"pause,omitempty"` // seconds to rest at the far end (alternate), default 0.15

	at float64 // resolved start within the loop, set by resolveTiming
}

// Table is a comparison grid. Columns[0] is the row-label column.
type Table struct {
	Columns      []string   `json:"columns"`
	ColumnColors []string   `json:"columnColors,omitempty"` // header accent per column
	Rows         [][]string `json:"rows"`
	FontSize     float64    `json:"fontSize,omitempty"` // default 15
}

// LegendItem maps a flow dot color to its meaning.
type LegendItem struct {
	Color string `json:"color"`
	Label string `json:"label"`
}

// Parse decodes a Card from JSON, rejecting unknown fields, applies defaults
// and validates it.
func Parse(r io.Reader) (*Card, error) { return parse(r, ".") }

func parse(r io.Reader, baseDir string) (*Card, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var c Card
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("card: decode: %w", err)
	}
	if err := c.compileSources(baseDir); err != nil {
		return nil, fmt.Errorf("card: %w", err)
	}
	c.ApplyDefaults()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// ParseFile is Parse on a file path.
func ParseFile(path string) (*Card, error) {
	f, err := os.Open(path) // #nosec G304 -- caller-supplied path
	if err != nil {
		return nil, err
	}
	c, err := parse(f, filepath.Dir(path))
	if err != nil {
		_ = f.Close() // parse error takes precedence
		return nil, err
	}
	return c, f.Close()
}

// ApplyDefaults fills zero values. It is idempotent.
func (c *Card) ApplyDefaults() {
	if c.Width == 0 {
		c.Width = DefaultWidth
	}
	if c.Height == 0 {
		c.Height = DefaultHeight
	}
	if c.StepGap == 0 {
		c.StepGap = DefaultStepGap
	}
	if c.LoopRest == 0 {
		c.LoopRest = DefaultLoopRest
	}
	t := &c.Theme
	setDefault(&t.Background, "#000000")
	setDefault(&t.Foreground, "#ffffff")
	setDefault(&t.Muted, "#b8b8b8")
	setDefault(&t.Stroke, "#d0d0d0")
	setDefault(&t.Fill, "#0c0c0c")
	setDefault(&t.Rule, "#ffffff")
	setDefault(&t.FontFamily, DefaultFontFamily)

	for i := range c.Panels {
		p := &c.Panels[i]
		if p.ID == "" {
			p.ID = fmt.Sprintf("p%d", i+1)
		}
		if p.Kind == "" {
			p.Kind = PanelDiagram
		}
		if p.TitleSize == 0 {
			p.TitleSize = DefaultTitleSize
		}
		if p.BodySize == 0 {
			p.BodySize = DefaultBodySize
		}
		if p.TextSide == "" {
			p.TextSide = "left"
		}
		if p.TextWidth == 0 {
			p.TextWidth = DefaultTextWidth
		}
		if p.Diagram == nil {
			continue
		}
		for j := range p.Diagram.Nodes {
			n := &p.Diagram.Nodes[j]
			if n.Shape == "" {
				n.Shape = ShapeCircle
			}
			if n.Size == 0 {
				n.Size = 13
			}
			if n.W == 0 {
				switch n.Shape {
				case ShapeCircle:
					n.W = 44
				default:
					n.W = 96
				}
			}
			if n.H == 0 {
				switch n.Shape {
				case ShapeCircle:
					n.H = n.W
				case ShapePill:
					n.H = 26
				default:
					n.H = 34
				}
			}
		}
		for j := range p.Diagram.Edges {
			e := &p.Diagram.Edges[j]
			if e.ID == "" {
				e.ID = e.From + "-" + e.To
			}
			if e.Style == "" {
				e.Style = StyleSolid
			}
			if e.Flow != nil {
				e.Flows = append([]Flow{*e.Flow}, e.Flows...)
				e.Flow = nil
			}
			for k := range e.Flows {
				e.Flows[k].applyDefaults(t.Foreground)
			}
		}
	}
}

func (f *Flow) applyDefaults(fg string) {
	if f.Mode == "" {
		f.Mode = FlowForward
	}
	setDefault(&f.Color, fg)
	if f.Size == 0 {
		f.Size = DefaultDotSize
	}
	if f.Speed == 0 {
		f.Speed = DefaultFlowSpeed
	}
	if f.Count == 0 {
		f.Count = 1
	}
	if f.Gap == 0 {
		f.Gap = 0.25
	}
	if f.Pause == 0 {
		f.Pause = 0.15
	}
}

func setDefault(s *string, v string) {
	if *s == "" {
		*s = v
	}
}

// Validate checks structural rules and returns every problem found.
func (c *Card) Validate() error {
	var errs []error
	add := func(format string, a ...any) { errs = append(errs, fmt.Errorf(format, a...)) }

	if len(c.Panels) == 0 {
		add("card has no panels")
	}
	if c.Loop < 0 {
		add("loop must be >= 0")
	}
	total := 0
	for i, p := range c.Panels {
		where := fmt.Sprintf("panel %d (%s)", i, p.ID)
		if p.Height <= 0 {
			add("%s: height must be > 0", where)
		}
		total += p.Height
		switch p.Kind {
		case PanelDiagram:
			if p.Diagram != nil {
				errs = append(errs, validateDiagram(where, p.Diagram)...)
			}
		case PanelTable:
			if p.Table == nil {
				add("%s: table panel has no table", where)
			} else if len(p.Table.Columns) == 0 {
				add("%s: table has no columns", where)
			} else {
				for r, row := range p.Table.Rows {
					if len(row) != len(p.Table.Columns) {
						add("%s: table row %d has %d cells, want %d", where, r, len(row), len(p.Table.Columns))
					}
				}
			}
		default:
			add("%s: unknown kind %q", where, p.Kind)
		}
		if p.TextSide != "left" && p.TextSide != "right" {
			add("%s: textSide must be left or right, got %q", where, p.TextSide)
		}
	}
	if avail := c.Height - c.footerHeight(); total > avail {
		add("panel heights sum to %d, exceeding the %d px available (card height %d minus footer)", total, avail, c.Height)
	}
	return errors.Join(errs...)
}

func validateDiagram(where string, d *Diagram) []error {
	var errs []error
	add := func(format string, a ...any) { errs = append(errs, fmt.Errorf(where+": "+format, a...)) }

	nodes := map[string]bool{}
	for _, n := range d.Nodes {
		if n.ID == "" {
			add("node with empty id")
			continue
		}
		if nodes[n.ID] {
			add("duplicate node id %q", n.ID)
		}
		nodes[n.ID] = true
		switch n.Shape {
		case ShapeCircle, ShapeBox, ShapePill:
		default:
			add("node %q: unknown shape %q", n.ID, n.Shape)
		}
	}
	edges := map[string]bool{}
	for _, e := range d.Edges {
		if !nodes[e.From] {
			add("edge %q: unknown from node %q", e.ID, e.From)
		}
		if !nodes[e.To] {
			add("edge %q: unknown to node %q", e.ID, e.To)
		}
		if e.From == e.To {
			add("edge %q: self-loop not supported", e.ID)
		}
		if edges[e.ID] {
			add("duplicate edge id %q", e.ID)
		}
		edges[e.ID] = true
		switch e.Style {
		case StyleSolid, StyleDashed:
		default:
			add("edge %q: unknown style %q", e.ID, e.Style)
		}
		for _, f := range e.Flows {
			if f.Step < 0 {
				add("edge %q: step must be >= 0", e.ID)
			}
			switch f.Mode {
			case FlowForward, FlowReverse, FlowAlternate, FlowBoth:
			default:
				add("edge %q: unknown flow mode %q", e.ID, f.Mode)
			}
		}
	}
	return errs
}

// footerHeight is the space reserved at the bottom for the legend and footer.
func (c *Card) footerHeight() int {
	if len(c.Legend) == 0 && c.Footer == "" {
		return 0
	}
	return 44
}
