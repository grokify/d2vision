package main

import (
	"fmt"
	"strings"

	"github.com/grokify/d2vision/format"
	"github.com/grokify/d2vision/generate"
	"github.com/spf13/cobra"
)

var (
	templateFormat        string
	templateClusters      int
	templateServices      int
	templateOutputD2      bool
	templatePipelineType  string
	templatePipelineStages int
)

var templateCmd = &cobra.Command{
	Use:   "template <name>",
	Short: "Generate diagram templates for common patterns",
	Long: `Generate diagram templates for common architectural patterns.

Available templates:
  network-boundary   Side-by-side network zones with services and datastores
  microservices      Service mesh with API gateway
  data-flow          ETL/data pipeline
  sequence           Request/response sequence diagram
  entity-relationship Database schema with SQL tables
  deployment         Cloud deployment architecture
  pipeline           Multi-stage process pipeline (LLM/agent workflow)

Pipeline Template (--pipeline-type):
  The pipeline template generates multi-stage process diagrams with:
  - Inputs: data, files, configs, prompts entering each stage
  - Executor: the process that runs (program, API, LLM, or agent)
  - Outputs: artifacts, data, files produced by the stage
  - Parallelism: fan-out/fan-in for concurrent execution

  Pipeline types:
    etl    - Extract/Transform/Load data pipeline (deterministic)
    llm    - LLM processing pipeline (document analysis, embeddings)
    agent  - Autonomous agent pipeline with parallel workers

  Executor types shown in diagrams:
    ⚙️ Program       - External program/binary
    🌐 API           - REST/gRPC API call
    📐 Deterministic - Custom code (same input = same output)
    🤖 LLM           - Language model inference (non-deterministic)
    🧠 Agent         - Autonomous agent execution

Output:
  - Default: TOON spec (can be modified and piped to 'generate')
  - With --d2: Ready-to-render D2 code

Examples:
  # Generate D2 code directly
  d2vision template network-boundary --d2

  # Customize and generate
  d2vision template network-boundary --clusters 3 --services 2 --d2 | d2 - output.svg

  # Generate ETL pipeline (default)
  d2vision template pipeline --d2

  # Generate LLM processing pipeline
  d2vision template pipeline --d2 --pipeline-type llm

  # Generate agent pipeline with parallel workers
  d2vision template pipeline --d2 --pipeline-type agent

  # Get pipeline spec as JSON for modification
  d2vision template pipeline --format json > my-pipeline.json
  # Edit my-pipeline.json, then generate D2:
  d2vision generate my-pipeline.json --format json > pipeline.d2

  # Full pipeline: generate and render to SVG
  d2vision template pipeline --d2 --pipeline-type agent | d2 - agent-workflow.svg
`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplate,
}

func init() {
	templateCmd.Flags().StringVarP(&templateFormat, "format", "f", "toon", "Output format: toon, json, yaml")
	templateCmd.Flags().IntVar(&templateClusters, "clusters", 2, "Number of clusters (for network-boundary)")
	templateCmd.Flags().IntVar(&templateServices, "services", 2, "Services per cluster (for network-boundary)")
	templateCmd.Flags().BoolVar(&templateOutputD2, "d2", false, "Output D2 code instead of spec")
	templateCmd.Flags().StringVar(&templatePipelineType, "pipeline-type", "etl", "Pipeline type: etl, llm, agent (for pipeline)")
	templateCmd.Flags().IntVar(&templatePipelineStages, "stages", 3, "Number of stages (for pipeline)")
}

