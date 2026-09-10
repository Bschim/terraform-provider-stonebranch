package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

// Generator handles exporting StoneBranch resources to Terraform HCL.
type Generator struct {
	dataSource   DataSource
	output       string
	noDeps       bool
	withImports  bool
	force        bool
	exported     map[string]bool // Track exported resources to avoid duplicates
	nameCounters map[string]int  // Counters for generating sequential resource names
	// buffers holds one *bytes.Buffer per output filename (e.g.
	// "tasks_unix.tf", "triggers.tf", "scripts.tf", ...), populated lazily
	// via bufferFor. Replaces the old single hclBuffer.
	buffers map[string]*bytes.Buffer
	nameMap map[string]string // Maps resource key to generated terraform name
}

// NewGenerator creates a new Generator backed by the given DataSource
// (either an APIDataSource wrapping a live UAC connection, or a
// LocalDataSource reading from a local export tree).
func NewGenerator(ds DataSource, output string, noDeps, withImports, force bool) *Generator {
	return &Generator{
		dataSource:   ds,
		output:       output,
		noDeps:       noDeps,
		withImports:  withImports,
		force:        force,
		exported:     make(map[string]bool),
		nameCounters: make(map[string]int),
		buffers:      make(map[string]*bytes.Buffer),
		nameMap:      make(map[string]string),
	}
}

// generateResourceName creates a sequential identifier like "task_unix_001"
func (g *Generator) generateResourceName(resourceType string) string {
	g.nameCounters[resourceType]++
	return fmt.Sprintf("%s_%03d", resourceType, g.nameCounters[resourceType])
}

// getOrCreateResourceName returns the terraform resource name for a given resource
func (g *Generator) getOrCreateResourceName(resourceType, name string) string {
	key := fmt.Sprintf("%s/%s", resourceType, name)
	if tfName, ok := g.nameMap[key]; ok {
		return tfName
	}
	tfName := g.generateResourceName(resourceType)
	g.nameMap[key] = tfName
	return tfName
}

// ExportResource exports a single resource by type and name, appending to the buffer.
func (g *Generator) ExportResource(ctx context.Context, resourceType, name string) error {
	if g.isExported(resourceType, name) {
		return nil
	}

	rt := GetResourceType(resourceType)
	if rt == nil {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	// Fetch the resource
	data, err := g.dataSource.Fetch(ctx, rt.APIEndpoint, rt.NameQueryParam, name)
	if err != nil {
		return err
	}

	// Generate sequential resource name
	tfName := g.getOrCreateResourceName(resourceType, name)

	// Add template fields
	data["_resourceName"] = tfName
	data["_terraformResource"] = rt.TerraformResource
	data["_originalName"] = name

	// Generate HCL
	hcl, err := g.generateHCLFromData(rt, data)
	if err != nil {
		return err
	}

	// Mark as exported
	g.markExported(resourceType, name)

	// Append to buffer with comment showing original name
	buf := g.bufferFor(rt)
	buf.WriteString(fmt.Sprintf("# %s: %s\n", resourceType, name))
	buf.WriteString(hcl)
	buf.WriteString("\n")
	g.writeImportBlock(rt, tfName, name)

	return nil
}

// ExportAll exports all resources of a given type matching the filter.
func (g *Generator) ExportAll(ctx context.Context, resourceType, filter string) error {
	rt := GetResourceType(resourceType)
	if rt == nil {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	items, err := g.dataSource.List(ctx, rt, filter)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "No %s resources found\n", resourceType)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d %s resources\n", len(items), resourceType)

	for _, item := range items {
		if item.Name == "" {
			continue
		}

		// For workflows, use the complete export
		if resourceType == "task_workflow" && !g.noDeps {
			if err := g.exportWorkflowComplete(ctx, item.Name); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to export workflow %s: %v\n", item.Name, err)
			}
		} else {
			if err := g.ExportResource(ctx, resourceType, item.Name); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to export %s/%s: %v\n", resourceType, item.Name, err)
			}
		}
	}

	return nil
}

