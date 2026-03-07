package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/grokify/d2vision/format"
	"github.com/grokify/d2vision/generate"
	"github.com/grokify/d2vision/render"
	"github.com/spf13/cobra"
)

var (
	pipelineFormat string
	pipelineOutput string
	pipelineSVG    bool
	pipelineSimple bool
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline <spec-file>",
	Short: "Generate D2 diagram from a PipelineSpec",
	Long: `Generate D2 code from a PipelineSpec file (JSON, TOON, or YAML).

PipelineSpec defines multi-stage process workflows with:
  - Stages: sequential or parallel execution steps
  - Executors: program, API, deterministic code, LLM, or agent
  - Inputs/Outputs: data, files, configs, prompts, models
  - Lanes: group stages by system/team (swimlane view)
  - Branches: conditional branching (decision flow)

The command reads a PipelineSpec and outputs D2 code that can be
rendered with the d2 command.

Rendering modes:
  - Detailed (default): Full I/O breakdown with inputs/executor/outputs
  - Simple (--simple): Just stage boxes with executor type badges
  - Swimlane (auto): Stages grouped by lane (when "lane" field present)
  - Decision (auto): Diamond shapes for stages with branches

Input formats (auto-detected or specify with --format):
  - json: JSON format
  - toon: Token-Oriented Object Notation
  - yaml: YAML format

Examples:
  # Generate D2 code from PipelineSpec
  d2vision pipeline workflow.json

  # Pipe to d2 for SVG rendering
  d2vision pipeline workflow.json | d2 - output.svg

  # Direct SVG output
  d2vision pipeline workflow.json --svg -o workflow.svg

  # Simple view (compact, no I/O)
  d2vision pipeline workflow.json --simple

  # Read from stdin
  cat workflow.json | d2vision pipeline -

  # Specify input format explicitly
  d2vision pipeline workflow.toon --format toon

PipelineSpec JSON structure:
  {
    "id": "my-pipeline",
    "label": "My Workflow",
    "direction": "right",
    "stages": [
      {
        "id": "step1",
        "label": "Process Data",
        "executor": {
          "name": "processor.py",
          "type": "deterministic"
        },
        "inputs": [
          {"id": "data", "label": "Input Data", "kind": "file", "required": true}
        ],
        "outputs": [
          {"id": "result", "label": "Result", "kind": "data"}
        ]
      }
    ]
  }

Executor types:
  - program:       External program/binary
  - api:           REST/gRPC API call
  - deterministic: Custom code (same input = same output)
  - llm:           Language model inference
  - agent:         Autonomous agent execution

Resource kinds:
  - data:     In-memory data structure
  - file:     File on disk
  - config:   Configuration
  - prompt:   Prompt template
  - model:    ML model weights
  - program:  Executable/script
  - artifact: Build artifact
`,
	Args: cobra.ExactArgs(1),
	RunE: runPipeline,
}

func init() {
	pipelineCmd.Flags().StringVarP(&pipelineFormat, "format", "f", "", "Input format: json, toon, yaml (auto-detected if not specified)")
	pipelineCmd.Flags().StringVarP(&pipelineOutput, "output", "o", "", "Output file (default: stdout)")
	pipelineCmd.Flags().BoolVar(&pipelineSVG, "svg", false, "Output SVG directly")
	pipelineCmd.Flags().BoolVar(&pipelineSimple, "simple", false, "Compact view with only stage boxes (no internal I/O)")
}

func runPipeline(cmd *cobra.Command, args []string) error {
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

	// Detect or parse format
	var f format.Format
	if pipelineFormat != "" {
		f, err = format.Parse(pipelineFormat)
		if err != nil {
			return err
		}
	} else {
		// Auto-detect based on content
		f = detectFormat(data)
	}

	// Unmarshal PipelineSpec
	var spec generate.PipelineSpec
	if err := format.Unmarshal(data, &spec, f); err != nil {
		return fmt.Errorf("parsing PipelineSpec: %w", err)
	}

	// Validate
	if len(spec.Stages) == 0 {
		return fmt.Errorf("PipelineSpec has no stages")
	}

	// Generate D2 code
	opts := generate.PipelineRenderOptions{
		Simple: pipelineSimple,
	}
	gen := generate.NewPipelineGenerator()
	d2Code := gen.GenerateWithOptions(&spec, opts)

	// Output
	if pipelineSVG {
		return outputSVG(d2Code, pipelineOutput)
	}

	if pipelineOutput != "" {
		if err := os.WriteFile(pipelineOutput, []byte(d2Code), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		fmt.Print(d2Code)
	}

	return nil
}

// detectFormat tries to detect the format from content.
func detectFormat(data []byte) format.Format {
	content := strings.TrimSpace(string(data))

	// JSON starts with { or [
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return format.JSON
	}

	// YAML often has key: value on first line without braces
	// TOON uses = for assignment
	if strings.Contains(content, ": ") && !strings.Contains(content, "=") {
		return format.YAML
	}

	// Default to TOON
	return format.TOON
}

// outputSVG renders D2 code to SVG using the embedded d2 library.
func outputSVG(d2Code, outputPath string) error {
	r, err := render.New()
	if err != nil {
		return fmt.Errorf("creating renderer: %w", err)
	}

	svg, err := r.RenderSVG(context.Background(), d2Code, nil)
	if err != nil {
		return fmt.Errorf("rendering SVG: %w", err)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, svg, 0644); err != nil {
			return fmt.Errorf("writing SVG: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(svg); err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}
	}

	return nil
}
