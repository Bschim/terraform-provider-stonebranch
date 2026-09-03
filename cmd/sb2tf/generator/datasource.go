package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// DataSource abstracts the operations Generator needs to fetch and list
// StoneBranch resources, so resources can be sourced either from a live UAC
// API connection (APIDataSource) or from a local JSON export tree produced
// by the sibling stonebranch-exporter-importer repo's
// `stonebranch_sync.py export` command (LocalDataSource).
//
// Every method takes ctx as its first parameter, matching the convention
// used throughout this codebase for client.Client and Generator methods.
type DataSource interface {
	// Fetch performs a generic fetch of a single resource, looking it up by
	// a name query parameter against the given API endpoint (API mode) or
	// the corresponding local export subtree (local mode). Replaces the
	// former Generator.fetchResourceByName.
	Fetch(ctx context.Context, endpoint, paramName, name string) (map[string]interface{}, error)

	// FetchTaskByName fetches a task (of any subtype) by name from the
	// generic task endpoint/subtree, used to discover a task's type before
	// dispatching to the correct resource-specific export path. Replaces
	// the former Generator.fetchResource's task usage (i.e.
	// fetchResourceByName(ctx, "/resources/task", "taskname", taskName)).
	FetchTaskByName(ctx context.Context, name string) (map[string]interface{}, error)

	// List returns all resources of the given resource type, optionally
	// filtered by name. Replaces the former
	// Generator.listResources/ListResources.
	List(ctx context.Context, rt *ResourceType, filter string) ([]ResourceItem, error)

	// FetchWorkflowVertices returns the vertices for the named workflow.
	// Replaces the former Generator.fetchWorkflowVertices.
	FetchWorkflowVertices(ctx context.Context, workflowName string) ([]WorkflowVertex, error)

	// FetchWorkflowEdges returns the edges for the named workflow. Replaces
	// the former Generator.fetchWorkflowEdges.
	FetchWorkflowEdges(ctx context.Context, workflowName string) ([]WorkflowEdge, error)
}

// ---------------------------------------------------------------------------
// APIDataSource
// ---------------------------------------------------------------------------

// APIDataSource implements DataSource against a live StoneBranch UAC
// instance over its REST API. Every method here moves the EXACT existing
// logic that used to live directly on Generator in generator.go (just
// adapted to the new receiver/method names and explicit ctx passthrough) -
// behavior is unchanged from the current generator.go implementation.
type APIDataSource struct {
	client *client.Client
}

// NewAPIDataSource creates a new APIDataSource wrapping the given API client.
func NewAPIDataSource(c *client.Client) *APIDataSource {
	return &APIDataSource{client: c}
}