// ExportTasks exports tasks matching a filter pattern.
// Workflows will include their contained tasks, vertices, and edges.
//
// NOTE: this method is not currently called from cli/export.go (which uses
// ExportAll/exportWorkflowComplete/ExportResource for its category/single
// export paths) but is kept and converted to use the DataSource for
// consistency with the rest of the Generator.
func (g *Generator) ExportTasks(ctx context.Context, filter string) error {
	// List all tasks matching the filter (any subtype), using a synthetic
	// ResourceType that mirrors the raw /resources/task/listadv listing
	// this method has always performed (no APITypeValue, so DataSource.List
	// applies no type filtering and returns every task type, matching the
	// old behavior exactly).
	items, err := g.dataSource.List(ctx, &ResourceType{
		ListEndpoint:   "/resources/task/listadv",
		NameQueryParam: "taskname",
		NameField:      "name",
	}, filter)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "No tasks found matching filter: %s\n", filter)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d tasks matching filter\n", len(items))

	// Separate workflows from other tasks
	var workflows []ResourceItem
	var otherTasks []ResourceItem

	for _, item := range items {
		if item.Type == "taskWorkflow" {
			workflows = append(workflows, item)
		} else {
			otherTasks = append(otherTasks, item)
		}
	}

	// First, export workflows (with their tasks, vertices, edges)
	for _, wf := range workflows {
		if wf.Name == "" {
			continue
		}
		if err := g.exportWorkflowComplete(ctx, wf.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to export workflow %s: %v\n", wf.Name, err)
		}
	}

	// Then export other tasks (that weren't already exported as part of a workflow)
	for _, task := range otherTasks {
		if task.Name == "" || task.Type == "" {
			continue
		}

		cliType, ok := APITypeToResourceType[task.Type]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: unsupported task type %s for task %s\n", task.Type, task.Name)
			continue
		}

		if g.isExported(cliType, task.Name) {
			continue
		}

		if err := g.ExportResource(ctx, cliType, task.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to export %s/%s: %v\n", cliType, task.Name, err)
		}
	}

	return nil
}

