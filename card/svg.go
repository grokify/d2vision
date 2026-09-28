package card

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
)

// Options controls rendering.
type Options struct {
	// Animate adds SMIL flow dots to edges that define a Flow. Without it the
	// output is a fully static SVG (the right input for a PNG fallback).
	Animate bool
}

const (
	textGap    = 12.0 // between the title and body blocks
	diagramGap = 10.0 // between the text column and the diagram
)

// Render draws the card as an SVG document.
func (c *Card) Render(opt Options) ([]byte, error) {
	c.ApplyDefaults()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	c.resolveTiming()
	th := c.Theme

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 %d %d" width="%d" height="%d" font-family="%s">`,
		c.Width, c.Height, c.Width, c.Height, esc(th.FontFamily))
	if c.Title != "" {
		fmt.Fprintf(&b, "<title>%s</title>", esc(c.Title))
	}
	fmt.Fprintf(&b, `<defs><marker id="arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10z" fill="%s"/></marker></defs>`, esc(th.Stroke))
	fmt.Fprintf(&b, `<rect id="background" width="%d" height="%d" fill="%s"/>`, c.Width, c.Height, esc(th.Background))

	y := 0
	for _, p := range c.Panels {
		if err := c.renderPanel(&b, p, y, opt); err != nil {
			return nil, fmt.Errorf("panel %s: %w", p.ID, err)
		}
		y += p.Height
		if p.RuleBelow {
			fmt.Fprintf(&b, `<rect class="rule" x="0" y="%d" width="%d" height="6" fill="%s"/>`, y-3, c.Width, esc(th.Rule))
		}
	}
	c.renderFooter(&b)
	b.WriteString("</svg>\n")
	return []byte(b.String()), nil
}

func (c *Card) renderPanel(b *strings.Builder, p Panel, y int, opt Options) error {
	fmt.Fprintf(b, `<g id="panel-%s" class="panel">`, esc(p.ID))
	defer b.WriteString("</g>")

	if p.Kind == PanelTable {
		return c.renderTable(b, p, y)
	}

	// Split the panel into a text column and a diagram area.
	m := float64(DefaultMargin)
	textW := float64(p.TextWidth)
	hasText := p.Title != "" || p.Body != ""
	if !hasText {
		textW = 0
	}
	var textX, diagX float64
	diagW := float64(c.Width) - 2*m - textW
	if hasText {
		diagW -= diagramGap
	}
	if p.TextSide == "right" {
		diagX = m
		textX = float64(c.Width) - m - textW
	} else {
		textX = m
		diagX = m + textW
		if hasText {
			diagX += diagramGap
		}
	}

	if hasText {
		if err := c.renderText(b, p, textX, float64(y), textW); err != nil {
			return err
		}
	}
	if p.Diagram != nil {
		return c.renderDiagram(b, p, diagX, float64(y), diagW, opt)
	}
	return nil
}

// renderText draws the title and body, vertically centered in the panel.
func (c *Card) renderText(b *strings.Builder, p Panel, x, y, w float64) error {
	th := c.Theme
	const titleLH, bodyLH = 1.18, 1.42

	titleLines := wrap(p.Title, p.TitleSize, w, true)
	bodyLines := wrap(p.Body, p.BodySize, w, false)
	titleH := float64(len(titleLines)) * p.TitleSize * titleLH
	bodyH := float64(len(bodyLines)) * p.BodySize * bodyLH
	gap := 0.0
	if p.Title != "" && p.Body != "" {
		gap = textGap
	}
	total := titleH + gap + bodyH
	if total > float64(p.Height)-16 {
		return fmt.Errorf("text needs %.0f px but panel is %d px tall; shorten the text or raise the height", total, p.Height)
	}

	anchor, tx := "start", x
	if p.Center {
		anchor, tx = "middle", x+w/2
	}
	cy := y + (float64(p.Height)-total)/2
	fmt.Fprintf(b, `<g class="text">`)
	for i, l := range titleLines {
		base := cy + (float64(i)+0.85)*p.TitleSize*titleLH
		fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" font-weight="700" fill="%s" text-anchor="%s">%s</text>`,
			num(tx), num(base), num(p.TitleSize), esc(th.Foreground), anchor, esc(l))
	}
	by := cy + titleH + gap
	for i, l := range bodyLines {
		base := by + (float64(i)+0.8)*p.BodySize*bodyLH
		fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" fill="%s" text-anchor="%s">%s</text>`,
			num(tx), num(base), num(p.BodySize), esc(th.Foreground), anchor, esc(l))
	}
	b.WriteString(`</g>`)
	return nil
}