// Fetch fetches a single resource by name from the given endpoint.
// Verbatim logic from the former Generator.fetchResourceByName.
func (d *APIDataSource) Fetch(ctx context.Context, endpoint, paramName, name string) (map[string]interface{}, error) {
	query := url.Values{}
	query.Set(paramName, name)

	respBody, err := d.client.Get(ctx, endpoint, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return data, nil
}

// FetchTaskByName fetches a task by name from the generic /resources/task
// endpoint. Same logic as the former Generator.fetchResourceByName(ctx,
// "/resources/task", "taskname", taskName) call sites in
// exportWorkflowComplete/ExportWorkflow.
func (d *APIDataSource) FetchTaskByName(ctx context.Context, name string) (map[string]interface{}, error) {
	return d.Fetch(ctx, "/resources/task", "taskname", name)
}

// List lists all resources of the given resource type. Verbatim logic from
// the former Generator.listResources/ListResources, with SysId/Summary now
// also populated (mirrors what cli/list.go's own listSpecificType already
// did with its separate ResourceItem struct, now unified here).
func (d *APIDataSource) List(ctx context.Context, rt *ResourceType, filter string) ([]ResourceItem, error) {
	query := url.Values{}
	// Don't filter by type in API - filter locally instead
	if filter != "" {
		query.Set(rt.NameQueryParam, filter)
	}

	respBody, err := d.client.Get(ctx, rt.ListEndpoint, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Parse as raw JSON to handle different field names
	var rawItems []map[string]interface{}
	if err := json.Unmarshal(respBody, &rawItems); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to ResourceItem, handling custom name fields
	nameField := rt.NameField
	if nameField == "" {
		nameField = "name"
	}

	var items []ResourceItem
	for _, raw := range rawItems {
		item := ResourceItem{}
		if name, ok := raw[nameField].(string); ok {
			item.Name = name
		}
		if t, ok := raw["type"].(string); ok {
			item.Type = t
		}
		if sysID, ok := raw["sysId"].(string); ok {
			item.SysId = sysID
		}
		if summary, ok := raw["summary"].(string); ok {
			item.Summary = summary
		}
		items = append(items, item)
	}

	// Filter by type locally if this resource type has a specific API type value
	if rt.APITypeValue != "" {
		var filtered []ResourceItem
		for _, item := range items {
			if item.Type == rt.APITypeValue {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	return items, nil
}

// FetchWorkflowVertices fetches all vertices for a workflow. Verbatim logic
// from the former Generator.fetchWorkflowVertices.
func (d *APIDataSource) FetchWorkflowVertices(ctx context.Context, workflowName string) ([]WorkflowVertex, error) {
	query := url.Values{}
	query.Set("workflowname", workflowName)

	respBody, err := d.client.Get(ctx, "/resources/workflow/vertices", query)
	if err != nil {
		return nil, err
	}

	var vertices []WorkflowVertex
	if err := json.Unmarshal(respBody, &vertices); err != nil {
		return nil, err
	}

	return vertices, nil
}

// FetchWorkflowEdges fetches all edges for a workflow. Verbatim logic from
// the former Generator.fetchWorkflowEdges.
func (d *APIDataSource) FetchWorkflowEdges(ctx context.Context, workflowName string) ([]WorkflowEdge, error) {
	query := url.Values{}
	query.Set("workflowname", workflowName)

	respBody, err := d.client.Get(ctx, "/resources/workflow/edges", query)
	if err != nil {
		return nil, err
	}

	var edges []WorkflowEdge
	if err := json.Unmarshal(respBody, &edges); err != nil {
		return nil, err
	}

	return edges, nil
}

// ---------------------------------------------------------------------------
// LocalDataSource
// ---------------------------------------------------------------------------

// LocalDataSource implements DataSource by reading from a local JSON export
// tree produced by the sibling stonebranch-exporter-importer repo's
// `stonebranch_sync.py export` command, instead of hitting a live UAC API.
//
// Local JSON files are the raw UAC API response bodies round-tripped
// verbatim (the Python exporter's pydantic BaseResource uses extra='allow'),
// so unmarshalling a local file into map[string]interface{} is a drop-in
// replacement for an API fetch - sb2tf's HCL templates already render off a
// plain map[string]interface{} keyed by raw field names.
type LocalDataSource struct {
	rootDir string
}

// NewLocalDataSource creates a new LocalDataSource rooted at rootDir - the
// directory produced by `stonebranch_sync.py export` (containing tasks/,
// triggers/, scripts/, calendars/, customdays/, email-templates/,
// universal-templates/, virtual-resources/ subdirectories, plus a root-level
// .index.json that is never treated as a resource file).
func NewLocalDataSource(rootDir string) *LocalDataSource {
	return &LocalDataSource{rootDir: rootDir}
}

// localKind describes how a category of resources is laid out in the local
// export tree: which glob pattern(s) to search and which JSON key holds the
// resource's name.
type localKind struct {
	globs     []string
	nameField string
}

// localKindForEndpoint maps a raw API endpoint (as used in
// ResourceType.APIEndpoint) to its local export subtree. Used by the generic
// Fetch method, which (like its API counterpart) only receives an endpoint
// string, not a *ResourceType.
func (d *LocalDataSource) localKindForEndpoint(endpoint string) (localKind, bool) {
	switch endpoint {
	case "/resources/task":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "tasks", "*", "*.json")},
			nameField: "name",
		}, true
	case "/resources/trigger":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "triggers", "*.json")},
			nameField: "name",
		}, true
	case "/resources/script":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "scripts", "*.json")},
			nameField: "scriptName",
		}, true
	case "/resources/calendar":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "calendars", "*.json")},
			nameField: "name",
		}, true
	case "/resources/customday":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "customdays", "*.json")},
			nameField: "name",
		}, true
	case "/resources/emailtemplate":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "email-templates", "*.json")},
			nameField: "templateName",
		}, true
	case "/resources/universaltemplate":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "universal-templates", "*.json")},
			nameField: "name",
		}, true
	case "/resources/virtual":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "virtual-resources", "*.json")},
			nameField: "name",
		}, true
	default:
		// credential/variable/databaseconnection/emailconnection/
		// agentcluster/businessservice/bundle have no local subtree in this
		// dataset - local mode simply doesn't support them.
		return localKind{}, false
	}
}

