# Stonebranch Virtual Resource Example
#
# This example demonstrates how to create virtual resources in Stonebranch
# for concurrency control.
#
# Usage:
#   export STONEBRANCH_API_TOKEN="your-token"
#   terraform init
#   terraform plan
#   terraform apply

terraform {
  required_providers {
    stonebranch = {
      source = "stonebranch/stonebranch"
    }
  }
}

provider "stonebranch" {
  # Uses STONEBRANCH_API_TOKEN environment variable
}

# Renewable virtual resource: usage is released when the task finishes,
# so it behaves as a simple concurrency limiter.
resource "stonebranch_virtual_resource" "db_connections" {
  name    = "tf-example-vr-db-connections"
  type    = "Renewable"
  limit   = 5
  summary = "Limits concurrent tasks connecting to the shared database"
}

# Depletable virtual resource: usage is consumed and not automatically
# released, useful for tracking a finite pool (e.g. license seats).
resource "stonebranch_virtual_resource" "license_seats" {
  name    = "tf-example-vr-license-seats"
  type    = "Depletable"
  limit   = 100
  summary = "Tracks available software license seats"
}

# Reference the virtual resource from a Unix task's variables to gate
# concurrency (actual task->virtual-resource binding is configured on the
# task's runtime/resource criteria within UAC).
resource "stonebranch_task_unix" "backup" {
  name    = "tf-example-task-backup"
  agent   = "example-agent"
  command = "run-backup.sh"

  variables = [
    {
      name  = "VIRTUAL_RESOURCE"
      value = stonebranch_virtual_resource.db_connections.name
    }
  ]
}