// exportWorkflowComplete exports a workflow with its tasks, vertices, and edges.
// Order: workflow definition -> tasks -> vertices -> edges
func (g *Generator) exportWorkflowComplete(ctx context.Context, name string) error {
	if g.isExported("task_workflow", name) {
		return nil
	}

	rt := GetResourceType("task_workflow")
	if rt == nil {
		return fmt.Errorf("unknown resource type: task_workflow")
	}

	// Fetch the workflow task
	data, err := g.dataSource.Fetch(ctx, rt.APIEndpoint, rt.NameQueryParam, name)
	if err != nil {
		return fmt.Errorf("failed to fetch workflow: %w", err)
	}

	// Generate workflow resource name
	wfTfName := g.getOrCreateResourceName("task_workflow", name)
	data["_resourceName"] = wfTfName
	data["_terraformResource"] = rt.TerraformResource
	data["_originalName"] = name

	// Generate workflow HCL
	workflowHCL, err := g.generateHCLFromData(rt, data)
	if err != nil {
		return fmt.Errorf("failed to generate workflow HCL: %w", err)
	}

	// Mark workflow as exported
	g.markExported("task_workflow", name)

	// Write workflow section header (into the task_workflow file - vertices
	// and edges below are routed into this same file, see bufferFor).
	wfBuf := g.bufferFor(rt)
	wfBuf.WriteString(fmt.Sprintf("\n# ============================================================\n"))
	wfBuf.WriteString(fmt.Sprintf("# Workflow: %s\n", name))
	wfBuf.WriteString(fmt.Sprintf("# ============================================================\n\n"))
	wfBuf.WriteString(fmt.Sprintf("# task_workflow: %s\n", name))
	wfBuf.WriteString(workflowHCL)
	wfBuf.WriteString("\n")
	g.writeImportBlock(rt, wfTfName, name)

	// Fetch workflow vertices
	vertices, err := g.dataSource.FetchWorkflowVertices(ctx, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch workflow vertices for %s: %v\n", name, err)
		return nil
	}

	if len(vertices) == 0 {
		return nil
	}

	// Export all tasks in the workflow. Each task is written into its own
	// resource-type file (e.g. tasks_unix.tf), not the workflow's file,
	// since these are regular task resources, not workflow-only constructs.
	wfBuf.WriteString("# --- Tasks in Workflow ---\n")
	for _, v := range vertices {
		taskName := v.Task.Value
		if taskName == "" {
			continue
		}

		// Fetch the task
		taskData, err := g.dataSource.FetchTaskByName(ctx, taskName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch task %s: %v\n", taskName, err)
			continue
		}

		taskType, ok := taskData["type"].(string)
		if !ok {
			continue
		}

		cliType, ok := APITypeToResourceType[taskType]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: unsupported task type %s for task %s\n", taskType, taskName)
			continue
		}

		// A vertex whose task is itself a workflow needs its own full
		// treatment (header + tasks + vertices + edges), not the generic
		// bare/leaf resource export below - otherwise it would be permanently
		// skipped (via isExported) when its own turn comes up later, missing
		// its own vertices/edges entirely. Recurse instead; the recursive
		// call's isExported guard (set before its own task loop runs) already
		// protects against duplicate work for diamond references and against
		// infinite recursion on direct cycles, so no extra visited-set is
		// needed here.
		if cliType == "task_workflow" {
			if err := g.exportWorkflowComplete(ctx, taskName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to export sub-workflow %s: %v\n", taskName, err)
			}
			continue
		}

		if g.isExported(cliType, taskName) {
			continue
		}

		taskRT := GetResourceType(cliType)
		if taskRT == nil {
			continue
		}

		// Generate task resource name
		taskTfName := g.getOrCreateResourceName(cliType, taskName)
		taskData["_resourceName"] = taskTfName
		taskData["_terraformResource"] = taskRT.TerraformResource
		taskData["_originalName"] = taskName

		taskHCL, err := g.generateHCLFromData(taskRT, taskData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to generate HCL for task %s: %v\n", taskName, err)
			continue
		}

		taskBuf := g.bufferFor(taskRT)
		taskBuf.WriteString(fmt.Sprintf("# %s: %s\n", cliType, taskName))
		taskBuf.WriteString(taskHCL)
		taskBuf.WriteString("\n")
		g.writeImportBlock(taskRT, taskTfName, taskName)

		g.markExported(cliType, taskName)
	}

	// Generate workflow vertices and track their terraform names. Vertices
	// only ever accompany a task_workflow export, so they're written into
	// the same tasks_workflow.tf file as the workflow itself (see
	// bufferFor's Category=="Workflow" case) rather than a separate file -
	// this keeps a workflow's full definition (task + vertices + edges)
	// readable in one place.
	vertexRT := GetResourceType("workflow_vertex")
	wfBuf.WriteString("# --- Workflow Vertices ---\n")
	vertexTfNames := make(map[string]string) // Maps vertexId to terraform resource name
	for _, v := range vertices {
		taskName := v.Task.Value
		if taskName == "" {
			continue
		}
		vertexTfName, vertexHCL, err := g.generateWorkflowVertexHCLNew(name, wfTfName, v)
		if err == nil {
			vertexTfNames[v.VertexId] = vertexTfName
			g.bufferFor(vertexRT).WriteString(vertexHCL)
			g.bufferFor(vertexRT).WriteString("\n")
			g.writeImportBlock(vertexRT, vertexTfName, fmt.Sprintf("%s/%s", name, v.VertexId))
		}
	}

	// Generate workflow edges
	edgeRT := GetResourceType("workflow_edge")
	edges, err := g.dataSource.FetchWorkflowEdges(ctx, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch workflow edges for %s: %v\n", name, err)
	} else if len(edges) > 0 {
		wfBuf.WriteString("# --- Workflow Edges ---\n")
		for _, e := range edges {
			edgeTfName, edgeHCL, err := g.generateWorkflowEdgeHCLNew(name, wfTfName, e, vertices, vertexTfNames)
			if err == nil {
				g.bufferFor(edgeRT).WriteString(edgeHCL)
				g.bufferFor(edgeRT).WriteString("\n")
				g.writeImportBlock(edgeRT, edgeTfName, fmt.Sprintf("%s/%s/%s", name, e.SourceId.Value, e.TargetId.Value))
			}
		}
	}

	return nil
}

