package generate

// PipelineSpec defines a multi-stage process pipeline with inputs, outputs,
// and various executor types including deterministic code and LLM/agents.
type PipelineSpec struct {
	ID          string      `json:"id" yaml:"id"`
	Label       string      `json:"label,omitempty" yaml:"label,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Direction   string      `json:"direction,omitempty" yaml:"direction,omitempty"` // right, down
	Stages      []StageSpec `json:"stages" yaml:"stages"`
	Flows       []FlowSpec  `json:"flows,omitempty" yaml:"flows,omitempty"` // Cross-stage data flows
}

// StageSpec defines a single stage in the pipeline.
type StageSpec struct {
	ID       string         `json:"id" yaml:"id"`
	Label    string         `json:"label,omitempty" yaml:"label,omitempty"`
	Executor ExecutorSpec   `json:"executor" yaml:"executor"`
	Inputs   []ResourceSpec `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs  []ResourceSpec `json:"outputs,omitempty" yaml:"outputs,omitempty"`

	// Parallelism support
	Parallel []StageSpec `json:"parallel,omitempty" yaml:"parallel,omitempty"` // Fan-out stages
	JoinFrom []string    `json:"joinFrom,omitempty" yaml:"joinFrom,omitempty"` // Fan-in from stage IDs

	// Swimlane support
	Lane string `json:"lane,omitempty" yaml:"lane,omitempty"` // Group stages by lane (system/team)

	// Decision node support
	Branches []BranchSpec `json:"branches,omitempty" yaml:"branches,omitempty"` // Conditional branches
}

// BranchSpec defines a conditional branch from a decision node.
type BranchSpec struct {
	Label     string `json:"label" yaml:"label"`         // Branch label (e.g., "Yes", "No", "> $1000")
	NextStage string `json:"nextStage" yaml:"nextStage"` // Target stage ID
}

// ExecutorSpec defines what runs in a stage.
type ExecutorSpec struct {
	Name string       `json:"name" yaml:"name"`
	Type ExecutorType `json:"type" yaml:"type"`

	// Type-specific metadata
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"` // For API
	Command  string `json:"command,omitempty" yaml:"command,omitempty"`   // For program
	Model    string `json:"model,omitempty" yaml:"model,omitempty"`       // For LLM
	Prompt   string `json:"prompt,omitempty" yaml:"prompt,omitempty"`     // For LLM/agent
	Agent    string `json:"agent,omitempty" yaml:"agent,omitempty"`       // For agent type
}

// ExecutorType represents the type of executor.
type ExecutorType string

const (
	// ExecutorProgram is an external program or binary.
	ExecutorProgram ExecutorType = "program"

	// ExecutorAPI is a REST/gRPC API call.
	ExecutorAPI ExecutorType = "api"

	// ExecutorDeterministic is custom code where same input = same output.
	ExecutorDeterministic ExecutorType = "deterministic"

	// ExecutorLLM is LLM inference (non-deterministic).
	ExecutorLLM ExecutorType = "llm"

	// ExecutorAgent is an autonomous agent execution.
	ExecutorAgent ExecutorType = "agent"
)

// Label returns a human-readable label for the executor type.
func (t ExecutorType) Label() string {
	switch t {
	case ExecutorProgram:
		return "Program"
	case ExecutorAPI:
		return "API"
	case ExecutorDeterministic:
		return "Deterministic"
	case ExecutorLLM:
		return "LLM"
	case ExecutorAgent:
		return "Agent"
	default:
		return string(t)
	}
}

// Color returns a suggested fill color for the executor type.
func (t ExecutorType) Color() string {
	switch t {
	case ExecutorProgram:
		return "#bbdefb" // Light blue
	case ExecutorAPI:
		return "#c8e6c9" // Light green
	case ExecutorDeterministic:
		return "#b3e5fc" // Lighter blue
	case ExecutorLLM:
		return "#e1bee7" // Light purple
	case ExecutorAgent:
		return "#f8bbd9" // Light pink
	default:
		return "#e0e0e0" // Gray
	}
}

// ResourceSpec defines an input or output resource.
type ResourceSpec struct {
	ID       string       `json:"id" yaml:"id"`
	Label    string       `json:"label,omitempty" yaml:"label,omitempty"`
	Kind     ResourceKind `json:"kind" yaml:"kind"`
	Schema   string       `json:"schema,omitempty" yaml:"schema,omitempty"`     // JSON Schema reference
	Required bool         `json:"required,omitempty" yaml:"required,omitempty"` // For inputs
}

// ResourceKind represents the type of resource.
type ResourceKind string

const (
	// ResourceData is in-memory data structure.
	ResourceData ResourceKind = "data"

	// ResourceFile is a file on disk.
	ResourceFile ResourceKind = "file"

	// ResourceProgram is an executable or script.
	ResourceProgram ResourceKind = "program"

	// ResourceConfig is configuration data.
	ResourceConfig ResourceKind = "config"

	// ResourceModel is ML model weights.
	ResourceModel ResourceKind = "model"

	// ResourcePrompt is a prompt template.
	ResourcePrompt ResourceKind = "prompt"

	// ResourceArtifact is a build artifact.
	ResourceArtifact ResourceKind = "artifact"
)

// Shape returns the D2 shape for the resource kind.
func (k ResourceKind) Shape() string {
	switch k {
	case ResourceData:
		return "cylinder"
	case ResourceFile:
		return "document"
	case ResourceProgram:
		return "hexagon"
	case ResourceConfig:
		return "page"
	case ResourceModel:
		return "package"
	case ResourcePrompt:
		return "page"
	case ResourceArtifact:
		return "package"
	default:
		return "rectangle"
	}
}

// Icon returns a suggested icon identifier for the resource kind.
// Currently returns empty strings to avoid embedding external icon URLs.
func (k ResourceKind) Icon() string {
	return ""
}

// FlowSpec defines data flow between stages.
type FlowSpec struct {
	From      string `json:"from" yaml:"from"`                               // stage.output ID path
	To        string `json:"to" yaml:"to"`                                   // stage.input ID path
	Label     string `json:"label,omitempty" yaml:"label,omitempty"`         // Edge label
	Transform string `json:"transform,omitempty" yaml:"transform,omitempty"` // Optional transformation description
	Async     bool   `json:"async,omitempty" yaml:"async,omitempty"`         // Async data flow
}
