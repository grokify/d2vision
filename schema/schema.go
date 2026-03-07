// Package schema provides embedded JSON Schema definitions for d2vision types.
package schema

import (
	_ "embed"
)

// PipelineSpecSchema is the JSON Schema for generate.PipelineSpec.
//
//go:embed pipeline-spec.schema.json
var PipelineSpecSchema []byte

// DiagramSpecSchema is the JSON Schema for generate.DiagramSpec.
//
//go:embed diagram-spec.schema.json
var DiagramSpecSchema []byte

// SequenceSpecSchema is the JSON Schema for generate.SequenceSpec.
//
//go:embed sequence-spec.schema.json
var SequenceSpecSchema []byte