// ExportWorkflow exports a workflow and all its components including dependent tasks.
func (g *Generator) ExportWorkflow(ctx context.Context, name string) error {
	// Export the workflow task itself
	if err := g.ExportResource(ctx, "task_workflow", name); err != nil {
		return fmt.Errorf("failed to export workflow: %w", err)
	}

	if g.noDeps {
		return nil
	}

	// Fetch workflow vertices
	vertices, err := g.dataSource.FetchWorkflowVertices(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to fetch workflow vertices: %w", err)
	}

	// Export each task in the workflow
	for _, v := range vertices {
		taskName := v.Task.Value
		if taskName == "" || g.isExported("task", taskName) {
			continue
		}

		// Determine task type and export
		taskData, err := g.dataSource.FetchTaskByName(ctx, taskName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch task %s: %v\n", taskName, err)
			continue
		}

		taskType, ok := taskData["type"].(string)
		if !ok {
			continue
		}

		cliType, ok := APITypeToResourceType[taskType]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: unsupported task type %s for task %s\n", taskType, taskName)
			continue
		}

		if err := g.ExportResource(ctx, cliType, taskName); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to export task %s: %v\n", taskName, err)
		}

		// Generate workflow vertex
		vertexHCL, err := g.generateWorkflowVertexHCL(name, v)
		if err == nil {
			g.outputHCL("workflow_vertex", fmt.Sprintf("%s_%s", name, taskName), vertexHCL)
		}
	}

	// Fetch and generate workflow edges
	edges, err := g.dataSource.FetchWorkflowEdges(ctx, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch workflow edges: %v\n", err)
	} else {
		for i, e := range edges {
			edgeHCL, err := g.generateWorkflowEdgeHCL(name, e, vertices)
			if err == nil {
				g.outputHCL("workflow_edge", fmt.Sprintf("%s_edge_%d", name, i), edgeHCL)
			}
		}
	}

	return nil
}

// ListResources lists all resources of a given type (exported for CLI use).
func (g *Generator) ListResources(ctx context.Context, rt *ResourceType, filter string) ([]ResourceItem, error) {
	return g.dataSource.List(ctx, rt, filter)
}