func runTemplate(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	var spec *generate.DiagramSpec

	switch templateName {
	case "network-boundary":
		spec = generateNetworkBoundaryTemplate(templateClusters, templateServices)
	case "microservices":
		spec = generateMicroservicesTemplate()
	case "data-flow":
		spec = generateDataFlowTemplate()
	case "sequence":
		spec = generateSequenceTemplate()
	case "entity-relationship", "er":
		spec = generateERTemplate()
	case "deployment":
		spec = generateDeploymentTemplate()
	case "pipeline":
		return runPipelineTemplate()
	case "list":
		fmt.Println("Available templates:")
		fmt.Println("  network-boundary     Side-by-side network zones with services and datastores")
		fmt.Println("  microservices        Service mesh with API gateway")
		fmt.Println("  data-flow            ETL/data pipeline")
		fmt.Println("  sequence             Request/response sequence diagram")
		fmt.Println("  entity-relationship  Database schema with SQL tables (alias: er)")
		fmt.Println("  deployment           Cloud deployment architecture")
		fmt.Println("  pipeline             Multi-stage process pipeline (LLM/agent workflow)")
		return nil
	default:
		return fmt.Errorf("unknown template: %s (use 'template list' to see available templates)", templateName)
	}

	// Output D2 code directly
	if templateOutputD2 {
		gen := generate.NewGenerator()
		fmt.Print(gen.Generate(spec))
		return nil
	}

	// Output spec in requested format
	f, err := format.Parse(templateFormat)
	if err != nil {
		return err
	}

	output, err := format.Marshal(spec, f)
	if err != nil {
		return fmt.Errorf("marshaling spec: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// generateNetworkBoundaryTemplate creates a network boundary template.
func generateNetworkBoundaryTemplate(numClusters, servicesPerCluster int) *generate.DiagramSpec {
	spec := &generate.DiagramSpec{
		GridColumns: numClusters,
		Containers:  make([]generate.ContainerSpec, numClusters),
	}

	for i := 0; i < numClusters; i++ {
		clusterNum := i + 1
		clusterID := fmt.Sprintf("cluster%d", clusterNum)

		container := generate.ContainerSpec{
			ID:        clusterID,
			Label:     fmt.Sprintf("Cluster %d", clusterNum),
			Direction: "down",
		}

		// Create services container if multiple services
		if servicesPerCluster > 1 {
			servicesContainer := generate.ContainerSpec{
				ID:        "services",
				Label:     "", // Invisible container
				Direction: "right",
				Style:     &generate.StyleSpec{StrokeWidth: generate.IntPtr(0)},
				Nodes:     make([]generate.NodeSpec, servicesPerCluster),
			}

			for j := 0; j < servicesPerCluster; j++ {
				serviceID := fmt.Sprintf("service%d%s", clusterNum, string(rune('a'+j)))
				servicesContainer.Nodes[j] = generate.NodeSpec{
					ID:    serviceID,
					Label: fmt.Sprintf("Service %d%s", clusterNum, strings.ToUpper(string(rune('a'+j)))),
				}
			}

			container.Containers = []generate.ContainerSpec{servicesContainer}

			// Add edges from services to datastore
			for j := 0; j < servicesPerCluster; j++ {
				serviceID := fmt.Sprintf("services.service%d%s", clusterNum, string(rune('a'+j)))
				container.Edges = append(container.Edges, generate.EdgeSpec{
					From: serviceID,
					To:   fmt.Sprintf("datastore%d", clusterNum),
				})
			}
		} else {
			// Single service
			container.Nodes = []generate.NodeSpec{
				{
					ID:    fmt.Sprintf("service%d", clusterNum),
					Label: fmt.Sprintf("Service %d", clusterNum),
				},
			}
			container.Edges = []generate.EdgeSpec{
				{
					From: fmt.Sprintf("service%d", clusterNum),
					To:   fmt.Sprintf("datastore%d", clusterNum),
				},
			}
		}

		// Add datastore
		container.Nodes = append(container.Nodes, generate.NodeSpec{
			ID:    fmt.Sprintf("datastore%d", clusterNum),
			Label: fmt.Sprintf("DataStore %d", clusterNum),
			Shape: "cylinder",
		})

		spec.Containers[i] = container
	}

	// Add cross-cluster replication edge if multiple clusters
	if numClusters >= 2 {
		spec.Edges = []generate.EdgeSpec{
			{
				From:  "cluster2.datastore2",
				To:    "cluster1.datastore1",
				Label: "replication",
			},
		}
	}

	return spec
}

// generateMicroservicesTemplate creates a microservices architecture template.
func generateMicroservicesTemplate() *generate.DiagramSpec {
	return &generate.DiagramSpec{
		Direction: "right",
		Nodes: []generate.NodeSpec{
			{ID: "client", Label: "Client", Shape: "person"},
		},
		Containers: []generate.ContainerSpec{
			{
				ID:        "gateway",
				Label:     "API Gateway",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "auth", Label: "Auth"},
					{ID: "rate_limit", Label: "Rate Limiter"},
					{ID: "router", Label: "Router"},
				},
				Edges: []generate.EdgeSpec{
					{From: "auth", To: "rate_limit"},
					{From: "rate_limit", To: "router"},
				},
			},
			{
				ID:          "services",
				Label:       "Services",
				Direction:   "down",
				GridColumns: 2,
				Nodes: []generate.NodeSpec{
					{ID: "user_svc", Label: "User Service"},
					{ID: "order_svc", Label: "Order Service"},
					{ID: "product_svc", Label: "Product Service"},
					{ID: "payment_svc", Label: "Payment Service"},
				},
			},
			{
				ID:        "data",
				Label:     "Data Layer",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "cache", Label: "Redis Cache", Shape: "cylinder"},
					{ID: "db", Label: "PostgreSQL", Shape: "cylinder"},
					{ID: "queue", Label: "Message Queue", Shape: "queue"},
				},
			},
		},
		Edges: []generate.EdgeSpec{
			{From: "client", To: "gateway.auth"},
			{From: "gateway.router", To: "services.user_svc"},
			{From: "gateway.router", To: "services.order_svc"},
			{From: "gateway.router", To: "services.product_svc"},
			{From: "gateway.router", To: "services.payment_svc"},
			{From: "services.user_svc", To: "data.db"},
			{From: "services.order_svc", To: "data.db"},
			{From: "services.order_svc", To: "data.queue"},
			{From: "services.product_svc", To: "data.cache"},
			{From: "services.payment_svc", To: "data.queue"},
		},
	}
}

