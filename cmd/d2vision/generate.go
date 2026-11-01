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
	generateTo     string
)

var generateCmd = &cobra.Command{
	Use:   "generate <spec-file>",
	Short: "Generate D2 or Mermaid from a structured specification",
	Long: `Generate D2 code, or a Mermaid flowchart, from a TOON, JSON, or YAML
specification file.

This command takes a structured diagram specification and outputs D2 code
that can be rendered with the d2 command. With --to mermaid it writes a Mermaid
flowchart instead.

Input formats (the format is not detected from the file name; name it with --format):
  - toon (default): Token-Oriented Object Notation
  - json: JSON format
  - yaml: YAML format

Output formats (--to):
  - d2 (default)
  - mermaid: a flowchart. Containers become subgraphs. Mermaid has no grid, so
    gridColumns and gridRows are ignored, and a spec with sequences or tables
    is refused.

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

  # Write a Mermaid flowchart instead
  d2vision generate spec.toon --to mermaid > diagram.mmd

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
	generateCmd.Flags().StringVar(&generateTo, "to", "d2", "Output format: d2, mermaid")
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

	f, err := format.Parse(generateFormat)
	if err != nil {
		return err
	}
	out, err := generateFromSpec(data, f, generateTo)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

// generateFromSpec reads a spec in the given input format and writes it as D2
// or as a Mermaid flowchart.
func generateFromSpec(data []byte, in format.Format, to string) (string, error) {
	var spec generate.DiagramSpec
	if err := format.Unmarshal(data, &spec, in); err != nil {
		return "", fmt.Errorf("parsing spec: %w", err)
	}
	gen := generate.NewGenerator()
	switch to {
	case "d2":
		return gen.Generate(&spec), nil
	case "mermaid":
		return gen.Mermaid(&spec)
	default:
		return "", fmt.Errorf("unknown output format %q (want d2 or mermaid)", to)
	}
}
