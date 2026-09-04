# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Open-sourced the project under MIT License
- Removed hardcoded default base URL; `base_url` or `STONEBRANCH_BASE_URL` is now required
- Added CI/CD workflows, contributing guidelines, and community files

### Fixed
- `stonebranch_task_file_transfer`: added missing UDM (Universal Data Mover) agent/broker/cluster fields (`primary_broker_choice`, `primary_broker`, `primary_cluster`, `primary_cluster_ref`, `primary_credentials`, `primary_cred_var`, `primary_filesys`, `primary_open_options`, `secondary_broker_choice`, `secondary_broker`, `secondary_cluster`, `secondary_cluster_ref`, `secondary_credentials`, `secondary_cred_var`, `secondary_filesys`, `secondary_open_options`, `udm_operation`, `udm_options`, `script`, `format`, `form_or_script`, `command`), which previously caused UAC to reject `terraform apply` for any UDM-based file transfer task with a misleading `primaryBrokerRef must not be blank` / `secondaryBrokerRef must not be blank` error
- `sb2tf`: `taskFileTransferTemplate` now exports the new UDM fields above
- `sb2tf`: fixed `notEmpty()` silently treating non-zero numeric fields as empty (JSON numbers decode as `float64`), which had dropped 14 fields (`stable_seconds`, `retry_maximum`, `retry_interval`, `smtp_port`, `limit`/`limit_amount`, `max_rows`, `timeout`, `time_interval`, `day`, `nth_amount`, `adjustment_amount`, `retention_duration_rt`, and universal-template field-slot attributes) from every export
- `sb2tf`: fixed `taskFileMonitorTemplate` never exporting `agentVar`/`agentClusterVar`, and `taskTimerTemplate` never exporting `sleep_amount`
- `sb2tf`: fixed the shared `actions` attribute (`system_operations`, `email_notifications`, `abort_actions`, `set_variable_actions`, `snmp_notifications`) being silently dropped from every task template's export
- `sb2tf`: fixed the top-level `variables` attribute being silently dropped from every task/trigger template's export
- `sb2tf`: `export --all` with dependency-following enabled (the default) could silently overwrite a sibling per-type `.tf` file with only the dependency-referenced subset of that type (e.g. `export task_workflow --all` truncating `tasks_unix.tf` to just the tasks referenced by workflows), losing previously-exported resources with no warning; `Finalize()` now refuses to overwrite an output file with fewer resources than it already has unless `--force` is passed

## [0.4.0] - 2026-03-13

### Added
- `stonebranch_task_universal_aws_s3` resource for AWS S3 Universal Tasks
- `sb2tf` CLI utility for exporting existing StoneBranch resources to Terraform

## [0.3.0]

### Added
- `stonebranch_task_stored_procedure` resource
- `stonebranch_task_web_service` resource
- `stonebranch_task_monitor` resource
- `stonebranch_trigger_task_monitor` resource
- `stonebranch_agent_cluster` resource
- `stonebranch_calendar` resource
- `stonebranch_task_file_monitor` resource
- `stonebranch_trigger_file_monitor` resource
- `stonebranch_business_service` resource

## [0.2.0]

### Added
- `stonebranch_task_workflow` resource
- `stonebranch_workflow_vertex` resource
- `stonebranch_workflow_edge` resource
- `stonebranch_task_sql` resource
- `stonebranch_task_email` resource
- `stonebranch_email_connection` resource
- `stonebranch_database_connection` resource
- `stonebranch_trigger_time` resource
- `stonebranch_trigger_cron` resource
- `stonebranch_variable` resource
- `stonebranch_credential` resource
- Data sources: `stonebranch_agents`, `stonebranch_agent_clusters`, `stonebranch_tasks`, `stonebranch_task_instances`, `stonebranch_task`, `stonebranch_trigger`

## [0.1.0]

### Added
- Initial release
- `stonebranch_task_unix` resource
- `stonebranch_task_windows` resource
- `stonebranch_task_file_transfer` resource
- `stonebranch_script` resource
- Provider configuration with `api_token` and `base_url`
- API client with Bearer token authentication