// localKindForResourceType maps a *ResourceType to its local export subtree,
// using the CLIName (e.g. "task_unix", "trigger_cron", "script",
// "customday", ...) rather than the API endpoint, since List callers always
// have the full ResourceType available.
func (d *LocalDataSource) localKindForResourceType(rt *ResourceType) (localKind, bool) {
	switch {
	case strings.HasPrefix(rt.CLIName, "task_"):
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "tasks", "*", "*.json")},
			nameField: "name",
		}, true
	case strings.HasPrefix(rt.CLIName, "trigger_"):
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "triggers", "*.json")},
			nameField: "name",
		}, true
	}

	switch rt.CLIName {
	case "script":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "scripts", "*.json")},
			nameField: "scriptName",
		}, true
	case "calendar":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "calendars", "*.json")},
			nameField: "name",
		}, true
	case "customday":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "customdays", "*.json")},
			nameField: "name",
		}, true
	case "emailtemplate":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "email-templates", "*.json")},
			nameField: "templateName",
		}, true
	case "universaltemplate":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "universal-templates", "*.json")},
			nameField: "name",
		}, true
	case "virtualresource":
		return localKind{
			globs:     []string{filepath.Join(d.rootDir, "virtual-resources", "*.json")},
			nameField: "name",
		}, true
	default:
		return localKind{}, false
	}
}

// globFiles expands the given glob patterns (path/filepath.Match semantics,
// same as filepath.Glob) into a sorted list of matching file paths, skipping
// the exporter's ".index.json" bookkeeping file wherever it appears.
func globFiles(globs []string) ([]string, error) {
	var files []string
	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err != nil {
			return nil, fmt.Errorf("failed to glob %s: %w", g, err)
		}
		for _, m := range matches {
			if filepath.Base(m) == ".index.json" {
				continue
			}
			files = append(files, m)
		}
	}
	sort.Strings(files)
	return files, nil
}

// readLocalJSONFile reads and unmarshals a single local JSON export file
// into a raw map, exactly as the exporter wrote it.
func readLocalJSONFile(path string) (map[string]interface{}, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return data, nil
}

// findByName searches the files matched by globs for one whose nameField
// key equals name, returning its parsed contents.
func findByName(globs []string, nameField, name string) (map[string]interface{}, error) {
	files, err := globFiles(globs)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		data, err := readLocalJSONFile(f)
		if err != nil {
			return nil, err
		}
		if n, ok := data[nameField].(string); ok && n == name {
			return data, nil
		}
	}

	return nil, fmt.Errorf("resource %q not found in local dataset", name)
}

// Fetch fetches a single resource by name, resolving endpoint to the
// corresponding local export subtree.
func (d *LocalDataSource) Fetch(ctx context.Context, endpoint, paramName, name string) (map[string]interface{}, error) {
	kind, ok := d.localKindForEndpoint(endpoint)
	if !ok {
		return nil, fmt.Errorf("local dataset has no resources for endpoint %q (resource type not present in local export tree)", endpoint)
	}
	return findByName(kind.globs, kind.nameField, name)
}

// FetchTaskByName fetches a task (of any subtype) by name, searching every
// tasks/<subdir>/*.json file for a matching "name" field. Tasks are
// discriminated by their embedded "type" field, not by which subdirectory
// they happen to live in, so no subtype filtering is needed here - the
// caller (exportWorkflowComplete/ExportWorkflow in generator.go) inspects
// data["type"] itself after the fetch, exactly as it does in API mode.
func (d *LocalDataSource) FetchTaskByName(ctx context.Context, name string) (map[string]interface{}, error) {
	kind, _ := d.localKindForEndpoint("/resources/task")
	return findByName(kind.globs, kind.nameField, name)
}

