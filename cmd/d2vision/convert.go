package main

import (
	"fmt"
	"io"
	"os"

	"github.com/grokify/d2vision/convert"
	_ "github.com/grokify/d2vision/convert/mermaid"  // Register Mermaid converter
	_ "github.com/grokify/d2vision/convert/plantuml" // Register PlantUML converter
	"github.com/grokify/d2vision/format"
	"github.com/spf13/cobra"
)

var (
	convertFrom     string
	convertFormat   string
	convertLintOnly bool
	convertStrict   bool
	convertOutput   string
)

var convertCmd = &cobra.Command{
	Use:   "convert <input-file>",
	Short: "Convert Mermaid or PlantUML diagrams to D2",
	Long: `Convert Mermaid or PlantUML diagram files to D2 format.

This command parses external diagram formats and converts them to D2 code
using the DiagramSpec intermediate representation.

Source formats (auto-detected or specify with --from):
  - mermaid, mmd: Mermaid diagrams
  - plantuml, puml: PlantUML diagrams

Output formats:
  - d2 (default): D2 source code
  - spec-toon: DiagramSpec in TOON format
  - spec-json: DiagramSpec in JSON format

Supported Mermaid diagram types:
  - flowchart, graph: Node-edge diagrams
  - sequenceDiagram: Sequence diagrams
  - classDiagram: Class diagrams (partial)

Supported PlantUML diagram types:
  - Sequence diagrams (@startuml with participant/actor)
  - Component diagrams (package, component)
  - Class diagrams (class, interface)

Examples:
  # Convert Mermaid to D2
  d2vision convert diagram.mmd > diagram.d2

  # Convert PlantUML with explicit format
  d2vision convert diagram.puml --from plantuml > diagram.d2

  # Lint before converting (shows unsupported features)
  d2vision convert --lint-only diagram.mmd

  # Strict mode: fail on any unsupported features
  d2vision convert --strict diagram.puml

  # Output spec instead of D2 code
  d2vision convert diagram.mmd --format spec-toon

  # Full pipeline: convert and render
  d2vision convert diagram.mmd | d2 - output.svg

  # Read from stdin
  cat diagram.mmd | d2vision convert --from mermaid -
`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

func init() {
	convertCmd.Flags().StringVarP(&convertFrom, "from", "f", "", "Source format: mermaid, plantuml (auto-detected if not specified)")
	convertCmd.Flags().StringVar(&convertFormat, "format", "d2", "Output format: d2, spec-toon, spec-json")
	convertCmd.Flags().BoolVar(&convertLintOnly, "lint-only", false, "Only lint, don't convert")
	convertCmd.Flags().BoolVar(&convertStrict, "strict", false, "Fail on any unsupported features")
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "", "Output file (default: stdout)")
}

func runConvert(cmd *cobra.Command, args []string) error {
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

	source := string(data)

	// Parse source format
	srcFormat := convert.ParseFormat(convertFrom)
	if srcFormat == convert.FormatUnknown && convertFrom != "" {
		return fmt.Errorf("unknown source format: %s", convertFrom)
	}

	// Auto-detect if not specified
	if srcFormat == convert.FormatUnknown {
		srcFormat = convert.DetectFormat(source)
		if srcFormat == convert.FormatUnknown {
			return fmt.Errorf("unable to detect source format; specify --from flag")
		}
	}

	// Get converter
	converter, err := convert.GetConverter(srcFormat)
	if err != nil {
		return err
	}

	// Lint only mode
	if convertLintOnly {
		return runLintOnly(converter, source)
	}

	// Lint first in strict mode
	if convertStrict {
		lintResult, err := converter.Lint(source)
		if err != nil {
			return fmt.Errorf("linting: %w", err)
		}
		if lintResult.HasUnsupported() {
			printLintResult(lintResult)
			return fmt.Errorf("conversion aborted: unsupported features found (use without --strict to convert anyway)")
		}
	}

	// Convert
	result, err := converter.Convert(source)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Output
	var output string
	switch convertFormat {
	case "d2":
		d2Code, _, err := convert.ConvertToD2(source, srcFormat)
		if err != nil {
			return err
		}
		output = d2Code

	case "spec-toon":
		data, err := format.Marshal(result.Spec, format.TOON)
		if err != nil {
			return fmt.Errorf("marshaling spec: %w", err)
		}
		output = string(data)

	case "spec-json":
		data, err := format.Marshal(result.Spec, format.JSON)
		if err != nil {
			return fmt.Errorf("marshaling spec: %w", err)
		}
		output = string(data)

	default:
		return fmt.Errorf("unknown output format: %s", convertFormat)
	}

	// Write output
	if convertOutput != "" {
		if err := os.WriteFile(convertOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	// Print warnings if any
	if result.HasWarnings() || result.HasSkipped() {
		printConversionWarnings(result)
	}

	return nil
}

func runLintOnly(converter convert.Converter, source string) error {
	lintResult, err := converter.Lint(source)
	if err != nil {
		return fmt.Errorf("linting: %w", err)
	}

	printLintResult(lintResult)

	if lintResult.HasUnsupported() {
		os.Exit(1)
	}
	return nil
}

func printLintResult(result *convert.LintResult) {
	fmt.Printf("Format: %s\n", result.Format)
	fmt.Printf("Diagram type: %s\n", result.DiagramType)
	fmt.Printf("Convertible: %v\n\n", result.Convertible)

	if len(result.Supported) > 0 {
		fmt.Println("Supported features:")
		for _, f := range result.Supported {
			fmt.Printf("  ✓ %s\n", f)
		}
		fmt.Println()
	}

	if len(result.Unsupported) > 0 {
		fmt.Println("Unsupported features:")
		for _, f := range result.Unsupported {
			fmt.Printf("  ✘ Line %d: %s\n", f.Line, f.Feature)
			fmt.Printf("    %s\n", f.Description)
			if f.Suggestion != "" {
				fmt.Printf("    → %s\n", f.Suggestion)
			}
		}
		fmt.Println()
	}
}

func printConversionWarnings(result *convert.ConversionResult) {
	if len(result.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  ⚠ Line %d [%s]: %s\n", w.Line, w.Feature, w.Message)
		}
	}

	if len(result.Skipped) > 0 {
		fmt.Fprintln(os.Stderr, "\nSkipped features:")
		for _, s := range result.Skipped {
			fmt.Fprintf(os.Stderr, "  - Line %d [%s]: %s\n", s.Line, s.Feature, s.Source)
		}
	}
}