// generateDataFlowTemplate creates an ETL/data pipeline template.
func generateDataFlowTemplate() *generate.DiagramSpec {
	return &generate.DiagramSpec{
		Direction: "right",
		Containers: []generate.ContainerSpec{
			{
				ID:        "sources",
				Label:     "Data Sources",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "api", Label: "REST API"},
					{ID: "db", Label: "Database", Shape: "cylinder"},
					{ID: "files", Label: "File System", Shape: "page"},
					{ID: "stream", Label: "Event Stream", Shape: "queue"},
				},
			},
			{
				ID:        "ingestion",
				Label:     "Ingestion",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "collector", Label: "Data Collector"},
					{ID: "validator", Label: "Validator"},
					{ID: "buffer", Label: "Buffer", Shape: "queue"},
				},
				Edges: []generate.EdgeSpec{
					{From: "collector", To: "validator"},
					{From: "validator", To: "buffer"},
				},
			},
			{
				ID:        "processing",
				Label:     "Processing",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "transform", Label: "Transform"},
					{ID: "enrich", Label: "Enrich"},
					{ID: "aggregate", Label: "Aggregate"},
				},
				Edges: []generate.EdgeSpec{
					{From: "transform", To: "enrich"},
					{From: "enrich", To: "aggregate"},
				},
			},
			{
				ID:        "storage",
				Label:     "Storage",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "lake", Label: "Data Lake", Shape: "cylinder"},
					{ID: "warehouse", Label: "Data Warehouse", Shape: "cylinder"},
					{ID: "mart", Label: "Data Mart", Shape: "cylinder"},
				},
				Edges: []generate.EdgeSpec{
					{From: "lake", To: "warehouse"},
					{From: "warehouse", To: "mart"},
				},
			},
			{
				ID:        "consumption",
				Label:     "Consumption",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "bi", Label: "BI Dashboard"},
					{ID: "reports", Label: "Reports", Shape: "page"},
					{ID: "ml", Label: "ML Models"},
				},
			},
		},
		Edges: []generate.EdgeSpec{
			{From: "sources.api", To: "ingestion.collector"},
			{From: "sources.db", To: "ingestion.collector"},
			{From: "sources.files", To: "ingestion.collector"},
			{From: "sources.stream", To: "ingestion.collector"},
			{From: "ingestion.buffer", To: "processing.transform"},
			{From: "processing.aggregate", To: "storage.lake"},
			{From: "storage.mart", To: "consumption.bi"},
			{From: "storage.mart", To: "consumption.reports"},
			{From: "storage.warehouse", To: "consumption.ml"},
		},
	}
}