// List returns all resources of the given resource type found in the local
// export tree, optionally filtered by name using path/filepath.Match
// glob-style wildcard semantics (matching the "supports wildcards" filter
// flag documented in cli/list.go).
func (d *LocalDataSource) List(ctx context.Context, rt *ResourceType, filter string) ([]ResourceItem, error) {
	kind, ok := d.localKindForResourceType(rt)
	if !ok {
		return nil, fmt.Errorf("local dataset does not support resource type %q", rt.CLIName)
	}

	files, err := globFiles(kind.globs)
	if err != nil {
		return nil, err
	}

	var items []ResourceItem
	for _, f := range files {
		data, err := readLocalJSONFile(f)
		if err != nil {
			return nil, err
		}

		// Discriminate tasks/triggers by their JSON "type" field rather than
		// by subdirectory/filename - the exporter's own subdirectory naming
		// (e.g. "sleep" for taskSleep, "ftp" for taskFtp) is not authoritative
		// and isn't a 1:1 mapping to CLI resource type names.
		if rt.APITypeValue != "" {
			t, _ := data["type"].(string)
			if t != rt.APITypeValue {
				continue
			}
		}

		name, _ := data[kind.nameField].(string)
		if name == "" {
			continue
		}

		if filter != "" {
			matched, err := filepath.Match(filter, name)
			if err != nil {
				return nil, fmt.Errorf("invalid filter pattern %q: %w", filter, err)
			}
			if !matched {
				continue
			}
		}

		item := ResourceItem{Name: name}
		if t, ok := data["type"].(string); ok {
			item.Type = t
		}
		if sysID, ok := data["sysId"].(string); ok {
			item.SysId = sysID
		}
		if summary, ok := data["summary"].(string); ok {
			item.Summary = summary
		}
		items = append(items, item)
	}

	return items, nil
}

// FetchWorkflowVertices reads the "workflowVertices" key embedded directly
// in the local workflow task's own JSON file (confirmed key name/shape
// against a sample export, e.g. tasks/workflow/1st_Month_Test_Parent.json,
// which contains both "workflowVertices" and "workflowEdges" arrays inline -
// there is no separate vertices/edges file/endpoint in local mode).
func (d *LocalDataSource) FetchWorkflowVertices(ctx context.Context, workflowName string) ([]WorkflowVertex, error) {
	data, err := d.FetchTaskByName(ctx, workflowName)
	if err != nil {
		return nil, fmt.Errorf("failed to locate workflow %q in local dataset: %w", workflowName, err)
	}

	raw, ok := data["workflowVertices"]
	if !ok || raw == nil {
		return nil, nil
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var vertices []WorkflowVertex
	if err := json.Unmarshal(b, &vertices); err != nil {
		return nil, fmt.Errorf("failed to parse workflowVertices for %q: %w", workflowName, err)
	}

	return vertices, nil
}

// FetchWorkflowEdges reads the "workflowEdges" key embedded directly in the
// local workflow task's own JSON file.
//
// Best-effort mapping note: each local edge object's sourceId/targetId also
// carries a convenience "taskName" field, and each edge has a
// "condition": {"value": "..."} object - neither is modeled by the current
// WorkflowEdge struct (Condition is only added in a later refactor per the
// sb2tf plan; SourceId/TargetId only expose "value" today). That data is
// present in the local file but is silently dropped by json.Unmarshal here,
// exactly as it would be if the API ever returned those same extra fields.
func (d *LocalDataSource) FetchWorkflowEdges(ctx context.Context, workflowName string) ([]WorkflowEdge, error) {
	data, err := d.FetchTaskByName(ctx, workflowName)
	if err != nil {
		return nil, fmt.Errorf("failed to locate workflow %q in local dataset: %w", workflowName, err)
	}

	raw, ok := data["workflowEdges"]
	if !ok || raw == nil {
		return nil, nil
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var edges []WorkflowEdge
	if err := json.Unmarshal(b, &edges); err != nil {
		return nil, fmt.Errorf("failed to parse workflowEdges for %q: %w", workflowName, err)
	}

	return edges, nil
}
