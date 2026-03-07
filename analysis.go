package d2vision

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// LayoutAnalysis contains analysis of a diagram's layout characteristics.
type LayoutAnalysis struct {
	// Layout type detection
	LayoutType       string   `json:"layoutType" toon:"LayoutType"`
	Direction        string   `json:"direction,omitempty" toon:"Direction"`
	GridColumns      int      `json:"gridColumns,omitempty" toon:"GridColumns"`
	GridRows         int      `json:"gridRows,omitempty" toon:"GridRows"`
	HasContainers    bool     `json:"hasContainers" toon:"HasContainers"`
	ContainerCount   int      `json:"containerCount" toon:"ContainerCount"`
	NestingDepth     int      `json:"nestingDepth" toon:"NestingDepth"`
	CrossContainerEdges int   `json:"crossContainerEdges" toon:"CrossContainerEdges"`

	// Layout insights
	Insights []string `json:"insights" toon:"Insights"`

	// Generation hints
	GenerationHints []string `json:"generationHints" toon:"GenerationHints"`
}

// AnalyzeLayout analyzes the diagram's layout characteristics.
func (d *Diagram) AnalyzeLayout() *LayoutAnalysis {
	analysis := &LayoutAnalysis{}

	// Count containers and analyze nesting
	containers := d.ContainerNodes()
	analysis.HasContainers = len(containers) > 0
	analysis.ContainerCount = len(containers)
	analysis.NestingDepth = d.calculateNestingDepth()

	// Detect layout type and direction
	analysis.LayoutType, analysis.Direction = d.detectLayoutType()

	// Detect grid layout
	analysis.GridColumns, analysis.GridRows = d.detectGridLayout()

	// Count cross-container edges
	analysis.CrossContainerEdges = d.countCrossContainerEdges()

	// Generate insights
	analysis.Insights = d.generateInsights(analysis)

	// Generate hints for recreation
	analysis.GenerationHints = d.generateHints(analysis)

	return analysis
}

