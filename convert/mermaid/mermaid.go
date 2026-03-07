// Package mermaid provides parsing and conversion of Mermaid diagrams to D2.
package mermaid

import (
	"fmt"
	"strings"

	"github.com/grokify/d2vision/convert"
	"github.com/grokify/d2vision/generate"
)

// Converter implements the convert.Converter interface for Mermaid diagrams.
type Converter struct{}

// NewConverter creates a new Mermaid converter.
func NewConverter() convert.Converter {
	return &Converter{}
}

// Format returns the source format.
func (c *Converter) Format() convert.SourceFormat {
	return convert.FormatMermaid
}

// Convert transforms Mermaid source into a DiagramSpec.
func (c *Converter) Convert(source string) (*convert.ConversionResult, error) {
	// Parse the source
	doc, err := Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parsing mermaid: %w", err)
	}

	result := &convert.ConversionResult{
		SourceType: string(doc.Type),
	}

	// Convert based on diagram type
	var spec *generate.DiagramSpec

	switch doc.Type {
	case DiagramFlowchart:
		converter := &FlowchartConverter{}
		spec = converter.Convert(doc)

	case DiagramSequence:
		converter := &SequenceConverter{}
		spec = converter.Convert(doc)

	case DiagramClass:
		converter := &ClassConverter{}
		spec = converter.Convert(doc)

	case DiagramState:
		// Convert state diagrams as flowcharts
		converter := &FlowchartConverter{}
		spec = converter.Convert(doc)
		result.Warnings = append(result.Warnings, convert.Warning{
			Line:    1,
			Feature: "stateDiagram",
			Message: "State diagram converted as flowchart; some features may be lost",
		})

	case DiagramER:
		// Convert ER diagrams as containers with edges
		converter := &FlowchartConverter{}
		spec = converter.Convert(doc)
		result.Warnings = append(result.Warnings, convert.Warning{
			Line:    1,
			Feature: "erDiagram",
			Message: "ER diagram converted as flowchart; cardinality notation lost",
		})

	default:
		return nil, fmt.Errorf("unsupported diagram type: %s", doc.Type)
	}

	result.Spec = spec

	// Add skipped features from parsing
	result.Skipped = collectSkippedFeatures(source)

	return result, nil
}

// Lint analyzes Mermaid source for D2 compatibility.
func (c *Converter) Lint(source string) (*convert.LintResult, error) {
	linter := &Linter{}
	return linter.Lint(source), nil
}

// Parse parses Mermaid source into a Document AST.
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
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Notes in sequence diagrams
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

		// Click handlers
		if strings.HasPrefix(lower, "click ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "click",
				Source:  trimmed,
			})
		}

		// Style definitions
		if strings.HasPrefix(lower, "style ") || strings.HasPrefix(lower, "classdef ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "style",
				Source:  trimmed,
			})
		}

		// Link styles
		if strings.HasPrefix(lower, "linkstyle ") {
			skipped = append(skipped, convert.SkippedFeature{
				Line:    lineNum,
				Feature: "linkStyle",
				Source:  trimmed,
			})
		}
	}

	return skipped
}

func init() {
	// Register the converter with the convert package
	convert.NewMermaidConverter = func() convert.Converter {
		return NewConverter()
	}
}
