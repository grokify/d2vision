package generate

import (
	"fmt"
	"strings"
)

// PipelineRenderOptions controls how the pipeline is rendered.
type PipelineRenderOptions struct {
	Simple bool // Hide internal I/O, show only stage boxes
}

// PipelineGenerator generates D2 code from PipelineSpec.
type PipelineGenerator struct {
	indent string
	sb     strings.Builder
	opts   PipelineRenderOptions
}

// NewPipelineGenerator creates a new pipeline generator.
func NewPipelineGenerator() *PipelineGenerator {
	return &PipelineGenerator{
		indent: "  ",
	}
}

// Generate produces D2 code from a PipelineSpec using default options.
func (g *PipelineGenerator) Generate(spec *PipelineSpec) string {
	return g.GenerateWithOptions(spec, PipelineRenderOptions{})
}

// GenerateWithOptions produces D2 code from a PipelineSpec with custom options.
func (g *PipelineGenerator) GenerateWithOptions(spec *PipelineSpec, opts PipelineRenderOptions) string {
	g.sb.Reset()
	g.opts = opts

	// Check for swimlane mode (auto-detected when lanes are present)
	if g.hasLanes(spec) {
		return g.generateWithLanes(spec)
	}

	// Build a map of parallel substage IDs to their parent stage ID
	// This is used to qualify joinFrom references
	parallelParent := make(map[string]string)
	for _, stage := range spec.Stages {
		for _, p := range stage.Parallel {
			parallelParent[p.ID] = stage.ID
		}
	}

	// Build set of decision node targets (stages that are branch targets)
	branchTargets := g.buildBranchTargets(spec)

	// Direction
	direction := spec.Direction
	if direction == "" {
		direction = "right"
	}
	g.writef("direction: %s\n\n", direction)

	// Generate stages
	for i, stage := range spec.Stages {
		if g.opts.Simple {
			g.generateStageSimple(&stage, 0)
		} else {
			g.generateStage(&stage, 0)
		}
		if i < len(spec.Stages)-1 {
			g.sb.WriteString("\n")
		}
	}

	// Generate cross-stage flows
	if len(spec.Flows) > 0 {
		g.sb.WriteString("\n# Data Flows\n")
		for _, flow := range spec.Flows {
			g.generateFlow(&flow)
		}
	}

	// Generate sequential stage connections (if no explicit flows)
	if len(spec.Flows) == 0 && len(spec.Stages) > 1 {
		g.sb.WriteString("\n# Stage Sequence\n")
		for i := 0; i < len(spec.Stages)-1; i++ {
			current := &spec.Stages[i]
			next := &spec.Stages[i+1]

			// Skip auto-connection if current is a decision node (it has its own branch edges)
			if g.isDecisionNode(current) {
				continue
			}

			// Skip auto-connection if next is a branch target (connected via decision node)
			if branchTargets[next.ID] {
				continue
			}

			// Handle fan-in
			if len(next.JoinFrom) > 0 {
				for _, fromID := range next.JoinFrom {
					// Qualify substage IDs with their parent stage ID
					qualifiedID := fromID
					if parentID, ok := parallelParent[fromID]; ok {
						qualifiedID = parentID + "." + fromID
					}
					g.writef("%s -> %s\n", qualifiedID, next.ID)
				}
			} else if len(current.Parallel) > 0 {
				// Fan-out to parallel stages, then fan-in to next
				// These are nested inside current, so qualify with current.ID
				for _, p := range current.Parallel {
					g.writef("%s.%s -> %s\n", current.ID, p.ID, next.ID)
				}
			} else {
				g.writef("%s -> %s\n", current.ID, next.ID)
			}
		}

		// Generate decision node branch edges
		g.generateDecisionEdges(spec.Stages, "")
	}

	return g.sb.String()
}

