//go:build ignore

// This file generates JSON Schema files from Go types.
// Run with: go run schema/generate.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grokify/d2vision/generate"
	"github.com/invopop/jsonschema"
)

func main() {
	schemaDir := "schema"

	// Generate PipelineSpec schema
	if err := generateSchema(generate.PipelineSpec{}, filepath.Join(schemaDir, "pipeline-spec.schema.json")); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating PipelineSpec schema: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated: schema/pipeline-spec.schema.json")

	// Generate DiagramSpec schema
	if err := generateSchema(generate.DiagramSpec{}, filepath.Join(schemaDir, "diagram-spec.schema.json")); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating DiagramSpec schema: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated: schema/diagram-spec.schema.json")

	// Generate SequenceSpec schema (standalone for sequence diagrams)
	if err := generateSchema(generate.SequenceSpec{}, filepath.Join(schemaDir, "sequence-spec.schema.json")); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating SequenceSpec schema: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated: schema/sequence-spec.schema.json")

	fmt.Println("\nAll schemas generated successfully!")
}

func generateSchema(v interface{}, outputPath string) error {
	r := jsonschema.Reflector{
		DoNotReference: false,
	}

	schema := r.Reflect(v)

	// Add schema metadata
	schema.Title = getTypeName(v)
	schema.Description = getTypeDescription(v)

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("writing schema file: %w", err)
	}

	return nil
}

func getTypeName(v interface{}) string {
	switch v.(type) {
	case generate.PipelineSpec:
		return "PipelineSpec"
	case generate.DiagramSpec:
		return "DiagramSpec"
	case generate.SequenceSpec:
		return "SequenceSpec"
	default:
		return ""
	}
}

func getTypeDescription(v interface{}) string {
	switch v.(type) {
	case generate.PipelineSpec:
		return "Defines a multi-stage process pipeline with inputs, outputs, and various executor types including deterministic code and LLM/agents."
	case generate.DiagramSpec:
		return "Defines the structure for generating D2 diagrams with nodes, containers, edges, and special diagram types."
	case generate.SequenceSpec:
		return "Defines a sequence diagram with actors and message steps."
	default:
		return ""
	}
}
