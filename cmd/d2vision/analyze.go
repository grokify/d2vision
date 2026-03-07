package main

import (
	"fmt"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/format"
	"github.com/spf13/cobra"
)

var (
	analyzeFormat string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <file.svg>",
	Short: "Analyze diagram layout and provide generation hints",
	Long: `Analyze a D2-generated SVG diagram's layout characteristics.

This command provides:
  - Layout type detection (side-by-side, stacked, hierarchical, flow)
  - Grid layout detection (columns and rows)
  - Container hierarchy analysis
  - Cross-container edge detection
  - Insights about what makes the layout work
  - Hints for recreating the diagram

Output formats:
  - text (default): Human-readable analysis
  - toon: TOON format (token-efficient)
  - json: JSON format

Examples:
  # Analyze a diagram
  d2vision analyze diagram.svg

  # Get analysis as JSON
  d2vision analyze diagram.svg --format json

  # Get analysis as TOON
  d2vision analyze diagram.svg --format toon
`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeFormat, "format", "f", "text", "Output format: text, toon, json")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	diagram, err := d2vision.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filePath, err)
	}

	analysis := diagram.AnalyzeLayout()

	switch analyzeFormat {
	case "text":
		printTextAnalysis(diagram, analysis)
		return nil
	default:
		f, err := format.Parse(analyzeFormat)
		if err != nil {
			return err
		}
		output, err := format.Marshal(analysis, f)
		if err != nil {
			return fmt.Errorf("marshaling output: %w", err)
		}
		fmt.Println(string(output))
	}

	return nil
}

func printTextAnalysis(diagram *d2vision.Diagram, analysis *d2vision.LayoutAnalysis) {
	fmt.Println("# Layout Analysis")
	fmt.Println()

	// Overview
	fmt.Println("## Overview")
	fmt.Printf("- Nodes: %d\n", len(diagram.Nodes))
	fmt.Printf("- Edges: %d\n", len(diagram.Edges))
	fmt.Printf("- Layout type: %s\n", analysis.LayoutType)
	if analysis.Direction != "" {
		fmt.Printf("- Direction: %s\n", analysis.Direction)
	}
	if analysis.GridColumns > 0 {
		fmt.Printf("- Grid: %d columns x %d rows\n", analysis.GridColumns, analysis.GridRows)
	}
	if analysis.ContainerCount > 0 {
		fmt.Printf("- Containers: %d (max depth: %d)\n", analysis.ContainerCount, analysis.NestingDepth)
	}
	if analysis.CrossContainerEdges > 0 {
		fmt.Printf("- Cross-container edges: %d\n", analysis.CrossContainerEdges)
	}
	fmt.Println()

	// Insights
	if len(analysis.Insights) > 0 {
		fmt.Println("## Insights")
		for _, insight := range analysis.Insights {
			fmt.Printf("- %s\n", insight)
		}
		fmt.Println()
	}

	// Generation hints
	if len(analysis.GenerationHints) > 0 {
		fmt.Println("## Generation Hints")
		for _, hint := range analysis.GenerationHints {
			fmt.Printf("- %s\n", hint)
		}
		fmt.Println()
	}
}