func (g *PipelineGenerator) generateStage(stage *StageSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	// Stage container
	label := stage.Label
	if label == "" {
		label = stage.ID
	}
	g.writef("%s%s: %s {\n", indent, g.escapeID(stage.ID), g.escapeLabel(label))

	innerIndent := indent + g.indent

	// Stage styling based on executor type
	g.writef("%sstyle.fill: \"%s\"\n", innerIndent, stage.Executor.Type.Color())
	g.writef("%sstyle.border-radius: 8\n", innerIndent)

	// Grid layout for inputs -> executor -> outputs
	hasInputs := len(stage.Inputs) > 0
	hasOutputs := len(stage.Outputs) > 0

	if hasInputs || hasOutputs {
		g.writef("%sgrid-columns: 3\n\n", innerIndent)
	}

	// Inputs container
	if hasInputs {
		g.writef("%sinputs: Inputs {\n", innerIndent)
		g.writef("%s  style.fill: \"#ffffff\"\n", innerIndent)
		g.writef("%s  style.border-radius: 4\n", innerIndent)
		for _, input := range stage.Inputs {
			g.generateResource(&input, depth+2, true)
		}
		g.writef("%s}\n\n", innerIndent)
	}

	// Executor
	g.generateExecutor(&stage.Executor, depth+1)

	// Outputs container
	if hasOutputs {
		g.writef("\n%soutputs: Outputs {\n", innerIndent)
		g.writef("%s  style.fill: \"#ffffff\"\n", innerIndent)
		g.writef("%s  style.border-radius: 4\n", innerIndent)
		for _, output := range stage.Outputs {
			g.generateResource(&output, depth+2, false)
		}
		g.writef("%s}\n", innerIndent)
	}

	// Internal connections
	if hasInputs {
		g.writef("\n%sinputs -> executor\n", innerIndent)
	}
	if hasOutputs {
		g.writef("%sexecutor -> outputs\n", innerIndent)
	}

	// Handle parallel stages (fan-out)
	if len(stage.Parallel) > 0 {
		g.writef("\n%s# Parallel Execution\n", innerIndent)
		for _, p := range stage.Parallel {
			g.generateStage(&p, depth+1)
		}

		// Connect executor to parallel stages
		for _, p := range stage.Parallel {
			g.writef("%sexecutor -> %s\n", innerIndent, g.escapeID(p.ID))
		}
	}

	g.writef("%s}\n", indent)
}

func (g *PipelineGenerator) generateExecutor(exec *ExecutorSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	// Executor node
	g.writef("%sexecutor: %s {\n", indent, g.escapeLabel(exec.Name))
	g.writef("%s  shape: hexagon\n", indent)
	g.writef("%s  style.fill: \"%s\"\n", indent, exec.Type.Color())
	g.writef("%s  style.font-size: 14\n", indent)
	g.writef("%s}\n", indent)

	// Type label below executor
	g.writef("%sexecutor.type: \"%s\"\n", indent, exec.Type.Label())

	// Additional metadata as tooltip/label
	if exec.Model != "" {
		g.writef("%sexecutor.model: \"Model: %s\"\n", indent, exec.Model)
	}
	if exec.Endpoint != "" {
		g.writef("%sexecutor.endpoint: \"%s\"\n", indent, exec.Endpoint)
	}
}

func (g *PipelineGenerator) generateResource(res *ResourceSpec, depth int, isInput bool) {
	indent := strings.Repeat(g.indent, depth)

	label := res.Label
	if label == "" {
		label = res.ID
	}

	g.writef("%s%s: %s {\n", indent, g.escapeID(res.ID), g.escapeLabel(label))
	g.writef("%s  shape: %s\n", indent, res.Kind.Shape())

	// Color coding for inputs vs outputs
	if isInput {
		if res.Required {
			g.writef("%s  style.fill: \"#ffebee\"\n", indent) // Light red for required
		} else {
			g.writef("%s  style.fill: \"#e8f5e9\"\n", indent) // Light green for optional
		}
	} else {
		g.writef("%s  style.fill: \"#e3f2fd\"\n", indent) // Light blue for outputs
	}

	g.writef("%s}\n", indent)
}

func (g *PipelineGenerator) generateFlow(flow *FlowSpec) {
	edge := fmt.Sprintf("%s -> %s", flow.From, flow.To)

	if flow.Label != "" || flow.Transform != "" || flow.Async {
		g.writef("%s: {\n", edge)
		if flow.Label != "" {
			g.writef("  label: %s\n", g.escapeLabel(flow.Label))
		}
		if flow.Async {
			g.writef("  style.stroke-dash: 5\n")
		}
		g.sb.WriteString("}\n")
	} else {
		g.writef("%s\n", edge)
	}
}

func (g *PipelineGenerator) writef(format string, args ...any) {
	fmt.Fprintf(&g.sb, format, args...)
}

func (g *PipelineGenerator) escapeID(id string) string {
	// Quote IDs with special characters
	if strings.ContainsAny(id, " -.:") {
		return fmt.Sprintf("%q", id)
	}
	return id
}

func (g *PipelineGenerator) escapeLabel(label string) string {
	// Always quote labels that might have special characters
	if strings.ContainsAny(label, " -.:{}[]<>") || strings.Contains(label, "\n") {
		return fmt.Sprintf("%q", label)
	}
	return label
}

// GeneratePipeline is a convenience function to generate D2 from a PipelineSpec.
func GeneratePipeline(spec *PipelineSpec) string {
	gen := NewPipelineGenerator()
	return gen.Generate(spec)
}

