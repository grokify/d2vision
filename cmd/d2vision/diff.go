package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/format"
	"github.com/spf13/cobra"
)

var (
	diffFormat         string
	diffIncludeBounds  bool
	diffBoundsThreshold float64
)

// DiffResult contains the differences between two diagrams.
type DiffResult struct {
	File1 string     `json:"file1" toon:"File1"`
	File2 string     `json:"file2" toon:"File2"`
	Nodes NodesDiff  `json:"nodes" toon:"Nodes"`
	Edges EdgesDiff  `json:"edges" toon:"Edges"`
	Same  bool       `json:"same" toon:"Same"`
}

// NodesDiff contains node-level differences.
type NodesDiff struct {
	Added    []string       `json:"added,omitempty" toon:"Added"`
	Removed  []string       `json:"removed,omitempty" toon:"Removed"`
	Modified []NodeModified `json:"modified,omitempty" toon:"Modified"`
}

// NodeModified describes changes to a node.
type NodeModified struct {
	ID      string   `json:"id" toon:"ID"`
	Changes []string `json:"changes" toon:"Changes"`
}

// EdgesDiff contains edge-level differences.
type EdgesDiff struct {
	Added    []string       `json:"added,omitempty" toon:"Added"`
	Removed  []string       `json:"removed,omitempty" toon:"Removed"`
	Modified []EdgeModified `json:"modified,omitempty" toon:"Modified"`
}

// EdgeModified describes changes to an edge.
type EdgeModified struct {
	ID      string   `json:"id" toon:"ID"`
	Changes []string `json:"changes" toon:"Changes"`
}

var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2>",
	Short: "Compare two diagrams and show differences",
	Long: `Compare two D2-generated SVG files and report structural differences.

Compares:
  - Node sets (added, removed, modified)
  - Edge sets (added, removed, modified)
  - Labels and shapes
  - Styles (optional)

Examples:
  # Compare two SVG files
  d2vision diff old.svg new.svg

  # JSON output for programmatic use
  d2vision diff old.svg new.svg --format json

  # TOON output
  d2vision diff old.svg new.svg --format toon

Exit codes:
  0: Files are identical
  1: Differences found or error occurred
`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func init() {
	diffCmd.Flags().StringVarP(&diffFormat, "format", "f", "text", "Output format: text, toon, json")
	diffCmd.Flags().BoolVar(&diffIncludeBounds, "bounds", false, "Include position/bounds comparison")
	diffCmd.Flags().Float64Var(&diffBoundsThreshold, "bounds-threshold", 5.0, "Minimum position change to report (pixels)")
}

func runDiff(cmd *cobra.Command, args []string) error {
	file1, file2 := args[0], args[1]

	// Parse both files
	diagram1, err := d2vision.ParseFile(file1)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", file1, err)
	}

	diagram2, err := d2vision.ParseFile(file2)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", file2, err)
	}

	// Compute diff
	result := computeDiff(file1, file2, diagram1, diagram2)

	// Output based on format
	switch diffFormat {
	case "text":
		printTextDiff(result)
	default:
		f, err := format.Parse(diffFormat)
		if err != nil {
			return err
		}
		output, err := format.Marshal(result, f)
		if err != nil {
			return fmt.Errorf("marshaling result: %w", err)
		}
		fmt.Println(string(output))
	}

	if !result.Same {
		// Exit with status 1 but don't print error (output already shown)
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		return fmt.Errorf("differences found")
	}
	return nil
}

func computeDiff(file1, file2 string, d1, d2 *d2vision.Diagram) DiffResult {
	result := DiffResult{
		File1: file1,
		File2: file2,
		Same:  true,
	}

	// Build node maps
	nodes1 := make(map[string]*d2vision.Node)
	nodes2 := make(map[string]*d2vision.Node)

	for i := range d1.Nodes {
		nodes1[d1.Nodes[i].ID] = &d1.Nodes[i]
	}
	for i := range d2.Nodes {
		nodes2[d2.Nodes[i].ID] = &d2.Nodes[i]
	}

	// Find added and removed nodes
	for id := range nodes2 {
		if _, exists := nodes1[id]; !exists {
			result.Nodes.Added = append(result.Nodes.Added, id)
			result.Same = false
		}
	}
	for id := range nodes1 {
		if _, exists := nodes2[id]; !exists {
			result.Nodes.Removed = append(result.Nodes.Removed, id)
			result.Same = false
		}
	}

	// Find modified nodes
	for id, node1 := range nodes1 {
		if node2, exists := nodes2[id]; exists {
			changes := compareNodes(node1, node2, diffIncludeBounds, diffBoundsThreshold)
			if len(changes) > 0 {
				result.Nodes.Modified = append(result.Nodes.Modified, NodeModified{
					ID:      id,
					Changes: changes,
				})
				result.Same = false
			}
		}
	}

	// Sort for consistent output
	sort.Strings(result.Nodes.Added)
	sort.Strings(result.Nodes.Removed)
	sort.Slice(result.Nodes.Modified, func(i, j int) bool {
		return result.Nodes.Modified[i].ID < result.Nodes.Modified[j].ID
	})

	// Build edge maps
	edges1 := make(map[string]*d2vision.Edge)
	edges2 := make(map[string]*d2vision.Edge)

	for i := range d1.Edges {
		key := edgeKey(&d1.Edges[i])
		edges1[key] = &d1.Edges[i]
	}
	for i := range d2.Edges {
		key := edgeKey(&d2.Edges[i])
		edges2[key] = &d2.Edges[i]
	}

	// Find added and removed edges
	for key := range edges2 {
		if _, exists := edges1[key]; !exists {
			result.Edges.Added = append(result.Edges.Added, key)
			result.Same = false
		}
	}
	for key := range edges1 {
		if _, exists := edges2[key]; !exists {
			result.Edges.Removed = append(result.Edges.Removed, key)
			result.Same = false
		}
	}

	// Find modified edges
	for key, edge1 := range edges1 {
		if edge2, exists := edges2[key]; exists {
			changes := compareEdges(edge1, edge2)
			if len(changes) > 0 {
				result.Edges.Modified = append(result.Edges.Modified, EdgeModified{
					ID:      key,
					Changes: changes,
				})
				result.Same = false
			}
		}
	}

	// Sort for consistent output
	sort.Strings(result.Edges.Added)
	sort.Strings(result.Edges.Removed)
	sort.Slice(result.Edges.Modified, func(i, j int) bool {
		return result.Edges.Modified[i].ID < result.Edges.Modified[j].ID
	})

	return result
}

