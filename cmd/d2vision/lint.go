package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/grokify/d2vision/format"
	"github.com/spf13/cobra"
)

var (
	lintFormat string
)

// LintIssue represents a potential problem in a D2 file.
type LintIssue struct {
	Line       int    `json:"line" toon:"Line"`
	Severity   string `json:"severity" toon:"Severity"`
	Code       string `json:"code" toon:"Code"`
	Message    string `json:"message" toon:"Message"`
	Suggestion string `json:"suggestion,omitempty" toon:"Suggestion"`
}

// LintResult contains all issues found in a D2 file.
type LintResult struct {
	File   string      `json:"file" toon:"File"`
	Issues []LintIssue `json:"issues" toon:"Issues"`
}

var lintCmd = &cobra.Command{
	Use:   "lint <file.d2>",
	Short: "Check D2 files for common layout issues",
	Long: `Analyze D2 files and report potential layout problems before rendering.

Checks:
  - Cross-container edges that may cause alignment issues
  - Missing grid-columns for side-by-side layouts
  - Inconsistent direction settings
  - Deeply nested containers (performance warning)
  - Duplicate node definitions

Examples:
  # Lint a D2 file
  d2vision lint diagram.d2

  # JSON output for CI integration
  d2vision lint diagram.d2 --format json

Exit codes:
  0: No issues found
  1: Issues found or error occurred
`,
	Args: cobra.ExactArgs(1),
	RunE: runLint,
}

func init() {
	lintCmd.Flags().StringVarP(&lintFormat, "format", "f", "text", "Output format: text, toon, json")
}