// GeneratePipelineWithOptions is a convenience function to generate D2 with options.
func GeneratePipelineWithOptions(spec *PipelineSpec, opts PipelineRenderOptions) string {
	gen := NewPipelineGenerator()
	return gen.GenerateWithOptions(spec, opts)
}

// hasLanes checks if any stage has a lane assigned.
func (g *PipelineGenerator) hasLanes(spec *PipelineSpec) bool {
	for _, stage := range spec.Stages {
		if stage.Lane != "" {
			return true
		}
	}
	return false
}

// isDecisionNode checks if a stage is a decision node (has branches).
func (g *PipelineGenerator) isDecisionNode(stage *StageSpec) bool {
	return len(stage.Branches) > 0
}

// buildBranchTargets returns a set of stage IDs that are targets of decision branches.
func (g *PipelineGenerator) buildBranchTargets(spec *PipelineSpec) map[string]bool {
	targets := make(map[string]bool)
	for _, stage := range spec.Stages {
		for _, branch := range stage.Branches {
			targets[branch.NextStage] = true
		}
	}
	return targets
}

// generateStageSimple generates a compact stage without I/O breakdown.
func (g *PipelineGenerator) generateStageSimple(stage *StageSpec, depth int) {
	indent := strings.Repeat(g.indent, depth)

	label := stage.Label
	if label == "" {
		label = stage.ID
	}

	// Check if this is a decision node
	if g.isDecisionNode(stage) {
		// Decision node gets diamond shape
		g.writef("%s%s: %s {\n", indent, g.escapeID(stage.ID), g.escapeLabel(label))
		g.writef("%s  shape: diamond\n", indent)
		g.writef("%s  style.fill: \"#fff9c4\"\n", indent) // Yellow for decision
		g.writef("%s}\n", indent)
	} else {
		// Regular stage box
		g.writef("%s%s: %s {\n", indent, g.escapeID(stage.ID), g.escapeLabel(label))
		g.writef("%s  shape: rectangle\n", indent)
		g.writef("%s  style.fill: \"%s\"\n", indent, stage.Executor.Type.Color())
		g.writef("%s  style.border-radius: 8\n", indent)
		g.writef("%s}\n", indent)

		// Add type badge below the stage
		g.writef("%s%s.type: \"%s\"\n", indent, g.escapeID(stage.ID), stage.Executor.Type.Label())
	}

	// Handle parallel stages (fan-out)
	if len(stage.Parallel) > 0 {
		for _, p := range stage.Parallel {
			g.generateStageSimple(&p, depth)
		}
		// Connect main stage to parallel stages
		for _, p := range stage.Parallel {
			g.writef("%s%s -> %s\n", indent, g.escapeID(stage.ID), g.escapeID(p.ID))
		}
	}
}

// generateWithLanes generates D2 with stages grouped into swimlane containers.
func (g *PipelineGenerator) generateWithLanes(spec *PipelineSpec) string {
	// Direction
	direction := spec.Direction
	if direction == "" {
		direction = "right"
	}
	g.writef("direction: %s\n\n", direction)

	// Group stages by lane
	lanes := make(map[string][]*StageSpec)
	var laneOrder []string
	for i := range spec.Stages {
		stage := &spec.Stages[i]
		lane := stage.Lane
		if lane == "" {
			lane = "Default"
		}
		if _, exists := lanes[lane]; !exists {
			laneOrder = append(laneOrder, lane)
		}
		lanes[lane] = append(lanes[lane], stage)
	}

	// Build branch targets for skipping auto-edges
	branchTargets := g.buildBranchTargets(spec)

	// Generate lane containers
	for _, lane := range laneOrder {
		stages := lanes[lane]
		g.writef("%s: %s {\n", g.escapeID(lane), g.escapeLabel(lane))
		g.writef("  style.fill: \"#f5f5f5\"\n")
		g.writef("  style.border-radius: 8\n\n")

		for i, stage := range stages {
			if g.opts.Simple {
				g.generateStageSimple(stage, 1)
			} else {
				g.generateStage(stage, 1)
			}
			if i < len(stages)-1 {
				g.sb.WriteString("\n")
			}
		}

		g.sb.WriteString("}\n\n")
	}

	// Generate cross-lane edges
	if len(spec.Flows) > 0 {
		g.sb.WriteString("# Data Flows\n")
		for _, flow := range spec.Flows {
			g.generateFlowWithLanes(&flow, spec)
		}
	}

	// Generate sequential connections with lane-qualified IDs
	if len(spec.Flows) == 0 && len(spec.Stages) > 1 {
		g.sb.WriteString("# Stage Sequence\n")

		// Build stage-to-lane mapping
		stageToLane := make(map[string]string)
		for i := range spec.Stages {
			stage := &spec.Stages[i]
			lane := stage.Lane
			if lane == "" {
				lane = "Default"
			}
			stageToLane[stage.ID] = lane
		}

		for i := 0; i < len(spec.Stages)-1; i++ {
			current := &spec.Stages[i]
			next := &spec.Stages[i+1]

			// Skip if current is a decision node
			if g.isDecisionNode(current) {
				continue
			}

			// Skip if next is a branch target
			if branchTargets[next.ID] {
				continue
			}

			currentLane := stageToLane[current.ID]
			nextLane := stageToLane[next.ID]

			currentQualified := fmt.Sprintf("%s.%s", g.escapeID(currentLane), g.escapeID(current.ID))
			nextQualified := fmt.Sprintf("%s.%s", g.escapeID(nextLane), g.escapeID(next.ID))

			g.writef("%s -> %s\n", currentQualified, nextQualified)
		}

		// Generate decision edges with lane qualification
		g.generateDecisionEdgesWithLanes(spec.Stages, stageToLane)
	}

	return g.sb.String()
}