func (d *Diagram) calculateNestingDepth() int {
	maxDepth := 0
	for _, node := range d.Nodes {
		depth := d.nodeDepth(node.ID)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

func (d *Diagram) nodeDepth(nodeID string) int {
	depth := 0
	currentID := nodeID
	visited := make(map[string]bool)

	for !visited[currentID] {
		visited[currentID] = true

		node := d.NodeByID(currentID)
		if node == nil || node.Parent == "" {
			break
		}
		depth++
		currentID = node.Parent
	}
	return depth
}

func (d *Diagram) detectLayoutType() (layoutType, direction string) {
	if len(d.Nodes) == 0 {
		return "empty", ""
	}

	containers := d.ContainerNodes()

	// Check for side-by-side layout
	if len(containers) >= 2 {
		if d.areSideBySide(containers) {
			return "side-by-side", "right"
		}
		if d.areStackedVertically(containers) {
			return "stacked", "down"
		}
	}

	// Check for hierarchical layout
	if d.ContainerCount() > 0 {
		return "hierarchical", "down"
	}

	// Simple flow
	if len(d.Edges) > 0 {
		direction = d.detectFlowDirection()
		return "flow", direction
	}

	return "simple", ""
}

func (d *Diagram) ContainerCount() int {
	count := 0
	for _, n := range d.Nodes {
		if n.HasChildren() {
			count++
		}
	}
	return count
}

func (d *Diagram) areSideBySide(containers []Node) bool {
	if len(containers) < 2 {
		return false
	}

	// Get root containers only
	var rootContainers []Node
	for _, c := range containers {
		if c.Parent == "" {
			rootContainers = append(rootContainers, c)
		}
	}

	if len(rootContainers) < 2 {
		return false
	}

	// Check if containers are arranged horizontally
	// They should have similar Y positions but different X positions
	var yPositions []float64
	for _, c := range rootContainers {
		yPositions = append(yPositions, c.Bounds.Y)
	}

	// Calculate variance in Y positions
	variance := calculateVariance(yPositions)
	avgHeight := d.ViewBox.Height / float64(len(rootContainers))

	// If variance is small relative to container height, they're side-by-side
	return variance < avgHeight*0.3
}

func (d *Diagram) areStackedVertically(containers []Node) bool {
	if len(containers) < 2 {
		return false
	}

	var rootContainers []Node
	for _, c := range containers {
		if c.Parent == "" {
			rootContainers = append(rootContainers, c)
		}
	}

	if len(rootContainers) < 2 {
		return false
	}

	// Check if containers are arranged vertically
	var xPositions []float64
	for _, c := range rootContainers {
		xPositions = append(xPositions, c.Bounds.X)
	}

	variance := calculateVariance(xPositions)
	avgWidth := d.ViewBox.Width / float64(len(rootContainers))

	return variance < avgWidth*0.3
}

func (d *Diagram) detectFlowDirection() string {
	if len(d.Edges) == 0 {
		return ""
	}

	// Analyze edge directions
	rightCount := 0
	downCount := 0
	leftCount := 0
	upCount := 0

	for _, edge := range d.Edges {
		sourceNode := d.NodeByID(edge.Source)
		targetNode := d.NodeByID(edge.Target)

		if sourceNode == nil || targetNode == nil {
			continue
		}

		dx := targetNode.Bounds.Center().X - sourceNode.Bounds.Center().X
		dy := targetNode.Bounds.Center().Y - sourceNode.Bounds.Center().Y

		if math.Abs(dx) > math.Abs(dy) {
			if dx > 0 {
				rightCount++
			} else {
				leftCount++
			}
		} else {
			if dy > 0 {
				downCount++
			} else {
				upCount++
			}
		}
	}

	maxCount := max(rightCount, downCount, leftCount, upCount)
	switch maxCount {
	case rightCount:
		return "right"
	case downCount:
		return "down"
	case leftCount:
		return "left"
	case upCount:
		return "up"
	}
	return "down"
}

func (d *Diagram) detectGridLayout() (cols, rows int) {
	containers := d.ContainerNodes()
	if len(containers) < 2 {
		return 0, 0
	}

	// Get root containers
	var rootContainers []Node
	for _, c := range containers {
		if c.Parent == "" {
			rootContainers = append(rootContainers, c)
		}
	}

	if len(rootContainers) < 2 {
		return 0, 0
	}

	// Sort by X position to detect columns
	sort.Slice(rootContainers, func(i, j int) bool {
		return rootContainers[i].Bounds.X < rootContainers[j].Bounds.X
	})

	// Detect distinct columns by X position clustering
	tolerance := d.ViewBox.Width * 0.1
	var colPositions []float64

	for _, c := range rootContainers {
		found := false
		for _, pos := range colPositions {
			if math.Abs(c.Bounds.X-pos) < tolerance {
				found = true
				break
			}
		}
		if !found {
			colPositions = append(colPositions, c.Bounds.X)
		}
	}

	cols = len(colPositions)

	// Detect rows similarly
	sort.Slice(rootContainers, func(i, j int) bool {
		return rootContainers[i].Bounds.Y < rootContainers[j].Bounds.Y
	})

	var rowPositions []float64
	tolerance = d.ViewBox.Height * 0.1

	for _, c := range rootContainers {
		found := false
		for _, pos := range rowPositions {
			if math.Abs(c.Bounds.Y-pos) < tolerance {
				found = true
				break
			}
		}
		if !found {
			rowPositions = append(rowPositions, c.Bounds.Y)
		}
	}

	rows = len(rowPositions)

	return cols, rows
}

func (d *Diagram) countCrossContainerEdges() int {
	count := 0
	for _, edge := range d.Edges {
		sourceNode := d.NodeByID(edge.Source)
		targetNode := d.NodeByID(edge.Target)

		if sourceNode == nil || targetNode == nil {
			continue
		}

		// Check if source and target have different root containers
		sourceRoot := d.getRootContainer(edge.Source)
		targetRoot := d.getRootContainer(edge.Target)

		if sourceRoot != "" && targetRoot != "" && sourceRoot != targetRoot {
			count++
		}
	}
	return count
}

func (d *Diagram) getRootContainer(nodeID string) string {
	node := d.NodeByID(nodeID)
	if node == nil {
		return ""
	}

	if node.Parent == "" {
		if node.HasChildren() {
			return node.ID
		}
		return ""
	}

	// Walk up to find root container
	current := node
	for current.Parent != "" {
		parent := d.NodeByID(current.Parent)
		if parent == nil {
			break
		}
		current = parent
	}

	if current.HasChildren() {
		return current.ID
	}
	return ""
}

func (d *Diagram) generateInsights(analysis *LayoutAnalysis) []string {
	var insights []string

	// Layout type insight
	switch analysis.LayoutType {
	case "side-by-side":
		insights = append(insights, "Containers are arranged horizontally (side-by-side)")
	case "stacked":
		insights = append(insights, "Containers are stacked vertically")
	case "hierarchical":
		insights = append(insights, "Diagram uses nested container hierarchy")
	case "flow":
		insights = append(insights, fmt.Sprintf("Flow diagram with %s direction", analysis.Direction))
	}

	// Grid detection
	if analysis.GridColumns >= 2 {
		insights = append(insights, fmt.Sprintf("Grid layout detected: %d columns", analysis.GridColumns))
	}

	// Container insights
	if analysis.ContainerCount > 0 {
		insights = append(insights, fmt.Sprintf("%d container(s) with max nesting depth of %d", analysis.ContainerCount, analysis.NestingDepth))
	}

	// Cross-container edges
	if analysis.CrossContainerEdges > 0 {
		insights = append(insights, fmt.Sprintf("%d edge(s) cross container boundaries", analysis.CrossContainerEdges))
	}

	// Shape diversity
	shapes := d.countShapes()
	if len(shapes) > 1 {
		var shapeList []string
		for shape, count := range shapes {
			shapeList = append(shapeList, fmt.Sprintf("%s (%d)", shape, count))
		}
		sort.Strings(shapeList)
		insights = append(insights, fmt.Sprintf("Shape types: %s", strings.Join(shapeList, ", ")))
	}

	return insights
}

func (d *Diagram) countShapes() map[string]int {
	shapes := make(map[string]int)
	for _, node := range d.Nodes {
		shapes[string(node.Shape)]++
	}
	return shapes
}

func (d *Diagram) generateHints(analysis *LayoutAnalysis) []string {
	var hints []string

	// Grid layout hint
	if analysis.GridColumns >= 2 && analysis.LayoutType == "side-by-side" {
		hints = append(hints, fmt.Sprintf("Use `grid-columns: %d` to arrange containers side-by-side", analysis.GridColumns))
	}

	// Direction hint
	if analysis.Direction != "" {
		hints = append(hints, fmt.Sprintf("Set `direction: %s` for primary flow", analysis.Direction))
	}

	// Container hints
	if analysis.ContainerCount > 0 {
		hints = append(hints, "Define containers first, then add nodes inside them")
		if analysis.NestingDepth > 2 {
			hints = append(hints, fmt.Sprintf("Warning: Deep nesting (%d levels) may affect rendering performance", analysis.NestingDepth))
		}
	}

	// Cross-container edge hints
	if analysis.CrossContainerEdges > 0 {
		hints = append(hints, "Cross-container edges may affect layout alignment")
		hints = append(hints, "Consider using fully-qualified IDs (e.g., container.node) for edges")
	}

	// Shape-specific hints
	shapes := d.countShapes()
	if shapes["cylinder"] > 0 {
		hints = append(hints, "Use `shape: cylinder` for database/storage nodes")
	}
	if shapes["circle"] > 0 || shapes["oval"] > 0 {
		hints = append(hints, "Use `shape: circle` or `shape: oval` for actors/processes")
	}

	return hints
}

// DescribeForGeneration generates output optimized for recreating the diagram.
func (d *Diagram) DescribeForGeneration() string {
	var sb strings.Builder

	analysis := d.AnalyzeLayout()

	sb.WriteString("# D2 Diagram Analysis for Recreation\n\n")

	// Summary
	sb.WriteString("## Overview\n")
	fmt.Fprintf(&sb, "- %d nodes, %d edges\n", len(d.Nodes), len(d.Edges))
	fmt.Fprintf(&sb, "- Layout type: %s\n", analysis.LayoutType)
	if analysis.Direction != "" {
		fmt.Fprintf(&sb, "- Direction: %s\n", analysis.Direction)
	}
	if analysis.GridColumns > 0 {
		fmt.Fprintf(&sb, "- Grid: %d columns x %d rows\n", analysis.GridColumns, analysis.GridRows)
	}
	sb.WriteString("\n")

	// Layout insights
	if len(analysis.Insights) > 0 {
		sb.WriteString("## Layout Analysis\n")
		for _, insight := range analysis.Insights {
			fmt.Fprintf(&sb, "- %s\n", insight)
		}
		sb.WriteString("\n")
	}

	// Generation hints
	if len(analysis.GenerationHints) > 0 {
		sb.WriteString("## Generation Hints\n")
		for _, hint := range analysis.GenerationHints {
			fmt.Fprintf(&sb, "- %s\n", hint)
		}
		sb.WriteString("\n")
	}

	// Structure
	sb.WriteString("## Structure\n\n")

	// Containers
	containers := d.ContainerNodes()
	if len(containers) > 0 {
		sb.WriteString("### Containers\n")
		for _, c := range containers {
			if c.Parent == "" {
				d.describeContainerTree(&sb, c, 0)
			}
		}
		sb.WriteString("\n")
	}

	// Leaf nodes (non-container nodes at root level)
	sb.WriteString("### Nodes\n")
	for _, node := range d.RootNodes() {
		if !node.HasChildren() {
			fmt.Fprintf(&sb, "- %s", node.ID)
			if node.Label != "" && node.Label != node.ID {
				fmt.Fprintf(&sb, " (label: %q)", node.Label)
			}
			if node.Shape != ShapeRectangle {
				fmt.Fprintf(&sb, " [%s]", node.Shape)
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Edges
	if len(d.Edges) > 0 {
		sb.WriteString("### Edges\n")
		for _, edge := range d.Edges {
			arrow := "->"
			if edge.IsBidirectional() {
				arrow = "<->"
			}
			fmt.Fprintf(&sb, "- %s %s %s", edge.Source, arrow, edge.Target)
			if edge.Label != "" {
				fmt.Fprintf(&sb, ": %q", edge.Label)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// D2 code skeleton
	sb.WriteString("## Suggested D2 Code Skeleton\n\n")
	sb.WriteString("```d2\n")
	sb.WriteString(d.generateD2Skeleton(analysis))
	sb.WriteString("```\n")

	return sb.String()
}

func (d *Diagram) describeContainerTree(sb *strings.Builder, container Node, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(sb, "%s- %s", indent, container.ID)
	if container.Label != "" && container.Label != container.ID {
		fmt.Fprintf(sb, " (label: %q)", container.Label)
	}
	sb.WriteString("\n")

	// List children
	for _, childID := range container.Children {
		child := d.NodeByID(childID)
		if child == nil {
			continue
		}
		if child.HasChildren() {
			d.describeContainerTree(sb, *child, depth+1)
		} else {
			childIndent := strings.Repeat("  ", depth+1)
			fmt.Fprintf(sb, "%s- %s", childIndent, child.ID)
			if child.Shape != ShapeRectangle {
				fmt.Fprintf(sb, " [%s]", child.Shape)
			}
			sb.WriteString("\n")
		}
	}
}

func (d *Diagram) generateD2Skeleton(analysis *LayoutAnalysis) string {
	var sb strings.Builder

	// Global directives
	if analysis.GridColumns >= 2 {
		fmt.Fprintf(&sb, "grid-columns: %d\n\n", analysis.GridColumns)
	}
	if analysis.Direction != "" && analysis.Direction != "down" {
		fmt.Fprintf(&sb, "direction: %s\n\n", analysis.Direction)
	}

	// Containers
	containers := d.ContainerNodes()
	for _, c := range containers {
		if c.Parent == "" {
			d.generateContainerD2(&sb, c, 0)
			sb.WriteString("\n")
		}
	}

	// Root-level nodes
	for _, node := range d.RootNodes() {
		if !node.HasChildren() {
			d.generateNodeD2(&sb, node, 0)
		}
	}

	// Edges
	if len(d.Edges) > 0 {
		sb.WriteString("\n# Edges\n")
		for _, edge := range d.Edges {
			arrow := " -> "
			if edge.IsBidirectional() {
				arrow = " <-> "
			}
			fmt.Fprintf(&sb, "%s%s%s", edge.Source, arrow, edge.Target)
			if edge.Label != "" {
				fmt.Fprintf(&sb, ": %s", edge.Label)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (d *Diagram) generateContainerD2(sb *strings.Builder, container Node, depth int) {
	indent := strings.Repeat("  ", depth)

	// Container declaration
	fmt.Fprintf(sb, "%s%s", indent, container.ID)
	if container.Label != "" && container.Label != container.ID {
		fmt.Fprintf(sb, ": %s", container.Label)
	}
	sb.WriteString(" {\n")

	// Children
	for _, childID := range container.Children {
		child := d.NodeByID(childID)
		if child == nil {
			continue
		}
		if child.HasChildren() {
			d.generateContainerD2(sb, *child, depth+1)
		} else {
			d.generateNodeD2(sb, *child, depth+1)
		}
	}

	fmt.Fprintf(sb, "%s}\n", indent)
}

func (d *Diagram) generateNodeD2(sb *strings.Builder, node Node, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(sb, "%s%s", indent, node.ID)

	if node.Label != "" && node.Label != node.ID {
		fmt.Fprintf(sb, ": %s", node.Label)
	}

	if node.Shape != ShapeRectangle && node.Shape != ShapeUnknown {
		fmt.Fprintf(sb, " {\n%s  shape: %s\n%s}", indent, node.Shape, indent)
	}

	sb.WriteString("\n")
}

// Helper function to calculate variance
func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	var variance float64
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	return variance / float64(len(values))
}