// generateSequenceTemplate creates a request/response sequence diagram template.
func generateSequenceTemplate() *generate.DiagramSpec {
	return &generate.DiagramSpec{
		Sequences: []generate.SequenceSpec{
			{
				ID:    "auth_flow",
				Label: "Authentication Flow",
				Actors: []generate.ActorSpec{
					{ID: "user", Label: "User", Shape: "person"},
					{ID: "client", Label: "Client App"},
					{ID: "gateway", Label: "API Gateway"},
					{ID: "auth", Label: "Auth Service"},
					{ID: "db", Label: "User DB"},
				},
				Steps: []generate.MessageSpec{
					{From: "user", To: "client", Label: "Enter credentials"},
					{From: "client", To: "gateway", Label: "POST /login"},
					{From: "gateway", To: "auth", Label: "Validate token"},
					{From: "auth", To: "db", Label: "Query user"},
					{From: "db", To: "auth", Label: "User data"},
					{From: "auth", To: "gateway", Label: "JWT token"},
					{From: "gateway", To: "client", Label: "200 OK + token"},
					{From: "client", To: "user", Label: "Login success"},
				},
				Groups: []generate.GroupSpec{
					{
						ID:    "error_case",
						Label: "Invalid Credentials",
						Messages: []generate.MessageSpec{
							{From: "auth", To: "gateway", Label: "401 Unauthorized"},
							{From: "gateway", To: "client", Label: "401 Error"},
						},
					},
				},
			},
		},
	}
}

// generateERTemplate creates an entity-relationship diagram template.
func generateERTemplate() *generate.DiagramSpec {
	return &generate.DiagramSpec{
		Direction: "right",
		Tables: []generate.TableSpec{
			{
				ID:    "users",
				Label: "users",
				Columns: []generate.ColumnSpec{
					{Name: "id", Type: "uuid", Constraints: []string{"PK"}},
					{Name: "email", Type: "varchar(255)", Constraints: []string{"UNQ", "NOT NULL"}},
					{Name: "name", Type: "varchar(100)"},
					{Name: "created_at", Type: "timestamp"},
					{Name: "updated_at", Type: "timestamp"},
				},
			},
			{
				ID:    "orders",
				Label: "orders",
				Columns: []generate.ColumnSpec{
					{Name: "id", Type: "uuid", Constraints: []string{"PK"}},
					{Name: "user_id", Type: "uuid", Constraints: []string{"FK"}},
					{Name: "status", Type: "varchar(20)"},
					{Name: "total", Type: "decimal(10,2)"},
					{Name: "created_at", Type: "timestamp"},
				},
			},
			{
				ID:    "products",
				Label: "products",
				Columns: []generate.ColumnSpec{
					{Name: "id", Type: "uuid", Constraints: []string{"PK"}},
					{Name: "name", Type: "varchar(255)", Constraints: []string{"NOT NULL"}},
					{Name: "price", Type: "decimal(10,2)"},
					{Name: "stock", Type: "int"},
				},
			},
			{
				ID:    "order_items",
				Label: "order_items",
				Columns: []generate.ColumnSpec{
					{Name: "id", Type: "uuid", Constraints: []string{"PK"}},
					{Name: "order_id", Type: "uuid", Constraints: []string{"FK"}},
					{Name: "product_id", Type: "uuid", Constraints: []string{"FK"}},
					{Name: "quantity", Type: "int"},
					{Name: "price", Type: "decimal(10,2)"},
				},
			},
		},
		Edges: []generate.EdgeSpec{
			{From: "orders.user_id", To: "users.id"},
			{From: "order_items.order_id", To: "orders.id"},
			{From: "order_items.product_id", To: "products.id"},
		},
	}
}

