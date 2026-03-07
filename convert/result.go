package convert

import "github.com/grokify/d2vision/generate"

// ConversionResult contains the output of a successful conversion.
type ConversionResult struct {
	// Spec is the converted diagram specification.
	Spec *generate.DiagramSpec `json:"spec"`

	// SourceType identifies the specific diagram type (flowchart, sequence, class, etc.).
	SourceType string `json:"sourceType"`

	// Warnings are non-fatal issues encountered during conversion.
	Warnings []Warning `json:"warnings,omitempty"`

	// Skipped contains features that were skipped during conversion.
	Skipped []SkippedFeature `json:"skipped,omitempty"`
}

// Warning represents a non-fatal issue during conversion.
type Warning struct {
	// Line is the source line number (1-indexed).
	Line int `json:"line"`

	// Feature is the feature name that triggered the warning.
	Feature string `json:"feature"`

	// Message describes the warning.
	Message string `json:"message"`
}

// SkippedFeature represents a feature that was skipped during conversion.
type SkippedFeature struct {
	// Line is the source line number (1-indexed).
	Line int `json:"line"`

	// Feature is the name of the skipped feature.
	Feature string `json:"feature"`

	// Source is the original source line.
	Source string `json:"source"`
}

// LintResult contains the output of linting source code.
type LintResult struct {
	// Format is the detected source format.
	Format SourceFormat `json:"format"`

	// DiagramType is the specific diagram type (flowchart, sequence, class, etc.).
	DiagramType string `json:"diagramType"`

	// Supported lists features that are fully supported.
	Supported []string `json:"supported"`

	// Unsupported lists features that are not supported or partially supported.
	Unsupported []UnsupportedFeature `json:"unsupported,omitempty"`

	// Convertible indicates whether the diagram can be converted.
	Convertible bool `json:"convertible"`
}

// UnsupportedFeature describes a feature that cannot be converted.
type UnsupportedFeature struct {
	// Line is the source line number (1-indexed).
	Line int `json:"line"`

	// Feature is the name of the unsupported feature.
	Feature string `json:"feature"`

	// Description explains why the feature is unsupported.
	Description string `json:"description"`

	// Suggestion provides an alternative approach, if any.
	Suggestion string `json:"suggestion,omitempty"`
}

// HasUnsupported returns true if there are any unsupported features.
func (r *LintResult) HasUnsupported() bool {
	return len(r.Unsupported) > 0
}

// HasWarnings returns true if there are any warnings.
func (r *ConversionResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// HasSkipped returns true if any features were skipped.
func (r *ConversionResult) HasSkipped() bool {
	return len(r.Skipped) > 0
}