// generateDecisionEdges generates edges for all decision node branches.
func (g *PipelineGenerator) generateDecisionEdges(stages []StageSpec, lanePrefix string) {
	for _, stage := range stages {
		if g.isDecisionNode(&stage) {
			fromID := stage.ID
			if lanePrefix != "" {
				fromID = lanePrefix + "." + stage.ID
			}

			for _, branch := range stage.Branches {
				toID := branch.NextStage
				if lanePrefix != "" {
					toID = lanePrefix + "." + branch.NextStage
				}

				if branch.Label != "" {
					g.writef("%s -> %s: %s\n", g.escapeID(fromID), g.escapeID(toID), g.escapeLabel(branch.Label))
				} else {
					g.writef("%s -> %s\n", g.escapeID(fromID), g.escapeID(toID))
				}
			}
		}
	}
}

// generateDecisionEdgesWithLanes generates decision edges with lane-qualified IDs.
func (g *PipelineGenerator) generateDecisionEdgesWithLanes(stages []StageSpec, stageToLane map[string]string) {
	for _, stage := range stages {
		if g.isDecisionNode(&stage) {
			fromLane := stageToLane[stage.ID]
			fromQualified := fmt.Sprintf("%s.%s", g.escapeID(fromLane), g.escapeID(stage.ID))

			for _, branch := range stage.Branches {
				toLane := stageToLane[branch.NextStage]
				toQualified := fmt.Sprintf("%s.%s", g.escapeID(toLane), g.escapeID(branch.NextStage))

				if branch.Label != "" {
					g.writef("%s -> %s: %s\n", fromQualified, toQualified, g.escapeLabel(branch.Label))
				} else {
					g.writef("%s -> %s\n", fromQualified, toQualified)
				}
			}
		}
	}
}

// generateFlowWithLanes generates a flow edge with lane-qualified IDs.
func (g *PipelineGenerator) generateFlowWithLanes(flow *FlowSpec, spec *PipelineSpec) {
	// Build stage-to-lane mapping
	stageToLane := make(map[string]string)
	for i := range spec.Stages {
		stage := &spec.Stages[i]
		lane := stage.Lane
		if lane == "" {
			lane = "Default"
		}
		stageToLane[stage.ID] = lane
	}

	// Parse stage IDs from flow paths (e.g., "stage1.outputs.output1")
	fromParts := strings.SplitN(flow.From, ".", 2)
	toParts := strings.SplitN(flow.To, ".", 2)

	fromStage := fromParts[0]
	toStage := toParts[0]

	fromLane := stageToLane[fromStage]
	toLane := stageToLane[toStage]

	// Reconstruct qualified paths
	qualifiedFrom := fmt.Sprintf("%s.%s", g.escapeID(fromLane), flow.From)
	qualifiedTo := fmt.Sprintf("%s.%s", g.escapeID(toLane), flow.To)

	edge := fmt.Sprintf("%s -> %s", qualifiedFrom, qualifiedTo)

	if flow.Label != "" || flow.Transform != "" || flow.Async {
		g.writef("%s: {\n", edge)
		if flow.Label != "" {
			g.writef("  label: %s\n", g.escapeLabel(flow.Label))
		}
		if flow.Async {
			g.writef("  style.stroke-dash: 5\n")
		}
		g.sb.WriteString("}\n")
	} else {
		g.writef("%s\n", edge)
	}
}
