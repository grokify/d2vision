package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/format"
	"github.com/spf13/cobra"
)

var (
	reviewFormat string
)

// ReviewResult contains the comprehensive diagram review.
type ReviewResult struct {
	File            string            `json:"file" toon:"File"`
	Overall         string            `json:"overall" toon:"Overall"`
	Dimensions      ReviewDimensions  `json:"dimensions" toon:"Dimensions"`
	Metrics         ReviewMetrics     `json:"metrics" toon:"Metrics"`
	Issues          []ReviewIssue     `json:"issues" toon:"Issues"`
	Recommendations []Recommendation  `json:"recommendations" toon:"Recommendations"`
}

// ReviewDimensions contains status for each review category.
type ReviewDimensions struct {
	Layout     DimensionStatus `json:"layout" toon:"Layout"`
	Whitespace DimensionStatus `json:"whitespace" toon:"Whitespace"`
	Legend     DimensionStatus `json:"legend" toon:"Legend"`
	Hierarchy  DimensionStatus `json:"hierarchy" toon:"Hierarchy"`
}

// DimensionStatus contains the status and details for a dimension.
type DimensionStatus struct {
	Status string   `json:"status" toon:"Status"`
	Issues []string `json:"issues,omitempty" toon:"Issues"`
}

// ReviewMetrics contains numeric measurements.
type ReviewMetrics struct {
	Width           float64 `json:"width" toon:"Width"`
	Height          float64 `json:"height" toon:"Height"`
	AspectRatio     float64 `json:"aspect_ratio" toon:"AspectRatio"`
	WhitespaceRatio float64 `json:"whitespace_ratio" toon:"WhitespaceRatio"`
	NodeCount       int     `json:"node_count" toon:"NodeCount"`
	EdgeCount       int     `json:"edge_count" toon:"EdgeCount"`
	ContainerCount  int     `json:"container_count" toon:"ContainerCount"`
	NestingDepth    int     `json:"nesting_depth" toon:"NestingDepth"`
}

// ReviewIssue represents a specific problem found.
type ReviewIssue struct {
	Severity   string `json:"severity" toon:"Severity"`
	Category   string `json:"category" toon:"Category"`
	Message    string `json:"message" toon:"Message"`
	Location   string `json:"location,omitempty" toon:"Location"`
	Suggestion string `json:"suggestion,omitempty" toon:"Suggestion"`
}

// Recommendation is a prioritized action item.
type Recommendation struct {
	Priority string `json:"priority" toon:"Priority"`
	Category string `json:"category" toon:"Category"`
	Action   string `json:"action" toon:"Action"`
	Impact   string `json:"impact,omitempty" toon:"Impact"`
}

// ColorUsage tracks where a color is used.
type ColorUsage struct {
	Color    string `json:"color" toon:"Color"`
	Section  string `json:"section" toon:"Section"`
	Item     string `json:"item" toon:"Item"`
	Property string `json:"property" toon:"Property"`
}

var reviewCmd = &cobra.Command{
	Use:   "review <file>",
	Short: "Comprehensive diagram quality review",
	Long: `Review a D2 diagram for visual quality issues including layout,
whitespace, legend clarity, and color conflicts.

Analyzes:
  - Layout: aspect ratio, direction consistency, grid usage
  - Whitespace: excessive empty space, margin balance
  - Legend: color uniqueness, element mapping clarity
  - Hierarchy: nesting depth, shape semantics

Supports both D2 source files and rendered SVGs.

Examples:
  # Review an SVG
  d2vision review diagram.svg

  # Review D2 source
  d2vision review diagram.d2

  # JSON output
  d2vision review diagram.svg --format json
`,
	Args: cobra.ExactArgs(1),
	RunE: runReview,
}

func init() {
	reviewCmd.Flags().StringVarP(&reviewFormat, "format", "f", "text", "Output format: text, json, toon, markdown")
}

