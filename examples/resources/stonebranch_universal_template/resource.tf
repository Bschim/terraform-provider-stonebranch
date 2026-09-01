# Stonebranch Universal Template Example
#
# This example demonstrates how to create universal template resources in
# Stonebranch. Universal templates define reusable, script-based custom task
# types that run via a Universal Agent plugin.
#
# Note: this resource models a curated core field set plus the custom UI
# form-builder (`fields`). The `commands`/`events` definitions and the
# top-level per-attribute UI restriction settings are not yet supported and
# must be managed outside Terraform.
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
# task-specific values (region, tags, notes) when launching a task built from
# this template. Each field maps to one of a fixed set of typed extension
# slots via `field_mapping` (see the resource documentation for the full
# list of confirmed slot names).
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
    {
      name          = "notes"
      label         = "Notes"
      field_mapping = "Text Field 1"
      field_type    = "Text"
      hint          = "Optional deployment notes"
    },
  ]
}