// generateDeploymentTemplate creates a cloud deployment architecture template.
func generateDeploymentTemplate() *generate.DiagramSpec {
	return &generate.DiagramSpec{
		GridColumns: 3,
		Containers: []generate.ContainerSpec{
			{
				ID:        "users",
				Label:     "Users",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "web", Label: "Web Users", Shape: "person"},
					{ID: "mobile", Label: "Mobile Users", Shape: "person"},
				},
			},
			{
				ID:        "edge",
				Label:     "Edge Layer",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "cdn", Label: "CDN", Shape: "cloud"},
					{ID: "waf", Label: "WAF"},
					{ID: "lb", Label: "Load Balancer"},
				},
				Edges: []generate.EdgeSpec{
					{From: "cdn", To: "waf"},
					{From: "waf", To: "lb"},
				},
			},
			{
				ID:        "compute",
				Label:     "Compute Layer",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "api", Label: "API Servers"},
					{ID: "workers", Label: "Worker Nodes"},
					{ID: "scheduler", Label: "Job Scheduler"},
				},
				Edges: []generate.EdgeSpec{
					{From: "api", To: "workers"},
					{From: "scheduler", To: "workers"},
				},
			},
			{
				ID:        "data",
				Label:     "Data Layer",
				Direction: "right",
				Nodes: []generate.NodeSpec{
					{ID: "primary", Label: "Primary DB", Shape: "cylinder"},
					{ID: "replica", Label: "Read Replica", Shape: "cylinder"},
					{ID: "cache", Label: "Cache", Shape: "cylinder"},
				},
				Edges: []generate.EdgeSpec{
					{From: "primary", To: "replica", Label: "replication"},
				},
			},
			{
				ID:        "storage",
				Label:     "Storage Layer",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "s3", Label: "Object Storage", Shape: "cylinder"},
					{ID: "logs", Label: "Log Storage", Shape: "cylinder"},
				},
			},
			{
				ID:        "monitoring",
				Label:     "Observability",
				Direction: "down",
				Nodes: []generate.NodeSpec{
					{ID: "metrics", Label: "Metrics"},
					{ID: "traces", Label: "Traces"},
					{ID: "alerts", Label: "Alerting"},
				},
			},
		},
		Edges: []generate.EdgeSpec{
			{From: "users.web", To: "edge.cdn"},
			{From: "users.mobile", To: "edge.cdn"},
			{From: "edge.lb", To: "compute.api"},
			{From: "compute.api", To: "data.cache"},
			{From: "compute.api", To: "data.primary"},
			{From: "compute.workers", To: "data.primary"},
			{From: "compute.api", To: "storage.s3"},
			{From: "compute.api", To: "monitoring.metrics"},
			{From: "compute.workers", To: "monitoring.traces"},
		},
	}
}