func compareNodes(n1, n2 *d2vision.Node, includeBounds bool, threshold float64) []string {
	var changes []string

	if n1.Label != n2.Label {
		changes = append(changes, fmt.Sprintf("label: %q → %q", n1.Label, n2.Label))
	}
	if n1.Shape != n2.Shape {
		changes = append(changes, fmt.Sprintf("shape: %s → %s", n1.Shape, n2.Shape))
	}
	if n1.Style.Fill != n2.Style.Fill {
		changes = append(changes, fmt.Sprintf("fill: %s → %s", n1.Style.Fill, n2.Style.Fill))
	}
	if n1.Style.Stroke != n2.Style.Stroke {
		changes = append(changes, fmt.Sprintf("stroke: %s → %s", n1.Style.Stroke, n2.Style.Stroke))
	}
	if n1.Parent != n2.Parent {
		changes = append(changes, fmt.Sprintf("parent: %s → %s", n1.Parent, n2.Parent))
	}

	// Position/bounds comparison (optional)
	if includeBounds {
		changes = append(changes, compareBounds(n1.Bounds, n2.Bounds, threshold)...)
	}

	return changes
}

func compareBounds(b1, b2 d2vision.Bounds, threshold float64) []string {
	var changes []string

	dx := math.Abs(b2.X - b1.X)
	dy := math.Abs(b2.Y - b1.Y)
	dw := math.Abs(b2.Width - b1.Width)
	dh := math.Abs(b2.Height - b1.Height)

	if dx > threshold || dy > threshold {
		changes = append(changes, fmt.Sprintf("position: (%.1f, %.1f) → (%.1f, %.1f)", b1.X, b1.Y, b2.X, b2.Y))
	}
	if dw > threshold || dh > threshold {
		changes = append(changes, fmt.Sprintf("size: %.1fx%.1f → %.1fx%.1f", b1.Width, b1.Height, b2.Width, b2.Height))
	}

	return changes
}

func compareEdges(e1, e2 *d2vision.Edge) []string {
	var changes []string

	if e1.Label != e2.Label {
		changes = append(changes, fmt.Sprintf("label: %q → %q", e1.Label, e2.Label))
	}
	if e1.SourceArrow != e2.SourceArrow {
		changes = append(changes, fmt.Sprintf("sourceArrow: %s → %s", e1.SourceArrow, e2.SourceArrow))
	}
	if e1.TargetArrow != e2.TargetArrow {
		changes = append(changes, fmt.Sprintf("targetArrow: %s → %s", e1.TargetArrow, e2.TargetArrow))
	}

	return changes
}

func edgeKey(e *d2vision.Edge) string {
	return fmt.Sprintf("%s -> %s", e.Source, e.Target)
}

func printTextDiff(result DiffResult) {
	if result.Same {
		fmt.Printf("✓ %s and %s are identical\n", result.File1, result.File2)
		return
	}

	fmt.Printf("Comparing %s → %s\n\n", result.File1, result.File2)

	// Nodes
	hasNodeChanges := len(result.Nodes.Added) > 0 || len(result.Nodes.Removed) > 0 || len(result.Nodes.Modified) > 0
	if hasNodeChanges {
		fmt.Println("Nodes:")
		for _, id := range result.Nodes.Added {
			fmt.Printf("  + %s (added)\n", id)
		}
		for _, id := range result.Nodes.Removed {
			fmt.Printf("  - %s (removed)\n", id)
		}
		for _, mod := range result.Nodes.Modified {
			fmt.Printf("  ~ %s:\n", mod.ID)
			for _, change := range mod.Changes {
				fmt.Printf("      %s\n", change)
			}
		}
		fmt.Println()
	}

	// Edges
	hasEdgeChanges := len(result.Edges.Added) > 0 || len(result.Edges.Removed) > 0 || len(result.Edges.Modified) > 0
	if hasEdgeChanges {
		fmt.Println("Edges:")
		for _, id := range result.Edges.Added {
			fmt.Printf("  + %s (added)\n", id)
		}
		for _, id := range result.Edges.Removed {
			fmt.Printf("  - %s (removed)\n", id)
		}
		for _, mod := range result.Edges.Modified {
			fmt.Printf("  ~ %s:\n", mod.ID)
			for _, change := range mod.Changes {
				fmt.Printf("      %s\n", change)
			}
		}
		fmt.Println()
	}

	// Summary
	var parts []string
	nodeCount := len(result.Nodes.Added) + len(result.Nodes.Removed) + len(result.Nodes.Modified)
	edgeCount := len(result.Edges.Added) + len(result.Edges.Removed) + len(result.Edges.Modified)

	if nodeCount > 0 {
		parts = append(parts, fmt.Sprintf("%d node change(s)", nodeCount))
	}
	if edgeCount > 0 {
		parts = append(parts, fmt.Sprintf("%d edge change(s)", edgeCount))
	}

	fmt.Printf("Summary: %s\n", strings.Join(parts, ", "))
}