// ResourceItem represents a resource in list responses.
type ResourceItem struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	Task struct {
		Value string `json:"value"`
	} `json:"task,omitempty"`
	// SysId and Summary are populated by DataSource.List implementations
	// (APIDataSource and LocalDataSource) so that cli/list.go can unify onto
	// this struct instead of maintaining its own separate ResourceItem type.
	SysId   string `json:"sysId,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// exportDependencies exports resources that this resource depends on.
func (g *Generator) exportDependencies(ctx context.Context, rt *ResourceType, data map[string]interface{}) error {
	for _, dep := range rt.Dependencies {
		// Check condition if present
		if dep.Condition != "" {
			parts := strings.Split(dep.Condition, "=")
			if len(parts) == 2 {
				if val, ok := data[parts[0]].(string); !ok || val != parts[1] {
					continue
				}
			}
		}

		// Get the referenced resource name
		refName, ok := data[dep.Field].(string)
		if !ok || refName == "" {
			continue
		}

		// Skip if already exported
		if g.isExported(dep.ResourceType, refName) {
			continue
		}

		// Export the dependency
		if err := g.ExportResource(ctx, dep.ResourceType, refName); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to export dependency %s/%s: %v\n", dep.ResourceType, refName, err)
		}
	}

	return nil
}

// WorkflowVertex represents a vertex in the workflow.
type WorkflowVertex struct {
	Task struct {
		Value string `json:"value"`
	} `json:"task"`
	VertexId string `json:"vertexId"`
	Alias    string `json:"alias,omitempty"`
	VertexX  string `json:"vertexX,omitempty"`
	VertexY  string `json:"vertexY,omitempty"`
}

// WorkflowEdge represents an edge in the workflow.
type WorkflowEdge struct {
	SourceId struct {
		Value string `json:"value"`
	} `json:"sourceId"`
	TargetId struct {
		Value string `json:"value"`
	} `json:"targetId"`
	StraightEdge bool `json:"straightEdge,omitempty"`
	Condition    struct {
		Value       string `json:"value,omitempty"`
		Type        string `json:"type,omitempty"`
		FirstValue  string `json:"firstValue,omitempty"`
		Operator    string `json:"operator,omitempty"`
		SecondValue string `json:"secondValue,omitempty"`
	} `json:"condition,omitempty"`
}

// generateHCL generates HCL for a resource (legacy - uses sanitized names).
func (g *Generator) generateHCL(rt *ResourceType, data map[string]interface{}) (string, error) {
	tmpl := GetTemplate(rt.CLIName)
	if tmpl == nil {
		return "", fmt.Errorf("no template for resource type: %s", rt.CLIName)
	}

	// Add computed fields to data
	name, _ := data["name"].(string)
	data["_resourceName"] = SanitizeName(name)
	data["_terraformResource"] = rt.TerraformResource

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// generateHCLFromData generates HCL for a resource using pre-set _resourceName.
func (g *Generator) generateHCLFromData(rt *ResourceType, data map[string]interface{}) (string, error) {
	tmpl := GetTemplate(rt.CLIName)
	if tmpl == nil {
		return "", fmt.Errorf("no template for resource type: %s", rt.CLIName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// generateWorkflowVertexHCL generates HCL for a workflow vertex.
func (g *Generator) generateWorkflowVertexHCL(workflowName string, v WorkflowVertex) (string, error) {
	data := map[string]interface{}{
		"workflowName":       workflowName,
		"taskName":           v.Task.Value,
		"alias":              v.Alias,
		"_resourceName":      SanitizeName(fmt.Sprintf("%s_%s", workflowName, v.Task.Value)),
		"_terraformResource": "stonebranch_workflow_vertex",
	}

	tmpl := GetTemplate("workflow_vertex")
	if tmpl == nil {
		return "", fmt.Errorf("no template for workflow_vertex")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateWorkflowEdgeHCL generates HCL for a workflow edge.
func (g *Generator) generateWorkflowEdgeHCL(workflowName string, e WorkflowEdge, vertices []WorkflowVertex) (string, error) {
	// Find task names for source and target vertex IDs
	var sourceTask, targetTask string
	for _, v := range vertices {
		if v.VertexId == e.SourceId.Value {
			sourceTask = v.Task.Value
		}
		if v.VertexId == e.TargetId.Value {
			targetTask = v.Task.Value
		}
	}

	data := map[string]interface{}{
		"workflowName":       workflowName,
		"sourceTask":         sourceTask,
		"targetTask":         targetTask,
		"straightEdge":       e.StraightEdge,
		"_resourceName":      SanitizeName(fmt.Sprintf("%s_%s_to_%s", workflowName, sourceTask, targetTask)),
		"_terraformResource": "stonebranch_workflow_edge",
	}
	if e.Condition.Value != "" || e.Condition.Type != "" {
		data["condition"] = map[string]interface{}{
			"type":        e.Condition.Type,
			"value":       e.Condition.Value,
			"firstValue":  e.Condition.FirstValue,
			"operator":    e.Condition.Operator,
			"secondValue": e.Condition.SecondValue,
		}
	}

	tmpl := GetTemplate("workflow_edge")
	if tmpl == nil {
		return "", fmt.Errorf("no template for workflow_edge")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateWorkflowVertexHCLNew generates HCL for a workflow vertex using sequential naming.
// Returns the terraform resource name and the HCL string.
func (g *Generator) generateWorkflowVertexHCLNew(workflowName, workflowTfName string, v WorkflowVertex) (string, string, error) {
	taskName := v.Task.Value
	vertexTfName := g.generateResourceName("workflow_vertex")

	data := map[string]interface{}{
		"workflowName":       workflowName,
		"workflowTfName":     workflowTfName,
		"taskName":           taskName,
		"alias":              v.Alias,
		"_resourceName":      vertexTfName,
		"_terraformResource": "stonebranch_workflow_vertex",
	}

	tmpl := GetTemplate("workflow_vertex")
	if tmpl == nil {
		return "", "", fmt.Errorf("no template for workflow_vertex")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", err
	}

	return vertexTfName, buf.String(), nil
}

// generateWorkflowEdgeHCLNew generates HCL for a workflow edge using sequential naming.
// Returns the edge's generated terraform resource name alongside its HCL so
// callers can also emit a matching import {} block.
func (g *Generator) generateWorkflowEdgeHCLNew(workflowName, workflowTfName string, e WorkflowEdge, vertices []WorkflowVertex, vertexTfNames map[string]string) (string, string, error) {
	// Find task names for source and target vertex IDs
	var sourceTask, targetTask string
	for _, v := range vertices {
		if v.VertexId == e.SourceId.Value {
			sourceTask = v.Task.Value
		}
		if v.VertexId == e.TargetId.Value {
			targetTask = v.Task.Value
		}
	}

	// Get the terraform resource names for source and target vertices
	sourceVertexTfName := vertexTfNames[e.SourceId.Value]
	targetVertexTfName := vertexTfNames[e.TargetId.Value]

	edgeTfName := g.generateResourceName("workflow_edge")

	data := map[string]interface{}{
		"workflowName":       workflowName,
		"workflowTfName":     workflowTfName,
		"sourceTask":         sourceTask,
		"targetTask":         targetTask,
		"sourceVertexTfName": sourceVertexTfName,
		"targetVertexTfName": targetVertexTfName,
		"straightEdge":       e.StraightEdge,
		"_resourceName":      edgeTfName,
		"_terraformResource": "stonebranch_workflow_edge",
	}
	if e.Condition.Value != "" || e.Condition.Type != "" {
		data["condition"] = map[string]interface{}{
			"type":        e.Condition.Type,
			"value":       e.Condition.Value,
			"firstValue":  e.Condition.FirstValue,
			"operator":    e.Condition.Operator,
			"secondValue": e.Condition.SecondValue,
		}
	}

	tmpl := GetTemplate("workflow_edge")
	if tmpl == nil {
		return "", "", fmt.Errorf("no template for workflow_edge")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", err
	}

	return edgeTfName, buf.String(), nil
}

// outputHCL appends HCL to the buffer for the given resource type (legacy compatibility).
func (g *Generator) outputHCL(resourceType, name, hcl string) error {
	rt := GetResourceType(resourceType)
	buf := g.bufferFor(rt)
	buf.WriteString(fmt.Sprintf("# %s: %s\n", resourceType, name))
	buf.WriteString(hcl)
	buf.WriteString("\n")
	return nil
}

// markExported marks a resource as exported.
func (g *Generator) markExported(resourceType, name string) {
	g.exported[fmt.Sprintf("%s/%s", resourceType, name)] = true
}

// isExported checks if a resource has already been exported.
func (g *Generator) isExported(resourceType, name string) bool {
	return g.exported[fmt.Sprintf("%s/%s", resourceType, name)]
}

// SanitizeName converts a StoneBranch resource name to a valid Terraform identifier.
func SanitizeName(name string) string {
	// Replace non-alphanumeric characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	sanitized := re.ReplaceAllString(name, "_")

	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)

	// Remove consecutive underscores
	re = regexp.MustCompile(`_+`)
	sanitized = re.ReplaceAllString(sanitized, "_")

	// Trim leading/trailing underscores
	sanitized = strings.Trim(sanitized, "_")

	// Ensure it starts with a letter
	if len(sanitized) > 0 && unicode.IsDigit(rune(sanitized[0])) {
		sanitized = "r_" + sanitized
	}

	if sanitized == "" {
		sanitized = "resource"
	}

	return sanitized
}

// GetExportedResources returns the list of exported resources.
func (g *Generator) GetExportedResources() []string {
	var result []string
	for key := range g.exported {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

// GetTemplate returns the template for a resource type.
// This is a placeholder that will be implemented in templates.go
var GetTemplate func(resourceType string) *template.Template

// ---------------------------------------------------------------------------
// Multi-file output
// ---------------------------------------------------------------------------

// otherFilenames maps CLINames outside the Tasks/Triggers/Workflow
// categories to their output filename. Enumerated explicitly (rather than
// derived via string pluralization) since the set of resource types is
// small and fixed, and this keeps filenames predictable/readable.
var otherFilenames = map[string]string{
	"script":              "scripts.tf",
	"variable":            "variables.tf",
	"credential":          "credentials.tf",
	"business_service":    "business_services.tf",
	"agent_cluster":       "agent_clusters.tf",
	"calendar":            "calendars.tf",
	"customday":           "custom_days.tf",
	"emailtemplate":       "email_templates.tf",
	"universaltemplate":   "universal_templates.tf",
	"virtualresource":     "virtual_resources.tf",
	"database_connection": "database_connections.tf",
	"email_connection":    "email_connections.tf",
}

// outputFilename maps a ResourceType to the .tf output filename its
// generated HCL belongs in.
//
//   - Task types (Category "Tasks", CLIName prefixed "task_") each get their
//     own file: tasks_<suffix>.tf, where <suffix> is the CLIName with the
//     "task_" prefix stripped (task_unix -> tasks_unix.tf, task_workflow ->
//     tasks_workflow.tf, task_universal -> tasks_universal.tf, ...).
//   - Trigger types (Category "Triggers") all share a single triggers.tf.
//   - workflow_vertex/workflow_edge (Category "Workflow") are routed into
//     tasks_workflow.tf, the same file as the task_workflow resource they
//     always accompany. They're never emitted on their own (only ever
//     alongside a task_workflow export via exportWorkflowComplete), so
//     giving them a separate workflows.tf file would just split a single
//     workflow's definition (task + vertices + edges) across two files for
//     no benefit - keeping them together in tasks_workflow.tf reads most
//     naturally and matches how exportWorkflowComplete already interleaves
//     them with header comments today.
//   - Everything else (Connections/Other categories) gets an explicit,
//     enumerated filename via otherFilenames.
func outputFilename(rt *ResourceType) string {
	switch rt.Category {
	case "Tasks":
		if strings.HasPrefix(rt.CLIName, "task_") {
			return "tasks_" + strings.TrimPrefix(rt.CLIName, "task_") + ".tf"
		}
	case "Triggers":
		return "triggers.tf"
	case "Workflow":
		return "tasks_workflow.tf"
	}

	if filename, ok := otherFilenames[rt.CLIName]; ok {
		return filename
	}

	// Fallback for any resource type not covered above - shouldn't happen
	// for any currently-registered type, but avoids silently dropping
	// output if a new type is added without also updating otherFilenames.
	return rt.CLIName + ".tf"
}

// bufferFor returns the buffer that HCL for the given ResourceType should be
// written to, lazily creating it (and registering it in g.buffers under its
// mapped filename) on first use.
func (g *Generator) bufferFor(rt *ResourceType) *bytes.Buffer {
	filename := outputFilename(rt)
	if buf, ok := g.buffers[filename]; ok {
		return buf
	}
	buf := &bytes.Buffer{}
	g.buffers[filename] = buf
	return buf
}

// writeImportBlock appends an `import {}` block to imports.tf for the given
// resource, if --with-imports was requested. `name` is the resource's import
// ID as expected by that resource type's ImportState (a plain name for most
// types; a composite "workflow_name/vertex_id" or
// "workflow_name/source_id/target_id" string for workflow_vertex/
// workflow_edge, which have no standalone UAC name of their own).
func (g *Generator) writeImportBlock(rt *ResourceType, tfName, name string) {
	if !g.withImports {
		return
	}
	buf := g.importsBuffer()
	buf.WriteString(fmt.Sprintf("import {\n  to = %s.%s\n  id = \"%s\"\n}\n\n",
		rt.TerraformResource, tfName, quote(name)))
}

func (g *Generator) importsBuffer() *bytes.Buffer {
	const filename = "imports.tf"
	if buf, ok := g.buffers[filename]; ok {
		return buf
	}
	buf := &bytes.Buffer{}
	g.buffers[filename] = buf
	return buf
}

// Finalize writes the buffered output (one section per output file) to
// stdout or to individual files in g.output.
func (g *Generator) Finalize() error {
	type section struct {
		filename string
		buf      *bytes.Buffer
	}

	var sections []section
	for filename, buf := range g.buffers {
		if buf.Len() == 0 {
			continue
		}
		sections = append(sections, section{filename: filename, buf: buf})
	}

	if len(sections) == 0 {
		return nil
	}

	sort.Slice(sections, func(i, j int) bool { return sections[i].filename < sections[j].filename })

	header := "# Generated by sb2tf from StoneBranch Universal Controller\n\n"

	if g.output == "" {
		// Write to stdout: concatenate every non-empty section, in a
		// stable (filename-sorted) order, each preceded by a header
		// comment identifying which file section it corresponds to -
		// preserving today's single-buffer stdout behavior while now
		// covering all sections.
		for _, s := range sections {
			fmt.Printf("# --- %s ---\n\n", s.filename)
			fmt.Print(header)
			fmt.Print(s.buf.String())
			fmt.Println()
		}
		return nil
	}

	// Write to file: one .tf file per non-empty buffer.
	if err := os.MkdirAll(g.output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var skipped []string

	for _, s := range sections {
		content := header + s.buf.String()
		filename := filepath.Join(g.output, s.filename)

		if !g.force {
			if existing, err := os.ReadFile(filename); err == nil {
				oldCount := countResourceBlocks(string(existing))
				newCount := countResourceBlocks(s.buf.String())
				if newCount < oldCount {
					fmt.Fprintf(os.Stderr,
						"Refusing to overwrite %s: it currently has %d resource block(s), "+
							"but this export would only write %d. This usually means dependency-following "+
							"(e.g. exporting a workflow without --no-deps) only touched a subset of this "+
							"file's resource type. Re-run with --no-deps for a standalone export of this "+
							"type, or pass --force to overwrite anyway.\n",
						filename, oldCount, newCount)
					skipped = append(skipped, filename)
					continue
				}
			}
		}

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
		fmt.Fprintf(os.Stderr, "Wrote %s\n", filename)
	}

	if len(skipped) > 0 {
		return fmt.Errorf("refused to overwrite %d file(s) that would have shrunk: %s", len(skipped), strings.Join(skipped, ", "))
	}

	return nil
}

// countResourceBlocks counts top-level `resource "..." "..." {` block
// declarations in HCL content, used by Finalize to detect a would-be
// shrinking overwrite (see the --force check above).
func countResourceBlocks(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "resource \"") {
			count++
		}
	}
	return count
}
