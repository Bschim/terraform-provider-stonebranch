# Stonebranch File Transfer Task Example
#
# This example demonstrates how to create file transfer tasks in Stonebranch.
#
# Usage:
#   export STONEBRANCH_API_TOKEN="your-token"
#   export TF_VAR_sftp_password="your-sftp-password"
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

# SFTP download task
resource "stonebranch_task_file_transfer" "sftp_download" {
  name    = "tf-example-sftp-download"
  summary = "Download a file from SFTP server"

  agent_var = var.agent_var

  server_type = "SFTP"

  remote_server      = var.sftp_server
  remote_filename    = "/reports/daily_report.csv"
  local_filename     = "/data/reports/daily_report.csv"
  remote_credentials = stonebranch_credential.sftp.name
}

# SFTP upload task
resource "stonebranch_task_file_transfer" "sftp_upload" {
  name    = "tf-example-sftp-upload"
  summary = "Upload a file to SFTP server"

  agent_var = var.agent_var

  server_type = "SFTP"

  remote_server      = var.sftp_server
  remote_filename    = "/incoming/data_export.csv"
  local_filename     = "/data/exports/data_export.csv"
  remote_credentials = stonebranch_credential.sftp.name
}

# Credential for SFTP authentication
resource "stonebranch_credential" "sftp" {
  name             = "tf-example-sftp-creds"
  description      = "SFTP credentials for file transfer"
  runtime_user     = var.sftp_user
  runtime_password = var.sftp_password
}

# UDM (Universal Data Mover) transfer task
resource "stonebranch_task_file_transfer" "udm_transfer" {
  name    = "tf-example-udm-transfer"
  summary = "Transfer a file via a UDM agent cluster"

  server_type = "UDM"

  primary_broker_choice   = "Agent"
  primary_broker          = "udm-agent-01"
  secondary_broker_choice = "Agent Cluster"
  secondary_cluster_ref   = "tf-example-udm-cluster"

  local_filename  = "/data/incoming"
  remote_filename = "/remote/outgoing/report.csv"

  command       = "GET"
  udm_operation = "Copy"
  format        = "Binary"

  form_or_script = "Script"
  script         = "tf-example-udm-transfer-script"
}

