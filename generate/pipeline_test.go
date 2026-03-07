package generate

import (
	"strings"
	"testing"
)

func TestPipelineGenerator_SimpleLinear(t *testing.T) {
	spec := &PipelineSpec{
		ID:        "etl-pipeline",
		Label:     "ETL Pipeline",
		Direction: "right",
		Stages: []StageSpec{
			{
				ID:    "extract",
				Label: "Extract Data",
				Executor: ExecutorSpec{
					Name: "extract.py",
					Type: ExecutorDeterministic,
				},
				Inputs: []ResourceSpec{
					{ID: "source_db", Label: "Source DB", Kind: ResourceData, Required: true},
					{ID: "config", Label: "Config", Kind: ResourceConfig},
				},
				Outputs: []ResourceSpec{
					{ID: "raw_data", Label: "Raw Data", Kind: ResourceData},
				},
			},
			{
				ID:    "transform",
				Label: "Transform",
				Executor: ExecutorSpec{
					Name:  "GPT-4",
					Type:  ExecutorLLM,
					Model: "gpt-4-turbo",
				},
				Inputs: []ResourceSpec{
					{ID: "data", Label: "Raw Data", Kind: ResourceData},
					{ID: "prompt", Label: "Transform Prompt", Kind: ResourcePrompt},
				},
				Outputs: []ResourceSpec{
					{ID: "transformed", Label: "Transformed Data", Kind: ResourceData},
				},
			},
			{
				ID:    "load",
				Label: "Load",
				Executor: ExecutorSpec{
					Name:     "Data API",
					Type:     ExecutorAPI,
					Endpoint: "https://api.example.com/load",
				},
				Inputs: []ResourceSpec{
					{ID: "data", Label: "Data", Kind: ResourceData},
				},
				Outputs: []ResourceSpec{
					{ID: "result", Label: "Load Result", Kind: ResourceArtifact},
				},
			},
		},
	}

	gen := NewPipelineGenerator()
	output := gen.Generate(spec)

	// Verify key elements
	if !strings.Contains(output, "direction: right") {
		t.Error("Expected direction: right")
	}
	if !strings.Contains(output, "extract:") {
		t.Error("Expected extract stage")
	}
	if !strings.Contains(output, "transform:") {
		t.Error("Expected transform stage")
	}
	if !strings.Contains(output, "load:") {
		t.Error("Expected load stage")
	}
	if !strings.Contains(output, "shape: hexagon") {
		t.Error("Expected hexagon shape for executor")
	}
	if !strings.Contains(output, `"LLM"`) {
		t.Error("Expected LLM type label")
	}
}

func TestPipelineGenerator_FanOutFanIn(t *testing.T) {
	spec := &PipelineSpec{
		ID:    "parallel-pipeline",
		Label: "Parallel Processing",
		Stages: []StageSpec{
			{
				ID:    "split",
				Label: "Split Data",
				Executor: ExecutorSpec{
					Name: "splitter",
					Type: ExecutorDeterministic,
				},
				Parallel: []StageSpec{
					{
						ID:    "process_a",
						Label: "Process A",
						Executor: ExecutorSpec{
							Name: "worker_a",
							Type: ExecutorProgram,
						},
					},
					{
						ID:    "process_b",
						Label: "Process B",
						Executor: ExecutorSpec{
							Name: "worker_b",
							Type: ExecutorProgram,
						},
					},
				},
			},
			{
				ID:    "merge",
				Label: "Merge Results",
				Executor: ExecutorSpec{
					Name: "merger",
					Type: ExecutorDeterministic,
				},
				JoinFrom: []string{"process_a", "process_b"},
			},
		},
	}

	gen := NewPipelineGenerator()
	output := gen.Generate(spec)

	// Verify parallel stages
	if !strings.Contains(output, "process_a:") {
		t.Error("Expected process_a parallel stage")
	}
	if !strings.Contains(output, "process_b:") {
		t.Error("Expected process_b parallel stage")
	}

	// Verify fan-in connections use fully qualified paths
	if !strings.Contains(output, "split.process_a -> merge") {
		t.Errorf("Expected qualified fan-in connection 'split.process_a -> merge', got:\n%s", output)
	}
	if !strings.Contains(output, "split.process_b -> merge") {
		t.Errorf("Expected qualified fan-in connection 'split.process_b -> merge', got:\n%s", output)
	}
}