func runReview(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	var result ReviewResult
	result.File = filePath

	// Determine file type and analyze accordingly
	if strings.HasSuffix(filePath, ".svg") {
		// Parse SVG for metrics
		diagram, err := d2vision.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("parsing SVG: %w", err)
		}
		result.Metrics = extractMetricsFromDiagram(diagram)
	} else if strings.HasSuffix(filePath, ".d2") {
		// Parse D2 source for legend analysis
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		// Try to find corresponding SVG for metrics
		svgPath := strings.TrimSuffix(filePath, ".d2") + ".svg"
		if _, err := os.Stat(svgPath); err == nil {
			diagram, err := d2vision.ParseFile(svgPath)
			if err == nil {
				result.Metrics = extractMetricsFromDiagram(diagram)
			}
		}

		// Analyze D2 source
		analyzeD2Source(&result, string(content))
	} else {
		return fmt.Errorf("unsupported file type: %s (use .d2 or .svg)", filePath)
	}

	// Evaluate dimensions
	evaluateDimensions(&result)

	// Calculate overall status
	result.Overall = calculateOverall(result.Dimensions)

	// Generate recommendations
	generateRecommendations(&result)

	// Output
	return outputReview(result)
}

func extractMetricsFromDiagram(diagram *d2vision.Diagram) ReviewMetrics {
	metrics := ReviewMetrics{
		Width:          diagram.ViewBox.Width,
		Height:         diagram.ViewBox.Height,
		NodeCount:      len(diagram.Nodes),
		EdgeCount:      len(diagram.Edges),
		ContainerCount: len(diagram.ContainerNodes()),
	}

	if metrics.Height > 0 {
		metrics.AspectRatio = metrics.Width / metrics.Height
	}

	// Calculate whitespace ratio
	totalArea := metrics.Width * metrics.Height
	contentArea := 0.0
	for _, node := range diagram.Nodes {
		contentArea += node.Bounds.Width * node.Bounds.Height
	}
	if totalArea > 0 {
		metrics.WhitespaceRatio = (totalArea - contentArea) / totalArea
	}

	// Calculate nesting depth
	maxDepth := 0
	for _, node := range diagram.Nodes {
		depth := strings.Count(node.ID, ".")
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	metrics.NestingDepth = maxDepth

	return metrics
}

func analyzeD2Source(result *ReviewResult, content string) {
	// Check for legend and color conflicts
	colorUsages := extractLegendColors(content)
	checkColorConflicts(result, colorUsages)

	// Check for legend positioning
	checkLegendPositioning(result, content)

	// Check for element mapping
	checkElementMapping(result, content)
}

func extractLegendColors(content string) []ColorUsage {
	var usages []ColorUsage

	// Pattern to find style definitions in legend
	// Matches: style.fill: "#color" or style.stroke: "#color"
	fillPattern := regexp.MustCompile(`(\w+):\s*\w+[^}]*style\.fill:\s*["']?(#[0-9a-fA-F]{6})["']?`)
	strokePattern := regexp.MustCompile(`(\w+):\s*\w+[^}]*style\.stroke:\s*["']?(#[0-9a-fA-F]{6})["']?`)

	// Track current section (legend container)
	sectionPattern := regexp.MustCompile(`(\w+):\s*(?:"[^"]*"|[\w\s\(\)]+)\s*\{`)
	lines := strings.Split(content, "\n")

	var currentSection string
	inLegend := false
	braceDepth := 0

	for _, line := range lines {
		// Check for legend start
		if strings.Contains(strings.ToLower(line), "legend") && strings.Contains(line, "{") {
			inLegend = true
			braceDepth = 1
			continue
		}

		if !inLegend {
			continue
		}

		// Track brace depth
		braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
		if braceDepth <= 0 {
			inLegend = false
			continue
		}

		// Check for section
		if match := sectionPattern.FindStringSubmatch(line); match != nil {
			currentSection = match[1]
		}

		// Extract fill colors
		if match := fillPattern.FindStringSubmatch(line); match != nil {
			usages = append(usages, ColorUsage{
				Color:    strings.ToLower(match[2]),
				Section:  currentSection,
				Item:     match[1],
				Property: "fill",
			})
		}

		// Extract stroke colors
		if match := strokePattern.FindStringSubmatch(line); match != nil {
			usages = append(usages, ColorUsage{
				Color:    strings.ToLower(match[2]),
				Section:  currentSection,
				Item:     match[1],
				Property: "stroke",
			})
		}
	}

	return usages
}

func checkColorConflicts(result *ReviewResult, usages []ColorUsage) {
	// Group by color
	colorMap := make(map[string][]ColorUsage)
	for _, usage := range usages {
		colorMap[usage.Color] = append(colorMap[usage.Color], usage)
	}

	// Check for conflicts (same color in different sections)
	for color, uses := range colorMap {
		if len(uses) <= 1 {
			continue
		}

		// Check if uses span multiple sections
		sections := make(map[string]bool)
		for _, use := range uses {
			sections[use.Section] = true
		}

		if len(sections) > 1 {
			var locations []string
			for _, use := range uses {
				locations = append(locations, fmt.Sprintf("%s.%s", use.Section, use.Item))
			}

			result.Issues = append(result.Issues, ReviewIssue{
				Severity:   "error",
				Category:   "legend",
				Message:    fmt.Sprintf("Color %s used in multiple sections: %s", color, strings.Join(locations, ", ")),
				Suggestion: "Use distinct color families for each legend section",
			})
		}
	}

	// Check for similar colors (within same color family)
	checkSimilarColors(result, usages)
}

func checkSimilarColors(result *ReviewResult, usages []ColorUsage) {
	// Group by color family
	families := make(map[string][]ColorUsage)

	for _, usage := range usages {
		family := getColorFamily(usage.Color)
		families[family] = append(families[family], usage)
	}

	// Check each family for cross-section usage
	for family, uses := range families {
		if len(uses) <= 1 {
			continue
		}

		sections := make(map[string]bool)
		for _, use := range uses {
			sections[use.Section] = true
		}

		if len(sections) > 1 && family != "neutral" {
			result.Issues = append(result.Issues, ReviewIssue{
				Severity:   "warning",
				Category:   "legend",
				Message:    fmt.Sprintf("%s color family used across multiple sections", family),
				Suggestion: "Consider using distinct color families per section for clarity",
			})
		}
	}
}

func getColorFamily(hexColor string) string {
	// Parse hex color
	hex := strings.TrimPrefix(hexColor, "#")
	if len(hex) != 6 {
		return "unknown"
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	// Determine family based on dominant channel
	max := maxInt64(r, maxInt64(g, b))
	min := minInt64(r, minInt64(g, b))

	// Check for neutral (gray/white/black)
	if max-min < 30 {
		return "neutral"
	}

	// Determine hue
	if r >= g && r >= b {
		if g > b {
			return "orange"
		}
		return "red"
	}
	if g >= r && g >= b {
		if b > r {
			return "teal"
		}
		return "green"
	}
	if r > g {
		return "purple"
	}
	return "blue"
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func checkLegendPositioning(result *ReviewResult, content string) {
	// Check if legend has 'near:' positioning
	hasNear := regexp.MustCompile(`legend[^}]*near:\s*\w+`).MatchString(content)

	if !hasNear && strings.Contains(content, "legend:") {
		result.Issues = append(result.Issues, ReviewIssue{
			Severity:   "warning",
			Category:   "layout",
			Message:    "Legend container missing 'near:' positioning",
			Suggestion: "Add 'near: bottom-center' to position legend without displacing main content",
		})
	}
}

func checkElementMapping(result *ReviewResult, content string) {
	// Check if legend labels describe what they color
	legendPattern := regexp.MustCompile(`(\w+):\s*["']?([^"'\n{]+)["']?\s*\{`)

	matches := legendPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		label := match[2]
		labelLower := strings.ToLower(label)

		// Skip if label already describes element type
		if strings.Contains(labelLower, "arrow") ||
			strings.Contains(labelLower, "box") ||
			strings.Contains(labelLower, "fill") ||
			strings.Contains(labelLower, "outline") ||
			strings.Contains(labelLower, "stroke") ||
			strings.Contains(labelLower, "data store") {
			continue
		}

		// Check if this is a legend section
		if strings.Contains(labelLower, "stride") ||
			strings.Contains(labelLower, "att&ck") ||
			strings.Contains(labelLower, "attack") ||
			strings.Contains(labelLower, "mitre") ||
			strings.Contains(labelLower, "asset") {
			result.Issues = append(result.Issues, ReviewIssue{
				Severity:   "info",
				Category:   "legend",
				Message:    fmt.Sprintf("Legend section '%s' doesn't describe what diagram elements it colors", label),
				Location:   match[1],
				Suggestion: "Add clarification like '(arrow color)' or '(box fill)' to the label",
			})
		}
	}
}

func evaluateDimensions(result *ReviewResult) {
	// Layout evaluation
	layoutIssues := []string{}
	layoutStatus := "pass"

	if result.Metrics.AspectRatio > 0 && result.Metrics.AspectRatio < 1.2 {
		layoutIssues = append(layoutIssues, fmt.Sprintf("Aspect ratio %.2f:1 suggests layout issues for horizontal flow", result.Metrics.AspectRatio))
		layoutStatus = "warning"
	}

	// Check for layout-related issues
	for _, issue := range result.Issues {
		if issue.Category == "layout" {
			if issue.Severity == "error" {
				layoutStatus = "fail"
			} else if layoutStatus != "fail" && issue.Severity == "warning" {
				layoutStatus = "warning"
			}
			layoutIssues = append(layoutIssues, issue.Message)
		}
	}

	result.Dimensions.Layout = DimensionStatus{Status: layoutStatus, Issues: layoutIssues}

	// Whitespace evaluation
	whitespaceIssues := []string{}
	whitespaceStatus := "pass"

	if result.Metrics.WhitespaceRatio > 0.5 {
		whitespaceIssues = append(whitespaceIssues, fmt.Sprintf("Whitespace ratio %.0f%% exceeds 50%% threshold", result.Metrics.WhitespaceRatio*100))
		whitespaceStatus = "fail"
	} else if result.Metrics.WhitespaceRatio > 0.4 {
		whitespaceIssues = append(whitespaceIssues, fmt.Sprintf("Whitespace ratio %.0f%% exceeds 40%% warning threshold", result.Metrics.WhitespaceRatio*100))
		whitespaceStatus = "warning"
	}

	result.Dimensions.Whitespace = DimensionStatus{Status: whitespaceStatus, Issues: whitespaceIssues}

	// Legend evaluation
	legendIssues := []string{}
	legendStatus := "pass"

	for _, issue := range result.Issues {
		if issue.Category == "legend" {
			if issue.Severity == "error" {
				legendStatus = "fail"
			} else if legendStatus != "fail" && issue.Severity == "warning" {
				legendStatus = "warning"
			}
			legendIssues = append(legendIssues, issue.Message)
		}
	}

	result.Dimensions.Legend = DimensionStatus{Status: legendStatus, Issues: legendIssues}

	// Hierarchy evaluation
	hierarchyIssues := []string{}
	hierarchyStatus := "pass"

	if result.Metrics.NestingDepth > 3 {
		hierarchyIssues = append(hierarchyIssues, fmt.Sprintf("Nesting depth %d exceeds recommended maximum of 3", result.Metrics.NestingDepth))
		hierarchyStatus = "warning"
	}

	result.Dimensions.Hierarchy = DimensionStatus{Status: hierarchyStatus, Issues: hierarchyIssues}
}

func calculateOverall(dims ReviewDimensions) string {
	if dims.Layout.Status == "fail" || dims.Whitespace.Status == "fail" ||
		dims.Legend.Status == "fail" || dims.Hierarchy.Status == "fail" {
		return "needs_work"
	}
	if dims.Layout.Status == "warning" || dims.Whitespace.Status == "warning" ||
		dims.Legend.Status == "warning" || dims.Hierarchy.Status == "warning" {
		return "acceptable"
	}
	return "good"
}

func generateRecommendations(result *ReviewResult) {
	// Layout recommendations
	if result.Metrics.AspectRatio > 0 && result.Metrics.AspectRatio < 1.5 {
		result.Recommendations = append(result.Recommendations, Recommendation{
			Priority: "high",
			Category: "layout",
			Action:   "Add 'near: bottom-center' to legend container",
			Impact:   "Reduce height by 30-50%",
		})
	}

	// Legend recommendations
	for _, issue := range result.Issues {
		if issue.Category == "legend" && issue.Severity == "error" {
			result.Recommendations = append(result.Recommendations, Recommendation{
				Priority: "high",
				Category: "legend",
				Action:   issue.Suggestion,
				Impact:   "Improve color clarity",
			})
		}
	}

	// Whitespace recommendations
	if result.Metrics.WhitespaceRatio > 0.4 {
		result.Recommendations = append(result.Recommendations, Recommendation{
			Priority: "medium",
			Category: "whitespace",
			Action:   "Add 'grid-columns' to containers with multiple children",
			Impact:   fmt.Sprintf("Reduce whitespace from %.0f%% to <40%%", result.Metrics.WhitespaceRatio*100),
		})
	}
}

func outputReview(result ReviewResult) error {
	switch reviewFormat {
	case "text":
		return outputReviewText(result)
	case "markdown":
		return outputReviewMarkdown(result)
	default:
		f, err := format.Parse(reviewFormat)
		if err != nil {
			return err
		}
		output, err := format.Marshal(result, f)
		if err != nil {
			return fmt.Errorf("marshaling result: %w", err)
		}
		fmt.Println(string(output))
	}
	return nil
}

func outputReviewText(result ReviewResult) error {
	fmt.Printf("Diagram Review: %s\n", result.File)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()

	// Overall
	overallIcon := "✓"
	if result.Overall == "needs_work" {
		overallIcon = "✘"
	} else if result.Overall == "acceptable" {
		overallIcon = "⚠"
	}
	fmt.Printf("Overall: %s %s\n\n", overallIcon, strings.ToUpper(result.Overall))

	// Dimensions
	printDimension("Layout", result.Dimensions.Layout)
	printDimension("Whitespace", result.Dimensions.Whitespace)
	printDimension("Legend", result.Dimensions.Legend)
	printDimension("Hierarchy", result.Dimensions.Hierarchy)

	// Metrics
	if result.Metrics.Width > 0 {
		fmt.Println("\nMetrics:")
		fmt.Printf("  Size: %.0f x %.0f (aspect ratio: %.2f:1)\n",
			result.Metrics.Width, result.Metrics.Height, result.Metrics.AspectRatio)
		fmt.Printf("  Whitespace: %.0f%%\n", result.Metrics.WhitespaceRatio*100)
		fmt.Printf("  Nodes: %d, Edges: %d, Containers: %d\n",
			result.Metrics.NodeCount, result.Metrics.EdgeCount, result.Metrics.ContainerCount)
	}

	// Recommendations
	if len(result.Recommendations) > 0 {
		fmt.Println("\nRecommendations:")
		for i, rec := range result.Recommendations {
			fmt.Printf("  %d. [%s] %s\n", i+1, strings.ToUpper(rec.Priority), rec.Action)
			if rec.Impact != "" {
				fmt.Printf("     Impact: %s\n", rec.Impact)
			}
		}
	}

	fmt.Println()
	return nil
}

func printDimension(name string, dim DimensionStatus) {
	icon := "✓"
	if dim.Status == "fail" {
		icon = "✘"
	} else if dim.Status == "warning" {
		icon = "⚠"
	}

	fmt.Printf("%s %s: %s\n", icon, name, strings.ToUpper(dim.Status))
	for _, issue := range dim.Issues {
		fmt.Printf("  - %s\n", issue)
	}
}

func outputReviewMarkdown(result ReviewResult) error {
	fmt.Printf("# Diagram Review: %s\n\n", result.File)

	// Summary table
	fmt.Print("## Summary\n\n")
	fmt.Println("| Dimension | Status |")
	fmt.Println("|-----------|--------|")
	fmt.Printf("| Layout | %s |\n", statusEmoji(result.Dimensions.Layout.Status))
	fmt.Printf("| Whitespace | %s |\n", statusEmoji(result.Dimensions.Whitespace.Status))
	fmt.Printf("| Legend | %s |\n", statusEmoji(result.Dimensions.Legend.Status))
	fmt.Printf("| Hierarchy | %s |\n", statusEmoji(result.Dimensions.Hierarchy.Status))
	fmt.Println()

	// Issues
	if len(result.Issues) > 0 {
		fmt.Print("## Issues\n\n")
		for _, issue := range result.Issues {
			fmt.Printf("- **[%s]** %s\n", strings.ToUpper(issue.Severity), issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("  - Suggestion: %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	}

	// Recommendations
	if len(result.Recommendations) > 0 {
		fmt.Print("## Recommendations\n\n")
		for i, rec := range result.Recommendations {
			fmt.Printf("%d. **[%s]** %s\n", i+1, strings.ToUpper(rec.Priority), rec.Action)
		}
		fmt.Println()
	}

	// Metrics
	if result.Metrics.Width > 0 {
		fmt.Print("## Metrics\n\n")
		fmt.Println("| Metric | Value |")
		fmt.Println("|--------|-------|")
		fmt.Printf("| Dimensions | %.0f x %.0f |\n", result.Metrics.Width, result.Metrics.Height)
		fmt.Printf("| Aspect Ratio | %.2f:1 |\n", result.Metrics.AspectRatio)
		fmt.Printf("| Whitespace | %.0f%% |\n", result.Metrics.WhitespaceRatio*100)
		fmt.Printf("| Nodes | %d |\n", result.Metrics.NodeCount)
		fmt.Printf("| Edges | %d |\n", result.Metrics.EdgeCount)
		fmt.Println()
	}

	return nil
}

func statusEmoji(status string) string {
	switch status {
	case "pass":
		return "PASS"
	case "warning":
		return "WARNING"
	case "fail":
		return "FAIL"
	default:
		return status
	}
}

// round rounds a float to n decimal places.
func round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
