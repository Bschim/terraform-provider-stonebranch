# Plan: sb2tf CLI Utility

## Overview

Add a command-line utility `sb2tf` to this Terraform provider project that reads existing resources from the StoneBranch Universal Controller API and generates Terraform configuration files (.tf).

## Use Cases

1. **Bootstrap a new Terraform project** - Export existing resources to start managing them with Terraform
2. **Test reproducibility** - Verify that Terraform configs can recreate existing infrastructure
3. **Migration** - Help migrate manually-created resources to Infrastructure as Code

## Design Decisions (Based on User Input)

| Decision | Choice |
|----------|--------|
| Import blocks | No - just resource definitions |
| File organization | Support both via `--format` flag (single/grouped) |
| Dependencies | Yes - automatically export dependent resources |
| CLI framework | Cobra |

---

## Project Structure

```
cmd/sb2tf/
├── main.go                          # CLI entry point, Cobra setup
├── cli/
│   ├── root.go                      # Root command, global flags (--token, --url, --output)
│   ├── list.go                      # List available resources by type
│   └── export.go                    # Export resources to Terraform HCL
└── generator/
    ├── generator.go                 # Core HCL generation logic
    ├── templates.go                 # HCL templates per resource type
    ├── resources.go                 # Resource type registry (API endpoint, TF resource name, fields)
    └── dependencies.go              # Dependency resolution logic
```

### Reused Code

- `internal/client/client.go` - API client (authentication, HTTP methods) - import directly

---

## CLI Interface

```bash
# Authentication (same env vars as provider)
export STONEBRANCH_API_TOKEN="your-token"
export STONEBRANCH_BASE_URL="https://your-instance.stonebranch.cloud"

# List resources
sb2tf list                                    # Show all resource types
sb2tf list tasks                              # List all tasks (shows name, type)
sb2tf list tasks --filter "prod-*"            # Filter by name pattern
sb2tf list triggers                           # List all triggers
sb2tf list variables                          # List all variables

# Export single resource
sb2tf export task_unix my_task                # Export one Unix task
sb2tf export task_workflow my_workflow        # Export workflow + dependencies

# Export multiple resources
sb2tf export tasks --filter "prod-*"          # Export tasks matching pattern
sb2tf export triggers --all                   # Export all triggers
sb2tf export --all                            # Export everything

# Output options
sb2tf export tasks --all --output ./terraform/           # Write to directory
sb2tf export tasks --all --format single                 # One file per resource (default)
sb2tf export tasks --all --format grouped                # Group by type

# Other flags
sb2tf export task_workflow my_wf --no-deps               # Skip dependency resolution
sb2tf export --dry-run                                   # Show what would be exported
```

---

## Supported Resource Types

| Category | CLI Name | TF Resource | API Endpoint | List Endpoint |
|----------|----------|-------------|--------------|---------------|
| **Tasks** | | | | |
| | task_unix | stonebranch_task_unix | /resources/task | /resources/task/listadv?type=taskUnix |
| | task_windows | stonebranch_task_windows | /resources/task | /resources/task/listadv?type=taskWindows |
| | task_sql | stonebranch_task_sql | /resources/task | /resources/task/listadv?type=taskSql |
| | task_email | stonebranch_task_email | /resources/task | /resources/task/listadv?type=taskEmail |
| | task_workflow | stonebranch_task_workflow | /resources/task | /resources/task/listadv?type=taskWorkflow |
| | task_file_monitor | stonebranch_task_file_monitor | /resources/task | /resources/task/listadv?type=taskFileMonitor |
| | task_file_transfer | stonebranch_task_file_transfer | /resources/task | /resources/task/listadv?type=taskFileTransfer |
| | task_timer | stonebranch_task_timer | /resources/task | /resources/task/listadv?type=taskTimer |
| | task_monitor | stonebranch_task_monitor | /resources/task | /resources/task/listadv?type=taskMonitor |
| | task_stored_procedure | stonebranch_task_stored_procedure | /resources/task | /resources/task/listadv?type=taskStoredProc |
| | task_web_service | stonebranch_task_web_service | /resources/task | /resources/task/listadv?type=taskWebService |
| | task_universal_aws_s3 | stonebranch_task_universal_aws_s3 | /resources/task | /resources/task/listadv?type=taskUniversal |
| **Triggers** | | | | |
| | trigger_time | stonebranch_trigger_time | /resources/trigger | /resources/trigger/listadv?type=triggerTime |
| | trigger_cron | stonebranch_trigger_cron | /resources/trigger | /resources/trigger/listadv?type=triggerCron |
| | trigger_file_monitor | stonebranch_trigger_file_monitor | /resources/trigger | /resources/trigger/listadv?type=triggerFm |
| | trigger_task_monitor | stonebranch_trigger_task_monitor | /resources/trigger | /resources/trigger/listadv?type=triggerTm |
| **Connections** | | | | |
| | database_connection | stonebranch_database_connection | /resources/databaseconnection | /resources/databaseconnection/list |
| | email_connection | stonebranch_email_connection | /resources/emailconnection | /resources/emailconnection/list |
| **Other** | | | | |
| | script | stonebranch_script | /resources/script | /resources/script/list |
| | variable | stonebranch_variable | /resources/variable | /resources/variable/list |
| | credential | stonebranch_credential | /resources/credential | /resources/credential/list |
| | business_service | stonebranch_business_service | /resources/businessservice | /resources/businessservice/list |
| | agent_cluster | stonebranch_agent_cluster | /resources/agentcluster | /resources/agentcluster/list |
| | calendar | stonebranch_calendar | /resources/calendar | /resources/calendar/list |
| **Workflow** | | | | |
| | workflow_vertex | stonebranch_workflow_vertex | /resources/workflow/vertices | (per-workflow) |
| | workflow_edge | stonebranch_workflow_edge | /resources/workflow/edges | (per-workflow) |

---

## Dependency Resolution

When exporting a workflow with `--deps` (default), sb2tf will:

1. **Export the workflow task itself**
2. **Fetch workflow vertices** via `GET /resources/workflow/vertices?workflowname=X`
3. **For each vertex**: Export the referenced task (and its dependencies)
4. **Fetch workflow edges** via `GET /resources/workflow/edges?workflowname=X`
5. **Export edges** with correct vertex references

### Task Dependencies

When exporting any task, also check and export:
- `credentials` → stonebranch_credential
- `script` → stonebranch_script (if command_or_script = "Script")
- `database_connection` → stonebranch_database_connection (for SQL/stored proc tasks)
- `email_connection` → stonebranch_email_connection (for email tasks)
- `agent_cluster` → stonebranch_agent_cluster

### Trigger Dependencies

- `tasks` → List of task names to export
- `calendar` → stonebranch_calendar
- `task_monitor` → For file_monitor and task_monitor triggers

---

## HCL Generation

### Template Approach

Use Go `text/template` to generate HCL. Example for a variable:

```go
const variableTemplate = `resource "stonebranch_variable" "{{.ResourceName}}" {
  name        = "{{.Name}}"
{{- if .Value}}
  value       = "{{.Value}}"
{{- end}}
{{- if .Description}}
  description = "{{.Description}}"
{{- end}}
{{- if .OpswiseGroups}}
  opswise_groups = [{{range $i, $v := .OpswiseGroups}}{{if $i}}, {{end}}"{{$v}}"{{end}}]
{{- end}}
}
`
```

### Resource Name Sanitization