// runPipelineTemplate generates a multi-stage process pipeline template.
func runPipelineTemplate() error {
	spec := generatePipelineSpec(templatePipelineType, templatePipelineStages)

	// Output D2 code directly
	if templateOutputD2 {
		gen := generate.NewPipelineGenerator()
		fmt.Print(gen.Generate(spec))
		return nil
	}

	// Output spec in requested format
	f, err := format.Parse(templateFormat)
	if err != nil {
		return err
	}

	output, err := format.Marshal(spec, f)
	if err != nil {
		return fmt.Errorf("marshaling spec: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// generatePipelineSpec creates a pipeline spec based on type.
func generatePipelineSpec(pipelineType string, stages int) *generate.PipelineSpec {
	switch pipelineType {
	case "llm":
		return generateLLMPipelineSpec()
	case "agent":
		return generateAgentPipelineSpec()
	default:
		return generateETLPipelineSpec()
	}
}

// generateETLPipelineSpec creates an ETL pipeline template.
func generateETLPipelineSpec() *generate.PipelineSpec {
	return &generate.PipelineSpec{
		ID:        "etl-pipeline",
		Label:     "ETL Data Pipeline",
		Direction: "right",
		Stages: []generate.StageSpec{
			{
				ID:    "extract",
				Label: "1. Extract",
				Executor: generate.ExecutorSpec{
					Name:    "extract.py",
					Type:    generate.ExecutorDeterministic,
					Command: "python extract.py",
				},
				Inputs: []generate.ResourceSpec{
					{ID: "source_db", Label: "Source Database", Kind: generate.ResourceData, Required: true},
					{ID: "config", Label: "Extract Config", Kind: generate.ResourceConfig, Required: true},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "raw_data", Label: "Raw Data", Kind: generate.ResourceData},
				},
			},
			{
				ID:    "transform",
				Label: "2. Transform",
				Executor: generate.ExecutorSpec{
					Name: "transform.py",
					Type: generate.ExecutorDeterministic,
				},
				Inputs: []generate.ResourceSpec{
					{ID: "data", Label: "Raw Data", Kind: generate.ResourceData, Required: true},
					{ID: "schema", Label: "Output Schema", Kind: generate.ResourceConfig},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "transformed", Label: "Transformed Data", Kind: generate.ResourceData},
				},
			},
			{
				ID:    "load",
				Label: "3. Load",
				Executor: generate.ExecutorSpec{
					Name:     "Data Warehouse API",
					Type:     generate.ExecutorAPI,
					Endpoint: "https://warehouse.example.com/api/load",
				},
				Inputs: []generate.ResourceSpec{
					{ID: "data", Label: "Transformed Data", Kind: generate.ResourceData, Required: true},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "result", Label: "Load Result", Kind: generate.ResourceArtifact},
					{ID: "metrics", Label: "Load Metrics", Kind: generate.ResourceData},
				},
			},
		},
		Flows: []generate.FlowSpec{
			{From: "extract.outputs.raw_data", To: "transform.inputs.data"},
			{From: "transform.outputs.transformed", To: "load.inputs.data"},
		},
	}
}

// generateLLMPipelineSpec creates an LLM processing pipeline template.
func generateLLMPipelineSpec() *generate.PipelineSpec {
	return &generate.PipelineSpec{
		ID:        "llm-pipeline",
		Label:     "LLM Processing Pipeline",
		Direction: "right",
		Stages: []generate.StageSpec{
			{
				ID:    "ingest",
				Label: "1. Ingest Documents",
				Executor: generate.ExecutorSpec{
					Name: "document_loader.py",
					Type: generate.ExecutorDeterministic,
				},
				Inputs: []generate.ResourceSpec{
					{ID: "documents", Label: "Source Documents", Kind: generate.ResourceFile, Required: true},
					{ID: "config", Label: "Parser Config", Kind: generate.ResourceConfig},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "chunks", Label: "Document Chunks", Kind: generate.ResourceData},
				},
			},
			{
				ID:    "embed",
				Label: "2. Generate Embeddings",
				Executor: generate.ExecutorSpec{
					Name:     "OpenAI Embeddings",
					Type:     generate.ExecutorAPI,
					Endpoint: "https://api.openai.com/v1/embeddings",
					Model:    "text-embedding-3-small",
				},
				Inputs: []generate.ResourceSpec{
					{ID: "chunks", Label: "Document Chunks", Kind: generate.ResourceData, Required: true},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "embeddings", Label: "Vector Embeddings", Kind: generate.ResourceData},
				},
			},
			{
				ID:    "analyze",
				Label: "3. LLM Analysis",
				Executor: generate.ExecutorSpec{
					Name:   "GPT-4 Analysis",
					Type:   generate.ExecutorLLM,
					Model:  "gpt-4-turbo",
					Prompt: "Analyze the following documents and extract key insights...",
				},
				Inputs: []generate.ResourceSpec{
					{ID: "chunks", Label: "Document Chunks", Kind: generate.ResourceData, Required: true},
					{ID: "embeddings", Label: "Embeddings", Kind: generate.ResourceData},
					{ID: "prompt", Label: "Analysis Prompt", Kind: generate.ResourcePrompt, Required: true},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "analysis", Label: "Analysis Results", Kind: generate.ResourceData},
					{ID: "summary", Label: "Summary", Kind: generate.ResourceData},
				},
			},
			{
				ID:    "output",
				Label: "4. Generate Report",
				Executor: generate.ExecutorSpec{
					Name: "report_generator.py",
					Type: generate.ExecutorDeterministic,
				},
				Inputs: []generate.ResourceSpec{
					{ID: "analysis", Label: "Analysis Results", Kind: generate.ResourceData, Required: true},
					{ID: "template", Label: "Report Template", Kind: generate.ResourceFile},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "report", Label: "Final Report", Kind: generate.ResourceFile},
				},
			},
		},
	}
}

