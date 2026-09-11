---
page_title: "Provider: Stonebranch"
description: |-
  The Stonebranch provider enables Terraform to manage resources in StoneBranch Universal Controller.
---

# Stonebranch Provider

The Stonebranch provider allows you to manage [StoneBranch Universal Controller](https://www.stonebranch.com/products/universal-controller/) resources using Terraform. Universal Controller is an enterprise workload automation platform that enables scheduling, orchestration, and monitoring of jobs across your infrastructure.

## Features

This provider supports managing:

**Tasks:**
- Unix/Linux command tasks
- Windows command tasks
- SQL database tasks
- Email notification tasks
- File transfer tasks
- File monitor tasks
- Workflow orchestration tasks

**Triggers:**
- Time-based triggers
- Cron expression triggers
- File monitor triggers

**Supporting Resources:**
- Scripts (reusable command scripts)
- Credentials (authentication)
- Variables (global/scoped)
- Database connections
- Email connections
- Business services (resource grouping)
- Calendars (business day definitions)

**Workflows:**
- Workflow tasks (DAG definitions)
- Workflow vertices (task nodes)
- Workflow edges (task dependencies)

**Data Sources:**
- Agents (query registered agents)
- Agent clusters (query agent groups)
- Tasks (search existing tasks)
- Task instances (query job execution history)

## Authentication

The provider authenticates to the Universal Controller API using a Bearer token.

### Obtaining an API Token

1. Log in to your Universal Controller web interface
2. Navigate to **Administration > Security > API Tokens**
3. Create a new token with appropriate permissions
4. Copy the token value

### Configuration

You can provide the API token in two ways:

**Environment Variable (Recommended):**

```shell
export STONEBRANCH_API_TOKEN="your-api-token"
export STONEBRANCH_BASE_URL="https://your-controller.stonebranch.cloud"  # Optional
```

**Provider Configuration:**

```terraform
provider "stonebranch" {
  api_token = var.stonebranch_token  # Use a variable, never hardcode
  base_url  = "https://your-controller.stonebranch.cloud"
}
```

-> **Note:** We recommend using environment variables for sensitive values like API tokens.

## Usage Tips

### Referencing Credentials

Many task resources require credentials for authentication. Credentials are referenced by **name** as a string, not as a Terraform resource reference. You can either reference an existing credential or create one with the `stonebranch_credential` resource:

```terraform
# Create a credential
resource "stonebranch_credential" "sftp_cred" {
  name             = "my-sftp-credential"
  runtime_user     = "sftpuser"
  runtime_password = var.sftp_password  # sensitive
}

# Reference it by name in a task
resource "stonebranch_task_file_transfer" "download" {
  name               = "download-files"
  remote_credentials = stonebranch_credential.sftp_cred.name
  # ...
}
```

Alternatively, use a `credentials_var` attribute to reference a StoneBranch variable containing the credential name at runtime.

### Using StoneBranch Variables in Strings

StoneBranch uses `${variable_name}` syntax for runtime variable substitution. Since Terraform uses the same syntax for interpolation, you must **escape the dollar sign** by doubling it:

```terraform
resource "stonebranch_task_file_transfer" "example" {
  name = "edi-download"
  # Use $$ to produce a literal $ in the output
  remote_filename = "/data/$${wf_date}_$${sequence}.txt"
  # Results in: /data/${wf_date}_${sequence}.txt
}
```

Without escaping, Terraform will try to resolve `${wf_date}` as a Terraform variable reference and fail.

## Example Usage

```terraform
# When using dev overrides, skip "terraform init" and run "terraform plan" directly.
#
# Usage:
#   cd examples/provider
#   export STONEBRANCH_API_TOKEN="your-token"
#   terraform plan

terraform {
  required_providers {
    stonebranch = {
      source = "stonebranch/stonebranch"
    }
  }
}

provider "stonebranch" {
  # api_token = var.stonebranch_token  # or set STONEBRANCH_API_TOKEN env var
  # base_url  = "https://your-instance.stonebranch.cloud"  # or set STONEBRANCH_BASE_URL env var
}

# Example: Simple Unix task that runs a command
resource "stonebranch_task_unix" "hello_world" {
  name    = "terraform-hello-world"
  summary = "A simple task managed by Terraform"

  # Agent configuration - ONE of these is REQUIRED:
  agent = var.agent_name  # or use agent_cluster for cluster

  # Command to execute
  command = "echo 'Hello from Terraform!'"

  # Exit code handling (optional)
  exit_codes           = "0"
  exit_code_processing = "Success Exitcode Range"
}

variable "agent_name" {
  description = "Name of the StoneBranch agent to run tasks on"
  type        = string
  default     = "your-agent-name"
}

# Example: Script resource
resource "stonebranch_script" "backup_script" {
  name        = "terraform-backup-script"
  description = "A backup script managed by Terraform"
  content     = <<-EOT
    #!/bin/bash
    echo "Starting backup..."
    date
    echo "Backup completed"
  EOT
}

# Example: Unix task that references a script resource
resource "stonebranch_task_unix" "script_task" {
  name              = "terraform-script-task"
  summary           = "Task that runs a script resource"
  command_or_script = "Script"
  script            = stonebranch_script.backup_script.name
  agent             = var.agent_name
  exit_codes        = "0"
}

# Example: Task with retry configuration
# resource "stonebranch_task_unix" "with_retry" {
#   name          = "terraform-retry-example"
#   summary       = "Task with retry configuration"
#   command       = "/opt/scripts/important-job.sh"
#   agent         = "your-agent-name"
#
#   retry_maximum  = 3
#   retry_interval = 300  # 5 minutes
# }

# Example: Time trigger that runs a task daily at 9:00 AM
resource "stonebranch_trigger_time" "daily_backup" {
  name        = "terraform-daily-backup-trigger"
  description = "Triggers the backup task every day at 9:00 AM"
  enabled     = false  # Set to true to activate

  # Reference the task(s) to run
  tasks = [stonebranch_task_unix.script_task.name]

  # Schedule configuration
  time      = "09:00"
  time_zone = "America/New_York"
}

# Example: Credential resource for SFTP authentication
resource "stonebranch_credential" "sftp_creds" {
  name             = "terraform-sftp-credentials"
  description      = "SFTP credentials managed by Terraform"
  runtime_user     = "sftpuser"
  runtime_password = var.sftp_password  # Use a variable for sensitive data
}

variable "sftp_password" {
  description = "Password for SFTP connection"
  type        = string
  sensitive   = true
  default     = "placeholder"  # Set via environment or tfvars
}

# Example: File Transfer task for SFTP download
resource "stonebranch_task_file_transfer" "download_report" {
  name    = "terraform-download-report"
  summary = "Download daily report from SFTP server"

  agent = var.agent_name

  # Transfer settings
  # transfer_direction = "GET"  # Only applies to UDM agents (GET=download, PUT=upload)
  server_type = "SFTP"

  # Remote server configuration
  remote_server   = "sftp.example.com"
  remote_filename = "/reports/daily_report.csv"

  # Local destination
  local_filename = "/data/reports/daily_report.csv"

  # Reference the credential resource
  remote_credentials = stonebranch_credential.sftp_creds.name
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_token` (String, Sensitive) Bearer token for StoneBranch API authentication. Can also be set via STONEBRANCH_API_TOKEN environment variable. Mutually exclusive with username/password.
- `base_url` (String) Base URL for the StoneBranch API. Can also be set via STONEBRANCH_BASE_URL environment variable.
- `password` (String, Sensitive) Password for StoneBranch API HTTP Basic Auth. Can also be set via STONEBRANCH_PASSWORD environment variable. Requires username. Mutually exclusive with api_token.
- `username` (String) Username for StoneBranch API HTTP Basic Auth, for UAC instances fronted by a proxy that requires Basic Auth instead of a Bearer token. Can also be set via STONEBRANCH_USERNAME environment variable. Requires password. Mutually exclusive with api_token.