func runLint(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	result := lintD2(filePath, string(content))

	// Output based on format
	switch lintFormat {
	case "text":
		if len(result.Issues) == 0 {
			fmt.Printf("✓ %s: no issues found\n", filePath)
			return nil
		}
		fmt.Printf("Found %d issue(s) in %s:\n\n", len(result.Issues), filePath)
		for _, issue := range result.Issues {
			var icon string
			switch issue.Severity {
			case "error":
				icon = "✘"
			case "info":
				icon = "ℹ"
			default:
				icon = "⚠"
			}
			fmt.Printf("  %s [%s] Line %d: %s\n", icon, issue.Code, issue.Line, issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("    → %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	default:
		f, err := format.Parse(lintFormat)
		if err != nil {
			return err
		}
		output, err := format.Marshal(result, f)
		if err != nil {
			return fmt.Errorf("marshaling result: %w", err)
		}
		fmt.Println(string(output))
	}

	if len(result.Issues) > 0 {
		os.Exit(1)
	}
	return nil
}

func lintD2(filePath, content string) LintResult {
	result := LintResult{
		File:   filePath,
		Issues: []LintIssue{},
	}

	lines := strings.Split(content, "\n")

	// Track state while parsing
	var (
		containers       []containerInfo
		currentContainer string
		nestingDepth     int
		maxNesting       int
		maxNestingLine   int
		hasGridColumns   bool
		nodes            = make(map[string]int) // node ID -> line number
		directions       = make(map[string]string)
	)

	// Patterns for detecting D2 constructs
	containerPattern := regexp.MustCompile(`^(\s*)(\w+):\s*(?:"[^"]*"|[^{]*)\{?\s*$`)
	nodePattern := regexp.MustCompile(`^(\s*)(\w+)(?:\s*:\s*(?:"[^"]*"|[^{\->]*))?$`)
	edgePattern := regexp.MustCompile(`(\w+(?:\.\w+)*)\s*->\s*(\w+(?:\.\w+)*)`)
	gridColumnsPattern := regexp.MustCompile(`^\s*grid-columns:\s*\d+`)
	directionPattern := regexp.MustCompile(`^\s*direction:\s*(\w+)`)
	closePattern := regexp.MustCompile(`^\s*}\s*$`)

	for i, line := range lines {
		lineNum := i + 1

		// Check for grid-columns at root level
		if nestingDepth == 0 && gridColumnsPattern.MatchString(line) {
			hasGridColumns = true
		}

		// Check for direction
		if match := directionPattern.FindStringSubmatch(line); match != nil {
			dir := match[1]
			if currentContainer == "" {
				directions["_root"] = dir
			} else {
				directions[currentContainer] = dir
			}
		}

		// Track container nesting
		if match := containerPattern.FindStringSubmatch(line); match != nil && strings.Contains(line, "{") {
			name := match[2]

			if currentContainer != "" {
				name = currentContainer + "." + name
			}
			containers = append(containers, containerInfo{name: name, line: lineNum, depth: nestingDepth})
			currentContainer = name
			nestingDepth++

			if nestingDepth > maxNesting {
				maxNesting = nestingDepth
				maxNestingLine = lineNum
			}
		}

		// Track container close
		if closePattern.MatchString(line) && nestingDepth > 0 {
			nestingDepth--
			if nestingDepth > 0 && len(containers) > 0 {
				// Find parent container
				for j := len(containers) - 1; j >= 0; j-- {
					if containers[j].depth == nestingDepth-1 {
						currentContainer = containers[j].name
						break
					}
				}
			} else {
				currentContainer = ""
			}
		}

		// Track node definitions
		if match := nodePattern.FindStringSubmatch(line); match != nil && !strings.Contains(line, "->") && !strings.Contains(line, "{") {
			nodeID := match[2]
			if currentContainer != "" {
				nodeID = currentContainer + "." + nodeID
			}
			if existingLine, exists := nodes[nodeID]; exists {
				result.Issues = append(result.Issues, LintIssue{
					Line:       lineNum,
					Severity:   "warning",
					Code:       "duplicate-node",
					Message:    fmt.Sprintf("Node '%s' was previously defined on line %d", nodeID, existingLine),
					Suggestion: "Consider consolidating node definitions",
				})
			}
			nodes[nodeID] = lineNum
		}

		// Check edges for cross-container issues
		if matches := edgePattern.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				source := match[1]
				target := match[2]

				sourceRoot := getRootContainer(source)
				targetRoot := getRootContainer(target)

				// Cross-container edge
				if sourceRoot != targetRoot && sourceRoot != "" && targetRoot != "" {
					if !hasGridColumns {
						result.Issues = append(result.Issues, LintIssue{
							Line:       lineNum,
							Severity:   "warning",
							Code:       "cross-container-edge",
							Message:    fmt.Sprintf("Cross-container edge '%s -> %s' may cause vertical stacking", source, target),
							Suggestion: "Add 'grid-columns: N' at root level to control horizontal layout",
						})
					}
				}
			}
		}
	}

	// Check for deep nesting
	if maxNesting > 3 {
		result.Issues = append(result.Issues, LintIssue{
			Line:       maxNestingLine,
			Severity:   "info",
			Code:       "deep-nesting",
			Message:    fmt.Sprintf("Container nesting depth of %d may impact layout performance", maxNesting),
			Suggestion: "Consider flattening the structure if possible",
		})
	}

	// Check for multiple root containers without grid-columns
	rootContainers := 0
	for _, c := range containers {
		if c.depth == 0 {
			rootContainers++
		}
	}
	if rootContainers > 1 && !hasGridColumns {
		result.Issues = append(result.Issues, LintIssue{
			Line:       1,
			Severity:   "info",
			Code:       "missing-grid",
			Message:    fmt.Sprintf("Found %d root-level containers without grid-columns", rootContainers),
			Suggestion: "Add 'grid-columns: N' to control horizontal arrangement",
		})
	}

	// Check for inconsistent directions
	directionCounts := make(map[string]int)
	for _, dir := range directions {
		directionCounts[dir]++
	}
	if len(directionCounts) > 1 && len(directions) > 2 {
		// Build warning message
		var parts []string
		for dir, count := range directionCounts {
			parts = append(parts, fmt.Sprintf("%s: %d", dir, count))
		}
		result.Issues = append(result.Issues, LintIssue{
			Line:       1,
			Severity:   "info",
			Code:       "mixed-directions",
			Message:    fmt.Sprintf("Mixed direction settings found: %s", strings.Join(parts, ", ")),
			Suggestion: "Consider using consistent directions for cleaner layout",
		})
	}

	return result
}

type containerInfo struct {
	name  string
	line  int
	depth int
}

// getRootContainer extracts the root container from a dotted path.
func getRootContainer(nodeID string) string {
	parts := strings.Split(nodeID, ".")
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}

