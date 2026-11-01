package generate

import (
	"fmt"
	"sort"
	"strings"
)

// Mermaid converts a DiagramSpec to a Mermaid flowchart. Containers become
// subgraphs, and shared classes become classDefs. Mermaid has no grid layout,
// so GridColumns and GridRows are ignored, and a diagram that holds sequences
// or tables is refused, since those are not flowcharts.
//
// Mermaid node IDs are global, so a node inside containers is named by its
// whole path. Output is the same for the same spec.
func (g *Generator) Mermaid(spec *DiagramSpec) (string, error) {
	if len(spec.Sequences) > 0 || len(spec.Tables) > 0 {
		return "", fmt.Errorf("generate: a Mermaid flowchart cannot hold sequences or tables")
	}
	m := &mermaidWriter{ids: map[string]string{}, used: map[string]bool{}}
	var body strings.Builder
	m.w = &body
	for _, c := range spec.Containers {
		m.container(c, "", 1)
	}
	for _, n := range spec.Nodes {
		m.node(n, "", 1)
	}
	for _, c := range spec.Containers {
		m.edges(c, "")
	}
	for _, e := range spec.Edges {
		if err := m.edge(e, "", 1); err != nil {
			return "", err
		}
	}
	if m.err != nil {
		return "", m.err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "flowchart %s\n", mermaidDirection(spec.Direction))
	names := make([]string, 0, len(spec.Classes))
	for name := range spec.Classes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if def := mermaidStyle(spec.Classes[name]); def != "" {
			fmt.Fprintf(&b, "  classDef %s %s\n", mermaidSafe(name), def)
		}
	}
	b.WriteString(body.String())
	for _, line := range m.classLines {
		b.WriteString(line)
	}
	return b.String(), nil
}

type mermaidWriter struct {
	w          *strings.Builder
	ids        map[string]string // element path -> Mermaid ID
	used       map[string]bool
	classLines []string
	err        error
}

func (m *mermaidWriter) line(depth int, format string, args ...any) {
	m.w.WriteString(strings.Repeat("  ", depth))
	fmt.Fprintf(m.w, format, args...)
	m.w.WriteString("\n")
}

// id returns the Mermaid ID for an element path, made unique and safe.
func (m *mermaidWriter) id(path string) string {
	if id, ok := m.ids[path]; ok {
		return id
	}
	base := "n_" + mermaidSafe(path)
	id := base
	for i := 2; m.used[id]; i++ {
		id = fmt.Sprintf("%s_%d", base, i)
	}
	m.used[id] = true
	m.ids[path] = id
	return id
}

func join(prefix, id string) string {
	if prefix == "" {
		return id
	}
	return prefix + "." + id
}

func (m *mermaidWriter) container(c ContainerSpec, prefix string, depth int) {
	path := join(prefix, c.ID)
	label := c.Label
	if label == "" {
		label = " "
	}
	m.line(depth, "subgraph %s[\"%s\"]", m.id(path), mermaidText(label))
	if c.Direction != "" {
		m.line(depth+1, "direction %s", mermaidDirection(c.Direction))
	}
	for _, nested := range c.Containers {
		m.container(nested, path, depth+1)
	}
	for _, n := range c.Nodes {
		m.node(n, path, depth+1)
	}
	m.line(depth, "end")
	m.element(path, c.Class, c.Style)
}

func (m *mermaidWriter) node(n NodeSpec, prefix string, depth int) {
	path := join(prefix, n.ID)
	label := n.Label
	if label == "" {
		label = n.ID
	}
	open, closing := mermaidShape(n.Shape)
	m.line(depth, "%s%s\"%s\"%s", m.id(path), open, mermaidText(label), closing)
	m.element(path, n.Class, n.Style)
}

// element records the class and the inline style of a node or container.
func (m *mermaidWriter) element(path, class string, style *StyleSpec) {
	if class != "" {
		m.classLines = append(m.classLines, fmt.Sprintf("  class %s %s\n", m.id(path), mermaidSafe(class)))
	}
	if style != nil {
		if def := mermaidStyle(*style); def != "" {
			m.classLines = append(m.classLines, fmt.Sprintf("  style %s %s\n", m.id(path), def))
		}
	}
}

// edges writes the edges held inside a container and its nested containers.
func (m *mermaidWriter) edges(c ContainerSpec, prefix string) {
	path := join(prefix, c.ID)
	for _, nested := range c.Containers {
		m.edges(nested, path)
	}
	for _, e := range c.Edges {
		if err := m.edge(e, path, 1); err != nil && m.err == nil {
			m.err = err
		}
	}
}

func (m *mermaidWriter) edge(e EdgeSpec, prefix string, depth int) error {
	from, ok := m.resolve(e.From, prefix)
	if !ok {
		return fmt.Errorf("generate: edge from %q: no such node", e.From)
	}
	to, ok := m.resolve(e.To, prefix)
	if !ok {
		return fmt.Errorf("generate: edge to %q: no such node", e.To)
	}
	arrow := "-->"
	if e.SourceArrow != "" && e.SourceArrow != "none" {
		arrow = "<-->"
	}
	if e.Label != "" {
		m.line(depth, "%s %s|\"%s\"| %s", from, arrow, mermaidText(e.Label), to)
	} else {
		m.line(depth, "%s %s %s", from, arrow, to)
	}
	return nil
}

// resolve finds the node an edge names: relative to the container the edge is
// written in, or else from the top.
func (m *mermaidWriter) resolve(ref, prefix string) (string, bool) {
	if prefix != "" {
		if id, ok := m.ids[prefix+"."+ref]; ok {
			return id, true
		}
	}
	id, ok := m.ids[ref]
	return id, ok
}

func mermaidDirection(d string) string {
	switch d {
	case "down":
		return "TB"
	case "left":
		return "RL"
	case "up":
		return "BT"
	default:
		return "LR"
	}
}

func mermaidShape(shape string) (open, closing string) {
	switch shape {
	case "cylinder":
		return "[(", ")]"
	case "circle", "oval":
		return "((", "))"
	case "diamond":
		return "{", "}"
	case "hexagon":
		return "{{", "}}"
	default:
		return "[", "]"
	}
}

// mermaidStyle writes a style as Mermaid style properties, or "" for none.
func mermaidStyle(s StyleSpec) string {
	var parts []string
	if s.Fill != "" {
		parts = append(parts, "fill:"+s.Fill)
	}
	if s.Stroke != "" {
		parts = append(parts, "stroke:"+s.Stroke)
	}
	if s.StrokeWidth != nil {
		parts = append(parts, fmt.Sprintf("stroke-width:%dpx", *s.StrokeWidth))
	}
	if s.FontColor != "" {
		parts = append(parts, "color:"+s.FontColor)
	}
	return strings.Join(parts, ",")
}

// mermaidSafe keeps letters, digits, and underscores.
func mermaidSafe(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, s)
}

// mermaidText escapes text for use inside a quoted Mermaid label.
func mermaidText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, "&", "#amp;")
	s = strings.ReplaceAll(s, `"`, "#quot;")
	s = strings.ReplaceAll(s, "<", "#lt;")
	s = strings.ReplaceAll(s, ">", "#gt;")
	return s
}