func (c *Card) renderDiagram(b *strings.Builder, p Panel, ox, oy, w float64, opt Options) error {
	th := c.Theme
	d := p.Diagram
	byID := make(map[string]Node, len(d.Nodes))
	for _, n := range d.Nodes {
		byID[n.ID] = n
		if err := checkFits(n, w, float64(p.Height)); err != nil {
			return err
		}
	}

	fmt.Fprintf(b, `<g class="diagram" transform="translate(%s %s)">`, num(ox), num(oy))

	// Edges first so nodes paint over their ends.
	var dots strings.Builder
	for _, e := range d.Edges {
		q := edgeQuad(byID[e.From], byID[e.To], e.Bend)
		pathID := fmt.Sprintf("path-%s-%s", p.ID, e.ID)
		dash := ""
		if e.Style == StyleDashed {
			dash = ` stroke-dasharray="5 4"`
		}
		marker := ""
		if e.Arrow {
			marker = ` marker-end="url(#arrow)"`
		}
		fmt.Fprintf(b, `<g id="edge-%s-%s" class="edge"><path id="%s" d="M%s %s Q%s %s %s %s" fill="none" stroke="%s" stroke-width="1.6"%s%s/></g>`,
			esc(p.ID), esc(e.ID), esc(pathID),
			num(q.P0.X), num(q.P0.Y), num(q.C.X), num(q.C.Y), num(q.P2.X), num(q.P2.Y),
			esc(th.Stroke), dash, marker)

		if opt.Animate {
			for k := range e.Flows {
				s, err := flowDots(esc(pathID), e, &e.Flows[k], q, c.loop)
				if err != nil {
					return err
				}
				dots.WriteString(s)
			}
		}
	}
	for _, n := range d.Nodes {
		c.renderNode(b, p.ID, n)
	}
	// Edge labels above nodes, then dots on top of everything.
	for _, e := range d.Edges {
		if e.Label == "" {
			continue
		}
		mid := edgeQuad(byID[e.From], byID[e.To], e.Bend).at(0.5)
		c.renderPill(b, mid, e.Label)
	}
	b.WriteString(dots.String())
	// Badges above the dots, so a dot passing underneath never hides a number.
	for _, e := range d.Edges {
		c.renderBadges(b, e, edgeQuad(byID[e.From], byID[e.To], e.Bend))
	}
	b.WriteString(`</g>`)
	return nil
}

// renderBadges numbers the sequenced flows on an edge. Forward badges sit
// toward the From end, reverse badges toward the To end, and alternating ones
// mid-edge, so opposite directions on one edge never collide.
func (c *Card) renderBadges(b *strings.Builder, e Edge, q quad) {
	const r = 9.0
	var fwd, rev int
	for _, f := range e.Flows {
		if f.Step == 0 {
			continue
		}
		var t float64
		switch f.Mode {
		case FlowReverse:
			t = 0.76 - 0.14*float64(rev)
			rev++
		case FlowAlternate, FlowBoth:
			t = 0.5
		default:
			t = 0.24 + 0.14*float64(fwd)
			fwd++
		}
		p := q.at(t)
		if t == 0.5 && e.Label != "" {
			// The label pill occupies the midpoint: sit just above it.
			tan := q.P2.sub(q.P0).unit()
			n := pt{-tan.Y, tan.X}
			if n.Y > 0 {
				n = n.scale(-1)
			}
			p = p.add(n.scale(r + 12))
		}
		fmt.Fprintf(b, `<g class="step-badge"><circle cx="%s" cy="%s" r="%s" fill="%s"/><text x="%s" y="%s" font-size="11" font-weight="700" fill="%s" text-anchor="middle">%d</text></g>`,
			num(p.X), num(p.Y), num(r), esc(f.Color), num(p.X), num(p.Y+3.9), esc(c.Theme.Background), f.Step)
	}
}