Convert StoneBranch names to valid Terraform identifiers:
- `My Task Name` → `my_task_name`
- `prod-task-01` → `prod_task_01`
- `123task` → `task_123` (prepend if starts with digit)

### Example Output

```hcl
# Generated by sb2tf from StoneBranch Universal Controller
# Source: my_unix_task

resource "stonebranch_task_unix" "my_unix_task" {
  name              = "my_unix_task"
  summary           = "Runs the daily batch job"
  agent             = "linux-prod-01"
  command           = "/opt/scripts/daily.sh"
  credentials       = "service_account"
  exit_codes        = "0"

  opswise_groups    = ["production", "batch-jobs"]
}

resource "stonebranch_credential" "service_account" {
  name        = "service_account"
  runtime_user = "svc_batch"
  # Note: password cannot be exported for security
}
```

---

## Implementation Steps

### Phase 1: CLI Skeleton
1. Create `cmd/sb2tf/main.go` - Cobra root command
2. Implement `cli/root.go` with global flags:
   - `--token` (env: STONEBRANCH_API_TOKEN)
   - `--url` (env: STONEBRANCH_BASE_URL)
   - `--output` (default: stdout)
   - `--format` (single|grouped)
3. Implement `cli/list.go` - list command

**Files to create:**
- `cmd/sb2tf/main.go`
- `cmd/sb2tf/cli/root.go`
- `cmd/sb2tf/cli/list.go`

### Phase 2: Generator Framework
1. Create `generator/resources.go` - resource type registry
2. Create `generator/generator.go` - main generation logic
3. Create `generator/templates.go` - HCL templates

**Files to create:**
- `cmd/sb2tf/generator/resources.go`
- `cmd/sb2tf/generator/generator.go`
- `cmd/sb2tf/generator/templates.go`

### Phase 3: Export Command
1. Implement `cli/export.go` - export command
2. Wire up generator to export command

**Files to create/modify:**
- `cmd/sb2tf/cli/export.go`

### Phase 4: Simple Resources
Implement templates and API models for:
1. variable
2. script
3. credential
4. business_service

### Phase 5: Task Resources
Implement templates for all task types:
1. task_unix
2. task_windows
3. task_sql
4. task_email
5. task_workflow
6. (remaining task types)

### Phase 6: Triggers and Connections
1. trigger_time, trigger_cron, trigger_file_monitor, trigger_task_monitor
2. database_connection, email_connection
3. agent_cluster, calendar

### Phase 7: Workflow Support
1. Implement workflow dependency resolution
2. Generate workflow_vertex and workflow_edge resources
3. Handle cross-references between resources

### Phase 8: Polish
1. Add `--dry-run` flag
2. Add `--filter` wildcards
3. Update Makefile with build targets
4. Add README section for sb2tf

---

## Verification

After implementation, verify with:

```bash
# Build the utility
make build-sb2tf

# List resources from a test instance
./bin/sb2tf list tasks

# Export a simple resource
./bin/sb2tf export variable my_var > test.tf
cat test.tf

# Export a workflow with dependencies
./bin/sb2tf export task_workflow my_workflow --output ./exported/

# Verify generated Terraform is valid
cd ./exported && terraform init && terraform validate
```

---

## Files to Modify

| File | Change |
|------|--------|
| `Makefile` | Add `build-sb2tf` and `install-sb2tf` targets |
| `go.mod` | Add Cobra dependency |
| `CLAUDE.md` | Document sb2tf utility |
| `.goreleaser.yaml` | Add sb2tf build configuration |

---

## GoReleaser Configuration

The project uses GoReleaser to build and publish releases to GitHub. To include `sb2tf` in releases, add a second build entry to `.goreleaser.yaml`:

```yaml
# GoReleaser configuration for terraform-provider-stonebranch
# Documentation: https://goreleaser.com

version: 2

before:
  hooks:
    - go mod tidy

builds:
  # Existing Terraform provider build
  - id: terraform-provider-stonebranch
    binary: terraform-provider-stonebranch_v{{ .Version }}
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X main.version={{ .Version }}

  # NEW: sb2tf CLI utility build
  - id: sb2tf
    main: ./cmd/sb2tf
    binary: sb2tf
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X main.version={{ .Version }}

archives:
  # Provider archive (unchanged)
  - id: provider-zip
    builds:
      - terraform-provider-stonebranch
    formats:
      - zip
    name_template: "terraform-provider-stonebranch_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

  # NEW: sb2tf archive (separate downloads)
  - id: sb2tf-zip
    builds:
      - sb2tf
    formats:
      - zip
    name_template: "sb2tf_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums_{{ .Version }}_SHA256SUMS"
  algorithm: sha256

snapshot:
  version_template: "{{ incpatch .Version }}-dev"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
      - Merge pull request
      - Merge branch

# Publish to GitHub Releases
release:
  github:
    owner: OptionMetrics
    name: terraform-provider-stonebranch
  draft: false
  prerelease: auto
  name_template: "v{{ .Version }}"
```

### Key Changes:

1. **Second build entry** (`id: sb2tf`) - builds from `./cmd/sb2tf` with binary name `sb2tf`
2. **Separate archive** (`id: sb2tf-zip`) - creates `sb2tf_v1.0.0_darwin_amd64.zip` etc.
3. **Updated checksum** - single checksum file covers both binaries
4. **Same platforms** - builds for darwin/linux/windows on amd64/arm64

### Release Assets

After running `goreleaser release`, GitHub releases will include:

```
terraform-provider-stonebranch_1.0.0_darwin_amd64.zip
terraform-provider-stonebranch_1.0.0_darwin_arm64.zip
terraform-provider-stonebranch_1.0.0_linux_amd64.zip
terraform-provider-stonebranch_1.0.0_linux_arm64.zip
terraform-provider-stonebranch_1.0.0_windows_amd64.zip
sb2tf_1.0.0_darwin_amd64.zip
sb2tf_1.0.0_darwin_arm64.zip
sb2tf_1.0.0_linux_amd64.zip
sb2tf_1.0.0_linux_arm64.zip
sb2tf_1.0.0_windows_amd64.zip
checksums_1.0.0_SHA256SUMS
```

### User Installation

Users can download the sb2tf binary directly from GitHub releases:

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/OptionMetrics/terraform-provider-stonebranch/releases/latest/download/sb2tf_1.0.0_darwin_arm64.zip
unzip sb2tf_1.0.0_darwin_arm64.zip
chmod +x sb2tf
sudo mv sb2tf /usr/local/bin/

