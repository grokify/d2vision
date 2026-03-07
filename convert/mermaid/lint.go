package mermaid

import (
	"strings"

	"github.com/grokify/d2vision/convert"
)

// Linter analyzes Mermaid source for D2 compatibility.
type Linter struct{}

// Lint analyzes the source code for unsupported features.
func (l *Linter) Lint(source string) *convert.LintResult {
	result := &convert.LintResult{
		Format:      convert.FormatMermaid,
		Convertible: true,
	}

	lines := strings.Split(source, "\n")

	// Detect diagram type
	diagramType := detectDiagramType(source)
	result.DiagramType = string(diagramType)

	// Check for unsupported diagram types
	switch diagramType {
	case DiagramGantt:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "gantt",
			Description: "Gantt charts are not supported in D2",
			Suggestion:  "Consider using a sequence diagram or flowchart to represent timeline",
		})
		result.Convertible = false

	case DiagramPie:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "pie",
			Description: "Pie charts are not supported in D2",
			Suggestion:  "D2 focuses on structural diagrams; consider external charting tools",
		})
		result.Convertible = false

	case DiagramGitGraph:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "gitGraph",
			Description: "Git graphs are not supported in D2",
			Suggestion:  "Consider using a flowchart to represent git branching",
		})
		result.Convertible = false

	case DiagramState:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "stateDiagram",
			Description: "State diagrams have partial support",
			Suggestion:  "States will be converted to nodes; some features may be lost",
		})
		// Still convertible, just with limitations

	case DiagramER:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "erDiagram",
			Description: "ER diagrams have partial support",
			Suggestion:  "Entities will be converted to containers; cardinality notation may be lost",
		})
		// Still convertible, just with limitations

	case DiagramFlowchart:
		result.Supported = append(result.Supported,
			"graph/flowchart declaration",
			"direction (TB, TD, BT, LR, RL)",
			"nodes with labels",
			"node shapes (rectangle, circle, diamond, cylinder, hexagon)",
			"edges with labels",
			"arrow styles (solid, dashed, thick)",
			"subgraphs",
		)

	case DiagramSequence:
		result.Supported = append(result.Supported,
			"sequenceDiagram declaration",
			"participant/actor declarations",
			"messages with labels",
			"message styles (solid, dashed, async)",
			"alt/opt/loop/par groups",
		)

	case DiagramClass:
		result.Supported = append(result.Supported,
			"classDiagram declaration",
			"class definitions",
			"attributes and methods",
			"relationships",
		)
	}

	// Line-by-line feature detection
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Check for click handlers
		if strings.HasPrefix(lower, "click ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "click",
				Description: "Click handlers are not supported in D2",
				Suggestion:  "D2 generates static SVGs; interactivity requires external tooling",
			})
		}

		// Check for callbacks
		if strings.Contains(lower, "callback ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "callback",
				Description: "Callbacks are not supported in D2",
			})
		}

		// Check for styling (partial support)
		if strings.HasPrefix(lower, "style ") || strings.HasPrefix(lower, "clasdef ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "style/classDef",
				Description: "Mermaid style definitions have limited support",
				Suggestion:  "Use D2's native style syntax instead",
			})
		}

		// Check for linkStyle
		if strings.HasPrefix(lower, "linkstyle ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "linkStyle",
				Description: "Link styling is not directly supported",
				Suggestion:  "Edge styles will use D2's default or stroke-dash for dashed",
			})
		}

		// Sequence diagram specific
		if diagramType == DiagramSequence {
			// Check for notes
			if strings.HasPrefix(lower, "note ") {
				result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
					Line:        lineNum,
					Feature:     "note",
					Description: "Notes are not supported in D2 sequence diagrams",
					Suggestion:  "Notes will be skipped during conversion",
				})
			}

			// Check for activate/deactivate
			if strings.HasPrefix(lower, "activate ") || strings.HasPrefix(lower, "deactivate ") {
				result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
					Line:        lineNum,
					Feature:     "activate/deactivate",
					Description: "Activation boxes are not supported in D2",
					Suggestion:  "Activation indicators will be skipped",
				})
			}

			// Check for autonumber
			if strings.HasPrefix(lower, "autonumber") {
				result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
					Line:        lineNum,
					Feature:     "autonumber",
					Description: "Auto-numbering is not supported in D2",
					Suggestion:  "Add numbers to message labels manually if needed",
				})
			}

			// Check for rect (highlighting)
			if strings.HasPrefix(lower, "rect ") {
				result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
					Line:        lineNum,
					Feature:     "rect",
					Description: "Background highlighting is not supported",
				})
			}
		}

		// Check for subgraph styling
		if strings.Contains(lower, ":::") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     ":::",
				Description: "Class application syntax is not supported",
				Suggestion:  "Use D2's native style syntax",
			})
		}

		// Check for FA icons
		if strings.Contains(line, "fa:") || strings.Contains(line, "fab:") || strings.Contains(line, "fas:") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "FontAwesome icons",
				Description: "FontAwesome icons are not directly supported",
				Suggestion:  "Use D2's icon syntax or SVG URLs instead",
			})
		}
	}

	return result
}

func detectDiagramType(source string) DiagramType {
	trimmed := strings.TrimSpace(source)
	lower := strings.ToLower(trimmed)

	// Skip any init directives
	if strings.HasPrefix(lower, "%%{") {
		endIdx := strings.Index(lower, "}%%")
		if endIdx != -1 {
			lower = strings.TrimSpace(lower[endIdx+3:])
		}
	}

	// Check first line
	firstLine := strings.Split(lower, "\n")[0]

	if strings.HasPrefix(firstLine, "graph ") || strings.HasPrefix(firstLine, "flowchart ") {
		return DiagramFlowchart
	}
	if strings.HasPrefix(firstLine, "sequencediagram") {
		return DiagramSequence
	}
	if strings.HasPrefix(firstLine, "classdiagram") {
		return DiagramClass
	}
	if strings.HasPrefix(firstLine, "statediagram") {
		return DiagramState
	}
	if strings.HasPrefix(firstLine, "erdiagram") {
		return DiagramER
	}
	if strings.HasPrefix(firstLine, "gantt") {
		return DiagramGantt
	}
	if strings.HasPrefix(firstLine, "pie") {
		return DiagramPie
	}
	if strings.HasPrefix(firstLine, "gitgraph") {
		return DiagramGitGraph
	}

	return DiagramUnknown
}