// checkFits reports nodes whose body or label would fall outside the diagram
// area, which usually means a typo in a coordinate.
func checkFits(n Node, w, h float64) error {
	if n.X-n.W/2 < 0 || n.X+n.W/2 > w || n.Y-n.H/2 < 0 || n.Y+n.H/2 > h {
		return fmt.Errorf("node %q at (%.0f,%.0f) falls outside the %.0fx%.0f diagram area", n.ID, n.X, n.Y, w, h)
	}
	return nil
}

func (c *Card) renderNode(b *strings.Builder, panelID string, n Node) {
	th := c.Theme
	stroke := th.Stroke
	if n.Color != "" {
		stroke = n.Color
	}
	fmt.Fprintf(b, `<g id="node-%s-%s" class="node">`, esc(panelID), esc(n.ID))
	lines := strings.Split(n.Label, "\n")
	lh := n.Size * 1.2

	switch n.Shape {
	case ShapeCircle:
		fmt.Fprintf(b, `<circle cx="%s" cy="%s" r="%s" fill="%s" stroke="%s" stroke-width="1.6"/>`,
			num(n.X), num(n.Y), num(n.W/2), esc(th.Fill), esc(stroke))
		if n.Icon != "" {
			fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" text-anchor="middle">%s</text>`,
				num(n.X), num(n.Y+n.W*0.17), num(n.W*0.5), esc(n.Icon))
		}
		if n.Label != "" {
			top := n.Y + n.W/2 + 4
			for i, l := range lines {
				fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" font-weight="600" fill="%s" text-anchor="middle">%s</text>`,
					num(n.X), num(top+(float64(i)+0.9)*lh), num(n.Size), esc(th.Foreground), esc(l))
			}
		}
	default:
		rx := 10.0
		if n.Shape == ShapePill {
			rx = n.H / 2
		}
		fmt.Fprintf(b, `<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" stroke="%s" stroke-width="1.6"/>`,
			num(n.X-n.W/2), num(n.Y-n.H/2), num(n.W), num(n.H), num(rx), esc(th.Fill), esc(stroke))
		tx := n.X
		if n.Icon != "" {
			fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" text-anchor="middle">%s</text>`,
				num(n.X-n.W/2+16), num(n.Y+n.Size*0.4), num(n.Size*1.3), esc(n.Icon))
			tx = n.X + 10
		}
		top := n.Y - float64(len(lines))*lh/2
		for i, l := range lines {
			fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" font-weight="600" fill="%s" text-anchor="middle">%s</text>`,
				num(tx), num(top+(float64(i)+0.8)*lh), num(n.Size), esc(th.Foreground), esc(l))
		}
	}
	b.WriteString(`</g>`)
}

// renderPill draws a small rounded label centered on p.
func (c *Card) renderPill(b *strings.Builder, p pt, label string) {
	th := c.Theme
	const size, padX, h = 11.0, 8.0, 18.0
	w := textWidth(label, size, false) + 2*padX
	fmt.Fprintf(b, `<g class="edge-label"><rect x="%s" y="%s" width="%s" height="%s" rx="6" fill="%s" stroke="%s" stroke-width="1"/><text x="%s" y="%s" font-size="%s" fill="%s" text-anchor="middle">%s</text></g>`,
		num(p.X-w/2), num(p.Y-h/2), num(w), num(h), esc(th.Background), esc(th.Stroke),
		num(p.X), num(p.Y+size*0.35), num(size), esc(th.Foreground), esc(label))
}