func TestExecutorType_Label(t *testing.T) {
	tests := []struct {
		typ  ExecutorType
		want string
	}{
		{ExecutorProgram, "Program"},
		{ExecutorAPI, "API"},
		{ExecutorDeterministic, "Deterministic"},
		{ExecutorLLM, "LLM"},
		{ExecutorAgent, "Agent"},
	}

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			if got := tt.typ.Label(); got != tt.want {
				t.Errorf("ExecutorType.Label() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceKind_Shape(t *testing.T) {
	tests := []struct {
		kind ResourceKind
		want string
	}{
		{ResourceData, "cylinder"},
		{ResourceFile, "document"},
		{ResourceProgram, "hexagon"},
		{ResourceConfig, "page"},
		{ResourcePrompt, "page"},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			if got := tt.kind.Shape(); got != tt.want {
				t.Errorf("ResourceKind.Shape() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineGenerator_WithFlows(t *testing.T) {
	spec := &PipelineSpec{
		ID: "flow-test",
		Stages: []StageSpec{
			{
				ID: "stage1",
				Executor: ExecutorSpec{
					Name: "step1",
					Type: ExecutorDeterministic,
				},
				Outputs: []ResourceSpec{
					{ID: "output1", Label: "Output 1", Kind: ResourceData},
				},
			},
			{
				ID: "stage2",
				Executor: ExecutorSpec{
					Name: "step2",
					Type: ExecutorDeterministic,
				},
				Inputs: []ResourceSpec{
					{ID: "input2", Label: "Input 2", Kind: ResourceData},
				},
			},
		},
		Flows: []FlowSpec{
			{
				From:  "stage1.outputs.output1",
				To:    "stage2.inputs.input2",
				Label: "processed data",
				Async: true,
			},
		},
	}

	gen := NewPipelineGenerator()
	output := gen.Generate(spec)

	// Verify flow
	if !strings.Contains(output, "stage1.outputs.output1 -> stage2.inputs.input2") {
		t.Error("Expected explicit flow connection")
	}
	if !strings.Contains(output, "stroke-dash") {
		t.Error("Expected dashed stroke for async flow")
	}
}

func TestPipelineGenerator_SimpleMode(t *testing.T) {
	spec := &PipelineSpec{
		ID:        "simple-test",
		Direction: "right",
		Stages: []StageSpec{
			{
				ID:    "extract",
				Label: "Extract",
				Executor: ExecutorSpec{
					Name: "extractor",
					Type: ExecutorDeterministic,
				},
				Inputs: []ResourceSpec{
					{ID: "source", Label: "Source", Kind: ResourceData},
				},
			},
			{
				ID:    "transform",
				Label: "Transform",
				Executor: ExecutorSpec{
					Name: "GPT-4",
					Type: ExecutorLLM,
				},
			},
		},
	}

	gen := NewPipelineGenerator()
	opts := PipelineRenderOptions{Simple: true}
	output := gen.GenerateWithOptions(spec, opts)

	// Verify simple mode output
	if !strings.Contains(output, "shape: rectangle") {
		t.Error("Expected rectangle shape in simple mode")
	}
	if !strings.Contains(output, `extract.type: "Deterministic"`) {
		t.Error("Expected type badge for extract stage")
	}
	if !strings.Contains(output, `transform.type: "LLM"`) {
		t.Error("Expected type badge for transform stage")
	}
	// Simple mode should NOT contain inputs/outputs containers
	if strings.Contains(output, "inputs: Inputs") {
		t.Error("Simple mode should not have inputs container")
	}
	if strings.Contains(output, "outputs: Outputs") {
		t.Error("Simple mode should not have outputs container")
	}
}

func TestPipelineGenerator_SwimlanesMode(t *testing.T) {
	spec := &PipelineSpec{
		ID:        "swimlane-test",
		Direction: "right",
		Stages: []StageSpec{
			{
				ID:    "receive",
				Label: "Receive Order",
				Lane:  "Sales",
				Executor: ExecutorSpec{
					Name: "order-api",
					Type: ExecutorAPI,
				},
			},
			{
				ID:    "validate",
				Label: "Validate",
				Lane:  "Sales",
				Executor: ExecutorSpec{
					Name: "validator",
					Type: ExecutorDeterministic,
				},
			},
			{
				ID:    "charge",
				Label: "Charge Card",
				Lane:  "Finance",
				Executor: ExecutorSpec{
					Name: "payment-api",
					Type: ExecutorAPI,
				},
			},
		},
	}

	gen := NewPipelineGenerator()
	output := gen.Generate(spec)

	// Verify lane containers
	if !strings.Contains(output, "Sales: Sales {") {
		t.Error("Expected Sales lane container")
	}
	if !strings.Contains(output, "Finance: Finance {") {
		t.Error("Expected Finance lane container")
	}

	// Verify cross-lane edges use qualified IDs
	if !strings.Contains(output, "Sales.validate -> Finance.charge") {
		t.Errorf("Expected cross-lane edge 'Sales.validate -> Finance.charge', got:\n%s", output)
	}
}

func TestPipelineGenerator_DecisionMode(t *testing.T) {
	spec := &PipelineSpec{
		ID:        "decision-test",
		Direction: "right",
		Stages: []StageSpec{
			{
				ID:    "receive",
				Label: "Receive Order",
				Executor: ExecutorSpec{
					Name: "order-api",
					Type: ExecutorAPI,
				},
			},
			{
				ID:    "check_stock",
				Label: "In Stock?",
				Executor: ExecutorSpec{
					Name: "inventory-check",
					Type: ExecutorAPI,
				},
				Branches: []BranchSpec{
					{Label: "Yes", NextStage: "ship"},
					{Label: "No", NextStage: "backorder"},
				},
			},
			{
				ID:    "ship",
				Label: "Ship Order",
				Executor: ExecutorSpec{
					Name: "shipping-api",
					Type: ExecutorAPI,
				},
			},
			{
				ID:    "backorder",
				Label: "Create Backorder",
				Executor: ExecutorSpec{
					Name: "backorder-api",
					Type: ExecutorAPI,
				},
			},
		},
	}

	gen := NewPipelineGenerator()
	opts := PipelineRenderOptions{Simple: true}
	output := gen.GenerateWithOptions(spec, opts)

	// Verify decision node has diamond shape
	if !strings.Contains(output, "shape: diamond") {
		t.Error("Expected diamond shape for decision node")
	}
	if !strings.Contains(output, "#fff9c4") {
		t.Error("Expected yellow fill for decision node")
	}

	// Verify branch edges
	if !strings.Contains(output, "check_stock -> ship: Yes") {
		t.Errorf("Expected 'check_stock -> ship: Yes' branch edge, got:\n%s", output)
	}
	if !strings.Contains(output, "check_stock -> backorder: No") {
		t.Errorf("Expected 'check_stock -> backorder: No' branch edge, got:\n%s", output)
	}

	// Verify no auto-edge from decision node
	if strings.Contains(output, "check_stock -> ship\n") && !strings.Contains(output, ": Yes") {
		t.Error("Should not have unlabeled auto-edge from decision node to next stage")
	}
}

func TestPipelineGenerator_SwimlanesWithDecisions(t *testing.T) {
	spec := &PipelineSpec{
		ID:        "combined-test",
		Direction: "right",
		Stages: []StageSpec{
			{
				ID:    "receive",
				Label: "Receive",
				Lane:  "Sales",
				Executor: ExecutorSpec{
					Name: "api",
					Type: ExecutorAPI,
				},
			},
			{
				ID:    "decide",
				Label: "Approve?",
				Lane:  "Sales",
				Executor: ExecutorSpec{
					Name: "approver",
					Type: ExecutorAgent,
				},
				Branches: []BranchSpec{
					{Label: "Yes", NextStage: "process"},
					{Label: "No", NextStage: "reject"},
				},
			},
			{
				ID:    "process",
				Label: "Process",
				Lane:  "Operations",
				Executor: ExecutorSpec{
					Name: "processor",
					Type: ExecutorDeterministic,
				},
			},
			{
				ID:    "reject",
				Label: "Reject",
				Lane:  "Sales",
				Executor: ExecutorSpec{
					Name: "rejector",
					Type: ExecutorDeterministic,
				},
			},
		},
	}

	gen := NewPipelineGenerator()
	output := gen.Generate(spec)

	// Verify lanes exist
	if !strings.Contains(output, "Sales: Sales {") {
		t.Error("Expected Sales lane")
	}
	if !strings.Contains(output, "Operations: Operations {") {
		t.Error("Expected Operations lane")
	}

	// Verify decision edges cross lanes correctly
	if !strings.Contains(output, "Sales.decide -> Operations.process: Yes") {
		t.Errorf("Expected cross-lane branch 'Sales.decide -> Operations.process: Yes', got:\n%s", output)
	}
	if !strings.Contains(output, "Sales.decide -> Sales.reject: No") {
		t.Errorf("Expected same-lane branch 'Sales.decide -> Sales.reject: No', got:\n%s", output)
	}
}

func TestBranchSpec(t *testing.T) {
	branch := BranchSpec{
		Label:     "Yes",
		NextStage: "approved",
	}

	if branch.Label != "Yes" {
		t.Errorf("Expected Label 'Yes', got '%s'", branch.Label)
	}
	if branch.NextStage != "approved" {
		t.Errorf("Expected NextStage 'approved', got '%s'", branch.NextStage)
	}
}
