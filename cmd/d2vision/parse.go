package main

import (
	"fmt"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/format"
	"github.com/spf13/cobra"
)

var (
	parseFormat        string
	parseForGeneration bool
)

var parseCmd = &cobra.Command{
	Use:   "parse <file.svg>",
	Short: "Parse a D2-generated SVG diagram",
	Long: `Parse a D2-generated SVG diagram and output its structure.

D2 encodes element IDs as base64 in CSS class names. This command decodes
those IDs to extract nodes, edges, shapes, positions, and hierarchy.

Output formats:
  - toon (default): Token-Oriented Object Notation, optimized for LLMs
  - json: Standard JSON with indentation
  - json-compact: Minified JSON
  - text: Human-readable description
  - summary: Brief one-line summary
  - llm: Detailed format for LLM analysis

Examples:
  d2vision parse diagram.svg
  d2vision parse diagram.svg --format json
  d2vision parse diagram.svg --format text
`,
	Args: cobra.ExactArgs(1),
	RunE: runParse,
}

func init() {
	parseCmd.Flags().StringVarP(&parseFormat, "format", "f", "toon", "Output format: toon, json, json-compact, text, summary, llm, analysis")
	parseCmd.Flags().BoolVar(&parseForGeneration, "for-generation", false, "Output hints for recreating the diagram")
}

func runParse(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	diagram, err := d2vision.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filePath, err)
	}

	// Handle --for-generation flag (overrides format)
	if parseForGeneration {
		fmt.Print(diagram.DescribeForGeneration())
		return nil
	}

	// Handle special text formats
	switch parseFormat {
	case "text":
		fmt.Print(diagram.DescribeDetailed())
		return nil
	case "summary":
		fmt.Println(diagram.DescribeSummary())
		return nil
	case "llm":
		fmt.Print(diagram.DescribeForLLM())
		return nil
	case "analysis":
		fmt.Print(diagram.DescribeForGeneration())
		return nil
	}

	// Handle structured formats
	f, err := format.Parse(parseFormat)
	if err != nil {
		return err
	}

	output, err := format.Marshal(diagram, f)
	if err != nil {
		return fmt.Errorf("marshaling output: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
