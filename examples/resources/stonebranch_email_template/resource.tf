# Stonebranch Email Template Example
#
# Email templates define reusable subject/body content and recipients for
# email notifications sent via a stonebranch_email_connection.
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

resource "stonebranch_email_connection" "notifications" {
  name          = "tf-example-notifications"
  smtp          = "smtp.example.com"
  smtp_port     = 25
  email_address = "notifications@example.com"
}

# At least one of "to", "cc", or "bcc" must be set.
resource "stonebranch_email_template" "job_failure" {
  name             = "tf-example-job-failure"
  email_connection = stonebranch_email_connection.notifications.name
  to               = "oncall@example.com"
  subject          = "Job Failed"
  body             = "A scheduled job has failed. Please check the Universal Controller for details."
  description      = "Sent when a critical job fails"
}

resource "stonebranch_email_template" "weekly_report" {
  name             = "tf-example-weekly-report"
  email_connection = stonebranch_email_connection.notifications.name
  cc               = "team-leads@example.com"
  bcc              = "audit@example.com"
  subject          = "Weekly Automation Report"
  body             = "Attached is the weekly automation summary."
}
