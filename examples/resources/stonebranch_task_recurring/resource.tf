# Stonebranch Recurring Task Example
#
# This example demonstrates how to create a Recurring task in Stonebranch.
# A recurring task repeatedly launches a target task/workflow on an
# interval, optionally restricted to a daily time window.
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
  # Optionally set base_url if not using the default
}

# Target task that the recurring task will launch repeatedly
resource "stonebranch_task_unix" "poll" {
  name    = "tf-example-recurring-target"
  summary = "Task launched repeatedly by the recurring task"

  agent_var = var.agent_var

  command    = "echo 'Polling...'"
  exit_codes = "0"
}

# Recurring task that launches the target task every 15 minutes,
# restricted to a daily time window (recurrence count is bounded by the
# window itself, so number_of_recurrences must not be set and
# indefinite_recurrences must be false when time_window is true)
resource "stonebranch_task_recurring" "poll_every_15_minutes" {
  name    = "tf-example-recurring-poll"
  summary = "Launches the poll task every 15 minutes between 01:00 and 23:40"

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
  name    = "tf-example-recurring-poll-limited"
  summary = "Launches the poll task every 10 minutes, 5 times total"

  target_task = stonebranch_task_unix.poll.name

  recurrence_type          = "Interval"
  recurrence_interval      = "10"
  recurrence_interval_unit = "Minutes"

  indefinite_recurrences = false
  number_of_recurrences  = "5"
}