# Linux (x86_64)
curl -LO https://github.com/OptionMetrics/terraform-provider-stonebranch/releases/latest/download/sb2tf_1.0.0_linux_amd64.zip
unzip sb2tf_1.0.0_linux_amd64.zip
chmod +x sb2tf
sudo mv sb2tf /usr/local/bin/
```

---

## Estimated Complexity

- **Phase 1-3 (Core)**: ~400-500 lines of Go
- **Phase 4-6 (Resources)**: ~100-150 lines per resource type (templates + models)
- **Phase 7 (Workflows)**: ~200-300 lines for dependency resolution
- **Phase 8 (Polish)**: ~100 lines

Total: ~2000-3000 lines of new Go code

---

# Extension: Local-Directory Data Source (IN PROGRESS)

**Status as of this writing: research/discovery complete, NO code written yet.**
Phases 1-7 above are already implemented and working (API-only). This section
plans turning sb2tf into also reading from a local JSON export tree (produced by
the sibling repo `stonebranch-exporter-importer`'s `stonebranch_sync.py export`)
instead of a live UAC connection. Bundles are explicitly out of scope.

Local dataset used for dev/testing:
`/Users/bschimmel/Documents/bng/schedular/stonebranch-definitions-2.0/definitions`

**Key simplifying fact:** local JSON files are the raw UAC API response bodies
(round-tripped verbatim by the Python exporter's pydantic `BaseResource` which
uses `extra='allow'`). sb2tf's HCL templates already render off a plain
`map[string]interface{}` keyed by raw field names. So `json.Unmarshal` of a local
file into `map[string]interface{}` is a drop-in replacement for an API fetch.

## Actual current state of `cmd/sb2tf` (verified by reading full files, not skimmed)

### `cmd/sb2tf/generator/generator.go` (858 lines)

- `Generator` struct: `client *client.Client`, `output string`, `noDeps bool`,
  `exported map[string]bool`, `nameCounters map[string]int`,
  `hclBuffer *bytes.Buffer` (single buffer, becomes multi-buffer in this plan),
  `nameMap map[string]string`.
- `NewGenerator(client *client.Client, output string, noDeps bool) *Generator`
- Methods that hit the network directly (all via `g.client.Get`) — these all
  move into `APIDataSource` per the DataSource plan below:
  - `fetchResourceByName(ctx, endpoint, paramName, name)` (line ~462) — generic GET+unmarshal
  - `fetchResource(ctx, rt, name)` (line ~458) — thin wrapper calling the above with `rt.APIEndpoint`/`rt.NameQueryParam`
  - `listResources(ctx, rt, filter)` (line ~485) — GET `rt.ListEndpoint`, parses raw items using `rt.NameField` (default `"name"`), then locally filters by `rt.APITypeValue` if set. Also exported publicly as `ListResources` (line ~480, used by `cli/export.go`'s `dryRunExport`).
  - `fetchWorkflowVertices(ctx, workflowName)` (line ~578) — GET `/resources/workflow/vertices?workflowname=X`
  - `fetchWorkflowEdges(ctx, workflowName)` (line ~596) — GET `/resources/workflow/edges?workflowname=X`
  - `ExportTasks(ctx, filter)` (line ~144) — **dead code, not called anywhere in cli/**; hits `/resources/task/listadv` directly. Still convert its body to use the DataSource for consistency (mechanical), even though unreachable today.
  - `ExportWorkflow(ctx, name)` (line ~359) — **also dead/unused by current cli/export.go** (export.go calls `exportWorkflowComplete` via `ExportAll`/single-export path, not `ExportWorkflow`). Uses legacy `generateWorkflowVertexHCL`/`generateWorkflowEdgeHCL` (also unused elsewhere). Still refactor for consistency; low risk since unreachable.
  - Inside `exportWorkflowComplete` (line ~220, **this one IS used**, called from `ExportAll` and from `runExport` in cli/export.go for the workflow single-export path) — calls `g.fetchResource(ctx, rt, name)` for the workflow itself, then `g.fetchResourceByName(ctx, "/resources/task", "taskname", taskName)` per vertex to resolve each task's type before dispatching to `ExportResource`. This second call is exactly what `DataSource.FetchTaskByName` is for.
- `exportDependencies(ctx, rt, data)` (line ~545): **defined but never called from anywhere** (grep confirmed zero call sites). No client access inside it (just reads `data` map + calls `g.ExportResource`), so no DataSource change needed there — leave as dead code, don't worry about wiring it in (out of scope; not mentioned as a goal in this plan).
- `ResourceItem` struct (line ~536): `Name string`, `Type string`, `Task struct{ Value string }`. **Does NOT currently have `SysId`/`Summary` fields** — cli/list.go has its own separate, richer `ResourceItem` struct (see below). Plan: add `SysId string` and `Summary string` (both `json:",omitempty"`) to generator's `ResourceItem` so `list.go` can be unified onto it (see CLI section).
- `WorkflowVertex` struct (line ~614): `Task struct{ Value string }`, `VertexId string`, `Alias`, `VertexX`, `VertexY` (all optional). Matches local JSON exactly (verified against `tasks/workflow/1st_Month_Child_1.json`).
- `WorkflowEdge` struct (line ~625): `SourceId struct{Value string}`, `TargetId struct{Value string}`, `StraightEdge bool`. **Missing `Condition` field** — local JSON's edges have a `condition.value` key (verified in sample file) that isn't modeled. Plan: add
  ```go
  Condition struct {
      Value string `json:"value"`
  } `json:"condition,omitempty"`
  ```
  This is needed for both API and local modes per the original top-level plan; do it while touching this struct.
- `Finalize()` (line ~428): writes single `main.tf` (or stdout if `g.output==""`). Plan: change to per-category multi-file (see Architecture section below). **Note:** stdout mode (`output == ""`) — need to decide how multi-file interacts with stdout-only mode; simplest: when `g.output == ""`, concatenate all buffers to stdout in a stable (sorted-by-filename) order with header comments per section, same as today's single-buffer behavior. When `g.output != ""`, write one file per buffer.
- `generateHCLFromData` / `generateWorkflowVertexHCLNew` / `generateWorkflowEdgeHCLNew` (the "New" suffixed ones are the ones actually used by `exportWorkflowComplete`; the non-"New" ones are legacy/unused) — these only touch `g.hclBuffer` — will need to route to the correct named buffer instead (see Multi-file section).
- `GetTemplate func(resourceType string) *template.Template` — package-level var, set from `templates.go`'s `init()`. No change needed.

### `cmd/sb2tf/generator/resources.go` (421 lines) — confirmed exact current content

- `ResourceType` struct fields: `CLIName`, `TerraformResource`, `APIEndpoint`, `ListEndpoint`, `APITypeValue`, `NameQueryParam`, `NameField`, `HasTypeField`, `Category`, `Dependencies []Dependency`.
- `Dependency{Field, ResourceType, Condition}`.
- `resourceTypes` map currently has (verified, exact CLINames): `task_unix`, `task_windows`, `task_sql`, `task_email`, `task_workflow`, `task_file_monitor`, `task_file_transfer`, `task_timer`, `task_monitor`, `task_stored_procedure`, `task_web_service`, `task_universal_aws_s3`, `trigger_time`, `trigger_cron`, `trigger_file_monitor`, `trigger_task_monitor`, `database_connection`, `email_connection`, `script`, `variable`, `credential`, `business_service`, `agent_cluster`, `calendar`, `workflow_vertex`, `workflow_edge`.
  - **Bug confirmed:** `task_timer` entry has `APITypeValue: "taskTimer"` (line ~160) — real UAC API type value is `"taskSleep"` (confirmed via local dataset: task files under `tasks/sleep/*.json` all have `"type": "taskSleep"`). Fix: change to `"taskSleep"`.
  - `task_universal_aws_s3` entry (line ~201) has `APITypeValue: "taskUniversal"` — this stays registered as-is (still usable explicitly), just no longer the *default* mapping target for bare `taskUniversal`.
- `APITypeToResourceType` map (line ~404, exact contents):
  ```go
  var APITypeToResourceType = map[string]string{
      "taskUnix":        "task_unix",
      "taskWindows":     "task_windows",
      "taskSql":         "task_sql",
      "taskEmail":       "task_email",
      "taskWorkflow":    "task_workflow",
      "taskFileMonitor": "task_file_monitor",
      "taskFtp":         "task_file_transfer",
      "taskTimer":       "task_timer",       // BUG: real API value is "taskSleep", not "taskTimer" — this key is essentially dead/wrong
      "taskMonitor":     "task_monitor",
      "taskStoredProc":  "task_stored_procedure",
      "taskWebService":  "task_web_service",
      "taskUniversal":   "task_universal_aws_s3", // change target to new generic "task_universal"
      "triggerTime":     "trigger_time",
      "triggerCron":     "trigger_cron",
      "triggerFm":       "trigger_file_monitor",
      "triggerTm":       "trigger_task_monitor",
  }
  ```
  Planned edits:
  - Remove/replace `"taskTimer": "task_timer"` entry with `"taskSleep": "task_timer"`.
  - Change `"taskUniversal": "task_universal_aws_s3"` → `"taskUniversal": "task_universal"` (new generic resource).
  - Add `"taskRecurring": "task_recurring"` (new).
- `GetResourceCategories()` groups by fixed category names: `Tasks`, `Triggers`, `Connections`, `Other`, `Workflow`. New non-task/trigger types (`customday`, `emailtemplate`, `universaltemplate`, `virtualresource`) should go in category `"Other"` (matches existing `script`/`variable`/`credential`/`calendar` placement — there is no dedicated category for them and the plan doesn't call for adding one).
- New registry entries needed (exact values determined from provider resource files, see below):
  ```go
  "task_recurring": {
      CLIName: "task_recurring", TerraformResource: "stonebranch_task_recurring",
      APIEndpoint: "/resources/task", ListEndpoint: "/resources/task/listadv",
      APITypeValue: "taskRecurring", NameQueryParam: "taskname",
      HasTypeField: true, Category: "Tasks",
  },
  "task_universal": {
      CLIName: "task_universal", TerraformResource: "stonebranch_task_universal",
      APIEndpoint: "/resources/task", ListEndpoint: "/resources/task/listadv",
      APITypeValue: "taskUniversal", NameQueryParam: "taskname",
      HasTypeField: true, Category: "Tasks",
      Dependencies: []Dependency{
          {Field: "credentials", ResourceType: "credential"},
          {Field: "agentCluster", ResourceType: "agent_cluster"},
      },
  },
  "customday": {
      CLIName: "customday", TerraformResource: "stonebranch_custom_day",
      APIEndpoint: "/resources/customday", ListEndpoint: "/resources/customday/list", // verify list endpoint suffix against a live/known-good pattern; not confirmed empirically, follow "script"-style /list convention used by similarly-simple resources
      NameQueryParam: "customdayname", NameField: "name",
      HasTypeField: false, Category: "Other",
  },
  "emailtemplate": {
      CLIName: "emailtemplate", TerraformResource: "stonebranch_email_template",
      APIEndpoint: "/resources/emailtemplate", ListEndpoint: "/resources/emailtemplate/list",
      NameQueryParam: "templatename", NameField: "templateName",
      HasTypeField: false, Category: "Other",
  },
  "universaltemplate": {
      CLIName: "universaltemplate", TerraformResource: "stonebranch_universal_template",
      APIEndpoint: "/resources/universaltemplate", ListEndpoint: "/resources/universaltemplate/list",
      NameQueryParam: "templatename", NameField: "name", // verify query param name empirically if API mode is ever exercised; local mode doesn't need it
      HasTypeField: false, Category: "Other",
  },
  "virtualresource": {
      CLIName: "virtualresource", TerraformResource: "stonebranch_virtual_resource",
      APIEndpoint: "/resources/virtual", ListEndpoint: "/resources/virtual/list", // NOTE: real endpoint is /resources/virtual, NOT /resources/virtualresource (confirmed in virtualresource.go comments)
      NameQueryParam: "resourcename", NameField: "name",
      HasTypeField: false, Category: "Other",
  },
  ```
  (List endpoints for the four new non-task types are **not empirically confirmed** since this plan's verification only exercises local-directory mode, which never calls `List`/`ListEndpoint` via the API. Follow the `/resources/<endpoint>/list` convention already used by `script`/`credential`/etc. If API-mode listing of these types is ever needed, verify against a live instance first.)

### `cmd/sb2tf/generator/templates.go` (934 lines) — confirmed exact current content

- `init()` registers templates via `registerTemplate(name, tmplString)` into `templates map[string]*template.Template`, and sets package-level `GetTemplate` func var (defined in generator.go).
- Template helpers (`quote`, `stringList`, `notEmpty`, `isTrue`) — defined here, registered into each template's FuncMap.
  - `notEmpty` only handles `string`/`[]interface{}`/`bool` cases — returns `false` for `nil` map (e.g. a numbered Universal Template field slot like `.textField1` is either `nil` or a `map[string]interface{}{"label":..,"name":..,"value":..}`). **Need a new helper** for task_universal's numbered slots — plan calls it `fieldValue`-style helper. Concretely, add:
    ```go
    // fieldSet reports whether a Universal Template numbered field slot
    // (an object with label/name/value keys) has a non-empty "value".
    func fieldSet(v interface{}) bool {
        m, ok := v.(map[string]interface{})
        if !ok {
            return false
        }
        s, ok := m["value"].(string)
        return ok && s != ""
    }
    ```
    and register it as `"fieldSet": fieldSet` in `registerTemplate`'s FuncMap. Template usage: `{{- if fieldSet .textField1}}` ... `{{quote .textField1.value}}`.
- **Confirmed bug in `taskTimerTemplate`** (line ~632-651): uses `.timerType`/`.timerDuration`/`.timerTime` and emits `timer_type`/`timer_duration`/`timer_time` HCL attributes. Real JSON fields (confirmed via local dataset `tasks/sleep/*.json`) are `sleepType`/`sleepDuration`/`sleepTime`/`sleepDayConstraint`, and real TF schema attributes (per `internal/provider/resources/task_timer.go`) are `sleep_type`/`sleep_duration`/`sleep_time`/`sleep_day_constraint`. **This bug currently has zero effect** because `APITypeValue: "taskTimer"` never matches any real data (see above) — so no taskSleep record was ever routed through this template. Once the `taskTimer`→`taskSleep` fix lands, this template bug becomes live and MUST be fixed too (needed for `terraform validate` to pass in the verification step), even though the original top-level plan text didn't call it out explicitly — it's a direct consequence of fixing the type-value bug. New corrected template body:
  ```
  {{- if notEmpty .sleepType}}
    sleep_type = "{{quote .sleepType}}"
  {{- end}}
  {{- if notEmpty .sleepDuration}}
    sleep_duration = "{{quote .sleepDuration}}"
  {{- end}}
  {{- if notEmpty .sleepTime}}
    sleep_time = "{{quote .sleepTime}}"
  {{- end}}
  {{- if notEmpty .sleepDayConstraint}}
    sleep_day_constraint = "{{quote .sleepDayConstraint}}"
  {{- end}}
  ```
  (Need to double check `task_timer.go`'s exact schema attribute list/required-ness before finalizing — it was read but exact required/optional flags for these four weren't specially noted; assume all Optional based on skim, verify when implementing.)
- Existing templates all follow the same shape: `resource "{{._terraformResource}}" "{{._resourceName}}" { name = "{{quote .name}}" ... }` with `{{- if notEmpty .field}} field = ... {{- end}}` blocks, string fields via `quote`, lists via `stringList`, booleans via `isTrue` (only emitted when true, since schema defaults are false), numbers emitted raw (no quotes) e.g. `{{.maxRows}}`.
- `workflow_vertex`/`workflow_edge` templates take a flat synthetic `map[string]interface{}` built in generator.go (not raw API data) — untouched by this plan except that `workflowEdgeTemplate` should gain a `condition`/branch-condition line once `WorkflowEdge.Condition` is added and `generateWorkflowEdgeHCLNew`'s `data` map includes it. **Need to check `internal/provider/resources` for the actual workflow_edge resource's schema attribute name for the branch condition** (not yet done — check `workflow_edge.go` or similar for the field name, likely `condition` or `branch_condition`, before writing this template change).

## Provider resource schemas confirmed (for new templates), exact findings

All read in full from `internal/provider/resources/`:

### `task_recurring.go`
- Metadata suffix: need to double check exact `resp.TypeName` suffix used (recorded as matching `stonebranch_task_recurring` per plan naming — confirm exact string when implementing, wasn't explicitly quoted in notes above but schema fields below were captured from the model).
- Confirmed schema/API fields (top-level, non-nested): `name` (Required), `target_task`/`targetTask` (Required), `number_of_recurrences`/? (conditionally required via ValidateConfig based on time_window/indefinite flags — just pass through whatever's in the source JSON, no extra logic needed), plus presumably `time_window`, `indefinite`, `interval`-style fields (exact tfsdk/json names need re-confirmation when writing the template — re-read `task_recurring.go` at implementation time to get the definitive attribute list, since these notes summarize but may not be 100% exhaustive).
- Skip (out of scope, consistent with how ALL existing templates skip these on task_unix etc.): `variables`, `exclusive_tasks`/`exclusiveTasks`, `virtual_resources`/`virtualResources`, `actions`.
- Sample local JSON confirmed at `tasks/recurring/*.json` (8 files) — matches expected shape (`type: "taskRecurring"`, `targetTask`, `name`, etc.)

### `task_universal.go` (generic Universal Template task resource)
- Metadata suffix: `_task_universal` → `stonebranch_task_universal`.
- `template` attribute: **Required**, string — the Universal Template's name (JSON field `template`, matches TF attribute `template`). Always emit unconditionally (or with quote but no notEmpty guard, since Required).
- Numbered field slots — confirmed via full file read: attribute names use underscore-before-number convention, e.g. `text_field_1`...`text_field_10` (NOT `text_field1`), `choice_field_1`...`choice_field_11`, `boolean_field_1`...`boolean_field_7`, `credential_field_1`...`credential_field_4`, `credential_var_field_1`, `credential_var_field_4` (only 1 and 4 exist per plan text), `custom_field_1`...`custom_field_2`, `int_field_1`, `large_text_field_1`...`large_text_field_4`, `script_field_1`...`script_field_2`, `script_var_field_2` (only field 2 exists).
  JSON field names use NO underscore before the number (confirmed via local sample `tasks/universal/*.json`): `textField1`, `choiceField1`, `booleanField1`, `credentialField1`, etc. Each is an object `{label, name, value}` in JSON; in the TF schema each is a nested object attribute with (at minimum) a `value` string sub-attribute (need to re-confirm exact nested attribute names — `label`/`name` might be Computed-only or might not be modeled at all in the TF schema, i.e. only `value` might be settable via TF while label/name are template-defined metadata. **Re ­read the relevant section of `task_universal.go` at implementation time** to confirm exactly which sub-attributes are user-settable vs computed, since only user-settable ones should be emitted.)
  Use the new `fieldSet`/`.value` pattern noted above for each slot: `{{- if fieldSet .textField1}} text_field_1 = { value = "{{quote .textField1.value}}" } {{- end}}` — **exact HCL shape (object literal vs separate nested block) depends on whether the schema attribute is `ObjectAttribute`/`SingleNestedAttribute`/etc. Re-confirm the schema attribute TYPE (not just name) for these slots before finalizing the template syntax**, since Terraform HCL for an object-typed attribute (`{ value = "x" }`) differs from a nested block (no `=`).
- Sample local JSON confirmed at a `taskUniversal`-typed file: top-level keys include `textField1`, `choiceField1`, `booleanField1`, `credentialField1`, `template`, `agent`, `agentCluster`, `credentials`, `summary`, `name`, `type`, plus the usual bookkeeping (`sysId`, `exportReleaseLevel`, `exportTable`, `version`). Each field-slot value looks like `{"label": "...", "name": "...", "value": "..."}` or `null`.

### `customday.go` (`CustomDayResource`)
- Metadata: `resp.TypeName = req.ProviderTypeName + "_custom_day"` → **`stonebranch_custom_day`**, NOT `stonebranch_customday`. CLIName can stay `customday` (directory-derived, doesn't need to match TF type name) but `TerraformResource` field MUST be `stonebranch_custom_day`.
- Endpoint confirmed: `/resources/customday` (Get/Post/Put), delete uses query param `customdayid`=sysId, read uses query param `customdayname`.
- Schema attributes (all confirmed, full file read):
  - Identity: `sys_id` (Computed), `name` (Required, json `name`), `version` (Computed).
  - `comments` (Optional, json `comments`).
  - `category` (Computed-only — **do NOT emit in template**, server-derived from holiday/period).
  - `ctype` (Required, json `ctype`) — one of `Single Date`/`List of Dates`/`Absolute Repeating Date`/`Relative Repeating Date`.
  - `date` (Optional, json `date`).
  - `date_list` (Optional list of strings, json `dateList`) → use `stringList`.
  - `month` (Optional/Computed, json `month`), `day` (Optional/Computed int64, json `day`), `dayofweek` (Optional/Computed, json `dayofweek` — note: all-lowercase, not camelCase), `relfreq` (Optional/Computed, json `relfreq`), `nth_amount` (Optional/Computed int64, json `nthAmount`), `nth_type` (Optional/Computed, json `nthType`).
  - `adjustment` (Optional/Computed, json `adjustment`), `adjustment_amount` (Optional/Computed int64, json `adjustmentAmount`), `adjustment_type` (Optional/Computed, json `adjustmentType`).
  - `holiday` (Optional/Computed bool, json `holiday`), `period` (Optional/Computed bool, json `period`).
  - `observed_rules` (Optional list-nested, json `observedRules`, each item `{actual_day_of_week (Required) / json actualDayOfWeek, observed_day_of_week (Required) / json observedDayOfWeek}`).
- Sample local JSON confirmed (`customdays/TARGET2_HOLIDAYS.json`): matches exactly, e.g. `"ctype": "List of Dates"`, `"dateList": [...]`, `"observedRules": []`, `"holiday": true`, `"period": false`.
- Since `date_list`/`observed_rules` are the only two nested/list attrs and both are simple, template should handle them like the existing `calendar`/`agentCluster` templates handle their string lists (`stringList` helper for `date_list`; for `observed_rules`, need a small `{{range .observedRules}}` block emitting nested `observed_rules { actual_day_of_week = ...; observed_day_of_week = ... }` — check whether schema uses **blocks** or **list attribute with `=`**; it's `ListNestedAttribute` so it's assigned with `=` and a list-of-objects literal: `observed_rules = [{ actual_day_of_week = "...", observed_day_of_week = "..." }, ...]`.

### `emailtemplate.go` (`EmailTemplateResource`)
- Metadata: `_email_template` → **`stonebranch_email_template`**.
- Endpoint: `/resources/emailtemplate`; read query param `templatename`; delete query param `templateid`=sysId.
- Schema (confirmed): `sys_id` (Computed), `name` (Required, **json `templateName`**, NOT `name`), `version` (Computed), `description` (Optional, json `description`), `email_connection` (Required, **json `connection`**, NOT `emailConnection`), `reply_to` (Optional, json `replyTo`), `to`/`cc`/`bcc` (all Optional strings, json `to`/`cc`/`bcc`; ValidateConfig requires at least one of the three to be set — trust source data, no extra logic needed), `subject` (Optional, json `subject`), `body` (Optional, json `body`), `opswise_groups` (Optional list of strings, json `opswiseGroups`).
- Sample local JSON confirmed (`email-templates/Notification_Email_Template.json`): exactly matches — `"templateName": "Notification_Email_Template"`, `"connection": "SBP_Email_Connection"`, `"to": "...@..."`, `"subject": "Task ${workflow} has failed"`, etc. **Note the literal `${workflow}` in `subject`** — the existing `quote` helper escapes `$` → `$$` specifically to prevent Terraform interpolation of such literal UAC placeholder syntax, so this is already handled correctly by reusing `quote`.

### `universaltemplate.go` (`UniversalTemplateResource`, 1553 lines — only read ~550 lines, NOT the full file)
- Metadata: `_universal_template` → **`stonebranch_universal_template`**.
- Top-level Required attributes (confirmed): `name` (json `name`), `variable_prefix` (json `variablePrefix`), `agent_type` (json `agentType`), `exit_codes` (json `exitCodes`) — **all four must be rendered unconditionally** (no `notEmpty` guard needed since Required just means "must have a value", and empty-string source data would still produce valid-if-unlikely HCL; but prefer rendering them plainly without conditionals to guarantee presence).
- Full confirmed top-level attribute list (tfsdk name → json name), all Optional unless noted:
  `description`/`description`, `extension`/`extension`, `template_type`/`templateType` (Optional+Computed), `use_common_script`/`useCommonScript` (bool, Optional+Computed), `script`/`script`, `script_unix`/`scriptUnix`, `script_windows`/`scriptWindows`, `script_type_windows`/`scriptTypeWindows`, `always_cancel_on_finish`/`alwaysCancelOnFinish` (bool), `credentials`/`credentials`, `credentials_var`/`credentialsVar`, `agent`/`agent`, `agent_var`/`agentVar`, `agent_cluster`/`agentCluster`, `agent_cluster_var`/`agentClusterVar`, `broadcast_cluster`/`broadcastCluster`, `broadcast_cluster_var`/`broadcastClusterVar`, `runtime_dir`/`runtimeDir`, `environment`/`environment` (list-nested of `{name (Required), value}` pairs — json `environment: [{name,value}]`), `send_environment`/`sendEnvironment` (Optional+Computed), `send_variables`/`sendVariables` (Optional+Computed), `exit_code_processing`/`exitCodeProcessing`, `exit_code_text`/`exitCodeText`, `exit_code_output`/`exitCodeOutput`, `output_type`/`outputType`, `output_content_type`/`outputContentType`, `output_path_expression`/`outputPathExpression`, `output_condition_operator`/`outputConditionOperator`, `output_condition_value`/`outputConditionValue`, `output_condition_strategy`/`outputConditionStrategy`, `auto_cleanup`/`autoCleanup` (bool), `output_return_type`/`outputReturnType`, `output_return_file`/`outputReturnFile`, `output_return_sline`/`outputReturnSline`, `output_return_nline`/`outputReturnNline`, `output_return_text`/`outputReturnText`, `wait_for_output`/`waitForOutput` (bool), `output_failure_only`/`outputFailureOnly` (bool), `elevate_user`/`elevateUser` (bool), `desktop_interact`/`desktopInteract` (bool), `create_console`/`createConsole` (bool), `log_level`/`logLevel`, `fields`/`fields` (list-nested, see below).
  (Sub-attribute Required/Optional flags for `environment` items were confirmed: `name` Required, `value` Optional. NOT yet confirmed line-by-line for every one of the ~30 top-level attrs above whether each is plain-Optional vs Optional+Computed — doesn't matter for template purposes since we only ever *emit* a value when present in source data via `notEmpty`, regardless of Computed-ness, EXCEPT: attributes that are **Computed-only without Optional** must never be emitted (would be a config error). None of the ones listed above are Computed-only-without-Optional based on what was read — but this was not exhaustively double-checked for every single attribute in the 1553-line file. Re-verify quickly at implementation time by grep'ing for `Computed:\s*true` blocks NOT paired with `Optional:\s*true` in `universaltemplate.go`.)
- `fields` nested list — each item confirmed (from `UniversalTemplateFieldModel`/`UniversalTemplateFieldAPIModel`, both fully read):
  tfsdk name → json name (all on one nested object per field):
  `name`/`name` (Required), `label`/`label` (Required), `field_mapping`/`fieldMapping` (Required), `field_type`/`fieldType`, `hint`/`hint`, `text_type`/`textType`, `field_value`/`fieldValue`, `default_list_view`/`defaultListView` (bool), `allow_variable`/`allowVariable` (bool), `field_restriction`/`fieldRestriction`, `preserve_output_on_rerun`/`preserveOutputOnRerun` (bool), `extension_status`/`extensionStatus` (bool), `boolean_value_type`/`booleanValueType`, `boolean_yes_value`/`booleanYesValue`, `boolean_no_value`/`booleanNoValue`, `choice_sort_option`/`choiceSortOption`, `choice_allow_empty`/`choiceAllowEmpty` (bool), `choice_allow_multiple`/`choiceAllowMultiple` (bool), `choice_dynamic`/`choiceDynamic` (bool), `choice_fields`/`choiceFields` (list of strings), `choices`/`choices` (list-nested of `{field_value/fieldValue, field_value_label/fieldValueLabel, use_field_value_for_label/useFieldValueForLabel (bool)}` — note JSON `choices[]` items ALSO have a `sequence` and `sysId` key not modeled in TF schema, skip those two), `required`/`required` (bool), `require_if_field`/`requireIfField`, `require_if_field_value`/`requireIfFieldValue`, `show_if_field`/`showIfField`, `show_if_field_value`/`showIfFieldValue`, `require_if_visible`/`requireIfVisible` (bool), `preserve_value_if_hidden`/`preserveValueIfHidden` (bool), `no_space_if_hidden`/`noSpaceIfHidden` (bool), `field_length`/`fieldLength` (int64), `int_field_min`/`intFieldMin` (int64), `int_field_max`/`intFieldMax` (int64), `field_regex`/`fieldRegex`, `form_column_span`/`formColumnSpan` (int64), `form_start_row`/`formStartRow` (bool), `form_end_row`/`formEndRow` (bool), `array_name_title`/`arrayNameTitle`, `array_value_title`/`arrayValueTitle`, `array_field_value`/`arrayFieldValue` (list-nested of `{name (Required), value}` — same shape as top-level `environment`).
  JSON field object also has a `sequence` key (int) not modeled in TF schema — skip.
- Sample local JSON confirmed at `universal-templates/CS Azure Blob Storage.json`: top-level keys match expectations; one `fields[0]` example fully inspected and matches the `UniversalTemplateFieldAPIModel` shape exactly, including a 9-item `choices` array with `fieldValue`/`fieldValueLabel`/`sequence`/`sysId`/`useFieldValueForLabel` per choice.
- **This is by far the most complex new template.** Plan: implement `fields` as a Go template `{{range .fields}}` block emitting a `fields { ... }` per the schema's actual HCL shape (list-nested attribute means `fields = [ { ... }, { ... } ]` object-list syntax, NOT repeated `fields {}` blocks — schema.ListNestedAttribute is always assigned with `=`), with a nested `{{range .choices}}`/`{{range .array_field_value}}`/`{{range .choice_fields}}` for the three sub-lists. Given the size, budget significant care to get comma placement right in the generated object-list literal (Go templates don't have great native support for "no trailing comma" — either accept trailing commas, which HCL permits, or use `{{if $i}}, {{end}}` index-based separators like the existing legacy `workflowEdgeTemplate`-adjacent code in generator.go does for its own inline lists).
- **Not yet done:** confirming the exact `template_type`/other Computed-vs-Optional edge cases (see above), and NOT yet read lines 550-1553 of the file (Configure/Create/Read/Update/Delete/toAPIModel/fromAPIModel — almost certainly irrelevant to template-writing, but the `ValidateConfig` if present might reveal additional cross-field constraints worth knowing, similar to customday's). Low risk to skip since we're just passing through already-valid exported data, not authoring new configs.

### `virtualresource.go` (`VirtualResourceResource`)
- Metadata: `_virtual_resource` → **`stonebranch_virtual_resource`**.
- **Real API endpoint is `/resources/virtual`, NOT `/resources/virtualresource`** (explicitly commented in source). Delete query param `resourceid`=sysId, read query param `resourcename`.
- Schema (confirmed, simple): `sys_id` (Computed), `name` (Required, json `name`), `version` (Computed), `limit` (Optional+Computed int64, json `limit`), `summary` (Optional, json `summary`), `type` (Optional+Computed, json `type` — values `Renewable`/`Boundary`/`Depletable`), `opswise_groups` (Optional list of strings, json `opswiseGroups`).
- Sample local JSON confirmed (`virtual-resources/udm_virtual_resource.json`): has `limit`, `name`, `type: "Renewable"`, `used` (int, **not in TF schema, read-only computed-by-usage, skip**), `usedByTasks` (**not in TF schema at all, skip entirely** — it's a list of `{amount, taskInstanceName}` showing current consumers, irrelevant to the resource definition), `version`.

### `task_timer.go` (existing resource, re-confirming the bug)
- Metadata suffix confirmed elsewhere to be `stonebranch_task_timer` (already registered correctly in resources.go's `TerraformResource` field — only `APITypeValue` is wrong).
- Real JSON/schema field names: `sleepType`/`sleep_type`, `sleepDuration`/`sleep_duration`, `sleepTime`/`sleep_time`, `sleepDayConstraint`/`sleep_day_constraint` (this last one wasn't in the original top-level plan's notes — discovered while reading `task_timer.go`; the current buggy template also doesn't handle it at all — include it in the fix).

## Local dataset directory layout (confirmed via `find`/`python3 -c` on the actual export tree)

Root: `/Users/bschimmel/Documents/bng/schedular/stonebranch-definitions-2.0/definitions`

- `tasks/<subdir>/*.json` — subdirs seen include `unix`, `workflow`, `universal`, `ftp`, `email`, `sleep` (⚠️ dir is named `sleep`, matching the real API type `taskSleep`, NOT `timer`), `filemonitor`(?), `recurring`, `sql`, `monitor`, `ucmd`. **Classification for LocalDataSource must use the JSON `type` field, NOT the subdirectory name** (already the plan's stated approach — confirmed correct/necessary since dir names don't reliably map 1:1 to CLI type strings anyway).
  Confirmed per-type counts via `python3` Counter over all `tasks/*/*.json`: matches the top-level plan's inventory table (taskUnix 145, taskWorkflow 174, taskUniversal 60, taskFtp 24, taskEmail 17, taskSleep 13, taskFileMonitor 8, taskRecurring 8, taskSql 8, taskMonitor 7, taskUcmd 1).
- `triggers/*.json` — flat, discriminated by `type` field (`triggerTime`, `triggerFm`, `triggerCron`, `triggerTm`).
- `scripts/*.json` — flat, name field `scriptName`.
- `calendars/*.json` — flat, name field `name`.
- `customdays/*.json` — flat, name field `name`. **1 file** (`TARGET2_HOLIDAYS.json`), fully inspected, matches `CustomDayAPIModel` exactly.
- `email-templates/*.json` — flat, name field `templateName`. **5 files**, one fully inspected, matches `EmailTemplateAPIModel` exactly.
- `universal-templates/*.json` — flat, name field `name`. **12 files**, one fully inspected (`CS Azure Blob Storage.json`).
- `virtual-resources/*.json` — flat, name field `name`. **1 file** (`udm_virtual_resource.json`), fully inspected, matches `VirtualResourceAPIModel` (plus extra read-only `used`/`usedByTasks` keys to ignore).
- No `credentials/`, `variables/`, `database-connections/`, `email-connections/`, `agent-clusters/`, `business-services/`, or `bundles/` dirs present in this dataset (confirmed via top-level `ls`/`find` — exact command was interrupted by the user before output was captured, but this matches the top-level plan's stated expectation and the `stonebranch-exporter-importer` repo's own `CLAUDE.md` resource-type list, so treat as confirmed by cross-reference even though the direct `ls` output wasn't captured in this session). **Before writing `LocalDataSource`'s `filepath.WalkDir`, re-run `find "$DIR" -maxdepth 1 -type d` once to get the definitive top-level directory list** (this was the exact command interrupted — just re-run it, trivial).
- `.index.json` exists at the root (per `stonebranch-exporter-importer`'s `FileIndex`) — must be skipped by `LocalDataSource`'s walk (per top-level plan, already noted).

## Remaining architecture work (all still TODO, nothing coded yet)

Follow the original top-level plan's Architecture section 1-5 exactly, using the corrected/confirmed details captured above. In particular:

1. **`datasource.go`**: use `ctx context.Context` as the first parameter on every `DataSource` interface method (deviates slightly from the top-level plan's pseudocode, which omitted `ctx` for brevity) — this matches every existing `Generator`/`client.Client` method signature in this codebase and is required since `APIDataSource` wraps HTTP calls. Move `fetchResourceByName`→`Fetch`/`FetchTaskByName`, `listResources`→`List`, `fetchWorkflowVertices`/`fetchWorkflowEdges`→same-named `DataSource` methods, verbatim logic, into `APIDataSource`. Add `SysId`/`Summary` fields to `generator.ResourceItem` (currently only `Name`/`Type`/`Task`) so `APIDataSource.List` and `LocalDataSource.List` can both populate them, enabling `cli/list.go` to drop its own separate `ResourceItem` struct and unify onto `generator.ResourceItem`.
2. **`Generator` struct**: replace `client *client.Client` field with `dataSource DataSource`; `NewGenerator(ds DataSource, output string, noDeps bool) *Generator`.
3. **Multi-file `Finalize`**: replace `hclBuffer *bytes.Buffer` with `buffers map[string]*bytes.Buffer` keyed by output filename (e.g. `tasks_unix.tf`, `tasks_workflow.tf`, `triggers.tf`, `scripts.tf`, `custom_days.tf`, `email_templates.tf`, `universal_templates.tf`, `virtual_resources.tf`, etc.) — need a small helper mapping `rt.Category` + `rt.CLIName` → filename (e.g. all `task_*` CLINames under category `Tasks` → `tasks_<suffix>.tf` where suffix strips the `task_` prefix; `trigger_*` → single shared `triggers.tf` per the top-level plan's file list; non-task/trigger types get their own file named after CLIName, e.g. `script`→`scripts.tf`, `calendar`→`calendars.tf`, `customday`→`custom_days.tf`, etc.). All the places currently writing to `g.hclBuffer` (`ExportResource`, `exportWorkflowComplete`, legacy `ExportWorkflow`/`outputHCL`) need to look up/create the right buffer via a helper like `g.bufferFor(rt *ResourceType) *bytes.Buffer`.
4. **Registry + template + CLI wiring**: as specified in the top-level plan, using the corrected values captured in this document (fixed `task_timer` APITypeValue, fixed `APITypeToResourceType` entries, new registry entries with confirmed `TerraformResource` names, new templates with confirmed attribute lists).
5. **`cli/list.go` (317 lines, fully read)**: the top-level plan understated this — `list.go` does NOT call `generator.NewGenerator` at all today; it calls `GetClient().Get(...)` directly in `listAllTasks`/`listAllTriggers`/`listAllConnections`/`listSpecificType`, using its own local `ResourceItem` struct (`SysId`/`Name`/`Type`/`Summary`). Plan (decided during this session, not yet implemented):
   - Delete `list.go`'s own `ResourceItem` struct; reuse `generator.ResourceItem` (after adding `SysId`/`Summary` to it per point 1 above).
   - `listAllTasks`/`listAllTriggers`: iterate `generator.GetResourceCategories()`'s `Tasks`/`Triggers` category types respectively, call `ds.List(ctx, rt, listFilter)` per type, concatenate results (each `rt.List` call already filters to its own `APITypeValue` subtype internally, so the union across all types in the category reproduces the old single-combined-endpoint behavior). Set `item.Type = rt.APITypeValue` when the item's `Type` came back empty (parity with old behavior where the combined `listadv` endpoint always populated `type`).
   - `listAllConnections`: iterate the `Connections` category types (`database_connection`, `email_connection`) via `ds.List`, setting `item.Type = rt.CLIName` (matches old hardcoded `"database_connection"`/`"email_connection"` type tagging).
   - `listSpecificType`: replace the raw `client.Get`+manual-parse block with a direct `ds.List(ctx, rt, listFilter)` call, keeping the existing "default Type to resourceType if empty" fallback loop and `printResourceTable` call.
   - This means `list.go` needs `GetDataSource()` (from `root.go`, see below) instead of `GetClient()`, and can drop its `encoding/json`/`net/url` imports.
   - Trade-off accepted: this changes API-mode `list` from 1 combined network call to N per-category calls for the `tasks`/`triggers` shortcuts. Acceptable per the top-level plan's explicit instruction to route `list.go` through `DataSource` too.
6. **`cli/root.go` (105 lines, fully read)**: add `--source-dir`/`-s` persistent string flag; in `initClient`, if `sourceDir != ""`, skip the "API token required" / URL checks entirely and don't construct `apiClient`; add `GetDataSource() generator.DataSource` that returns a lazily-constructed `*generator.LocalDataSource` (built once, cached) when `sourceDir != ""`, else `generator.NewAPIDataSource(apiClient)`. Keep `GetClient()` around only if still needed elsewhere (grep confirmed it's used by `export.go` and `list.go` only — both being migrated to `GetDataSource()` — so `GetClient()` can likely be removed entirely once both callers are migrated; double check no other callers before deleting).
7. **`cli/export.go` (224 lines, fully read)**: single change — `client := GetClient()` + `generator.NewGenerator(client, output, noDeps)` (line 59/63) becomes `ds := GetDataSource()` + `generator.NewGenerator(ds, output, noDeps)`. Everything else (`exportAllResources`, `exportCategory`, `dryRunExport`, category shortcuts `tasks`/`triggers`/`connections`) is unchanged — confirmed these already only call `gen.ExportAll`/`gen.ExportResource`/`gen.ExportWorkflow`/`gen.Finalize`/`gen.ListResources`/`gen.GetExportedResources`, none of which need CLI-level changes since the DataSource swap happens inside `Generator`.

## Open questions / things to verify empirically at implementation time (not blocking, just flagged)

- Exact `task_recurring.go` full attribute list beyond `name`/`target_task` (re-read the file fresh when writing the template — it WAS fully read this session but exact field names for the recurrence-interval/time-window attributes weren't transcribed verbatim into this doc; re-read takes one tool call).
- `universaltemplate.go` lines 550-1553 not read (Configure/CRUD/toAPIModel/fromAPIModel + possible `ValidateConfig`) — low risk to skip, but if a `ValidateConfig` exists with cross-field constraints worth knowing about (like `customday.go`'s), a quick grep for `func.*ValidateConfig` would confirm.
- `workflow_edge`'s branch-condition attribute name in the TF schema — not yet located; grep `internal/provider/resources` for the workflow edge resource (likely `workflow_edge.go` or similar) to get the exact attribute name before wiring `WorkflowEdge.Condition` through to `workflowEdgeTemplate`.
- List/query-param details for the 4 new non-task registry entries (`customday`, `emailtemplate`, `universaltemplate`, `virtualresource`) are best-guess-by-convention, not empirically verified against a live UAC instance — irrelevant for local-mode verification (this plan's actual verification path) but should be double-checked before this code path is ever exercised in API mode.
- Exact HCL literal syntax needed for `fields`/`environment`/`array_field_value`/`choices`/`observed_rules` (list-nested-attribute-as-object-list vs repeated blocks) should be sanity-checked against ONE existing example of a `ListNestedAttribute` being both written by a human and consumed successfully, if one exists elsewhere in the repo's example `.tf` files/docs, to nail the exact comma/bracket formatting Terraform expects (trailing commas in `[...]` lists are fine in HCL2, so this is low-risk, but glancing at `docs/resources/*.md` examples for `email_connection`/similar list-nested-attribute resources would remove all doubt).

## Next steps when resuming (in order)

1. Re-run `find "$DIR" -maxdepth 1 -type d` (was interrupted) to nail the definitive top-level directory list for `LocalDataSource`'s classification table.
2. Write `cmd/sb2tf/generator/datasource.go` (DataSource interface + APIDataSource + LocalDataSource), per Architecture section 1 of the top-level plan and the confirmed details above.
3. Refactor `generator.go` (Generator struct, NewGenerator, all `g.client.*`/`g.fetch*`/`g.listResources` call sites, add `WorkflowEdge.Condition`, multi-file `Finalize`+buffer-routing helper).
4. Apply the registry fixes/additions to `resources.go`.
5. Add the 6 new templates + `fieldSet` helper + `taskTimerTemplate` fix to `templates.go`.
6. Wire `root.go` (`--source-dir`, `GetDataSource`), `export.go` (one-line swap), `list.go` (bigger refactor per point 5 above).
7. `go build ./... && go vet ./...`, then run the verification commands from the top-level plan's Verification section against the local dataset, spot-checking the resource types called out there plus the newly-added ones (`task_recurring`, `task_universal`, `customday`, `emailtemplate`, `universaltemplate`, `virtualresource`).