// generateAgentPipelineSpec creates an autonomous agent pipeline template.
func generateAgentPipelineSpec() *generate.PipelineSpec {
	return &generate.PipelineSpec{
		ID:        "agent-pipeline",
		Label:     "Autonomous Agent Pipeline",
		Direction: "right",
		Stages: []generate.StageSpec{
			{
				ID:    "plan",
				Label: "1. Planning",
				Executor: generate.ExecutorSpec{
					Name:   "Planning Agent",
					Type:   generate.ExecutorAgent,
					Agent:  "planner",
					Model:  "claude-3-opus",
					Prompt: "Create an execution plan for the given task...",
				},
				Inputs: []generate.ResourceSpec{
					{ID: "task", Label: "Task Description", Kind: generate.ResourceData, Required: true},
					{ID: "context", Label: "Context", Kind: generate.ResourceData},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "plan", Label: "Execution Plan", Kind: generate.ResourceData},
					{ID: "subtasks", Label: "Subtasks", Kind: generate.ResourceData},
				},
			},
			{
				ID:    "execute",
				Label: "2. Parallel Execution",
				Executor: generate.ExecutorSpec{
					Name:  "Orchestrator",
					Type:  generate.ExecutorDeterministic,
				},
				Parallel: []generate.StageSpec{
					{
						ID:    "worker_a",
						Label: "Worker A",
						Executor: generate.ExecutorSpec{
							Name:  "Research Agent",
							Type:  generate.ExecutorAgent,
							Agent: "researcher",
							Model: "claude-3-sonnet",
						},
						Inputs: []generate.ResourceSpec{
							{ID: "subtask", Label: "Subtask", Kind: generate.ResourceData},
						},
						Outputs: []generate.ResourceSpec{
							{ID: "result", Label: "Research Result", Kind: generate.ResourceData},
						},
					},
					{
						ID:    "worker_b",
						Label: "Worker B",
						Executor: generate.ExecutorSpec{
							Name:  "Code Agent",
							Type:  generate.ExecutorAgent,
							Agent: "coder",
							Model: "claude-3-sonnet",
						},
						Inputs: []generate.ResourceSpec{
							{ID: "subtask", Label: "Subtask", Kind: generate.ResourceData},
						},
						Outputs: []generate.ResourceSpec{
							{ID: "result", Label: "Code Result", Kind: generate.ResourceData},
						},
					},
				},
			},
			{
				ID:    "synthesize",
				Label: "3. Synthesize Results",
				Executor: generate.ExecutorSpec{
					Name:   "Synthesis Agent",
					Type:   generate.ExecutorAgent,
					Agent:  "synthesizer",
					Model:  "claude-3-opus",
					Prompt: "Combine the results from all workers into a coherent output...",
				},
				JoinFrom: []string{"worker_a", "worker_b"},
				Inputs: []generate.ResourceSpec{
					{ID: "results", Label: "Worker Results", Kind: generate.ResourceData, Required: true},
					{ID: "plan", Label: "Original Plan", Kind: generate.ResourceData},
				},
				Outputs: []generate.ResourceSpec{
					{ID: "output", Label: "Final Output", Kind: generate.ResourceData},
					{ID: "report", Label: "Execution Report", Kind: generate.ResourceFile},
				},
			},
		},
	}
}
