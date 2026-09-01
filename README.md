# Terraform Provider for StoneBranch

A community-developed Terraform provider for managing resources in [StoneBranch Universal Controller](https://www.stonebranch.com/).

> **Disclaimer:** This is an independent, community-driven project and is **not affiliated with, endorsed by, or sponsored by Stonebranch GmbH**. "Stonebranch", "Universal Controller", "Universal Automation Center", and related product names are trademarks or registered trademarks of Stonebranch GmbH. All trademarks are the property of their respective owners.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- A StoneBranch Universal Controller instance with API access
- A valid API token with appropriate permissions

## Installation

Download the latest release from [GitHub Releases](https://github.com/OptionMetrics/terraform-provider-stonebranch/releases) and install to your Terraform plugins directory.

### macOS (Apple Silicon)

```bash
VERSION=0.4.0
curl -LO "https://github.com/OptionMetrics/terraform-provider-stonebranch/releases/download/v${VERSION}/terraform-provider-stonebranch_${VERSION}_darwin_arm64.zip"
unzip terraform-provider-stonebranch_${VERSION}_darwin_arm64.zip
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/stonebranch/stonebranch/${VERSION}/darwin_arm64
mv terraform-provider-stonebranch_v${VERSION} ~/.terraform.d/plugins/registry.terraform.io/stonebranch/stonebranch/${VERSION}/darwin_arm64/
rm terraform-provider-stonebranch_${VERSION}_darwin_arm64.zip
```

### macOS (Intel)

```bash
VERSION=0.4.0
curl -LO "https://github.com/OptionMetrics/terraform-provider-stonebranch/releases/download/v${VERSION}/terraform-provider-stonebranch_${VERSION}_darwin_amd64.zip"
unzip terraform-provider-stonebranch_${VERSION}_darwin_amd64.zip
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/stonebranch/stonebranch/${VERSION}/darwin_amd64
mv terraform-provider-stonebranch_v${VERSION} ~/.terraform.d/plugins/registry.terraform.io/stonebranch/stonebranch/${VERSION}/darwin_amd64/
rm terraform-provider-stonebranch_${VERSION}_darwin_amd64.zip
```

### Linux (x86_64)

```bash
VERSION=0.4.0
curl -LO "https://github.com/OptionMetrics/terraform-provider-stonebranch/releases/download/v${VERSION}/terraform-provider-stonebranch_${VERSION}_linux_amd64.zip"
unzip terraform-provider-stonebranch_${VERSION}_linux_amd64.zip
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/stonebranch/stonebranch/${VERSION}/linux_amd64
mv terraform-provider-stonebranch_v${VERSION} ~/.terraform.d/plugins/registry.terraform.io/stonebranch/stonebranch/${VERSION}/linux_amd64/
rm terraform-provider-stonebranch_${VERSION}_linux_amd64.zip
```

### Configure Terraform to Use Local Provider

After installing the provider binary, create or edit `~/.terraformrc` to tell Terraform to use the local filesystem mirror:

```hcl
provider_installation {
  filesystem_mirror {
    path    = "/Users/YOUR_USERNAME/.terraform.d/plugins"
    include = ["stonebranch/stonebranch"]
  }
  direct {
    exclude = ["stonebranch/stonebranch"]
  }
}
```

Replace `YOUR_USERNAME` with your actual username, or use the full path from `echo $HOME`.

**For Linux**, use:
```hcl
provider_installation {
  filesystem_mirror {
    path    = "/home/YOUR_USERNAME/.terraform.d/plugins"
    include = ["stonebranch/stonebranch"]
  }
  direct {
    exclude = ["stonebranch/stonebranch"]
  }
}
```

Then in your Terraform configuration, reference the provider:

```hcl
terraform {
  required_providers {
    stonebranch = {
      source  = "stonebranch/stonebranch"
      version = "0.4.0"
    }
  }
}
```

Run `terraform init` to verify the provider is found.

## Building from Source

Requires [Go](https://golang.org/doc/install) >= 1.24.

```bash
git clone https://github.com/OptionMetrics/terraform-provider-stonebranch.git
cd terraform-provider-stonebranch
make build
```

This creates the `terraform-provider-stonebranch` binary in the project root.

See [RELEASE.md](RELEASE.md) for the full release process.

## sb2tf - Export Existing Resources to Terraform (Experimental)

> **Note:** `sb2tf` is an **experimental** utility shipped separately from the provider. It may have incomplete support for some resource types, and its output may require manual adjustments. Use at your own risk.

The `sb2tf` utility exports existing StoneBranch resources to Terraform configuration files. Use it to bootstrap a new Terraform project from existing resources or migrate manually-created resources to Infrastructure as Code.

### Installation

Download from [GitHub Releases](https://github.com/OptionMetrics/terraform-provider-stonebranch/releases) or build from source:

```bash
make build-sb2tf
# Binary created at ./bin/sb2tf
```

### Authentication

```bash
export STONEBRANCH_API_TOKEN="your-token"
export STONEBRANCH_BASE_URL="https://your-instance.stonebranch.cloud"
```

### Usage Examples

```bash
# List available resource types
sb2tf list

# List all tasks (shows name, type, summary)
sb2tf list tasks

# List tasks matching a pattern (supports * and ? wildcards)
sb2tf list tasks --filter "prod-*"

# Export a single resource
sb2tf export task_unix my_task

# Export a workflow with all its tasks, vertices, and edges
sb2tf export task_workflow my_workflow

# Export all tasks matching a pattern
sb2tf export tasks --all --filter "prod-*"

# Export to a directory (creates main.tf)
sb2tf export tasks --all --output ./terraform/

# Show what would be exported without writing files
sb2tf export tasks --all --dry-run
```

### Workflow Export

When exporting workflows, sb2tf automatically includes:
- The workflow definition
- All tasks contained in the workflow
- Workflow vertices (task instances)
- Workflow edges (task dependencies)

The output is organized logically with proper Terraform references:

```hcl
# Workflow definition
resource "stonebranch_task_workflow" "task_workflow_001" {
  name = "My Workflow"
  ...
}

# Tasks in the workflow
resource "stonebranch_task_unix" "task_unix_001" {
  name = "Task A"
  ...
}

# Vertices
resource "stonebranch_workflow_vertex" "workflow_vertex_001" {
  workflow_name = "My Workflow"
  task_name     = "Task A"
}

# Edges with proper vertex references
resource "stonebranch_workflow_edge" "workflow_edge_001" {
  workflow_name = "My Workflow"
  source_id     = stonebranch_workflow_vertex.workflow_vertex_001.vertex_id
  target_id     = stonebranch_workflow_vertex.workflow_vertex_002.vertex_id
}
```

## Local Development Setup

### 1. Configure Development Overrides

Create or edit `~/.terraformrc` to point Terraform to your local build:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/stonebranch/stonebranch" = "/path/to/terraform-provider-stonebranch"
  }
  direct {}
}
```

Or use the provided dev config:

```bash
export TF_CLI_CONFIG_FILE=/path/to/terraform-provider-stonebranch/examples/dev.tfrc
```

### 2. Set Environment Variables

```bash
export STONEBRANCH_API_TOKEN="your-bearer-token"
export STONEBRANCH_BASE_URL="https://your-instance.stonebranch.cloud"  # optional
```

### 3. Build and Test

```bash
# Build the provider
make build

# Navigate to example directory
cd examples/provider

# IMPORTANT: Skip "terraform init" when using dev overrides!
# Just run plan/apply directly:

# Plan changes
terraform plan

# Apply changes
terraform apply
```

## Provider Configuration

```hcl
terraform {
  required_providers {
    stonebranch = {
      source = "registry.terraform.io/stonebranch/stonebranch"
    }
  }
}

provider "stonebranch" {
  # API token for authentication (required)
  # Can also use STONEBRANCH_API_TOKEN environment variable
  api_token = var.stonebranch_token

  # Base URL for the StoneBranch API (required)
  # Can also use STONEBRANCH_BASE_URL environment variable
  base_url = "https://your-instance.stonebranch.cloud"
}
```

## Authentication

The provider uses Bearer token authentication. Obtain your token from the StoneBranch Universal Controller:

1. Log into your StoneBranch instance
2. Navigate to user settings or API token management
3. Generate or copy your API token

You can provide the token via:
- The `api_token` provider attribute
- The `STONEBRANCH_API_TOKEN` environment variable (recommended for security)

## Resources

### Tasks
- [stonebranch_task_unix](#stonebranch_task_unix) - Unix/Linux command tasks
- [stonebranch_task_windows](#stonebranch_task_windows) - Windows command tasks
- [stonebranch_task_file_transfer](#stonebranch_task_file_transfer) - File transfer tasks
- [stonebranch_task_sql](#stonebranch_task_sql) - SQL database tasks
- [stonebranch_task_email](#stonebranch_task_email) - Email notification tasks
- [stonebranch_task_workflow](#stonebranch_task_workflow) - Workflow orchestration tasks
- [stonebranch_task_recurring](#stonebranch_task_recurring) - Recurring tasks that repeatedly launch a target task on an interval

### Workflows
- [stonebranch_workflow_vertex](#stonebranch_workflow_vertex) - Tasks within workflows
- [stonebranch_workflow_edge](#stonebranch_workflow_edge) - Task dependencies in workflows

### Triggers
- [stonebranch_trigger_time](#stonebranch_trigger_time) - Time-based triggers
- [stonebranch_trigger_cron](#stonebranch_trigger_cron) - Cron expression triggers

### Connections
- [stonebranch_database_connection](#stonebranch_database_connection) - Database connections
- [stonebranch_email_connection](#stonebranch_email_connection) - Email server connections

### Supporting Resources
- [stonebranch_script](#stonebranch_script) - Reusable scripts
- [stonebranch_credential](#stonebranch_credential) - Authentication credentials
- [stonebranch_variable](#stonebranch_variable) - Global variables
- [stonebranch_business_service](#stonebranch_business_service) - Business service groups
- [stonebranch_email_template](#stonebranch_email_template) - Reusable email notification templates
- [stonebranch_custom_day](#stonebranch_custom_day) - Calendar exception dates and holidays
- [stonebranch_virtual_resource](#stonebranch_virtual_resource) - Concurrency control resources
- [stonebranch_universal_template](#stonebranch_universal_template) - Custom script-based task templates

### stonebranch_task_unix

Manages a StoneBranch Unix/Linux task.

#### Example Usage

```hcl
# Simple task with a command
resource "stonebranch_task_unix" "hello" {
  name    = "terraform-hello-world"
  summary = "A simple task managed by Terraform"
  command = "echo 'Hello from Terraform!'"
  agent   = "my-linux-agent"
}

# Task with script content
resource "stonebranch_task_unix" "script" {
  name              = "terraform-script-task"
  summary           = "Task that runs a script"
  command_or_script = "Script"
  script            = <<-EOT
    #!/bin/bash
    echo "Starting..."
    date
    echo "Done"
  EOT
  agent = "my-linux-agent"
}

# Task with retry configuration
resource "stonebranch_task_unix" "with_retry" {
  name           = "terraform-retry-task"
  command        = "/opt/scripts/job.sh"
  agent          = "my-linux-agent"
  retry_maximum  = 3
  retry_interval = 300
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the task |
| `summary` | string | No | Description of the task |
| `agent` | string | No* | Agent to run the task on |
| `agent_cluster` | string | No* | Agent cluster to run the task on |
| `command` | string | No | Command to execute |
| `command_or_script` | string | No | `Command` or `Script` |
| `script` | string | No | Script content (when `command_or_script = "Script"`) |
| `runtime_dir` | string | No | Working directory |
| `parameters` | string | No | Parameters to pass |
| `credentials` | string | No | Credentials to use |
| `exit_codes` | string | No | Success exit codes (e.g., `"0"` or `"0,1,2"`) |
| `exit_code_processing` | string | No | `Success Exitcode Range` or `Failure Exitcode Range` |
| `retry_maximum` | int | No | Max retry attempts |
| `retry_interval` | int | No | Seconds between retries |
| `retry_indefinitely` | bool | No | Retry forever |
| `run_as_sudo` | bool | No | Run with sudo |
| `opswise_groups` | list | No | Business service names |

*One of `agent` or `agent_cluster` is required by the API.

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |
| `version` | Version number for optimistic locking |

#### Import

Tasks can be imported using the task name:

```bash
terraform import stonebranch_task_unix.example "task-name"
```

### stonebranch_task_windows

Manages a StoneBranch Windows task.

#### Example Usage

```hcl
# Simple Windows task with a command
resource "stonebranch_task_windows" "hello" {
  name    = "terraform-windows-hello"
  summary = "A simple Windows task managed by Terraform"
  command = "echo Hello from Terraform!"
  agent   = "my-windows-agent"
}

# Windows task with elevated privileges
resource "stonebranch_task_windows" "admin_task" {
  name         = "terraform-admin-task"
  command      = "net user"
  agent        = "my-windows-agent"
  elevate_user = true
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the task |
| `summary` | string | No | Description of the task |
| `agent` | string | No* | Agent to run the task on |
| `agent_cluster` | string | No* | Agent cluster to run the task on |
| `command` | string | No | Command to execute |
| `command_or_script` | string | No | `Command` or `Script` |
| `script` | string | No | Script resource name (when `command_or_script = "Script"`) |
| `runtime_dir` | string | No | Working directory |
| `parameters` | string | No | Parameters to pass |
| `credentials` | string | No | Credentials to use |
| `exit_codes` | string | No | Success exit codes (e.g., `"0"` or `"0,1,2"`) |
| `exit_code_processing` | string | No | `Success Exitcode Range` or `Failure Exitcode Range` |
| `retry_maximum` | int | No | Max retry attempts |
| `retry_interval` | int | No | Seconds between retries |
| `retry_indefinitely` | bool | No | Retry forever |
| `elevate_user` | bool | No | Run with administrator privileges |
| `desktop_interact` | bool | No | Allow desktop interaction |
| `create_console` | bool | No | Create console window |
| `opswise_groups` | list | No | Business service names |

*One of `agent` or `agent_cluster` is required by the API.

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |
| `version` | Version number for optimistic locking |

#### Import

Tasks can be imported using the task name:

```bash
terraform import stonebranch_task_windows.example "task-name"
```

### stonebranch_task_recurring

Manages a StoneBranch Task (Recurring task type). A recurring task repeatedly launches a target task/workflow on an interval, within an optional time window.

#### Example Usage

```hcl
# Recurring task that launches a target task every 15 minutes,
# restricted to a daily time window
resource "stonebranch_task_recurring" "poll_every_15_minutes" {
  name        = "poll-every-15-minutes"
  summary     = "Launches the poll task every 15 minutes between 01:00 and 23:40"
  target_task = stonebranch_task_unix.poll.name

  recurrence_type          = "Interval"
  recurrence_interval      = "15"
  recurrence_interval_unit = "Minutes"

  time_window            = true
  interval_start_time    = "01:00"
  interval_end_time      = "23:40"
  indefinite_recurrences = false
}

# Recurring task that stops after a fixed number of recurrences
resource "stonebranch_task_recurring" "poll_five_times" {
  name        = "poll-five-times"
  target_task = stonebranch_task_unix.poll.name

  recurrence_type          = "Interval"
  recurrence_interval      = "10"
  recurrence_interval_unit = "Minutes"

  indefinite_recurrences = false
  number_of_recurrences  = "5"
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the task |
| `target_task` | string | Yes | Name of the task or workflow to launch on each recurrence |
| `summary` | string | No | Description of the task |
| `target_task_monitor_condition` | string | No | `None`, `First Recurrence`, `Last Recurrence`, or `All Recurrences` |
| `target_task_status_text` | string | No | Status text to display for the target task |
| `recurrence_type` | string | No | `Interval` or `On` |
| `recurrence_interval` | string | No | Interval amount between recurrences |
| `recurrence_interval_unit` | string | No | `Seconds`, `Minutes`, `Hours`, or `Days` |
| `time_window` | bool | No | Whether recurrences are restricted to `interval_start_time`/`interval_end_time` |
| `interval_start_time` | string | No | Time of day (HH:MM) the interval window starts |
| `interval_end_time` | string | No | Time of day (HH:MM) the interval window ends |
| `interval_start_day_constraint` | string | No | `None` or `Same Day` |
| `interval_end_day_constraint` | string | No | `None` or `Same Day` |
| `indefinite_recurrences` | bool | No | Whether the task recurs indefinitely |
| `number_of_recurrences` | string | No* | Number of times to recur before stopping |
| `skip_condition` | string | No | `None`, `Active`, or `Active By Recurring Task Instance` |
| `retention_duration_rt` | int | No | How long to retain completed recurrence instances |
| `retention_duration_unit_rt` | string | No | Unit for `retention_duration_rt` (e.g. `Days`) |
| `retention_duration_purge_rt` | bool | No | Whether to purge retained recurrence instances after `retention_duration_rt` elapses |
| `rd_exclude_backup_rt` | bool | No | Whether to exclude retained recurrence instances from backup |
| `opswise_groups` | list | No | Business service names |

*`time_window` and `indefinite_recurrences` cannot both be `true`. `number_of_recurrences` must not be set when `time_window` is `true`, and is required when both `time_window` and `indefinite_recurrences` are `false`.

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |
| `version` | Version number for optimistic locking |

#### Import

Tasks can be imported using the task name:

```bash
terraform import stonebranch_task_recurring.example "task-name"
```

### stonebranch_script

Manages a reusable script resource that can be referenced by tasks.

#### Example Usage

```hcl
resource "stonebranch_script" "backup" {
  name    = "backup-script"
  content = <<-EOT
    #!/bin/bash
    tar -czf /backup/data.tar.gz /data
  EOT
}

# Reference the script in a task
resource "stonebranch_task_unix" "backup_job" {
  name              = "backup-job"
  command_or_script = "Script"
  script            = stonebranch_script.backup.name
  agent             = "my-linux-agent"
}
```

### stonebranch_trigger_time

Manages a time-based trigger for scheduling task execution.

#### Example Usage

```hcl
resource "stonebranch_trigger_time" "daily" {
  name      = "daily-trigger"
  tasks     = [stonebranch_task_unix.my_task.name]
  time      = "08:00"
  time_zone = "America/New_York"
}
```

### stonebranch_credential

Manages authentication credentials for task execution.

#### Example Usage

```hcl
resource "stonebranch_credential" "service_account" {
  name             = "service-account-creds"
  runtime_user     = "svc_user"
  runtime_password = var.service_password
}
```

### stonebranch_variable

Manages global variables that can be referenced by tasks and triggers.

#### Example Usage

```hcl
resource "stonebranch_variable" "environment" {
  name        = "APP_ENVIRONMENT"
  value       = "production"
  description = "Current application environment"
}
```

### stonebranch_business_service

Manages business service groups for organizing resources.

#### Example Usage

```hcl
resource "stonebranch_business_service" "production" {
  name        = "Production Services"
  description = "Business service for production workloads"
}

# Reference in other resources via opswise_groups
resource "stonebranch_variable" "app_env" {
  name           = "APP_ENVIRONMENT"
  value          = "production"
  opswise_groups = [stonebranch_business_service.production.name]
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the business service |
| `description` | string | No | Description of the business service |

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |
| `version` | Version number for optimistic locking |

#### Import

Business services can be imported using the name:

```bash
terraform import stonebranch_business_service.example "service-name"
```

### stonebranch_universal_template

Manages a StoneBranch Universal Template. Universal templates define reusable, script-based custom task types that run via a Universal Agent plugin.

> **Known limitation (v1):** this resource models a curated core field set plus the custom UI form-builder (`fields`). The `commands`/`events` (event-metric) definitions and the top-level per-attribute UI restriction (`*FieldsRestriction`) settings are not yet supported and must be managed outside Terraform (e.g. via the UAC UI).
>
> **Known API limitation:** once a template has `fields`, the server rejects updates that would reduce the list to empty (`fields = []` or omitting the attribute) — you can only ever replace fields with a different non-empty set.

#### Example Usage

```hcl
# Cross-platform ("Any" agent type) template using a single common script.
resource "stonebranch_universal_template" "health_check" {
  name              = "tf-example-health-check"
  description       = "Runs a health check script on any agent platform"
  variable_prefix   = "HC"
  agent_type        = "Any"
  use_common_script = true
  script            = "curl -sf $HC_URL || exit 1"
  exit_codes        = "0"

  environment = [
    { name = "HC_URL", value = "https://example.com/health" },
  ]
}

# Windows-specific template with platform script and elevated privileges.
resource "stonebranch_universal_template" "windows_cleanup" {
  name            = "tf-example-windows-cleanup"
  description     = "Cleans up temp files on a Windows agent"
  variable_prefix = "WC"
  agent_type      = "Windows"
  script_windows  = "Remove-Item -Path $env:TEMP\\* -Recurse -Force"
  exit_codes      = "0"
  elevate_user    = true
}

# Template exposing a custom UI form ("fields") so operators can fill in
# task-specific values when launching a task built from this template.
resource "stonebranch_universal_template" "aws_deploy" {
  name              = "tf-example-aws-deploy"
  description       = "Deploys to AWS with operator-supplied region and tags"
  variable_prefix   = "AWS"
  agent_type        = "Any"
  use_common_script = true
  script            = "echo Deploying to $AWS_region with tags $AWS_tags"
  exit_codes        = "0"

  fields = [
    {
      name          = "region"
      label         = "Region"
      field_mapping = "Choice Field 1"
      field_type    = "Choice"
      choices = [
        { field_value = "us-east-1", field_value_label = "US East 1" },
        { field_value = "us-west-2", field_value_label = "US West 2" },
      ]
    },
    {
      name              = "tags"
      label             = "Tags"
      field_mapping     = "Array Field 1"
      field_type        = "Array"
      array_name_title  = "Key"
      array_value_title = "Value"
      array_field_value = [
        { name = "env", value = "prod" },
      ]
    },
  ]
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the universal template |
| `variable_prefix` | string | Yes | Prefix used for variables exposed by this template |
| `agent_type` | string | Yes | Agent platform this template targets. Confirmed values: `Windows`, `Any` |
| `exit_codes` | string | Yes | Exit codes that indicate success (e.g. `0` or `0,1,2`); no server default |
| `description` | string | No | Description of the universal template |
| `use_common_script` / `script` | bool / string | No | Single script shared across all agent types |
| `script_unix` / `script_windows` | string | No | Platform-specific script content |
| `script_type_windows` | string | No | Script type/interpreter for the Windows script |
| `agent` / `agent_var` / `agent_cluster` / `agent_cluster_var` | string | No | Agent targeting for tasks based on this template |
| `broadcast_cluster` / `broadcast_cluster_var` | string | No | Broadcast cluster targeting |
| `credentials` / `credentials_var` | string | No | Credentials used for task execution |
| `environment` | list of objects | No | `name`/`value` environment variable pairs. Full-replace on update — omitting this attribute clears any previously-set values |
| `exit_code_processing` | string | No | Confirmed values: `Success Exitcode Range`, `Failure Exitcode Range`. Default: `Success Exitcode Range` |
| `elevate_user` / `desktop_interact` / `create_console` | bool | No | Windows-only execution options |
| `template_type` | string | No | Confirmed value: `Script` |
| `fields` | list of objects | No | Custom UI form fields; each maps to a typed extension slot via `field_mapping`. Full-replace on update, but the server rejects reducing the list to empty (see known API limitation above) |

See `docs/resources/universal_template.md` for the full attribute reference.

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |

#### Import

Universal templates can be imported using the name:

```bash
terraform import stonebranch_universal_template.example "template-name"
```

### stonebranch_email_template

Manages reusable email notification templates. Templates define subject/body content and recipients, and reference a `stonebranch_email_connection` used to send the email.

#### Example Usage

```hcl
resource "stonebranch_email_connection" "notifications" {
  name          = "notifications"
  smtp          = "smtp.example.com"
  smtp_port     = 25
  email_address = "notifications@example.com"
}

# At least one of "to", "cc", or "bcc" must be set.
resource "stonebranch_email_template" "job_failure" {
  name             = "job-failure"
  email_connection = stonebranch_email_connection.notifications.name
  to               = "oncall@example.com"
  subject          = "Job Failed"
  body             = "A scheduled job has failed. Please check the Universal Controller for details."
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the email template |
| `email_connection` | string | Yes | Name of the `stonebranch_email_connection` used to send emails from this template |
| `to` / `cc` / `bcc` | string | No* | Comma-separated recipient lists |
| `subject` | string | No | Subject line of the email |
| `body` | string | No | Body content of the email |
| `reply_to` | string | No | Reply-To address |
| `description` | string | No | Description of the email template |
| `opswise_groups` | list | No | Business service names |

*At least one of `to`, `cc`, or `bcc` must be set; validated at plan time.

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |
| `version` | Version number for optimistic locking |

#### Import

Email templates can be imported using the name:

```bash
terraform import stonebranch_email_template.example "template-name"
```

### stonebranch_custom_day

Manages calendar exception dates (single dates, date lists, or yearly repeating dates), optionally marked as holidays with weekend-observance rules.

#### Example Usage

```hcl
# A single, specific exception date
resource "stonebranch_custom_day" "maintenance_window" {
  name  = "maintenance-window"
  ctype = "Single Date"
  date  = "2026-04-15"
}

# The same month+day every year, marked as a holiday with
# weekend-observance rules
resource "stonebranch_custom_day" "christmas" {
  name    = "christmas"
  ctype   = "Absolute Repeating Date"
  month   = "Dec"
  day     = 25
  holiday = true

  observed_rules = [
    { actual_day_of_week = "Sat", observed_day_of_week = "Fri" },
    { actual_day_of_week = "Sun", observed_day_of_week = "Mon" },
  ]
}

# The nth weekday of a month, every year (e.g. 4th Thursday of November)
resource "stonebranch_custom_day" "thanksgiving" {
  name      = "thanksgiving"
  ctype     = "Relative Repeating Date"
  month     = "Nov"
  dayofweek = "Thu"
  relfreq   = "4th"
  holiday   = true
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the custom day |
| `ctype` | string | Yes | Definition style: `Single Date`, `List of Dates`, `Absolute Repeating Date`, or `Relative Repeating Date` |
| `date` | string | No | Specific date (`yyyy-MM-dd`), used when `ctype` is `Single Date` |
| `date_list` | list(string) | No | List of specific dates (`yyyy-MM-dd`), used when `ctype` is `List of Dates` |
| `month` | string | No | Month (`Jan`-`Dec`), used when `ctype` is `Absolute Repeating Date` or `Relative Repeating Date` |
| `day` | number | No | Day of month, used when `ctype` is `Absolute Repeating Date` |
| `dayofweek` | string | No | Day of week (`Sun`-`Sat`), used when `ctype` is `Relative Repeating Date` |
| `relfreq` | string | No | Relative frequency (`1st`, `2nd`, `3rd`, `4th`, `Last`, `Every`, `Nth`, `Last Day`, `Last Business Day`), used when `ctype` is `Relative Repeating Date` |
| `nth_amount` / `nth_type` | number / string | No | Nth day-of-month value/type, used when `relfreq` is `Nth` |
| `adjustment` / `adjustment_amount` / `adjustment_type` | string / number / string | No | Offset applied to the resolved date (`None`, `Less`, `Plus`) |
| `holiday` | bool | No | Marks this custom day as a holiday, enabling `observed_rules` |
| `period` | bool | No | Marks this custom day as a period (not allowed when `ctype` is `Single Date`) |
| `observed_rules` | list(object) | No | Weekend-observance rules (`actual_day_of_week` / `observed_day_of_week`), used when `holiday` is `true` |
| `comments` | string | No | Description of the custom day |

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |
| `version` | Version number for optimistic locking |
| `category` | Server-computed classification: `Day`, `Holiday`, or `Period`, derived from `holiday`/`period` |

#### Import

Custom days can be imported using the name:

```bash
terraform import stonebranch_custom_day.example "custom-day-name"
```

### stonebranch_virtual_resource

Manages a StoneBranch Virtual Resource for concurrency control. Note: the underlying UAC API endpoint for this resource is `/resources/virtual`, not `/resources/virtualresource`.

#### Example Usage

```hcl
resource "stonebranch_virtual_resource" "db_connections" {
  name    = "db-connections"
  type    = "Renewable"
  limit   = 5
  summary = "Limits concurrent tasks connecting to the shared database"
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Unique name of the virtual resource |
| `type` | string | No | Type of virtual resource: `Renewable`, `Boundary`, or `Depletable` |
| `limit` | number | No | Maximum concurrent usage allowed |
| `summary` | string | No | Description of the virtual resource |
| `opswise_groups` | list(string) | No | Business services this virtual resource belongs to |

#### Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `sys_id` | System ID assigned by StoneBranch |
| `version` | Version number for optimistic locking |

#### Import

Virtual resources can be imported using the name:

```bash
terraform import stonebranch_virtual_resource.example "resource-name"
```

## Development

### Project Structure

```
terraform-provider-stonebranch/
├── main.go                          # Provider entry point
├── cmd/
│   └── sb2tf/                       # sb2tf CLI utility
│       ├── main.go                  # CLI entry point
│       ├── cli/                     # Command implementations
│       │   ├── root.go              # Root command, global flags
│       │   ├── list.go              # List resources command
│       │   └── export.go            # Export resources command
│       └── generator/               # HCL generation
│           ├── generator.go         # Core generation logic
│           ├── resources.go         # Resource type registry
│           └── templates.go         # HCL templates
├── internal/
│   ├── provider/
│   │   ├── provider.go              # Provider configuration and schema
│   │   ├── resources/               # Resource implementations
│   │   └── data_sources/            # Data source implementations
│   ├── acctest/
│   │   └── acctest.go               # Acceptance test helpers
│   └── client/
│       └── client.go                # StoneBranch API HTTP client
├── examples/
│   ├── dev.tfrc                     # Development override config
│   └── provider/
│       └── main.tf                  # Example Terraform configuration
├── docs/                            # Generated documentation
├── Makefile                         # Build automation
└── openapi.yaml                     # StoneBranch API specification
```

### Useful Commands

```bash
# Provider
make build            # Build the provider binary
make test             # Run tests
make testacc          # Run acceptance tests (requires API credentials)
make fmt              # Format Go code
make clean            # Remove built binaries
make docs             # Generate provider documentation

# sb2tf utility
make build-sb2tf      # Build the sb2tf binary
make install-sb2tf    # Install sb2tf to $GOPATH/bin

# Releases
make release-snapshot # Build release artifacts (no tag required)
make publish          # Build and publish to GitHub Releases
```

### Generating Documentation

Documentation is auto-generated from provider schemas using [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs).

```bash
# Generate/update documentation
make docs
```

The generated docs are written to `docs/` and include:
- Provider overview (`docs/index.md`)
- Resource documentation (`docs/resources/*.md`)
- Data source documentation (`docs/data-sources/*.md`)

Examples are pulled from `examples/resources/*/resource.tf` and `examples/data-sources/*/data-source.tf`.

**Always regenerate docs after schema changes:**
```bash
# After modifying resource schemas
make docs
git add docs/
git commit -m "Update generated documentation"
```

### Releasing

See [RELEASE.md](RELEASE.md) for the complete release and publishing process.

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires valid credentials)
export STONEBRANCH_API_TOKEN="your-token"
export STONEBRANCH_BASE_URL="https://your-instance.stonebranch.cloud"
TF_ACC=1 go test -v ./...
```

## Troubleshooting

### "Provider not found" error

Ensure your `~/.terraformrc` or `TF_CLI_CONFIG_FILE` points to the correct binary location.

### Authentication errors

1. Verify your token is valid and not expired
2. Check that the token has appropriate permissions
3. Ensure the base URL is correct (no trailing slash)

### API errors

Enable debug logging:

```bash
export TF_LOG=DEBUG
terraform plan
```

## License

This project is licensed under the [MIT License](LICENSE).

## Trademarks

"Stonebranch", "Universal Controller", "Universal Automation Center", and related product names are trademarks or registered trademarks of Stonebranch GmbH. "Terraform" is a trademark of HashiCorp, Inc. This project is not endorsed by or affiliated with either company.
