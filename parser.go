package d2vision

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Parser parses D2-generated SVG diagrams.
//
// D2 encodes element IDs as base64 in CSS class names. For example:
//   - "YQ==" decodes to "a" (a node)
//   - "KGEgLT4gYilbMF0=" decodes to "(a -> b)[0]" (an edge)
//
// The parser decodes these class names to reconstruct the diagram structure.
type Parser struct {
	// IncludePaths controls whether edge path coordinates are included in output.
	IncludePaths bool
	// IncludeStyles controls whether style information (fill, stroke) is extracted.
	IncludeStyles bool
}

// NewParser creates a new Parser with default settings.
func NewParser() *Parser {
	return &Parser{
		IncludePaths:  true,
		IncludeStyles: true,
	}
}

// Parse parses an SVG from the given reader.
func (p *Parser) Parse(r io.Reader) (*Diagram, error) {
	decoder := xml.NewDecoder(r)
	// Use lenient parsing to handle HTML entities in SVG
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity

	var svg svgRoot
	if err := decoder.Decode(&svg); err != nil {
		return nil, fmt.Errorf("failed to parse SVG: %w", err)
	}

	return p.parseSVG(&svg)
}

// ParseBytes parses an SVG from bytes.
func (p *Parser) ParseBytes(data []byte) (*Diagram, error) {
	return p.Parse(strings.NewReader(string(data)))
}

// ParseString parses an SVG from a string.
func (p *Parser) ParseString(s string) (*Diagram, error) {
	return p.Parse(strings.NewReader(s))
}

// =============================================================================
// SVG XML Structure Types
// =============================================================================
//
// D2 generates SVGs with a specific structure:
//
//   <svg ...>                          <!-- outer SVG with viewBox, version -->
//     <svg class="d2-xxx d2-svg" ...>  <!-- inner SVG containing actual diagram -->
//       <g class="YQ==">...</g>        <!-- node "a" (base64 encoded) -->
//       <g class="Yg==">...</g>        <!-- node "b" -->
//       <g class="KGEgLT4gYilbMF0=">   <!-- edge "(a -> b)[0]" -->
//         ...
//       </g>
//     </svg>
//   </svg>
//
// Key insight: D2 uses NESTED <svg> elements. The diagram content is inside
// an inner <svg>, not directly under the root. We must handle this nesting
// to extract bounds and other properties correctly.

// svgRoot represents the root SVG element.
type svgRoot struct {
	XMLName    xml.Name    `xml:"svg"`
	ViewBox    string      `xml:"viewBox,attr"`
	Width      string      `xml:"width,attr"`
	Height     string      `xml:"height,attr"`
	Version    string      `xml:"data-d2-version,attr"`
	Children   []svgGroup  `xml:"g"`
	NestedSVGs []svgNested `xml:"svg"` // D2 nests diagram content in inner <svg>
	Defs       svgDefs     `xml:"defs"`
	InnerXML   string      `xml:",innerxml"`
}

// svgNested represents a nested SVG element.
// D2 wraps the actual diagram content in an inner <svg> element, which
// contains the <g> groups for nodes and edges. Without parsing this nested
// structure, we would miss the actual diagram elements and their bounds.
type svgNested struct {
	ViewBox  string      `xml:"viewBox,attr"`
	Children []svgGroup  `xml:"g"`
}

type svgDefs struct {
	Markers []svgMarker `xml:"marker"`
}

type svgMarker struct {
	ID string `xml:"id,attr"`
}

// svgGroup represents a <g> element which may contain nodes or edges.
// D2 encodes the element ID as a base64 string in the class attribute.
type svgGroup struct {
	Class    string       `xml:"class,attr"`
	ID       string       `xml:"id,attr"`
	Children []svgGroup   `xml:"g"`
	Rects    []svgRect    `xml:"rect"`
	Circles  []svgCircle  `xml:"circle"`
	Ellipses []svgEllipse `xml:"ellipse"`
	Paths    []svgPath    `xml:"path"`
	Texts    []svgText    `xml:"text"`
	InnerXML string       `xml:",innerxml"`
}

