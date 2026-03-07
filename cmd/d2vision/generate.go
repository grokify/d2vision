package main

import (
	"fmt"
	"io"
	"os"

	"github.com/grokify/d2vision/format"
	"github.com/grokify/d2vision/generate"
	"github.com/spf13/cobra"
)

var (
	generateFormat string
)

var generateCmd = &cobra.Command{
	Use:   "generate <spec-file>",
	Short: "Generate D2 code from a structured specification",
	Long: `Generate D2 code from a TOON, JSON, or YAML specification file.

This command takes a structured diagram specification and outputs D2 code
that can be rendered with the d2 command.

Input formats (auto-detected or specify with --format):
  - toon (default): Token-Oriented Object Notation
  - json: JSON format
  - yaml: YAML format

The specification defines:
  - Layout settings (direction, grid-columns, grid-rows)
  - Containers (clusters, boundaries) with nested children
  - Nodes with shapes, labels, and styles
  - Edges with labels and arrow styles

Examples:
  # Generate from TOON spec
  d2vision generate spec.toon > diagram.d2

  # Generate from JSON
  d2vision generate spec.json --format json > diagram.d2

  # Pipe from stdin
  cat spec.toon | d2vision generate - > diagram.d2

  # Generate and render in one pipeline
  d2vision generate spec.toon | d2 - output.svg
`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVarP(&generateFormat, "format", "f", "toon", "Input format: toon, json, yaml")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Read input
	var data []byte
	var err error

	if filePath == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(filePath)
	}
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	// Parse format
	f, err := format.Parse(generateFormat)
	if err != nil {
		return err
	}

	// Unmarshal spec
	var spec generate.DiagramSpec
	if err := format.Unmarshal(data, &spec, f); err != nil {
		return fmt.Errorf("parsing spec: %w", err)
	}

	// Generate D2 code
	gen := generate.NewGenerator()
	d2Code := gen.Generate(&spec)

	fmt.Print(d2Code)
	return nil
}
