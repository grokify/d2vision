package main

import (
	"fmt"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/format"
	"github.com/grokify/d2vision/generate"
	"github.com/spf13/cobra"
)

var (
	learnFormat   string
	learnOutputD2 bool
)

var learnCmd = &cobra.Command{
	Use:   "learn <file.svg>",
	Short: "Reverse engineer D2 code from an SVG diagram",
	Long: `Analyze a D2-generated SVG and output D2 code or a spec that recreates it.

This command helps AI assistants:
  - Understand existing diagram patterns
  - Learn from examples
  - Modify existing diagrams

Output modes:
  - Default: TOON spec (can be modified and piped to 'generate')
  - With --d2: Ready-to-render D2 code

Examples:
  # Get spec from existing diagram
  d2vision learn diagram.svg

  # Get D2 code directly
  d2vision learn diagram.svg --d2

  # Round-trip: modify and regenerate
  d2vision learn diagram.svg > spec.toon
  # ... edit spec.toon ...
  d2vision generate spec.toon | d2 - new_diagram.svg

  # JSON output for programmatic use
  d2vision learn diagram.svg --format json
`,
	Args: cobra.ExactArgs(1),
	RunE: runLearn,
}

func init() {
	learnCmd.Flags().StringVarP(&learnFormat, "format", "f", "toon", "Output format: toon, json, yaml")
	learnCmd.Flags().BoolVar(&learnOutputD2, "d2", false, "Output D2 code instead of spec")
}