func (c *Card) renderTable(b *strings.Builder, p Panel, y int) error {
	th := c.Theme
	t := p.Table
	fs := t.FontSize
	if fs == 0 {
		fs = 15
	}
	const padY, lh = 8.0, 1.3
	m := float64(DefaultMargin)
	full := float64(c.Width) - 2*m
	labelW := 150.0
	colW := (full - labelW) / float64(len(t.Columns)-1)
	if len(t.Columns) == 1 {
		colW = full
		labelW = 0
	}
	x := func(col int) float64 {
		if col == 0 {
			return m
		}
		return m + labelW + float64(col-1)*colW
	}
	width := func(col int) float64 {
		if col == 0 {
			return labelW
		}
		return colW
	}

	rowH := func(cells []string, size float64, bold bool) (float64, [][]string) {
		tallest := 1
		wrapped := make([][]string, len(cells))
		for i, cell := range cells {
			wrapped[i] = wrap(cell, size, width(i)-12, bold)
			if len(wrapped[i]) > tallest {
				tallest = len(wrapped[i])
			}
		}
		return float64(tallest)*size*lh + 2*padY, wrapped
	}

	// Title above the table, if any.
	cy := float64(y) + 10
	if p.Title != "" {
		fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" font-weight="700" fill="%s">%s</text>`,
			num(m), num(cy+p.TitleSize*0.85), num(p.TitleSize), esc(th.Foreground), esc(p.Title))
		cy += p.TitleSize*1.3 + 6
	}

	hh, hw := rowH(t.Columns, fs+1, true)
	drawRow := func(top float64, cells [][]string, size float64, bold bool, colorFor func(int) string) {
		for i, lines := range cells {
			for j, l := range lines {
				fmt.Fprintf(b, `<text x="%s" y="%s" font-size="%s" fill="%s"%s>%s</text>`,
					num(x(i)+6), num(top+padY+(float64(j)+0.85)*size*lh), num(size), esc(colorFor(i)), boldAttr(bold), esc(l))
			}
		}
	}
	drawRow(cy, hw, fs+1, true, func(i int) string {
		if i < len(t.ColumnColors) && t.ColumnColors[i] != "" {
			return t.ColumnColors[i]
		}
		return th.Foreground
	})
	cy += hh
	fmt.Fprintf(b, `<rect class="rule" x="%s" y="%s" width="%s" height="2" fill="%s"/>`, num(m), num(cy-1), num(full), esc(th.Stroke))

	for _, row := range t.Rows {
		h, wr := rowH(row, fs, false)
		drawRow(cy, wr, fs, false, func(i int) string {
			if i == 0 {
				return th.Foreground
			}
			return th.Muted
		})
		cy += h
		fmt.Fprintf(b, `<rect class="rule" x="%s" y="%s" width="%s" height="1" fill="%s" fill-opacity="0.25"/>`, num(m), num(cy-0.5), num(full), esc(th.Stroke))
	}
	if cy > float64(y+p.Height) {
		return fmt.Errorf("table needs %.0f px but panel is %d px tall", cy-float64(y), p.Height)
	}
	return nil
}

func (c *Card) renderFooter(b *strings.Builder) {
	if len(c.Legend) == 0 && c.Footer == "" {
		return
	}
	th := c.Theme
	m := float64(DefaultMargin)
	y := float64(c.Height) - 26
	x := m
	b.WriteString(`<g id="legend">`)
	for _, it := range c.Legend {
		fmt.Fprintf(b, `<circle cx="%s" cy="%s" r="5" fill="%s"/><text x="%s" y="%s" font-size="12" fill="%s">%s</text>`,
			num(x+5), num(y), esc(it.Color), num(x+16), num(y+4), esc(th.Muted), esc(it.Label))
		x += 16 + textWidth(it.Label, 12, false) + 20
	}
	b.WriteString(`</g>`)
	if c.Footer != "" {
		fmt.Fprintf(b, `<text id="footer" x="%s" y="%s" font-size="11" fill="%s" text-anchor="end" fill-opacity="0.7">%s</text>`,
			num(float64(c.Width)-m), num(y+4), esc(th.Muted), esc(c.Footer))
	}
}

func boldAttr(bold bool) string {
	if bold {
		return ` font-weight="700"`
	}
	return ""
}

func esc(s string) string { return html.EscapeString(s) }

// num formats a float compactly for SVG attributes.
func num(v float64) string {
	if math.Abs(v) < 0.005 {
		return "0"
	}
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

func joinNums(vs []float64) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = strconv.FormatFloat(math.Round(v*10000)/10000, 'f', -1, 64)
	}
	return strings.Join(parts, ";")
}
