package generate

import (
	"fmt"
	"strings"
)

// Generator converts DiagramSpec to D2 code.
type Generator struct {
	indent string
}

// NewGenerator creates a new D2 code generator.
func NewGenerator() *Generator {
	return &Generator{indent: "  "}
}

// Generate converts a DiagramSpec to D2 code.
func (g *Generator) Generate(spec *DiagramSpec) string {
	var b strings.Builder

	// Root-level layout settings
	if spec.GridColumns > 0 {
		fmt.Fprintf(&b, "grid-columns: %d\n", spec.GridColumns)
	}
	if spec.GridRows > 0 {
		fmt.Fprintf(&b, "grid-rows: %d\n", spec.GridRows)
	}
	if spec.Direction != "" {
		fmt.Fprintf(&b, "direction: %s\n", spec.Direction)
	}

	// Add blank line after root settings if any were written
	if spec.GridColumns > 0 || spec.GridRows > 0 || spec.Direction != "" {
		b.WriteString("\n")
	}

	// Generate containers
	for _, container := range spec.Containers {
		g.generateContainer(&b, container, 0)
		b.WriteString("\n")
	}

	// Generate top-level nodes
	for _, node := range spec.Nodes {
		g.generateNode(&b, node, 0)
	}

	// Generate top-level edges
	for _, edge := range spec.Edges {
		g.generateEdge(&b, edge, 0)
	}

	// Generate sequence diagrams
	for _, seq := range spec.Sequences {
		g.generateSequence(&b, seq, 0)
		b.WriteString("\n")
	}

	// Generate SQL tables
	for _, table := range spec.Tables {
		g.generateTable(&b, table, 0)
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func (g *Generator) generateContainer(b *strings.Builder, c ContainerSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	// Container declaration
	if c.Label != "" && c.Label != c.ID {
		fmt.Fprintf(b, "%s%s: %s {\n", indent, g.escapeID(c.ID), c.Label)
	} else if c.Label == "" {
		// Empty label - invisible container for layout purposes
		fmt.Fprintf(b, "%s%s: \"\" {\n", indent, g.escapeID(c.ID))
	} else {
		fmt.Fprintf(b, "%s%s: %s {\n", indent, g.escapeID(c.ID), c.Label)
	}

	innerIndent := strings.Repeat(g.indent, depth+1)

	// Container layout settings
	if c.Direction != "" {
		fmt.Fprintf(b, "%sdirection: %s\n", innerIndent, c.Direction)
	}
	if c.GridColumns > 0 {
		fmt.Fprintf(b, "%sgrid-columns: %d\n", innerIndent, c.GridColumns)
	}
	if c.GridRows > 0 {
		fmt.Fprintf(b, "%sgrid-rows: %d\n", innerIndent, c.GridRows)
	}

	// Container style
	g.generateStyle(b, c.Style, depth+1)

	// Add blank line after settings if any were written
	hasSettings := c.Direction != "" || c.GridColumns > 0 || c.GridRows > 0 || !g.isEmptyStyle(c.Style)
	if hasSettings && (len(c.Nodes) > 0 || len(c.Containers) > 0) {
		b.WriteString("\n")
	}

	// Nested containers
	for _, nested := range c.Containers {
		g.generateContainer(b, nested, depth+1)
	}

	// Nodes inside container
	for _, node := range c.Nodes {
		g.generateNode(b, node, depth+1)
	}

	// Add blank line before edges if there are nodes
	if len(c.Nodes) > 0 && len(c.Edges) > 0 {
		b.WriteString("\n")
	}

	// Edges inside container
	for _, edge := range c.Edges {
		g.generateEdge(b, edge, depth+1)
	}

	fmt.Fprintf(b, "%s}\n", indent)
}

func (g *Generator) generateNode(b *strings.Builder, n NodeSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	// Check if node needs a block (has shape, icon, or style)
	needsBlock := n.Shape != "" || n.Icon != "" || !g.isEmptyStyle(n.Style)

	if !needsBlock {
		// Simple node: just ID and optional label
		if n.Label != "" && n.Label != n.ID {
			fmt.Fprintf(b, "%s%s: %s\n", indent, g.escapeID(n.ID), n.Label)
		} else {
			fmt.Fprintf(b, "%s%s\n", indent, g.escapeID(n.ID))
		}
		return
	}

	// Node with properties
	if n.Label != "" && n.Label != n.ID {
		fmt.Fprintf(b, "%s%s: %s {\n", indent, g.escapeID(n.ID), n.Label)
	} else {
		fmt.Fprintf(b, "%s%s: {\n", indent, g.escapeID(n.ID))
	}

	innerIndent := strings.Repeat(g.indent, depth+1)

	if n.Shape != "" {
		fmt.Fprintf(b, "%sshape: %s\n", innerIndent, n.Shape)
	}
	if n.Icon != "" {
		fmt.Fprintf(b, "%sicon: %s\n", innerIndent, n.Icon)
	}

	g.generateStyle(b, n.Style, depth+1)

	fmt.Fprintf(b, "%s}\n", indent)
}

func (g *Generator) generateEdge(b *strings.Builder, e EdgeSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	// Determine arrow direction
	arrow := "->"
	if e.SourceArrow != "" && e.SourceArrow != "none" {
		arrow = "<->"
	}

	// Edge declaration
	edgeStr := fmt.Sprintf("%s%s %s %s", indent, g.escapeID(e.From), arrow, g.escapeID(e.To))

	// Add label if present
	if e.Label != "" {
		edgeStr += fmt.Sprintf(": %s", e.Label)
	}

	// Check if edge needs a block for style
	if !g.isEmptyStyle(e.Style) {
		b.WriteString(edgeStr + " {\n")
		g.generateStyle(b, e.Style, depth+1)
		fmt.Fprintf(b, "%s}\n", indent)
	} else {
		b.WriteString(edgeStr + "\n")
	}
}

func (g *Generator) generateStyle(b *strings.Builder, s *StyleSpec, depth int) {
	if s == nil {
		return
	}
	indent := strings.Repeat(g.indent, depth)

	if s.Fill != "" {
		fmt.Fprintf(b, "%sstyle.fill: %q\n", indent, s.Fill)
	}
	if s.Stroke != "" {
		fmt.Fprintf(b, "%sstyle.stroke: %q\n", indent, s.Stroke)
	}
	if s.StrokeWidth != nil {
		fmt.Fprintf(b, "%sstyle.stroke-width: %d\n", indent, *s.StrokeWidth)
	}
	if s.BorderRadius != nil && *s.BorderRadius > 0 {
		fmt.Fprintf(b, "%sstyle.border-radius: %d\n", indent, *s.BorderRadius)
	}
	if s.FontSize != nil && *s.FontSize > 0 {
		fmt.Fprintf(b, "%sstyle.font-size: %d\n", indent, *s.FontSize)
	}
	if s.Opacity != nil && *s.Opacity > 0 && *s.Opacity < 1 {
		fmt.Fprintf(b, "%sstyle.opacity: %.2f\n", indent, *s.Opacity)
	}
}

func (g *Generator) isEmptyStyle(s *StyleSpec) bool {
	if s == nil {
		return true
	}
	return s.Fill == "" && s.Stroke == "" && s.StrokeWidth == nil &&
		s.BorderRadius == nil && s.FontSize == nil && s.Opacity == nil
}

// escapeID escapes a D2 ID if it contains special characters.
func (g *Generator) escapeID(id string) string {
	// Check if ID needs quoting (contains spaces, dots, or special chars)
	needsQuote := false
	for _, c := range id {
		if c == ' ' || c == '-' || c == '.' || c == ':' {
			needsQuote = true
			break
		}
	}
	if needsQuote {
		return fmt.Sprintf("%q", id)
	}
	return id
}

// generateSequence generates D2 code for a sequence diagram.
func (g *Generator) generateSequence(b *strings.Builder, seq SequenceSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	// Sequence diagram container
	if seq.Label != "" {
		fmt.Fprintf(b, "%s%s: %s {\n", indent, g.escapeID(seq.ID), seq.Label)
	} else {
		fmt.Fprintf(b, "%s%s {\n", indent, g.escapeID(seq.ID))
	}

	innerIndent := strings.Repeat(g.indent, depth+1)

	// Set shape to sequence_diagram
	fmt.Fprintf(b, "%sshape: sequence_diagram\n\n", innerIndent)

	// Define actors (optional, but helps control order)
	for _, actor := range seq.Actors {
		if actor.Label != "" && actor.Label != actor.ID {
			fmt.Fprintf(b, "%s%s: %s", innerIndent, g.escapeID(actor.ID), actor.Label)
		} else {
			fmt.Fprintf(b, "%s%s", innerIndent, g.escapeID(actor.ID))
		}
		if actor.Shape != "" {
			fmt.Fprintf(b, " { shape: %s }", actor.Shape)
		}
		b.WriteString("\n")
	}

	if len(seq.Actors) > 0 {
		b.WriteString("\n")
	}

	// Generate messages in order
	for _, msg := range seq.Steps {
		if msg.Label != "" {
			fmt.Fprintf(b, "%s%s -> %s: %s\n", innerIndent, g.escapeID(msg.From), g.escapeID(msg.To), msg.Label)
		} else {
			fmt.Fprintf(b, "%s%s -> %s\n", innerIndent, g.escapeID(msg.From), g.escapeID(msg.To))
		}
	}

	// Generate groups
	for _, group := range seq.Groups {
		b.WriteString("\n")
		if group.Label != "" {
			fmt.Fprintf(b, "%s%s: %s {\n", innerIndent, g.escapeID(group.ID), group.Label)
		} else {
			fmt.Fprintf(b, "%s%s {\n", innerIndent, g.escapeID(group.ID))
		}
		groupIndent := strings.Repeat(g.indent, depth+2)
		for _, msg := range group.Messages {
			if msg.Label != "" {
				fmt.Fprintf(b, "%s%s -> %s: %s\n", groupIndent, g.escapeID(msg.From), g.escapeID(msg.To), msg.Label)
			} else {
				fmt.Fprintf(b, "%s%s -> %s\n", groupIndent, g.escapeID(msg.From), g.escapeID(msg.To))
			}
		}
		fmt.Fprintf(b, "%s}\n", innerIndent)
	}

	fmt.Fprintf(b, "%s}\n", indent)
}

// generateTable generates D2 code for an SQL table.
func (g *Generator) generateTable(b *strings.Builder, table TableSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	// Table declaration
	if table.Label != "" && table.Label != table.ID {
		fmt.Fprintf(b, "%s%s: %s {\n", indent, g.escapeID(table.ID), table.Label)
	} else {
		fmt.Fprintf(b, "%s%s {\n", indent, g.escapeID(table.ID))
	}

	innerIndent := strings.Repeat(g.indent, depth+1)

	// Set shape to sql_table
	fmt.Fprintf(b, "%sshape: sql_table\n\n", innerIndent)

	// Generate columns
	for _, col := range table.Columns {
		fmt.Fprintf(b, "%s%s: %s", innerIndent, g.escapeID(col.Name), col.Type)
		if len(col.Constraints) > 0 {
			if len(col.Constraints) == 1 {
				fmt.Fprintf(b, " { constraint: %s }", col.Constraints[0])
			} else {
				fmt.Fprintf(b, " { constraint: [%s] }", strings.Join(col.Constraints, ", "))
			}
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(b, "%s}\n", indent)
}