func runLearn(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Parse the SVG
	diagram, err := d2vision.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filePath, err)
	}

	// Convert Diagram to DiagramSpec
	spec := diagramToSpec(diagram)

	// Output D2 code directly
	if learnOutputD2 {
		gen := generate.NewGenerator()
		fmt.Print(gen.Generate(spec))
		return nil
	}

	// Output spec in requested format
	f, err := format.Parse(learnFormat)
	if err != nil {
		return err
	}

	output, err := format.Marshal(spec, f)
	if err != nil {
		return fmt.Errorf("marshaling spec: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// diagramToSpec converts a parsed Diagram to a DiagramSpec for generation.
func diagramToSpec(d *d2vision.Diagram) *generate.DiagramSpec {
	spec := &generate.DiagramSpec{}

	// Build a map of node IDs to their data for quick lookup
	nodeMap := make(map[string]*d2vision.Node)
	for i := range d.Nodes {
		nodeMap[d.Nodes[i].ID] = &d.Nodes[i]
	}

	// Analyze layout to determine grid settings
	spec.GridColumns, spec.GridRows = detectGridLayout(d)

	// Process root-level containers and nodes
	rootNodes := d.RootNodes()

	for _, node := range rootNodes {
		if len(node.Children) > 0 {
			// It's a container
			container := nodeToContainerSpec(&node, nodeMap, d)
			spec.Containers = append(spec.Containers, container)
		} else {
			// It's a simple node
			spec.Nodes = append(spec.Nodes, nodeToNodeSpec(&node))
		}
	}

	// Process root-level edges (cross-container edges)
	for _, edge := range d.Edges {
		// Check if this is a cross-container edge
		sourceNode := nodeMap[edge.Source]
		targetNode := nodeMap[edge.Target]

		if sourceNode == nil || targetNode == nil {
			continue
		}

		// Cross-container if different top-level parents or both are root
		sourceRoot := getRootParent(edge.Source, nodeMap)
		targetRoot := getRootParent(edge.Target, nodeMap)

		if sourceRoot != targetRoot || (sourceNode.Parent == "" && targetNode.Parent == "") {
			spec.Edges = append(spec.Edges, edgeToEdgeSpec(&edge))
		}
	}

	return spec
}

// nodeToContainerSpec converts a container node to ContainerSpec.
func nodeToContainerSpec(node *d2vision.Node, nodeMap map[string]*d2vision.Node, d *d2vision.Diagram) generate.ContainerSpec {
	container := generate.ContainerSpec{
		ID:    node.ID,
		Label: node.Label,
	}

	// Detect direction from child layout
	container.Direction = detectContainerDirection(node, nodeMap)

	// Process children
	for _, childID := range node.Children {
		childNode := nodeMap[childID]
		if childNode == nil {
			continue
		}

		if len(childNode.Children) > 0 {
			// Nested container
			nestedContainer := nodeToContainerSpec(childNode, nodeMap, d)
			// Use relative ID for nested containers
			nestedContainer.ID = d2vision.ExtractBaseName(childNode.ID)
			container.Containers = append(container.Containers, nestedContainer)
		} else {
			// Simple node
			nodeSpec := nodeToNodeSpec(childNode)
			// Use relative ID
			nodeSpec.ID = d2vision.ExtractBaseName(childNode.ID)
			container.Nodes = append(container.Nodes, nodeSpec)
		}
	}

	// Process edges within this container
	for _, edge := range d.Edges {
		// Check if edge is within this container
		if isEdgeWithinContainer(edge, node.ID, nodeMap) {
			edgeSpec := edgeToEdgeSpec(&edge)
			// Make edge IDs relative to container
			edgeSpec.From = makeRelativeID(edge.Source, node.ID)
			edgeSpec.To = makeRelativeID(edge.Target, node.ID)
			container.Edges = append(container.Edges, edgeSpec)
		}
	}

	return container
}

// nodeToNodeSpec converts a Node to NodeSpec.
func nodeToNodeSpec(node *d2vision.Node) generate.NodeSpec {
	spec := generate.NodeSpec{
		ID:    node.ID,
		Label: node.Label,
	}

	// Only set shape if not default rectangle
	if node.Shape != "" && node.Shape != d2vision.ShapeRectangle {
		spec.Shape = string(node.Shape)
	}

	// Extract style if present
	if node.Style.Fill != "" || node.Style.Stroke != "" {
		spec.Style.Fill = node.Style.Fill
		spec.Style.Stroke = node.Style.Stroke
	}

	return spec
}

// edgeToEdgeSpec converts an Edge to EdgeSpec.
func edgeToEdgeSpec(edge *d2vision.Edge) generate.EdgeSpec {
	return generate.EdgeSpec{
		From:  edge.Source,
		To:    edge.Target,
		Label: edge.Label,
	}
}

// detectGridLayout analyzes the diagram to detect grid settings.
func detectGridLayout(d *d2vision.Diagram) (columns, rows int) {
	rootNodes := d.RootNodes()
	if len(rootNodes) < 2 {
		return 0, 0
	}

	// Check if root nodes are horizontally aligned (similar Y values)
	var yValues []float64
	for _, node := range rootNodes {
		if node.Bounds.Height > 0 {
			yValues = append(yValues, node.Bounds.Y)
		}
	}

	if len(yValues) < 2 {
		return 0, 0
	}

	// If Y values are similar (within 20px), assume horizontal grid
	allSimilarY := true
	firstY := yValues[0]
	for _, y := range yValues[1:] {
		if abs(y-firstY) > 20 {
			allSimilarY = false
			break
		}
	}

	if allSimilarY {
		return len(rootNodes), 0
	}

	return 0, 0
}

// detectContainerDirection analyzes child positions to detect direction.
func detectContainerDirection(node *d2vision.Node, nodeMap map[string]*d2vision.Node) string {
	if len(node.Children) < 2 {
		return ""
	}

	var xValues, yValues []float64
	for _, childID := range node.Children {
		child := nodeMap[childID]
		if child != nil && child.Bounds.Width > 0 {
			xValues = append(xValues, child.Bounds.X)
			yValues = append(yValues, child.Bounds.Y)
		}
	}

	if len(xValues) < 2 {
		return ""
	}

	// Calculate variance in X and Y
	xVariance := variance(xValues)
	yVariance := variance(yValues)

	// If X varies more than Y, children are horizontal → direction: right
	// If Y varies more than X, children are vertical → direction: down
	if xVariance > yVariance*2 {
		return "right"
	} else if yVariance > xVariance*2 {
		return "down"
	}

	return "down" // Default
}

// isEdgeWithinContainer checks if an edge is contained within a container.
func isEdgeWithinContainer(edge d2vision.Edge, containerID string, nodeMap map[string]*d2vision.Node) bool {
	sourceNode := nodeMap[edge.Source]
	targetNode := nodeMap[edge.Target]

	if sourceNode == nil || targetNode == nil {
		return false
	}

	// Check if both source and target are descendants of the container
	return isDescendantOf(edge.Source, containerID) && isDescendantOf(edge.Target, containerID)
}

// isDescendantOf checks if nodeID is a descendant of containerID.
func isDescendantOf(nodeID, containerID string) bool {
	// Simple prefix check: "container.child" is descendant of "container"
	if len(nodeID) <= len(containerID) {
		return false
	}
	return nodeID[:len(containerID)] == containerID && nodeID[len(containerID)] == '.'
}

// getRootParent returns the top-level parent ID for a node.
func getRootParent(nodeID string, nodeMap map[string]*d2vision.Node) string {
	node := nodeMap[nodeID]
	if node == nil || node.Parent == "" {
		return nodeID
	}

	// Walk up to root
	current := node.Parent
	for {
		parent := nodeMap[current]
		if parent == nil || parent.Parent == "" {
			return current
		}
		current = parent.Parent
	}
}

// makeRelativeID removes the container prefix from an ID.
func makeRelativeID(fullID, containerID string) string {
	if len(fullID) > len(containerID)+1 && fullID[:len(containerID)] == containerID {
		return fullID[len(containerID)+1:]
	}
	return fullID
}

// Helper functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	var sqDiff float64
	for _, v := range values {
		diff := v - mean
		sqDiff += diff * diff
	}

	return sqDiff / float64(len(values))
}