type svgRect struct {
	X      string `xml:"x,attr"`
	Y      string `xml:"y,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
	Fill   string `xml:"fill,attr"`
	Stroke string `xml:"stroke,attr"`
	Class  string `xml:"class,attr"`
}

type svgCircle struct {
	CX     string `xml:"cx,attr"`
	CY     string `xml:"cy,attr"`
	R      string `xml:"r,attr"`
	Fill   string `xml:"fill,attr"`
	Stroke string `xml:"stroke,attr"`
}

type svgEllipse struct {
	CX     string `xml:"cx,attr"`
	CY     string `xml:"cy,attr"`
	RX     string `xml:"rx,attr"`
	RY     string `xml:"ry,attr"`
	Fill   string `xml:"fill,attr"`
	Stroke string `xml:"stroke,attr"`
}

type svgPath struct {
	D         string `xml:"d,attr"`
	Fill      string `xml:"fill,attr"`
	Stroke    string `xml:"stroke,attr"`
	MarkerEnd string `xml:"marker-end,attr"`
	Class     string `xml:"class,attr"`
}

type svgText struct {
	X       string     `xml:"x,attr"`
	Y       string     `xml:"y,attr"`
	Content string     `xml:",chardata"`
	TSpans  []svgTSpan `xml:"tspan"`
}

type svgTSpan struct {
	Content string `xml:",chardata"`
}

// =============================================================================
// Parsing Implementation
// =============================================================================

func (p *Parser) parseSVG(svg *svgRoot) (*Diagram, error) {
	diagram := &Diagram{
		Version: svg.Version,
		Nodes:   []Node{},
		Edges:   []Edge{},
	}

	// Parse viewBox for diagram dimensions
	if svg.ViewBox != "" {
		vb, err := ParseViewBox(svg.ViewBox)
		if err != nil {
			return nil, err
		}
		diagram.ViewBox = vb
	}

	// Track nodes/edges by ID to avoid duplicates
	nodeMap := make(map[string]*Node)
	edgeMap := make(map[string]*Edge)

	// Process groups directly under root SVG (rarely used by D2)
	p.processGroups(svg.Children, diagram, nodeMap, edgeMap)

	// Process groups inside nested SVG elements.
	// IMPORTANT: D2 wraps diagram content in an inner <svg> element.
	// Without this, we would miss all nodes/edges and their bounds.
	for _, nested := range svg.NestedSVGs {
		p.processGroups(nested.Children, diagram, nodeMap, edgeMap)
	}

	// Fallback: scan raw XML for any elements we might have missed
	p.processInnerXML(svg.InnerXML, diagram, nodeMap, edgeMap)

	// Build container hierarchy from ID prefixes (e.g., "container.node")
	p.buildHierarchy(diagram)

	return diagram, nil
}

// processGroups recursively processes SVG groups to extract nodes and edges.
func (p *Parser) processGroups(groups []svgGroup, diagram *Diagram, nodeMap map[string]*Node, edgeMap map[string]*Edge) {
	for _, g := range groups {
		p.processGroup(g, diagram, nodeMap, edgeMap)
		// Recurse into child groups
		p.processGroups(g.Children, diagram, nodeMap, edgeMap)
	}
}

// processGroup attempts to decode a single SVG group as a D2 element.
func (p *Parser) processGroup(g svgGroup, diagram *Diagram, nodeMap map[string]*Node, edgeMap map[string]*Edge) {
	if g.Class == "" {
		return
	}

	// Attempt to decode the class as a base64-encoded D2 ID
	decodedID, err := DecodeBase64ID(g.Class)
	if err != nil {
		// Not a base64 class - could be a regular CSS class like "shape"
		return
	}

	decodedID = NormalizeID(decodedID)

	// Determine if this is an edge or a node based on ID format
	if IsEdgeID(decodedID) {
		edge := p.parseEdge(decodedID, g)
		if edge != nil && edgeMap[edge.ID] == nil {
			diagram.Edges = append(diagram.Edges, *edge)
			edgeMap[edge.ID] = edge
		}
	} else {
		node := p.parseNode(decodedID, g)
		if node != nil && nodeMap[node.ID] == nil {
			diagram.Nodes = append(diagram.Nodes, *node)
			nodeMap[node.ID] = node
		}
	}
}

// parseNode extracts node information from an SVG group.
func (p *Parser) parseNode(id string, g svgGroup) *Node {
	node := &Node{
		ID:    id,
		Shape: ShapeUnknown,
	}

	// Extract label from <text> elements
	node.Label = p.extractLabel(g)

	// Detect shape from SVG primitives (rect, circle, ellipse, path)
	node.Shape = p.detectShape(g)

	// Extract bounding box - searches recursively through nested groups
	// D2 often nests the actual shape inside a <g class="shape"> child
	node.Bounds = p.extractBounds(g)

	if p.IncludeStyles {
		node.Style = p.extractNodeStyle(g)
	}

	return node
}

// parseEdge extracts edge information from an SVG group.
func (p *Parser) parseEdge(id string, g svgGroup) *Edge {
	endpoints, ok := ParseEdgeID(id)
	if !ok {
		return nil
	}

	edge := &Edge{
		ID:          id,
		Source:      endpoints.Source,
		Target:      endpoints.Target,
		TargetArrow: ArrowTriangle, // D2 default
	}

	edge.Label = p.extractLabel(g)
	edge.SourceArrow, edge.TargetArrow = p.detectArrows(g)

	if p.IncludePaths {
		edge.Path = p.extractPath(g)
	}

	if p.IncludeStyles {
		edge.Style = p.extractEdgeStyle(g)
	}

	return edge
}

// =============================================================================
// Element Extraction Helpers
// =============================================================================

// extractLabel finds text content within an SVG group.
// Searches both direct <text> children and nested groups.
func (p *Parser) extractLabel(g svgGroup) string {
	// Check direct text elements
	for _, t := range g.Texts {
		label := p.textContent(t)
		if label != "" {
			return html.UnescapeString(label)
		}
	}

	// Recursively check nested groups
	for _, child := range g.Children {
		label := p.extractLabel(child)
		if label != "" {
			return label
		}
	}

	return ""
}

// textContent extracts text from an svgText element.
// Handles both direct content and <tspan> children.
func (p *Parser) textContent(t svgText) string {
	// Check tspans first (D2 often uses these)
	var parts []string
	for _, ts := range t.TSpans {
		content := strings.TrimSpace(ts.Content)
		if content != "" {
			parts = append(parts, content)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, " ")
	}

	// Fall back to direct content
	return strings.TrimSpace(t.Content)
}

// detectShape determines the shape type from SVG primitives.
// Checks rect, circle, ellipse, and path elements.
func (p *Parser) detectShape(g svgGroup) ShapeType {
	// Check rectangles
	if len(g.Rects) > 0 {
		rect := g.Rects[0]
		w, _ := strconv.ParseFloat(rect.Width, 64)
		h, _ := strconv.ParseFloat(rect.Height, 64)
		if w > 0 && h > 0 {
			if w == h {
				return ShapeSquare
			}
			return ShapeRectangle
		}
	}

	// Check circles
	if len(g.Circles) > 0 {
		return ShapeCircle
	}

	// Check ellipses (could be circle or oval)
	if len(g.Ellipses) > 0 {
		e := g.Ellipses[0]
		rx, _ := strconv.ParseFloat(e.RX, 64)
		ry, _ := strconv.ParseFloat(e.RY, 64)
		if rx == ry {
			return ShapeCircle
		}
		return ShapeOval
	}

	// Check paths for special shapes (cylinder, diamond, hexagon)
	for _, path := range g.Paths {
		shape := p.detectShapeFromPath(path.D)
		if shape != ShapeUnknown {
			return shape
		}
	}

	// Recursively check nested groups
	// D2 often nests shapes inside <g class="shape">
	for _, child := range g.Children {
		shape := p.detectShape(child)
		if shape != ShapeUnknown {
			return shape
		}
	}

	return ShapeRectangle // Default
}

// detectShapeFromPath uses heuristics to detect special shapes from SVG paths.
//
// D2 renders special shapes using SVG paths:
//   - Cylinder: Uses cubic bezier (C) curves for rounded top/bottom + vertical (V) lines
//   - Diamond: 4 vertices forming a closed rhombus
//   - Hexagon: 6 vertices forming a closed hexagon
//
// The heuristics analyze command frequency and patterns to identify shapes.
func (p *Parser) detectShapeFromPath(d string) ShapeType {
	if d == "" {
		return ShapeUnknown
	}

	// Count SVG path commands using regex to avoid false matches in numbers
	// Each pattern matches the command letter followed by coordinates
	moveCount := len(regexp.MustCompile(`[Mm]\s*[-\d]`).FindAllString(d, -1))
	cubicCount := len(regexp.MustCompile(`[Cc]\s*[-\d]`).FindAllString(d, -1))
	vertCount := len(regexp.MustCompile(`[Vv]\s*[-\d]`).FindAllString(d, -1))
	horizCount := len(regexp.MustCompile(`[Hh]\s*[-\d]`).FindAllString(d, -1))
	lineCount := len(regexp.MustCompile(`[Ll]\s*[-\d]`).FindAllString(d, -1))
	isClosed := regexp.MustCompile(`[Zz]\s*$`).MatchString(strings.TrimSpace(d))

	// Total vertices = M commands (starting points) + L commands (line endpoints)
	// For polygons, M establishes first vertex, each L adds another
	vertexCount := moveCount + lineCount

	// Cylinder detection:
	// D2 cylinders use cubic bezier curves (C) for the rounded top and bottom,
	// and vertical lines (V) for the sides. Pattern:
	//   M start C curve C curve V down C curve C curve Z
	// Typically has 4+ cubic curves and 1+ vertical lines
	if cubicCount >= 4 && vertCount >= 1 && isClosed {
		return ShapeCylinder
	}

	// Also detect cylinders that use arcs (A commands) instead of cubic curves
	arcCount := len(regexp.MustCompile(`[Aa]\s*[-\d]`).FindAllString(d, -1))
	if arcCount >= 2 && isClosed {
		return ShapeCylinder
	}

	// Diamond: exactly 4 vertices forming a closed path
	// No curves, just straight lines (M + L commands only)
	if vertexCount == 4 && cubicCount == 0 && isClosed {
		return ShapeDiamond
	}

	// Hexagon: 6 vertices forming a closed path
	if vertexCount == 6 && cubicCount == 0 && isClosed {
		return ShapeHexagon
	}

	// Parallelogram: 4 vertices with horizontal line components
	if vertexCount == 4 && horizCount > 0 && isClosed {
		return ShapeParallelogram
	}

	return ShapeUnknown
}

// extractBounds determines the bounding box of a node.
// Searches recursively because D2 often nests shapes inside child groups.
//
// SVG Structure example:
//
//	<g class="YQ==">              <!-- node group with base64 ID -->
//	  <g class="shape">           <!-- nested shape group -->
//	    <rect x="0" y="0" .../>   <!-- actual shape with bounds -->
//	  </g>
//	  <text>label</text>
//	</g>
//
// For path-based shapes (cylinders, diamonds, hexagons), we parse the SVG path
// d attribute to extract coordinates and calculate the bounding box.
func (p *Parser) extractBounds(g svgGroup) Bounds {
	// Try rectangles
	if len(g.Rects) > 0 {
		rect := g.Rects[0]
		x, _ := strconv.ParseFloat(rect.X, 64)
		y, _ := strconv.ParseFloat(rect.Y, 64)
		w, _ := strconv.ParseFloat(rect.Width, 64)
		h, _ := strconv.ParseFloat(rect.Height, 64)
		return Bounds{X: x, Y: y, Width: w, Height: h}
	}

	// Try circles (convert center+radius to bounding box)
	if len(g.Circles) > 0 {
		c := g.Circles[0]
		cx, _ := strconv.ParseFloat(c.CX, 64)
		cy, _ := strconv.ParseFloat(c.CY, 64)
		r, _ := strconv.ParseFloat(c.R, 64)
		return Bounds{X: cx - r, Y: cy - r, Width: r * 2, Height: r * 2}
	}

	// Try ellipses
	if len(g.Ellipses) > 0 {
		e := g.Ellipses[0]
		cx, _ := strconv.ParseFloat(e.CX, 64)
		cy, _ := strconv.ParseFloat(e.CY, 64)
		rx, _ := strconv.ParseFloat(e.RX, 64)
		ry, _ := strconv.ParseFloat(e.RY, 64)
		return Bounds{X: cx - rx, Y: cy - ry, Width: rx * 2, Height: ry * 2}
	}

	// Try paths (for cylinders, diamonds, hexagons, and other path-based shapes)
	// D2 renders these shapes using SVG <path> elements, not rect/circle/ellipse
	for _, path := range g.Paths {
		bounds := extractBoundsFromPath(path.D)
		if bounds.Width > 0 || bounds.Height > 0 {
			return bounds
		}
	}

	// Recursively check nested groups
	// This is essential for D2 SVGs where shapes are nested
	for _, child := range g.Children {
		bounds := p.extractBounds(child)
		if bounds.Width > 0 || bounds.Height > 0 {
			return bounds
		}
	}

	return Bounds{}
}

// extractBoundsFromPath calculates bounding box from an SVG path d attribute.
// This handles path-based shapes like cylinders, diamonds, and hexagons.
//
// SVG path commands supported:
//   - M/m x y: moveto (absolute/relative)
//   - L/l x y: lineto
//   - H/h x: horizontal line
//   - V/v y: vertical line
//   - C/c x1 y1 x2 y2 x y: cubic bezier (uses all control points)
//   - S/s x2 y2 x y: smooth cubic bezier
//   - Q/q x1 y1 x y: quadratic bezier
//   - T/t x y: smooth quadratic bezier
//   - A/a rx ry rotation large-arc sweep x y: arc (uses endpoint)
//   - Z/z: close path (no coordinates)
func extractBoundsFromPath(d string) Bounds {
	if d == "" {
		return Bounds{}
	}

	// Extract all numeric values from the path
	// This regex finds all numbers (including negative and decimals)
	numPattern := regexp.MustCompile(`[-+]?(?:\d+\.?\d*|\.\d+)(?:[eE][-+]?\d+)?`)
	matches := numPattern.FindAllString(d, -1)

	if len(matches) < 2 {
		return Bounds{}
	}

	// Parse numbers and separate X/Y coordinates
	// Most SVG path commands use x,y pairs, so we alternate
	var allX, allY []float64
	var currentX, currentY float64

	// Parse path commands to properly interpret coordinates
	cmdPattern := regexp.MustCompile(`([MmLlHhVvCcSsQqTtAaZz])([^MmLlHhVvCcSsQqTtAaZz]*)`)
	cmdMatches := cmdPattern.FindAllStringSubmatch(d, -1)

	for _, cmdMatch := range cmdMatches {
		cmd := cmdMatch[1]
		args := cmdMatch[2]

		// Extract numbers from this command's arguments
		nums := numPattern.FindAllString(args, -1)
		var vals []float64
		for _, n := range nums {
			v, err := strconv.ParseFloat(n, 64)
			if err == nil {
				vals = append(vals, v)
			}
		}

		isRelative := len(cmd) > 0 && cmd[0] >= 'a' && cmd[0] <= 'z'
		cmdUpper := strings.ToUpper(cmd)

		switch cmdUpper {
		case "M", "L", "T":
			// x,y pairs
			for i := 0; i+1 < len(vals); i += 2 {
				x, y := vals[i], vals[i+1]
				if isRelative {
					x += currentX
					y += currentY
				}
				allX = append(allX, x)
				allY = append(allY, y)
				currentX, currentY = x, y
			}
		case "H":
			// horizontal line: x only
			for _, x := range vals {
				if isRelative {
					x += currentX
				}
				allX = append(allX, x)
				currentX = x
			}
		case "V":
			// vertical line: y only
			for _, y := range vals {
				if isRelative {
					y += currentY
				}
				allY = append(allY, y)
				currentY = y
			}
		case "C":
			// cubic bezier: x1,y1 x2,y2 x,y
			for i := 0; i+5 < len(vals); i += 6 {
				for j := 0; j < 6; j += 2 {
					x, y := vals[i+j], vals[i+j+1]
					if isRelative {
						x += currentX
						y += currentY
					}
					allX = append(allX, x)
					allY = append(allY, y)
				}
				if isRelative {
					currentX += vals[i+4]
					currentY += vals[i+5]
				} else {
					currentX = vals[i+4]
					currentY = vals[i+5]
				}
			}
		case "S", "Q":
			// smooth cubic/quadratic: x2,y2 x,y or x1,y1 x,y
			for i := 0; i+3 < len(vals); i += 4 {
				for j := 0; j < 4; j += 2 {
					x, y := vals[i+j], vals[i+j+1]
					if isRelative {
						x += currentX
						y += currentY
					}
					allX = append(allX, x)
					allY = append(allY, y)
				}
				if isRelative {
					currentX += vals[i+2]
					currentY += vals[i+3]
				} else {
					currentX = vals[i+2]
					currentY = vals[i+3]
				}
			}
		case "A":
			// arc: rx ry rotation large-arc sweep x y
			for i := 0; i+6 < len(vals); i += 7 {
				x, y := vals[i+5], vals[i+6]
				if isRelative {
					x += currentX
					y += currentY
				}
				allX = append(allX, x)
				allY = append(allY, y)
				currentX, currentY = x, y
			}
		case "Z":
			// close path - no coordinates
		}
	}

	if len(allX) == 0 || len(allY) == 0 {
		return Bounds{}
	}

	// Find min/max
	minX, maxX := allX[0], allX[0]
	for _, x := range allX {
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
	}

	minY, maxY := allY[0], allY[0]
	for _, y := range allY {
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	return Bounds{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// detectArrows determines arrow types from SVG markers.
func (p *Parser) detectArrows(g svgGroup) (source, target ArrowType) {
	source = ArrowNone
	target = ArrowTriangle // D2 default

	for _, path := range g.Paths {
		if path.MarkerEnd != "" {
			target = p.parseMarkerType(path.MarkerEnd)
		}
	}

	// Check nested groups
	for _, child := range g.Children {
		s, t := p.detectArrows(child)
		if s != ArrowNone {
			source = s
		}
		if t != ArrowNone && t != ArrowTriangle {
			target = t
		}
	}

	return source, target
}

// parseMarkerType determines arrow type from a marker reference URL.
func (p *Parser) parseMarkerType(markerRef string) ArrowType {
	markerRef = strings.ToLower(markerRef)

	if strings.Contains(markerRef, "diamond") {
		return ArrowDiamond
	}
	if strings.Contains(markerRef, "circle") {
		return ArrowCircle
	}
	if strings.Contains(markerRef, "crowfoot") || strings.Contains(markerRef, "cf-") {
		return ArrowCrowfoot
	}

	return ArrowTriangle
}

// extractPath extracts edge path coordinates from SVG path elements.
func (p *Parser) extractPath(g svgGroup) []Point {
	for _, path := range g.Paths {
		if path.D != "" && path.Stroke != "" && path.Stroke != "none" {
			return parseSVGPath(path.D)
		}
	}

	// Check nested groups
	for _, child := range g.Children {
		points := p.extractPath(child)
		if len(points) > 0 {
			return points
		}
	}

	return nil
}

// parseSVGPath parses an SVG path d attribute into coordinate points.
// This is a simplified parser that handles M (moveto) and L (lineto) commands.
func parseSVGPath(d string) []Point {
	var points []Point

	coordPattern := regexp.MustCompile(`[ML]\s*([-\d.]+)[,\s]+([-\d.]+)`)
	matches := coordPattern.FindAllStringSubmatch(d, -1)

	for _, match := range matches {
		x, err1 := strconv.ParseFloat(match[1], 64)
		y, err2 := strconv.ParseFloat(match[2], 64)
		if err1 == nil && err2 == nil {
			points = append(points, Point{X: x, Y: y})
		}
	}

	return points
}

// extractNodeStyle extracts visual styling from a node's SVG elements.
func (p *Parser) extractNodeStyle(g svgGroup) NodeStyle {
	style := NodeStyle{}

	if len(g.Rects) > 0 {
		rect := g.Rects[0]
		style.Fill = rect.Fill
		style.Stroke = rect.Stroke
	}

	return style
}

// extractEdgeStyle extracts visual styling from an edge's SVG elements.
func (p *Parser) extractEdgeStyle(g svgGroup) EdgeStyle {
	style := EdgeStyle{}

	for _, path := range g.Paths {
		if path.Stroke != "" {
			style.Stroke = path.Stroke
			break
		}
	}

	return style
}

// =============================================================================
// Hierarchy Building
// =============================================================================

// buildHierarchy establishes parent-child relationships between nodes.
// D2 uses dot-separated IDs for hierarchy: "container.child" means "child"
// is inside "container". This function populates Parent and Children fields.
//
// Note: We use an index map rather than pointer map to avoid issues with
// slice reallocation when modifying nodes during iteration.
func (p *Parser) buildHierarchy(diagram *Diagram) {
	// Build a map from ID to slice index
	idxMap := make(map[string]int, len(diagram.Nodes))
	for i := range diagram.Nodes {
		idxMap[diagram.Nodes[i].ID] = i
	}

	// Establish parent-child relationships
	for i := range diagram.Nodes {
		node := &diagram.Nodes[i]
		parentID := ExtractParentID(node.ID)
		if parentID != "" {
			node.Parent = parentID
			if parentIdx, ok := idxMap[parentID]; ok {
				diagram.Nodes[parentIdx].Children = append(diagram.Nodes[parentIdx].Children, node.ID)
			}
		}
	}
}

// =============================================================================
// Fallback XML Processing
// =============================================================================

// processInnerXML scans raw XML for any base64-encoded classes we might have
// missed during structured parsing. This is a fallback mechanism.
func (p *Parser) processInnerXML(innerXML string, diagram *Diagram, nodeMap map[string]*Node, edgeMap map[string]*Edge) {
	classPattern := regexp.MustCompile(`class="([^"]+)"`)
	matches := classPattern.FindAllStringSubmatch(innerXML, -1)

	for _, match := range matches {
		className := match[1]

		// Skip non-base64 classes
		if !mightBeBase64(className) {
			continue
		}

		decodedID, err := DecodeBase64ID(className)
		if err != nil {
			continue
		}

		decodedID = NormalizeID(decodedID)

		if IsEdgeID(decodedID) {
			if edgeMap[decodedID] == nil {
				endpoints, ok := ParseEdgeID(decodedID)
				if ok {
					edge := &Edge{
						ID:          decodedID,
						Source:      endpoints.Source,
						Target:      endpoints.Target,
						TargetArrow: ArrowTriangle,
					}
					diagram.Edges = append(diagram.Edges, *edge)
					edgeMap[decodedID] = edge
				}
			}
		} else {
			if nodeMap[decodedID] == nil {
				node := &Node{
					ID:    decodedID,
					Shape: ShapeRectangle,
				}
				diagram.Nodes = append(diagram.Nodes, *node)
				nodeMap[decodedID] = node
			}
		}
	}
}

// mightBeBase64 checks if a string could be base64-encoded.
// Base64 uses only: A-Z, a-z, 0-9, +, /, and = (padding).
func mightBeBase64(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		isUpper := c >= 'A' && c <= 'Z'
		isLower := c >= 'a' && c <= 'z'
		isDigit := c >= '0' && c <= '9'
		isSpecial := c == '+' || c == '/' || c == '='
		if !isUpper && !isLower && !isDigit && !isSpecial {
			return false
		}
	}
	return true
}
