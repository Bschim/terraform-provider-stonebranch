# Stonebranch Workflow Task Example
#
# This example demonstrates how to create workflow task resources in Stonebranch.
# Workflow tasks orchestrate the execution of multiple tasks in sequence.
#
# Note: Tasks within the workflow (vertices) and their dependencies (edges) are
# typically managed through the StoneBranch UI after the workflow is created.
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

# Simple workflow
resource "stonebranch_task_workflow" "simple" {
  name    = "tf-example-simple-workflow"
  summary = "A simple workflow for demonstration"
}

# Workflow with critical path calculation
resource "stonebranch_task_workflow" "with_critical_path" {
  name                    = "tf-example-critical-path-workflow"
  summary                 = "Workflow that calculates critical path"
  calculate_critical_path = true
}

# Workflow with skip handling
resource "stonebranch_task_workflow" "with_skip_handling" {
  name           = "tf-example-skip-workflow"
  summary        = "Workflow with custom skip behavior"
  skipped_option = "Skip Successors On Skip"
}

# Workflow with instance wait
resource "stonebranch_task_workflow" "with_instance_wait" {
  name                 = "tf-example-instance-wait-workflow"
  summary              = "Workflow that waits for other instances"
  instance_wait        = "Wait For All"
  instance_wait_lookup = "Oldest Instance"
}

# Workflow with retry configuration
resource "stonebranch_task_workflow" "with_retry" {
  name                    = "tf-example-retry-workflow"
  summary                 = "Workflow with retry settings"
  retry_maximum           = 3
  retry_interval          = 300
  retry_suppress_failure  = false
}

# Production workflow example
resource "stonebranch_task_workflow" "production" {
  name                    = "tf-example-production-workflow"
  summary                 = "Production data processing pipeline"
  calculate_critical_path = true
  skipped_option          = "Run Successors On Skip"
  layout_option           = "Vertical"

  retry_maximum  = 2
  retry_interval = 600
}

# Workflow with run_criteria referencing a task within it.
#
# Note: the Stonebranch API requires the referenced task to already exist as
# a vertex in the workflow before run_criteria can reference it, so the
# stonebranch_workflow_vertex resource must be applied before (or alongside,
# on a subsequent apply) the run_criteria that references it.
resource "stonebranch_task_unix" "step" {
  name       = "tf-example-workflow-step"
  agent_var  = var.agent_var
  command    = "echo step"
  exit_codes = "0"
}

resource "stonebranch_task_workflow" "with_run_criteria" {
  name    = "tf-example-run-criteria-workflow"
  summary = "Workflow with a Skip Criteria on one of its tasks"

  run_criteria = [
    {
      type         = "Skip Criteria"
      task         = stonebranch_task_unix.step.name
      business_day = true
    }
  ]
}

resource "stonebranch_workflow_vertex" "step" {
  workflow_name = stonebranch_task_workflow.with_run_criteria.name
  task_name     = stonebranch_task_unix.step.name
}
