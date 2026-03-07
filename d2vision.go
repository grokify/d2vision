// Package d2vision parses D2-generated SVG diagrams and outputs
// structured JSON and natural language descriptions.
//
// D2 (https://d2lang.com) is a modern diagram scripting language.
// This package enables programmatic analysis of D2 diagrams by
// extracting node and edge information from the generated SVG files.
//
// # Basic Usage
//
//	diagram, err := d2vision.ParseFile("diagram.svg")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get JSON output
//	json, _ := diagram.JSONIndent("", "  ")
//	fmt.Println(string(json))
//
//	// Get natural language description
//	fmt.Println(diagram.Describe())
//
// # How It Works
//
// D2 encodes node and edge IDs as base64 in CSS class names within the
// generated SVG. For example, the class "YQ==" decodes to "a", and
// "(a -> b)[0]" is encoded as "KGEgLT4gYilbMF0=". By decoding these
// class names, d2vision reconstructs the diagram structure including:
//
//   - Node IDs and labels
//   - Edge connections (source → target)
//   - Container/hierarchy relationships
//   - Shape types (rectangle, circle, cylinder, etc.)
//   - Visual styling information
//
// # Output Formats
//
// The package supports multiple output formats:
//
//   - JSON: Structured data suitable for programmatic consumption
//   - Text: Human-readable natural language description
//   - Summary: Brief overview of diagram contents
//   - LLM: Optimized format for large language model consumption
package d2vision

import (
	"io"
	"os"
)

// Version is the version of the d2vision package.
// This is set at build time by goreleaser via ldflags.
var Version = "0.1.0"

// ParseFile parses a D2 SVG file and returns a Diagram.
func ParseFile(path string) (diagram *Diagram, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	return Parse(f)
}

// Parse parses a D2 SVG from an io.Reader and returns a Diagram.
func Parse(r io.Reader) (*Diagram, error) {
	parser := NewParser()
	return parser.Parse(r)
}

// ParseBytes parses a D2 SVG from bytes and returns a Diagram.
func ParseBytes(data []byte) (*Diagram, error) {
	parser := NewParser()
	return parser.ParseBytes(data)
}

// ParseString parses a D2 SVG from a string and returns a Diagram.
func ParseString(s string) (*Diagram, error) {
	parser := NewParser()
	return parser.ParseString(s)
}
