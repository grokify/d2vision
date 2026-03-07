package plantuml

import (
	"strings"

	"github.com/grokify/d2vision/convert"
)

// Linter analyzes PlantUML source for D2 compatibility.
type Linter struct{}

// Lint analyzes the source code for unsupported features.
func (l *Linter) Lint(source string) *convert.LintResult {
	result := &convert.LintResult{
		Format:      convert.FormatPlantUML,
		Convertible: true,
	}

	lines := strings.Split(source, "\n")

	// Detect diagram type
	diagramType := detectDiagramType(source)
	result.DiagramType = string(diagramType)

	// Check for unsupported diagram types
	switch diagramType {
	case DiagramActivity:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "activity",
			Description: "Activity diagrams have limited support in D2",
			Suggestion:  "Consider converting to a flowchart manually",
		})
		// Still convertible with limitations

	case DiagramUseCase:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "usecase",
			Description: "Use case diagrams have partial support",
			Suggestion:  "Use cases will be converted to simple nodes",
		})

	case DiagramState:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "state",
			Description: "State diagrams have partial support",
			Suggestion:  "States will be converted to nodes; some features may be lost",
		})

	case DiagramObject:
		result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
			Line:        1,
			Feature:     "object",
			Description: "Object diagrams have partial support",
			Suggestion:  "Objects will be converted to containers",
		})

	case DiagramSequence:
		result.Supported = append(result.Supported,
			"participant/actor declarations",
			"messages with labels",
			"message styles (solid, dashed, async)",
			"alt/opt/loop/par groups",
			"participant types (database, boundary, control, entity)",
		)

	case DiagramClass:
		result.Supported = append(result.Supported,
			"class definitions",
			"interface/abstract modifiers",
			"attributes and methods",
			"inheritance (--|>)",
			"composition (--*)",
			"aggregation (--o)",
			"dependency (..>)",
			"packages/namespaces",
		)

	case DiagramComponent:
		result.Supported = append(result.Supported,
			"component declarations ([Name] and component keyword)",
			"package/node/folder/frame containers",
			"relations between components",
			"stereotypes",
		)
	}

	// Line-by-line feature detection
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Check for notes
		if strings.HasPrefix(lower, "note ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "note",
				Description: "Notes are not supported in D2",
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

		// Check for skinparam
		if strings.HasPrefix(lower, "skinparam ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "skinparam",
				Description: "Skin parameters are not supported",
				Suggestion:  "Use D2's native style syntax instead",
			})
		}

		// Check for colors in arrows
		if strings.Contains(line, "[#") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "arrow color",
				Description: "Arrow colors are not directly supported",
				Suggestion:  "Use D2's edge style.stroke for coloring",
			})
		}

		// Check for create/destroy
		if strings.HasPrefix(lower, "create ") || strings.HasPrefix(lower, "destroy ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "create/destroy",
				Description: "Dynamic creation/destruction is not supported",
			})
		}

		// Check for ref over
		if strings.HasPrefix(lower, "ref over ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "ref over",
				Description: "Reference boxes are not supported",
			})
		}

		// Check for divider
		if strings.HasPrefix(trimmed, "==") && strings.HasSuffix(trimmed, "==") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "divider",
				Description: "Dividers are not supported in D2",
			})
		}

		// Check for delay
		if trimmed == "..." || strings.HasPrefix(lower, "...") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "delay",
				Description: "Delay notation is not supported",
			})
		}

		// Check for swimlanes
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "swimlane",
				Description: "Activity swimlanes are not supported",
				Suggestion:  "Use containers to group activities instead",
			})
			result.Convertible = false
		}

		// Check for hide/show
		if strings.HasPrefix(lower, "hide ") || strings.HasPrefix(lower, "show ") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "hide/show",
				Description: "Visibility control is not supported",
			})
		}

		// Check for preprocessing
		if strings.HasPrefix(trimmed, "!") {
			result.Unsupported = append(result.Unsupported, convert.UnsupportedFeature{
				Line:        lineNum,
				Feature:     "preprocessor",
				Description: "Preprocessor directives are not supported",
			})
		}
	}

	return result
}

func detectDiagramType(source string) DiagramType {
	lower := strings.ToLower(source)

	// Check for explicit type hints after @startuml
	if strings.Contains(lower, "participant ") || strings.Contains(lower, "actor ") {
		if strings.Contains(lower, "->") || strings.Contains(lower, "-->") {
			return DiagramSequence
		}
	}

	if strings.Contains(lower, "class ") || strings.Contains(lower, "interface ") {
		return DiagramClass
	}

	if strings.Contains(lower, "[") && strings.Contains(lower, "]") {
		if strings.Contains(lower, "component ") || strings.Contains(lower, "package ") {
			return DiagramComponent
		}
	}

	if strings.Contains(lower, "start") && strings.Contains(lower, "stop") {
		return DiagramActivity
	}

	if strings.Contains(lower, "usecase ") {
		return DiagramUseCase
	}

	if strings.Contains(lower, "state ") || strings.Contains(lower, "[*]") {
		return DiagramState
	}

	if strings.Contains(lower, "object ") {
		return DiagramObject
	}

	// Default detection based on arrow patterns
	if strings.Contains(lower, "->") || strings.Contains(lower, "-->") {
		// Check if it looks more like sequence (has colon for message)
		lines := strings.Split(source, "\n")
		for _, line := range lines {
			if (strings.Contains(line, "->") || strings.Contains(line, "-->")) && strings.Contains(line, ":") {
				return DiagramSequence
			}
		}
		return DiagramComponent
	}

	return DiagramUnknown
}
