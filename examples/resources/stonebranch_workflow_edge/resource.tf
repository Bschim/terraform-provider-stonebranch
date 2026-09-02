# Stonebranch Workflow Edge Example
#
# This example demonstrates how to create dependencies (edges) between
# tasks in a workflow. Edges define the execution order of tasks.
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

# Create a workflow
resource "stonebranch_task_workflow" "example" {
  name    = "tf-example-workflow-with-edges"
  summary = "Workflow demonstrating edge management"
}

# Create tasks for the workflow
resource "stonebranch_task_unix" "start" {
  name       = "tf-example-edge-start"
  summary    = "Starting task"
  agent_var  = "agent_name"
  command    = "echo 'Starting...'"
  exit_codes = "0"
}

resource "stonebranch_task_unix" "process_a" {
  name       = "tf-example-edge-process-a"
  summary    = "Process A - runs in parallel"
  agent_var  = "agent_name"
  command    = "echo 'Process A'"
  exit_codes = "0"
}

resource "stonebranch_task_unix" "process_b" {
  name       = "tf-example-edge-process-b"
  summary    = "Process B - runs in parallel"
  agent_var  = "agent_name"
  command    = "echo 'Process B'"
  exit_codes = "0"
}

resource "stonebranch_task_unix" "finish" {
  name       = "tf-example-edge-finish"
  summary    = "Finishing task"
  agent_var  = "agent_name"
  command    = "echo 'Finished!'"
  exit_codes = "0"
}

# Add tasks to workflow as vertices
resource "stonebranch_workflow_vertex" "start" {
  workflow_name = stonebranch_task_workflow.example.name
  task_name     = stonebranch_task_unix.start.name
  vertex_x      = "200"
  vertex_y      = "50"
}

resource "stonebranch_workflow_vertex" "process_a" {
  workflow_name = stonebranch_task_workflow.example.name
  task_name     = stonebranch_task_unix.process_a.name
  vertex_x      = "100"
  vertex_y      = "150"
}

resource "stonebranch_workflow_vertex" "process_b" {
  workflow_name = stonebranch_task_workflow.example.name
  task_name     = stonebranch_task_unix.process_b.name
  vertex_x      = "300"
  vertex_y      = "150"
}

resource "stonebranch_workflow_vertex" "finish" {
  workflow_name = stonebranch_task_workflow.example.name
  task_name     = stonebranch_task_unix.finish.name
  vertex_x      = "200"
  vertex_y      = "250"
}

# Create edges to define the workflow execution order
#
# This creates a diamond pattern:
#
#        start
#       /     \
#  process_a  process_b
#       \     /
#        finish
#

# start -> process_a
resource "stonebranch_workflow_edge" "start_to_a" {
  workflow_name = stonebranch_task_workflow.example.name
  source_id     = stonebranch_workflow_vertex.start.vertex_id
  target_id     = stonebranch_workflow_vertex.process_a.vertex_id
}

# start -> process_b
resource "stonebranch_workflow_edge" "start_to_b" {
  workflow_name = stonebranch_task_workflow.example.name
  source_id     = stonebranch_workflow_vertex.start.vertex_id
  target_id     = stonebranch_workflow_vertex.process_b.vertex_id
}

# process_a -> finish
#
# `condition` controls when the edge is followed. When omitted, UAC
# defaults to a Status condition of "Success". Here we require process_a
# to have exited with code 0 specifically.
resource "stonebranch_workflow_edge" "a_to_finish" {
  workflow_name = stonebranch_task_workflow.example.name
  source_id     = stonebranch_workflow_vertex.process_a.vertex_id
  target_id     = stonebranch_workflow_vertex.finish.vertex_id

  condition = {
    type      = "Exit Code"
    exit_code = "0"
  }
}

# process_b -> finish
#
# A Variable condition compares two runtime variable values. `first_value`
# and `second_value` typically reference `$${variable}` expressions.
resource "stonebranch_workflow_edge" "b_to_finish" {
  workflow_name = stonebranch_task_workflow.example.name
  source_id     = stonebranch_workflow_vertex.process_b.vertex_id
  target_id     = stonebranch_workflow_vertex.finish.vertex_id

  condition = {
    type         = "Variable"
    first_value  = "$${OUTPUT_LINE_COUNT}"
    operator     = "!="
    second_value = "$${EXPECTED_LINE_COUNT}"
  }
}

# Output workflow details
output "workflow_name" {
  value = stonebranch_task_workflow.example.name
}
