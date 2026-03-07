// Package convert provides converters for transforming Mermaid and PlantUML
// diagrams into D2 format via the generate.DiagramSpec intermediate representation.
package convert

import (
	"fmt"
	"strings"

	"github.com/grokify/d2vision/generate"
)

// SourceFormat represents the source diagram format.
type SourceFormat string

const (
	FormatMermaid  SourceFormat = "mermaid"
	FormatPlantUML SourceFormat = "plantuml"
	FormatUnknown  SourceFormat = "unknown"
)

// String returns the string representation of the format.
func (f SourceFormat) String() string {
	return string(f)
}

// ParseFormat parses a format string into a SourceFormat.
func ParseFormat(s string) SourceFormat {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "mermaid", "mmd":
		return FormatMermaid
	case "plantuml", "puml", "uml":
		return FormatPlantUML
	default:
		return FormatUnknown
	}
}

// Converter defines the interface for diagram format converters.
type Converter interface {
	// Convert transforms source code into a DiagramSpec.
	Convert(source string) (*ConversionResult, error)

	// Lint analyzes source code for unsupported features without converting.
	Lint(source string) (*LintResult, error)

	// Format returns the source format this converter handles.
	Format() SourceFormat
}

// DetectFormat attempts to detect the source format from content.
func DetectFormat(source string) SourceFormat {
	trimmed := strings.TrimSpace(source)

	// Check for Mermaid indicators
	if isMermaid(trimmed) {
		return FormatMermaid
	}

	// Check for PlantUML indicators
	if isPlantUML(trimmed) {
		return FormatPlantUML
	}

	return FormatUnknown
}

// isMermaid checks if the source appears to be Mermaid syntax.
func isMermaid(source string) bool {
	lower := strings.ToLower(source)

	// Mermaid diagram type declarations
	mermaidPrefixes := []string{
		"graph ",
		"flowchart ",
		"sequencediagram",
		"classdiagram",
		"statediagram",
		"erdiagram",
		"gantt",
		"pie ",
		"journey",
		"gitgraph",
		"mindmap",
		"timeline",
		"quadrantchart",
		"requirementdiagram",
		"c4context",
	}

	for _, prefix := range mermaidPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	// Check for %%{init directive
	if strings.HasPrefix(source, "%%{") {
		return true
	}

	return false
}

// isPlantUML checks if the source appears to be PlantUML syntax.
func isPlantUML(source string) bool {
	lower := strings.ToLower(source)

	// PlantUML start marker
	if strings.HasPrefix(lower, "@startuml") {
		return true
	}

	// PlantUML component indicators without @startuml
	plantUMLIndicators := []string{
		"participant ",
		"actor ",
		"usecase ",
		"class ",
		"interface ",
		"package ",
		"component ",
		"database ",
		"cloud ",
		"node ",
		"frame ",
		"folder ",
		"rectangle ",
	}

	for _, indicator := range plantUMLIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	return false
}

// ConvertToD2 converts source code to D2 format.
// It auto-detects the format if not specified.
func ConvertToD2(source string, formatHint SourceFormat) (string, *ConversionResult, error) {
	format := formatHint
	if format == FormatUnknown || format == "" {
		format = DetectFormat(source)
	}

	if format == FormatUnknown {
		return "", nil, fmt.Errorf("unable to detect source format; specify --from flag")
	}

	converter, err := GetConverter(format)
	if err != nil {
		return "", nil, err
	}

	result, err := converter.Convert(source)
	if err != nil {
		return "", nil, err
	}

	gen := generate.NewGenerator()
	d2Code := gen.Generate(result.Spec)

	return d2Code, result, nil
}

// GetConverter returns a converter for the specified format.
func GetConverter(format SourceFormat) (Converter, error) {
	switch format {
	case FormatMermaid:
		return NewMermaidConverter(), nil
	case FormatPlantUML:
		return NewPlantUMLConverter(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// NewMermaidConverter creates a new Mermaid converter.
// This is implemented in the mermaid subpackage.
var NewMermaidConverter func() Converter

// NewPlantUMLConverter creates a new PlantUML converter.
// This is implemented in the plantuml subpackage.
var NewPlantUMLConverter func() Converter
