// Command d2vision parses D2-generated SVG diagrams, generates D2 code,
// and provides tools for AI-assisted diagram creation.
package main

import (
	"os"

	"github.com/grokify/d2vision"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "d2vision",
	Short: "Tools for D2 diagram parsing and generation",
	Long: `d2vision provides tools for working with D2 diagrams:

  Parse:    Extract structure from D2-generated SVGs
  Generate: Create D2 code from structured specifications
  Pipeline: Generate D2 from PipelineSpec (workflow diagrams)
  Template: Generate common diagram patterns
  Learn:    Reverse engineer D2 code from SVGs
  Lint:     Check D2 files for layout issues
  Diff:     Compare two diagrams
  Watch:    Auto-render D2 files on changes
  Analyze:  Analyze layout and provide generation hints
  Convert:  Convert Mermaid/PlantUML diagrams to D2
  Rotate:   Rotate SVG by 90° increments (for portrait/landscape conversion)

Output formats:
  - toon (default): Token-Oriented Object Notation (~40% fewer tokens than JSON)
  - json: Standard JSON with indentation
  - json-compact: Minified JSON
  - yaml: YAML format

Examples:
  # Parse an SVG diagram (default: TOON output)
  d2vision parse diagram.svg

  # Generate D2 code from a spec file
  d2vision generate spec.toon > diagram.d2

  # Use templates
  d2vision template network-boundary

  # Learn D2 from existing SVG
  d2vision learn diagram.svg --d2

  # Lint D2 files for issues
  d2vision lint diagram.d2

  # Compare two diagrams
  d2vision diff old.svg new.svg
`,
	Version: d2vision.Version,
}

func init() {
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(iconsCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(pipelineCmd)
	rootCmd.AddCommand(rotateCmd)
}
