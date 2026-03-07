// Package plantuml provides parsing and conversion of PlantUML diagrams to D2.
package plantuml

import (
	"fmt"
	"strings"

	"github.com/grokify/d2vision/convert"
	"github.com/grokify/d2vision/generate"
)

// Converter implements the convert.Converter interface for PlantUML diagrams.
type Converter struct{}

// NewConverter creates a new PlantUML converter.
func NewConverter() convert.Converter {
	return &Converter{}
}

// Format returns the source format.
func (c *Converter) Format() convert.SourceFormat {
	return convert.FormatPlantUML
}

// Convert transforms PlantUML source into a DiagramSpec.
func (c *Converter) Convert(source string) (*convert.ConversionResult, error) {
	// Parse the source
	doc, err := Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parsing plantuml: %w", err)
	}

	result := &convert.ConversionResult{
		SourceType: string(doc.Type),
	}

	// Convert based on diagram type
	var spec *generate.DiagramSpec

	switch doc.Type {
	case DiagramSequence:
		converter := &SequenceConverter{}
		spec = converter.Convert(doc)

	case DiagramClass:
		converter := &ClassConverter{}
		spec = converter.Convert(doc)

	case DiagramComponent:
		converter := &ComponentConverter{}
		spec = converter.Convert(doc)

	case DiagramActivity:
		// Convert activity diagrams as component diagrams
		converter := &ComponentConverter{}
		spec = converter.Convert(doc)
		result.Warnings = append(result.Warnings, convert.Warning{
			Line:    1,
			Feature: "activity",
			Message: "Activity diagram converted as component diagram; flow semantics may be lost",
		})

	case DiagramUseCase:
		// Convert use case diagrams as component diagrams
		converter := &ComponentConverter{}
		spec = converter.Convert(doc)
		result.Warnings = append(result.Warnings, convert.Warning{
			Line:    1,
			Feature: "usecase",
			Message: "Use case diagram converted as component diagram; use case notation lost",
		})

	case DiagramState:
		// Convert state diagrams as component diagrams
		converter := &ComponentConverter{}
		spec = converter.Convert(doc)
		result.Warnings = append(result.Warnings, convert.Warning{
			Line:    1,
			Feature: "state",
			Message: "State diagram converted as component diagram; state notation lost",
		})

	case DiagramObject:
		// Convert object diagrams as class diagrams
		converter := &ClassConverter{}
		spec = converter.Convert(doc)
		result.Warnings = append(result.Warnings, convert.Warning{
			Line:    1,
			Feature: "object",
			Message: "Object diagram converted as class diagram",
		})

	default:
		// Try component converter as fallback
		converter := &ComponentConverter{}
		spec = converter.Convert(doc)
		result.Warnings = append(result.Warnings, convert.Warning{
			Line:    1,
			Feature: "unknown",
			Message: "Unknown diagram type; converted as component diagram",
		})
	}

	result.Spec = spec

	// Add skipped features from parsing
	result.Skipped = collectSkippedFeatures(source)

	return result, nil
}

// Lint analyzes PlantUML source for D2 compatibility.
func (c *Converter) Lint(source string) (*convert.LintResult, error) {
	linter := &Linter{}
	return linter.Lint(source), nil
}

// Parse parses PlantUML source into a Document AST.
func Parse(source string) (*Document, error) {
	// Store original lines for diagnostics
	lines := strings.Split(source, "\n")

	// Tokenize
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	// Parse
	parser := NewParser(tokens)
	doc, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	doc.Lines = lines
	return doc, nil
}

// collectSkippedFeatures identifies features that were skipped during conversion.
func collectSkippedFeatures(source string) []convert.SkippedFeature {
	var skipped []convert.SkippedFeature

	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "'") {
			continue
		}

		// Notes
		if strings.HasPrefix(lower, "note ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "note",
				Source:  trimmed,
			})
		}

		// Activation
		if strings.HasPrefix(lower, "activate ") || strings.HasPrefix(lower, "deactivate ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "activation",
				Source:  trimmed,
			})
		}

		// Skinparam
		if strings.HasPrefix(lower, "skinparam ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "skinparam",
				Source:  trimmed,
			})
		}

		// Hide/show
		if strings.HasPrefix(lower, "hide ") || strings.HasPrefix(lower, "show ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "visibility",
				Source:  trimmed,
			})
		}

		// Autonumber
		if strings.HasPrefix(lower, "autonumber") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "autonumber",
				Source:  trimmed,
			})
		}

		// Create/destroy
		if strings.HasPrefix(lower, "create ") || strings.HasPrefix(lower, "destroy ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "lifecycle",
				Source:  trimmed,
			})
		}
	}

	return skipped
}

func init() {
	// Register the converter with the convert package
	convert.NewPlantUMLConverter = func() convert.Converter {
		return NewConverter()
	}
}
