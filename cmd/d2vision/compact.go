package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/grokify/d2vision/render"
	"github.com/spf13/cobra"
)

var (
	compactOutput   string
	compactLegend   string
	compactFontSize int64
)

var compactCmd = &cobra.Command{
	Use:   "compact <input.d2>",
	Short: "Render D2 with long node labels replaced by short numeric keys",
	Long: `Render a D2 diagram after replacing long leaf-node labels with short
sequential numeric keys ("1", "2", "3", …). Because labels are shortened before
layout, shapes are sized for the keys and the diagram renders much narrower.

A legend mapping each key back to its original label is written as a JSON array
of {key, label} objects, so the caller can place the full labels above the image.

Container/boundary labels and edge labels are left untouched. Only SVG output is
supported.

Examples:
  # Render compact SVG and write the legend to a file
  d2vision compact diagram.d2 -o compact.svg --legend legend.json

  # Render compact SVG, legend to stdout
  d2vision compact diagram.d2 -o compact.svg
`,
	Args: cobra.ExactArgs(1),
	RunE: runCompact,
}

func init() {
	compactCmd.Flags().StringVarP(&compactOutput, "output", "o", "", "Output SVG file (required)")
	compactCmd.Flags().StringVar(&compactLegend, "legend", "", "Output JSON file for the legend (default: stdout)")
	compactCmd.Flags().Int64Var(&compactFontSize, "font-size", 0, "Global font-size floor in px (0 = unchanged)")
}

func runCompact(cmd *cobra.Command, args []string) error {
	if compactOutput == "" {
		return fmt.Errorf("--output/-o is required")
	}

	d2Code, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	r, err := render.New()
	if err != nil {
		return fmt.Errorf("creating renderer: %w", err)
	}

	opts := render.DefaultOptions()
	opts.FontSize = compactFontSize

	svg, legend, err := r.RenderCompact(context.Background(), string(d2Code), render.FormatSVG, opts)
	if err != nil {
		return fmt.Errorf("rendering compact SVG: %w", err)
	}

	if err := os.WriteFile(compactOutput, svg, 0644); err != nil {
		return fmt.Errorf("writing SVG: %w", err)
	}

	legendJSON, err := json.MarshalIndent(legend, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling legend: %w", err)
	}
	legendJSON = append(legendJSON, '\n')

	if compactLegend != "" {
		if err := os.WriteFile(compactLegend, legendJSON, 0644); err != nil {
			return fmt.Errorf("writing legend: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(legendJSON); err != nil {
			return fmt.Errorf("writing legend to stdout: %w", err)
		}
	}

	return nil
}
