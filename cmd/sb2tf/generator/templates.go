package generator

import (
	"strings"
	"text/template"
)

// templates maps resource types to their HCL templates.
var templates = map[string]*template.Template{}

func init() {
	// Register all templates
	registerTemplate("variable", variableTemplate)
	registerTemplate("script", scriptTemplate)
	registerTemplate("credential", credentialTemplate)
	registerTemplate("business_service", businessServiceTemplate)
	registerTemplate("calendar", calendarTemplate)
	registerTemplate("agent_cluster", agentClusterTemplate)
	registerTemplate("database_connection", databaseConnectionTemplate)
	registerTemplate("email_connection", emailConnectionTemplate)
	registerTemplate("task_unix", taskUnixTemplate)
	registerTemplate("task_windows", taskWindowsTemplate)
	registerTemplate("task_sql", taskSQLTemplate)
	registerTemplate("task_email", taskEmailTemplate)
	registerTemplate("task_workflow", taskWorkflowTemplate)
	registerTemplate("task_file_monitor", taskFileMonitorTemplate)
	registerTemplate("task_file_transfer", taskFileTransferTemplate)
	registerTemplate("task_timer", taskTimerTemplate)
	registerTemplate("task_monitor", taskMonitorTemplate)
	registerTemplate("task_stored_procedure", taskStoredProcedureTemplate)
	registerTemplate("task_web_service", taskWebServiceTemplate)
	registerTemplate("task_universal_aws_s3", taskUniversalAWSS3Template)
	registerTemplate("task_recurring", taskRecurringTemplate)
	registerTemplate("task_universal", taskUniversalTemplate)
	registerTemplate("trigger_time", triggerTimeTemplate)
	registerTemplate("trigger_cron", triggerCronTemplate)
	registerTemplate("trigger_file_monitor", triggerFileMonitorTemplate)
	registerTemplate("trigger_task_monitor", triggerTaskMonitorTemplate)
	registerTemplate("customday", customDayTemplate)
	registerTemplate("emailtemplate", emailTemplateTemplate)
	registerTemplate("universaltemplate", universalTemplateTemplate)
	registerTemplate("virtualresource", virtualResourceTemplate)
	registerTemplate("workflow_vertex", workflowVertexTemplate)
	registerTemplate("workflow_edge", workflowEdgeTemplate)

	// Set the GetTemplate function in generator.go
	GetTemplate = func(resourceType string) *template.Template {
		return templates[resourceType]
	}
}

func registerTemplate(name, tmpl string) {
	t := template.New(name).Funcs(template.FuncMap{
		"quote":         quote,
		"stringList":    stringList,
		"notEmpty":      notEmpty,
		"isTrue":        isTrue,
		"fieldSet":      fieldSet,
		"hasActions":    hasActions,
		"heredocEscape": heredocEscape,
	})
	// Parse the shared actions_block/*_item sub-templates into t's set
	// first (actionsTemplateDefs contains only {{define}} blocks, so this
	// is a no-op on t's own body), then parse the resource's own template
	// body. Every registered template gets these sub-templates available
	// regardless of whether it invokes {{template "actions_block" ...}} -
	// harmless for the ones that don't.
	template.Must(t.Parse(actionsTemplateDefs))
	template.Must(t.Parse(taskVariablesTemplateDefs))
	template.Must(t.Parse(taskWaitDelayTemplateDefs))
	templates[name] = template.Must(t.Parse(tmpl))
}

// Template helper functions
func quote(s interface{}) string {
	if s == nil {
		return ""
	}
	str, ok := s.(string)
	if !ok {
		return ""
	}
	// Escape backslashes and quotes
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	// Escape $ to $$ for Terraform (prevents interpolation)
	str = strings.ReplaceAll(str, "$", "$$")
	// Escape literal newlines/carriage returns - HCL quoted strings
	// cannot span multiple raw lines.
	str = strings.ReplaceAll(str, "\r\n", "\\n")
	str = strings.ReplaceAll(str, "\n", "\\n")
	str = strings.ReplaceAll(str, "\r", "\\r")
	return str
}

// heredocEscape escapes only the characters that Terraform's HCL heredoc
// (<<-EOT ... EOT) syntax treats specially - "${" and "%{" interpolation/
// directive markers - by doubling the leading "$"/"%". Unlike quote(), it
// must NOT escape backslashes, quotes, or newlines, since heredoc content
// is embedded verbatim (raw multi-line text, e.g. shell script source that
// legitimately contains "${...}" parameter expansion).
func heredocEscape(s interface{}) string {
	if s == nil {
		return ""
	}
	str, ok := s.(string)
	if !ok {
		return ""
	}
	str = strings.ReplaceAll(str, "$", "$$")
	str = strings.ReplaceAll(str, "%", "%%")
	return str
}

func stringList(v interface{}) string {
	if v == nil {
		return ""
	}
	list, ok := v.([]interface{})
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range list {
		if s, ok := item.(string); ok {
			parts = append(parts, "\""+quote(s)+"\"")
		}
	}
	return strings.Join(parts, ", ")
}

func notEmpty(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	case bool:
		return true // booleans are always "not empty" for template purposes
	case float64:
		return val != 0 // JSON numbers decode as float64 in map[string]interface{}
	case float32:
		return val != 0
	case int:
		return val != 0
	case int64:
		return val != 0
	default:
		return false
	}
}

func isTrue(v interface{}) bool {
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// fieldSet reports whether a Universal Template numbered field slot
// (an object with label/name/value keys, e.g. .textField1, .booleanField3,
// .intField1) has a non-nil, non-empty "value". Unlike notEmpty, this also
// accepts non-string values (boolean_field_* slots carry a bool value,
// int_field_1 carries a number), since the slot's underlying JSON shape is
// always an object regardless of the field's declared type.
func fieldSet(v interface{}) bool {
	m, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	val, exists := m["value"]
	if !exists || val == nil {
		return false
	}
	if s, ok := val.(string); ok {
		return s != ""
	}
	return true
}

// hasActions reports whether a raw "actions" JSON object has at least one
// non-empty action sub-list. Mirrors TaskActionsFromAPI's null-collapsing
// logic in internal/provider/resources/task_actions.go, so an `actions` key
// that is present but entirely empty (as seen on tasks with no configured
// actions) does not produce an empty `actions = {}` stub in the generated HCL.
func hasActions(v interface{}) bool {
	m, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	for _, k := range []string{"systemOperations", "emailNotifications", "abortActions", "setVariableActions", "snmpNotifications"} {
		if notEmpty(m[k]) {
			return true
		}
	}
	return false
}

// actionsTemplateDefs contains only {{define}} blocks (no top-level text)
// for rendering the shared `actions` attribute present on every task
// resource type. It is parsed into every registered template's set by
// registerTemplate, and invoked by each task template via
// {{template "actions_block" .actions}}, gated on {{if hasActions .actions}}.
//
// Field names below use raw camelCase JSON keys (matching how every other
// template already accesses fields on the generic map[string]interface{}
// data); the snake_case only appears in the *emitted* HCL attribute names.
// Cross-checked against internal/provider/resources/task_actions.go:
//   - "sys_id" is intentionally omitted everywhere (Computed-only, never
//     Optional, so it's never something a generated .tf file should set).
//   - "variable_name" in set_variable_action_item is unconditional (schema:
//     Required: true).
const actionsTemplateDefs = `
{{define "common_action_attrs"}}
{{- if notEmpty .status}}
      status = "{{quote .status}}"
{{- end}}
{{- if isTrue .notifyOnLateStart}}
      notify_on_late_start = true
{{- end}}
{{- if isTrue .notifyOnLateFinish}}
      notify_on_late_finish = true
{{- end}}
{{- if isTrue .notifyOnEarlyFinish}}
      notify_on_early_finish = true
{{- end}}
{{- if isTrue .notifyOnProjectedLate}}
      notify_on_projected_late = true
{{- end}}
{{- if notEmpty .exitCodes}}
      exit_codes = "{{quote .exitCodes}}"
{{- end}}
{{- if notEmpty .description}}
      description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .inheritance}}
      inheritance = "{{quote .inheritance}}"
{{- end}}
{{end}}

{{define "abort_action_item"}}
    {
{{- template "common_action_attrs" .}}
{{- if isTrue .cancelProcess}}
      cancel_process = true
{{- end}}
{{- if notEmpty .overrideExitCode}}
      override_exit_code = "{{quote .overrideExitCode}}"
{{- end}}
{{- if isTrue .haltOnFinish}}
      halt_on_finish = true
{{- end}}
    },
{{end}}

{{define "email_notification_item"}}
    {
{{- template "common_action_attrs" .}}
{{- if notEmpty .subject}}
      subject = "{{quote .subject}}"
{{- end}}
{{- if notEmpty .body}}
      body = "{{quote .body}}"
{{- end}}
{{- if notEmpty .to}}
      to = "{{quote .to}}"
{{- end}}
{{- if notEmpty .cc}}
      cc = "{{quote .cc}}"
{{- end}}
{{- if notEmpty .bcc}}
      bcc = "{{quote .bcc}}"
{{- end}}
{{- if notEmpty .replyTo}}
      reply_to = "{{quote .replyTo}}"
{{- end}}
{{- if notEmpty .emailTemplate}}
      email_template = "{{quote .emailTemplate}}"
{{- end}}
{{- if notEmpty .emailTemplateVar}}
      email_template_var = "{{quote .emailTemplateVar}}"
{{- end}}
{{- if notEmpty .emailConnection}}
      email_connection = "{{quote .emailConnection}}"
{{- end}}
{{- if isTrue .attachStdError}}
      attach_std_error = true
{{- end}}
{{- if isTrue .attachStdOut}}
      attach_std_out = true
{{- end}}
{{- if isTrue .attachFile}}
      attach_file = true
{{- end}}
{{- if notEmpty .fileName}}
      file_name = "{{quote .fileName}}"
{{- end}}
{{- if notEmpty .fileNumLines}}
      file_num_lines = {{.fileNumLines}}
{{- end}}
{{- if notEmpty .fileStartLine}}
      file_start_line = {{.fileStartLine}}
{{- end}}
{{- if notEmpty .fileScanText}}
      file_scan_text = "{{quote .fileScanText}}"
{{- end}}
{{- if notEmpty .stderrNumLines}}
      stderr_num_lines = {{.stderrNumLines}}
{{- end}}
{{- if notEmpty .stderrStartLine}}
      stderr_start_line = {{.stderrStartLine}}
{{- end}}
{{- if notEmpty .stderrScanText}}
      stderr_scan_text = "{{quote .stderrScanText}}"
{{- end}}
{{- if notEmpty .stdoutNumLines}}
      stdout_num_lines = {{.stdoutNumLines}}
{{- end}}
{{- if notEmpty .stdoutStartLine}}
      stdout_start_line = {{.stdoutStartLine}}
{{- end}}
{{- if notEmpty .stdoutScanText}}
      stdout_scan_text = "{{quote .stdoutScanText}}"
{{- end}}
{{- if isTrue .attachJobLog}}
      attach_job_log = true
{{- end}}
{{- if notEmpty .joblogStartLine}}
      joblog_start_line = {{.joblogStartLine}}
{{- end}}
{{- if notEmpty .joblogNumLines}}
      joblog_num_lines = {{.joblogNumLines}}
{{- end}}
{{- if notEmpty .joblogScanText}}
      joblog_scan_text = "{{quote .joblogScanText}}"
{{- end}}
{{- if notEmpty .reportId}}
      report_id = "{{quote .reportId}}"
{{- end}}
{{- if notEmpty .reportVar}}
      report_var = "{{quote .reportVar}}"
{{- end}}
{{- if notEmpty .useReportVar}}
      use_report_var = "{{quote .useReportVar}}"
{{- end}}
{{- if notEmpty .listReportFormat}}
      list_report_format = "{{quote .listReportFormat}}"
{{- end}}
{{- if isTrue .attachLocalFile}}
      attach_local_file = true
{{- end}}
{{- if notEmpty .localAttachment}}
      local_attachment = "{{quote .localAttachment}}"
{{- end}}
{{- if notEmpty .localAttachmentsPath}}
      local_attachments_path = "{{quote .localAttachmentsPath}}"
{{- end}}
{{- if .report}}
      report = {
{{- if notEmpty .report.title}}
        title = "{{quote .report.title}}"
{{- end}}
{{- if notEmpty .report.userName}}
        user_name = "{{quote .report.userName}}"
{{- end}}
{{- if notEmpty .report.groupName}}
        group_name = "{{quote .report.groupName}}"
{{- end}}
{{- if notEmpty .report.groupNames}}
        group_names = [{{stringList .report.groupNames}}]
{{- end}}
      }
{{- end}}
    },
{{end}}

{{define "set_variable_action_item"}}
    {
{{- template "common_action_attrs" .}}
{{- if notEmpty .variableScope}}
      variable_scope = "{{quote .variableScope}}"
{{- end}}
      variable_name = "{{quote .variableName}}"
{{- if notEmpty .variableDescription}}
      variable_description = "{{quote .variableDescription}}"
{{- end}}
{{- if notEmpty .variableValue}}
      variable_value = "{{quote .variableValue}}"
{{- end}}
{{- if notEmpty .notificationOption}}
      notification_option = "{{quote .notificationOption}}"
{{- end}}
    },
{{end}}

{{define "snmp_notification_item"}}
    {
{{- template "common_action_attrs" .}}
{{- if notEmpty .snmpManager}}
      snmp_manager = "{{quote .snmpManager}}"
{{- end}}
{{- if notEmpty .severity}}
      severity = "{{quote .severity}}"
{{- end}}
    },
{{end}}

{{define "system_operation_item"}}
    {
{{- template "common_action_attrs" .}}
{{- if notEmpty .operation}}
      operation = "{{quote .operation}}"
{{- end}}
{{- if notEmpty .task}}
      task = "{{quote .task}}"
{{- end}}
{{- if notEmpty .taskVar}}
      task_var = "{{quote .taskVar}}"
{{- end}}
{{- if notEmpty .taskLimitType}}
      task_limit_type = "{{quote .taskLimitType}}"
{{- end}}
{{- if notEmpty .limit}}
      limit = "{{quote .limit}}"
{{- end}}
{{- if notEmpty .virtualResource}}
      virtual_resource = "{{quote .virtualResource}}"
{{- end}}
{{- if notEmpty .virtualResourceVar}}
      virtual_resource_var = "{{quote .virtualResourceVar}}"
{{- end}}
{{- if notEmpty .execCommand}}
      exec_command = "{{quote .execCommand}}"
{{- end}}
{{- if notEmpty .execCriteria}}
      exec_criteria = "{{quote .execCriteria}}"
{{- end}}
{{- if notEmpty .execLookupOption}}
      exec_lookup_option = "{{quote .execLookupOption}}"
{{- end}}
{{- if notEmpty .execName}}
      exec_name = "{{quote .execName}}"
{{- end}}
{{- if notEmpty .execId}}
      exec_id = "{{quote .execId}}"
{{- end}}
{{- if notEmpty .execWorkflowName}}
      exec_workflow_name = "{{quote .execWorkflowName}}"
{{- end}}
{{- if notEmpty .execWorkflowNameCond}}
      exec_workflow_name_cond = "{{quote .execWorkflowNameCond}}"
{{- end}}
{{- if notEmpty .agent}}
      agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentVar}}
      agent_var = "{{quote .agentVar}}"
{{- end}}
{{- if notEmpty .agentCluster}}
      agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .agentClusterVar}}
      agent_cluster_var = "{{quote .agentClusterVar}}"
{{- end}}
{{- if notEmpty .trigger}}
      trigger = "{{quote .trigger}}"
{{- end}}
{{- if notEmpty .triggerVar}}
      trigger_var = "{{quote .triggerVar}}"
{{- end}}
{{- if notEmpty .overrideTriggerTime}}
      override_trigger_time = "{{quote .overrideTriggerTime}}"
{{- end}}
{{- if isTrue .overrideTriggerDateTime}}
      override_trigger_date_time = true
{{- end}}
{{- if notEmpty .overrideTriggerDateOffset}}
      override_trigger_date_offset = "{{quote .overrideTriggerDateOffset}}"
{{- end}}
{{- if isTrue .vertexSelection}}
      vertex_selection = true
{{- end}}
{{- if notEmpty .vertices}}
      vertices = [
{{- range .vertices}}
        {
{{- if notEmpty .taskName}}
          task_name = "{{quote .taskName}}"
{{- end}}
{{- if notEmpty .vertexName}}
          vertex_name = "{{quote .vertexName}}"
{{- end}}
{{- if notEmpty .vertexId}}
          vertex_id = "{{quote .vertexId}}"
{{- end}}
        },
{{- end}}
      ]
{{- end}}
{{- if notEmpty .notificationOption}}
      notification_option = "{{quote .notificationOption}}"
{{- end}}
{{- if isTrue .variablesUnresolved}}
      variables_unresolved = true
{{- end}}
{{- if notEmpty .variables}}
      variables = [
{{- range .variables}}
        {
          name = "{{quote .name}}"
{{- if notEmpty .value}}
          value = "{{quote .value}}"
{{- end}}
{{- if notEmpty .description}}
          description = "{{quote .description}}"
{{- end}}
        },
{{- end}}
      ]
{{- end}}
    },
{{end}}

{{define "actions_block"}}
  actions = {
{{- if notEmpty .systemOperations}}
    system_operations = [
{{- range .systemOperations}}
{{template "system_operation_item" .}}
{{- end}}
    ]
{{- end}}
{{- if notEmpty .emailNotifications}}
    email_notifications = [
{{- range .emailNotifications}}
{{template "email_notification_item" .}}
{{- end}}
    ]
{{- end}}
{{- if notEmpty .abortActions}}
    abort_actions = [
{{- range .abortActions}}
{{template "abort_action_item" .}}
{{- end}}
    ]
{{- end}}
{{- if notEmpty .setVariableActions}}
    set_variable_actions = [
{{- range .setVariableActions}}
{{template "set_variable_action_item" .}}
{{- end}}
    ]
{{- end}}
{{- if notEmpty .snmpNotifications}}
    snmp_notifications = [
{{- range .snmpNotifications}}
{{template "snmp_notification_item" .}}
{{- end}}
    ]
{{- end}}
  }
{{end}}
`

// taskVariablesTemplateDefs renders the task/trigger-level "variables"
// attribute (distinct from actions.set_variable_actions and the nested
// variables inside a system_operation_item - same JSON key, different
// scope). See TaskVariablesSchema() in internal/provider/resources/helpers.go.
const taskVariablesTemplateDefs = `
{{define "task_variables_block"}}
  variables = [
{{- range .}}
    {
      name = "{{quote .name}}"
{{- if notEmpty .value}}
      value = "{{quote .value}}"
{{- end}}
{{- if notEmpty .description}}
      description = "{{quote .description}}"
{{- end}}
    },
{{- end}}
  ]
{{end}}
`

// taskWaitDelayTemplateDefs is shared by every task_* resource template -
// they all embed the same wait_to_start/delay_on_start fields
// (task_common_fields.go), backed by the same "tw*"-prefixed raw JSON keys
// on every task API model.
const taskWaitDelayTemplateDefs = `
{{define "task_wait_delay_block"}}
{{- if notEmpty .twWaitType}}
  wait_to_start = "{{quote .twWaitType}}"
{{- end}}
{{- if notEmpty .twWaitTime}}
  wait_time = "{{quote .twWaitTime}}"
{{- end}}
{{- if notEmpty .twWaitDuration}}
  wait_duration = "{{quote .twWaitDuration}}"
{{- end}}
{{- if notEmpty .twWaitAmount}}
  wait_amount = "{{quote .twWaitAmount}}"
{{- end}}
{{- if notEmpty .twWaitDayConstraint}}
  wait_day_constraint = "{{quote .twWaitDayConstraint}}"
{{- end}}
{{- if notEmpty .twDelayType}}
  delay_on_start = "{{quote .twDelayType}}"
{{- end}}
{{- if notEmpty .twDelayDuration}}
  delay_duration = "{{quote .twDelayDuration}}"
{{- end}}
{{- if notEmpty .twDelayAmount}}
  delay_amount = "{{quote .twDelayAmount}}"
{{- end}}
{{- if notEmpty .twWorkflowOnly}}
  workflow_only = "{{quote .twWorkflowOnly}}"
{{- end}}
{{end}}
`

// ============================================================================
// Simple Resources
// ============================================================================

const variableTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .value}}
  value = "{{quote .value}}"
{{- end}}
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

const scriptTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .scriptName}}"
{{- if notEmpty .scriptType}}
  script_type = "{{quote .scriptType}}"
{{- end}}
  content = <<-EOT
{{heredocEscape .content}}
EOT
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if isTrue .resolveVariables}}
  resolve_variables = true
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

const credentialTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .runtimeUser}}
  runtime_user = "{{quote .runtimeUser}}"
{{- end}}
  # Note: password/key_location cannot be exported for security reasons
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

const businessServiceTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
}
`

const calendarTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .comments}}
  comments = "{{quote .comments}}"
{{- end}}
{{- if notEmpty .businessDays}}
  business_days = "{{quote .businessDays}}"
{{- end}}
{{- if notEmpty .firstDayOfWeek}}
  first_day_of_week = "{{quote .firstDayOfWeek}}"
{{- end}}
{{- if notEmpty .firstQuarterMonth}}
  first_quarter_month = "{{quote .firstQuarterMonth}}"
  first_quarter_day = {{.firstQuarterDay}}
{{- end}}
{{- if notEmpty .secondQuarterMonth}}
  second_quarter_month = "{{quote .secondQuarterMonth}}"
  second_quarter_day = {{.secondQuarterDay}}
{{- end}}
{{- if notEmpty .thirdQuarterMonth}}
  third_quarter_month = "{{quote .thirdQuarterMonth}}"
  third_quarter_day = {{.thirdQuarterDay}}
{{- end}}
{{- if notEmpty .fourthQuarterMonth}}
  fourth_quarter_month = "{{quote .fourthQuarterMonth}}"
  fourth_quarter_day = {{.fourthQuarterDay}}
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

const agentClusterTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .type}}
  type = "{{quote .type}}"
{{- end}}
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .distribution}}
  distribution = "{{quote .distribution}}"
{{- end}}
{{- if notEmpty .limitType}}
  limit_type = "{{quote .limitType}}"
{{- end}}
{{- if notEmpty .limitAmount}}
  limit_amount = {{.limitAmount}}
{{- end}}
{{- if notEmpty .agents}}
  agents = [{{stringList .agents}}]
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

// ============================================================================
// Connections
// ============================================================================

const databaseConnectionTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .dbDriver}}
  db_driver = "{{quote .dbDriver}}"
{{- end}}
{{- if notEmpty .dbUrl}}
  db_url = "{{quote .dbUrl}}"
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{- if notEmpty .credentialsVar}}
  credentials_var = "{{quote .credentialsVar}}"
{{- end}}
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .maxRows}}
  max_rows = {{.maxRows}}
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

const emailConnectionTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .smtp}}
  smtp = "{{quote .smtp}}"
{{- end}}
{{- if notEmpty .smtpPort}}
  smtp_port = {{.smtpPort}}
{{- end}}
{{- if isTrue .smtpSsl}}
  smtp_ssl = true
{{- end}}
{{- if isTrue .smtpStarttls}}
  smtp_starttls = true
{{- end}}
{{- if notEmpty .emailAddress}}
  email_address = "{{quote .emailAddress}}"
{{- end}}
{{- if notEmpty .authentication}}
  authentication = "{{quote .authentication}}"
{{- end}}
{{- if notEmpty .authenticationType}}
  authentication_type = "{{quote .authenticationType}}"
{{- end}}
{{- if notEmpty .defaultUser}}
  default_user = "{{quote .defaultUser}}"
{{- end}}
  # Note: default_password cannot be exported for security reasons
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

// ============================================================================
// Task Resources
// ============================================================================

const taskUnixTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .agent}}
  agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentCluster}}
  agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .agentVar}}
  agent_var = "{{quote .agentVar}}"
{{- end}}
{{- if notEmpty .agentClusterVar}}
  agent_cluster_var = "{{quote .agentClusterVar}}"
{{- end}}
{{- if notEmpty .command}}
  command = "{{quote .command}}"
{{- end}}
{{- if notEmpty .commandOrScript}}
  command_or_script = "{{quote .commandOrScript}}"
{{- end}}
{{- if notEmpty .script}}
  script = "{{quote .script}}"
{{- end}}
{{- if notEmpty .runtimeDir}}
  runtime_dir = "{{quote .runtimeDir}}"
{{- end}}
{{- if notEmpty .parameters}}
  parameters = "{{quote .parameters}}"
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{- if notEmpty .credentialsVar}}
  credentials_var = "{{quote .credentialsVar}}"
{{- end}}
{{- if notEmpty .exitCodes}}
  exit_codes = "{{quote .exitCodes}}"
{{- end}}
{{- if notEmpty .exitCodeProcessing}}
  exit_code_processing = "{{quote .exitCodeProcessing}}"
{{- end}}
{{- if notEmpty .exitCodeText}}
  exit_code_text = "{{quote .exitCodeText}}"
{{- end}}
{{- if isTrue .runAsSudo}}
  run_as_sudo = true
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskWindowsTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .agent}}
  agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentCluster}}
  agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .agentVar}}
  agent_var = "{{quote .agentVar}}"
{{- end}}
{{- if notEmpty .agentClusterVar}}
  agent_cluster_var = "{{quote .agentClusterVar}}"
{{- end}}
{{- if notEmpty .command}}
  command = "{{quote .command}}"
{{- end}}
{{- if notEmpty .commandOrScript}}
  command_or_script = "{{quote .commandOrScript}}"
{{- end}}
{{- if notEmpty .script}}
  script = "{{quote .script}}"
{{- end}}
{{- if notEmpty .runtimeDir}}
  runtime_dir = "{{quote .runtimeDir}}"
{{- end}}
{{- if notEmpty .parameters}}
  parameters = "{{quote .parameters}}"
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{- if notEmpty .credentialsVar}}
  credentials_var = "{{quote .credentialsVar}}"
{{- end}}
{{- if notEmpty .exitCodes}}
  exit_codes = "{{quote .exitCodes}}"
{{- end}}
{{- if notEmpty .exitCodeProcessing}}
  exit_code_processing = "{{quote .exitCodeProcessing}}"
{{- end}}
{{- if notEmpty .exitCodeText}}
  exit_code_text = "{{quote .exitCodeText}}"
{{- end}}
{{- if isTrue .elevateUser}}
  elevate_user = true
{{- end}}
{{- if isTrue .desktopInteract}}
  desktop_interact = true
{{- end}}
{{- if isTrue .createConsole}}
  create_console = true
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskSQLTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .databaseConnection}}
  database_connection = "{{quote .databaseConnection}}"
{{- end}}
{{- if notEmpty .sqlStatement}}
  sql_statement = "{{quote .sqlStatement}}"
{{- end}}
{{- if notEmpty .sqlCommand}}
  sql_command = "{{quote .sqlCommand}}"
{{- end}}
{{- if notEmpty .columnType}}
  column_type = "{{quote .columnType}}"
{{- end}}
{{- if notEmpty .columnValue}}
  column_value = "{{quote .columnValue}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskEmailTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .connection}}
  email_connection = "{{quote .connection}}"
{{- end}}
{{- if notEmpty .connectionVar}}
  email_connection_var = "{{quote .connectionVar}}"
{{- end}}
{{- if notEmpty .template}}
  template = "{{quote .template}}"
{{- end}}
{{- if notEmpty .templateVar}}
  template_var = "{{quote .templateVar}}"
{{- end}}
{{- if notEmpty .toRecipients}}
  to_recipients = "{{quote .toRecipients}}"
{{- end}}
{{- if notEmpty .ccRecipients}}
  cc_recipients = "{{quote .ccRecipients}}"
{{- end}}
{{- if notEmpty .bccRecipients}}
  bcc_recipients = "{{quote .bccRecipients}}"
{{- end}}
{{- if notEmpty .replyTo}}
  reply_to = "{{quote .replyTo}}"
{{- end}}
{{- if notEmpty .subject}}
  subject = "{{quote .subject}}"
{{- end}}
{{- if notEmpty .body}}
  body = "{{quote .body}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskWorkflowTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if isTrue .calculateCriticalPath}}
  calculate_critical_path = true
{{- end}}
{{- if notEmpty .skippedOption}}
  skipped_option = "{{quote .skippedOption}}"
{{- end}}
{{- if notEmpty .instanceWait}}
  instance_wait = "{{quote .instanceWait}}"
{{- end}}
{{- if notEmpty .instanceWaitLookup}}
  instance_wait_lookup = "{{quote .instanceWaitLookup}}"
{{- end}}
{{- if notEmpty .layoutOption}}
  layout_option = "{{quote .layoutOption}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskFileMonitorTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .agent}}
  agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentCluster}}
  agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .agentVar}}
  agent_var = "{{quote .agentVar}}"
{{- end}}
{{- if notEmpty .agentClusterVar}}
  agent_cluster_var = "{{quote .agentClusterVar}}"
{{- end}}
{{- if notEmpty .fileName}}
  file_name = "{{quote .fileName}}"
{{- end}}
{{- if isTrue .useRegex}}
  use_regex = true
{{- end}}
{{- if notEmpty .stableSeconds}}
  stable_seconds = {{.stableSeconds}}
{{- end}}
{{- if notEmpty .fmType}}
  fm_type = "{{quote .fmType}}"
{{- end}}
{{- if isTrue .recursive}}
  recursive = true
{{- end}}
{{- if isTrue .triggerOnExist}}
  trigger_on_exist = true
{{- end}}
{{- if isTrue .triggerOnCreate}}
  trigger_on_create = true
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskFileTransferTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .agent}}
  agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentCluster}}
  agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .agentVar}}
  agent_var = "{{quote .agentVar}}"
{{- end}}
{{- if notEmpty .agentClusterVar}}
  agent_cluster_var = "{{quote .agentClusterVar}}"
{{- end}}
{{- if notEmpty .transferDirection}}
  transfer_direction = "{{quote .transferDirection}}"
{{- end}}
{{- if notEmpty .transferMode}}
  transfer_mode = "{{quote .transferMode}}"
{{- end}}
{{- if notEmpty .serverType}}
  server_type = "{{quote .serverType}}"
{{- end}}
{{- if notEmpty .remoteServer}}
  remote_server = "{{quote .remoteServer}}"
{{- end}}
{{- if notEmpty .remoteFilename}}
  remote_filename = "{{quote .remoteFilename}}"
{{- end}}
{{- if notEmpty .remoteCredentials}}
  remote_credentials = "{{quote .remoteCredentials}}"
{{- end}}
{{- if notEmpty .remoteCredVar}}
  remote_credentials_var = "{{quote .remoteCredVar}}"
{{- end}}
{{- if notEmpty .localFilename}}
  local_filename = "{{quote .localFilename}}"
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{- if notEmpty .credentialsVar}}
  credentials_var = "{{quote .credentialsVar}}"
{{- end}}
{{- if notEmpty .exitCodes}}
  exit_codes = "{{quote .exitCodes}}"
{{- end}}
{{- if notEmpty .exitCodeProcessing}}
  exit_code_processing = "{{quote .exitCodeProcessing}}"
{{- end}}
{{- if notEmpty .exitCodeText}}
  exit_code_text = "{{quote .exitCodeText}}"
{{- end}}
{{- if isTrue .useRegex}}
  use_regex = true
{{- end}}
{{- if notEmpty .encrypt}}
  encrypt = "{{quote .encrypt}}"
{{- end}}
{{- if notEmpty .compress}}
  compress = "{{quote .compress}}"
{{- end}}
{{- if notEmpty .primaryBrokerChoice}}
  primary_broker_choice = "{{quote .primaryBrokerChoice}}"
{{- end}}
{{- if notEmpty .primaryBroker}}
  primary_broker = "{{quote .primaryBroker}}"
{{- end}}
{{- if notEmpty .primaryCluster}}
  primary_cluster = "{{quote .primaryCluster}}"
{{- end}}
{{- if notEmpty .primaryClusterRef}}
  primary_cluster_ref = "{{quote .primaryClusterRef}}"
{{- end}}
{{- if notEmpty .primaryCredentials}}
  primary_credentials = "{{quote .primaryCredentials}}"
{{- end}}
{{- if notEmpty .primaryCredVar}}
  primary_cred_var = "{{quote .primaryCredVar}}"
{{- end}}
{{- if notEmpty .primaryFilesys}}
  primary_filesys = "{{quote .primaryFilesys}}"
{{- end}}
{{- if notEmpty .primaryOpenOptions}}
  primary_open_options = "{{quote .primaryOpenOptions}}"
{{- end}}
{{- if notEmpty .secondaryBrokerChoice}}
  secondary_broker_choice = "{{quote .secondaryBrokerChoice}}"
{{- end}}
{{- if notEmpty .secondaryBroker}}
  secondary_broker = "{{quote .secondaryBroker}}"
{{- end}}
{{- if notEmpty .secondaryCluster}}
  secondary_cluster = "{{quote .secondaryCluster}}"
{{- end}}
{{- if notEmpty .secondaryClusterRef}}
  secondary_cluster_ref = "{{quote .secondaryClusterRef}}"
{{- end}}
{{- if notEmpty .secondaryCredentials}}
  secondary_credentials = "{{quote .secondaryCredentials}}"
{{- end}}
{{- if notEmpty .secondaryCredVar}}
  secondary_cred_var = "{{quote .secondaryCredVar}}"
{{- end}}
{{- if notEmpty .secondaryFilesys}}
  secondary_filesys = "{{quote .secondaryFilesys}}"
{{- end}}
{{- if notEmpty .secondaryOpenOptions}}
  secondary_open_options = "{{quote .secondaryOpenOptions}}"
{{- end}}
{{- if notEmpty .udmOperation}}
  udm_operation = "{{quote .udmOperation}}"
{{- end}}
{{- if notEmpty .udmOptions}}
  udm_options = "{{quote .udmOptions}}"
{{- end}}
{{- if notEmpty .script}}
  script = "{{quote .script}}"
{{- end}}
{{- if notEmpty .format}}
  format = "{{quote .format}}"
{{- end}}
{{- if notEmpty .formOrScript}}
  form_or_script = "{{quote .formOrScript}}"
{{- end}}
{{- if notEmpty .command}}
  command = "{{quote .command}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskTimerTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .sleepType}}
  sleep_type = "{{quote .sleepType}}"
{{- end}}
{{- if notEmpty .sleepAmount}}
  sleep_amount = "{{quote .sleepAmount}}"
{{- end}}
{{- if notEmpty .sleepDuration}}
  sleep_duration = "{{quote .sleepDuration}}"
{{- end}}
{{- if notEmpty .sleepTime}}
  sleep_time = "{{quote .sleepTime}}"
{{- end}}
{{- if notEmpty .sleepDayConstraint}}
  sleep_day_constraint = "{{quote .sleepDayConstraint}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskMonitorTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .taskMonName}}
  task_mon_name = "{{quote .taskMonName}}"
{{- end}}
{{- if notEmpty .statusText}}
  status_text = "{{quote .statusText}}"
{{- end}}
{{- if notEmpty .monType}}
  mon_type = "{{quote .monType}}"
{{- end}}
{{- if notEmpty .typeText}}
  type_text = "{{quote .typeText}}"
{{- end}}
{{- if notEmpty .timeScope}}
  time_scope = "{{quote .timeScope}}"
{{- end}}
{{- if isTrue .monitorLateStart}}
  monitor_late_start = true
{{- end}}
{{- if isTrue .monitorLateFinish}}
  monitor_late_finish = true
{{- end}}
{{- if isTrue .monitorEarlyFinish}}
  monitor_early_finish = true
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskStoredProcedureTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .storedProcName}}
  stored_proc_name = "{{quote .storedProcName}}"
{{- end}}
{{- if notEmpty .databaseConnection}}
  database_connection = "{{quote .databaseConnection}}"
{{- end}}
{{- if notEmpty .connectionVar}}
  connection_var = "{{quote .connectionVar}}"
{{- end}}
{{- if notEmpty .maxRows}}
  max_rows = {{.maxRows}}
{{- end}}
{{- if isTrue .autoCleanup}}
  auto_cleanup = true
{{- end}}
{{- if notEmpty .resultProcessing}}
  result_processing = "{{quote .resultProcessing}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskWebServiceTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .url}}
  url = "{{quote .url}}"
{{- end}}
{{- if notEmpty .protocol}}
  protocol = "{{quote .protocol}}"
{{- end}}
{{- if notEmpty .httpMethod}}
  http_method = "{{quote .httpMethod}}"
{{- end}}
{{- if notEmpty .mimeType}}
  mime_type = "{{quote .mimeType}}"
{{- end}}
{{- if notEmpty .httpAuth}}
  http_auth = "{{quote .httpAuth}}"
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{- if notEmpty .payload}}
  payload = "{{quote .payload}}"
{{- end}}
{{- if notEmpty .responseProcessingType}}
  response_processing_type = "{{quote .responseProcessingType}}"
{{- end}}
{{- if notEmpty .statusCodeRange}}
  status_code_range = "{{quote .statusCodeRange}}"
{{- end}}
{{- if notEmpty .timeout}}
  timeout = {{.timeout}}
{{- end}}
{{- if isTrue .insecure}}
  insecure = true
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskUniversalAWSS3Template = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .agent}}
  agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentCluster}}
  agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
  # Note: Universal task template fields may require manual adjustment
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const taskRecurringTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
  target_task = "{{quote .targetTask}}"
{{- if notEmpty .targetTaskMonitorCondition}}
  target_task_monitor_condition = "{{quote .targetTaskMonitorCondition}}"
{{- end}}
{{- if notEmpty .targetTaskStatusText}}
  target_task_status_text = "{{quote .targetTaskStatusText}}"
{{- end}}
{{- if notEmpty .recurrenceType}}
  recurrence_type = "{{quote .recurrenceType}}"
{{- end}}
{{- if notEmpty .recurrenceInterval}}
  recurrence_interval = "{{quote .recurrenceInterval}}"
{{- end}}
{{- if notEmpty .recurrenceIntervalUnit}}
  recurrence_interval_unit = "{{quote .recurrenceIntervalUnit}}"
{{- end}}
{{- if notEmpty .intervalStartTime}}
  interval_start_time = "{{quote .intervalStartTime}}"
{{- end}}
{{- if notEmpty .intervalEndTime}}
  interval_end_time = "{{quote .intervalEndTime}}"
{{- end}}
{{- if notEmpty .intervalStartDayConstraint}}
  interval_start_day_constraint = "{{quote .intervalStartDayConstraint}}"
{{- end}}
{{- if notEmpty .intervalEndDayConstraint}}
  interval_end_day_constraint = "{{quote .intervalEndDayConstraint}}"
{{- end}}
{{- if isTrue .timeWindow}}
  time_window = true
{{- end}}
{{- if isTrue .indefiniteRecurrences}}
  indefinite_recurrences = true
{{- end}}
{{- if notEmpty .numberOfRecurrences}}
  number_of_recurrences = "{{quote .numberOfRecurrences}}"
{{- end}}
{{- if notEmpty .skipCondition}}
  skip_condition = "{{quote .skipCondition}}"
{{- end}}
{{- if notEmpty .retentionDurationRt}}
  retention_duration_rt = {{.retentionDurationRt}}
{{- end}}
{{- if notEmpty .retentionDurationUnitRt}}
  retention_duration_unit_rt = "{{quote .retentionDurationUnitRt}}"
{{- end}}
{{- if isTrue .retentionDurationPurgeRt}}
  retention_duration_purge_rt = true
{{- end}}
{{- if isTrue .rdExcludeBackupRt}}
  rd_exclude_backup_rt = true
{{- end}}
{{- if isTrue .holdResources}}
  hold_resources = true
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

// taskUniversalTemplate handles the generic Universal Template task
// (stonebranch_task_universal). Unlike other task_* resources, the numbered
// field slots (text_field_1, boolean_field_1, etc.) are declared as plain
// scalar attributes (StringAttribute/BoolAttribute/Int64Attribute) in the
// provider schema, NOT as nested objects - only the raw JSON/API
// representation nests each slot as {label, name, value}. The fieldSet
// helper checks the raw JSON shape while the emitted HCL assigns the
// unwrapped scalar .value.
const taskUniversalTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
  template = "{{quote .template}}"
{{- if notEmpty .agent}}
  agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentCluster}}
  agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .agentVar}}
  agent_var = "{{quote .agentVar}}"
{{- end}}
{{- if notEmpty .agentClusterVar}}
  agent_cluster_var = "{{quote .agentClusterVar}}"
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{- if notEmpty .credentialsVar}}
  credentials_var = "{{quote .credentialsVar}}"
{{- end}}
{{- if notEmpty .exitCodes}}
  exit_codes = "{{quote .exitCodes}}"
{{- end}}
{{- if notEmpty .exitCodeProcessing}}
  exit_code_processing = "{{quote .exitCodeProcessing}}"
{{- end}}
{{- if notEmpty .exitCodeText}}
  exit_code_text = "{{quote .exitCodeText}}"
{{- end}}
{{- if notEmpty .retryMaximum}}
  retry_maximum = {{.retryMaximum}}
{{- end}}
{{- if isTrue .retryIndefinitely}}
  retry_indefinitely = true
{{- end}}
{{- if notEmpty .retryInterval}}
  retry_interval = {{.retryInterval}}
{{- end}}
{{- if isTrue .retrySuppressFailure}}
  retry_suppress_failure = true
{{- end}}
{{- if isTrue .holdResources}}
  hold_resources = true
{{- end}}
{{- if fieldSet .textField1}}
  text_field_1 = "{{quote .textField1.value}}"
{{- end}}
{{- if fieldSet .textField2}}
  text_field_2 = "{{quote .textField2.value}}"
{{- end}}
{{- if fieldSet .textField3}}
  text_field_3 = "{{quote .textField3.value}}"
{{- end}}
{{- if fieldSet .textField4}}
  text_field_4 = "{{quote .textField4.value}}"
{{- end}}
{{- if fieldSet .textField5}}
  text_field_5 = "{{quote .textField5.value}}"
{{- end}}
{{- if fieldSet .textField6}}
  text_field_6 = "{{quote .textField6.value}}"
{{- end}}
{{- if fieldSet .textField7}}
  text_field_7 = "{{quote .textField7.value}}"
{{- end}}
{{- if fieldSet .textField8}}
  text_field_8 = "{{quote .textField8.value}}"
{{- end}}
{{- if fieldSet .textField9}}
  text_field_9 = "{{quote .textField9.value}}"
{{- end}}
{{- if fieldSet .textField10}}
  text_field_10 = "{{quote .textField10.value}}"
{{- end}}
{{- if fieldSet .choiceField1}}
  choice_field_1 = "{{quote .choiceField1.value}}"
{{- end}}
{{- if fieldSet .choiceField2}}
  choice_field_2 = "{{quote .choiceField2.value}}"
{{- end}}
{{- if fieldSet .choiceField3}}
  choice_field_3 = "{{quote .choiceField3.value}}"
{{- end}}
{{- if fieldSet .choiceField4}}
  choice_field_4 = "{{quote .choiceField4.value}}"
{{- end}}
{{- if fieldSet .choiceField5}}
  choice_field_5 = "{{quote .choiceField5.value}}"
{{- end}}
{{- if fieldSet .choiceField6}}
  choice_field_6 = "{{quote .choiceField6.value}}"
{{- end}}
{{- if fieldSet .choiceField7}}
  choice_field_7 = "{{quote .choiceField7.value}}"
{{- end}}
{{- if fieldSet .choiceField8}}
  choice_field_8 = "{{quote .choiceField8.value}}"
{{- end}}
{{- if fieldSet .choiceField9}}
  choice_field_9 = "{{quote .choiceField9.value}}"
{{- end}}
{{- if fieldSet .choiceField10}}
  choice_field_10 = "{{quote .choiceField10.value}}"
{{- end}}
{{- if fieldSet .choiceField11}}
  choice_field_11 = "{{quote .choiceField11.value}}"
{{- end}}
{{- if and (fieldSet .booleanField1) (isTrue .booleanField1.value)}}
  boolean_field_1 = true
{{- end}}
{{- if and (fieldSet .booleanField2) (isTrue .booleanField2.value)}}
  boolean_field_2 = true
{{- end}}
{{- if and (fieldSet .booleanField3) (isTrue .booleanField3.value)}}
  boolean_field_3 = true
{{- end}}
{{- if and (fieldSet .booleanField4) (isTrue .booleanField4.value)}}
  boolean_field_4 = true
{{- end}}
{{- if and (fieldSet .booleanField5) (isTrue .booleanField5.value)}}
  boolean_field_5 = true
{{- end}}
{{- if and (fieldSet .booleanField6) (isTrue .booleanField6.value)}}
  boolean_field_6 = true
{{- end}}
{{- if and (fieldSet .booleanField7) (isTrue .booleanField7.value)}}
  boolean_field_7 = true
{{- end}}
{{- if fieldSet .credentialField1}}
  credential_field_1 = "{{quote .credentialField1.value}}"
{{- end}}
{{- if fieldSet .credentialField2}}
  credential_field_2 = "{{quote .credentialField2.value}}"
{{- end}}
{{- if fieldSet .credentialField3}}
  credential_field_3 = "{{quote .credentialField3.value}}"
{{- end}}
{{- if fieldSet .credentialField4}}
  credential_field_4 = "{{quote .credentialField4.value}}"
{{- end}}
{{- if fieldSet .credentialVarField1}}
  credential_var_field_1 = "{{quote .credentialVarField1.value}}"
{{- end}}
{{- if fieldSet .credentialVarField4}}
  credential_var_field_4 = "{{quote .credentialVarField4.value}}"
{{- end}}
{{- if fieldSet .customField1}}
  custom_field_1 = "{{quote .customField1.value}}"
{{- end}}
{{- if fieldSet .customField2}}
  custom_field_2 = "{{quote .customField2.value}}"
{{- end}}
{{- if fieldSet .intField1}}
  int_field_1 = {{.intField1.value}}
{{- end}}
{{- if fieldSet .largeTextField1}}
  large_text_field_1 = "{{quote .largeTextField1.value}}"
{{- end}}
{{- if fieldSet .largeTextField2}}
  large_text_field_2 = "{{quote .largeTextField2.value}}"
{{- end}}
{{- if fieldSet .largeTextField3}}
  large_text_field_3 = "{{quote .largeTextField3.value}}"
{{- end}}
{{- if fieldSet .largeTextField4}}
  large_text_field_4 = "{{quote .largeTextField4.value}}"
{{- end}}
{{- if fieldSet .scriptField1}}
  script_field_1 = "{{quote .scriptField1.value}}"
{{- end}}
{{- if fieldSet .scriptField2}}
  script_field_2 = "{{quote .scriptField2.value}}"
{{- end}}
{{- if fieldSet .scriptVarField2}}
  script_var_field_2 = "{{quote .scriptVarField2.value}}"
{{- end}}
{{template "task_wait_delay_block" .}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if hasActions .actions}}
{{template "actions_block" .actions}}
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

// ============================================================================
// Trigger Resources
// ============================================================================

const triggerTimeTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .tasks}}
  tasks = [{{stringList .tasks}}]
{{- end}}
{{- if notEmpty .time}}
  time = "{{quote .time}}"
{{- end}}
{{- if notEmpty .timeZone}}
  time_zone = "{{quote .timeZone}}"
{{- end}}
{{- if notEmpty .timeStyle}}
  time_style = "{{quote .timeStyle}}"
{{- end}}
{{- if notEmpty .timeInterval}}
  time_interval = {{.timeInterval}}
{{- end}}
{{- if notEmpty .timeIntervalUnits}}
  time_interval_units = "{{quote .timeIntervalUnits}}"
{{- end}}
{{- if notEmpty .dayStyle}}
  day_style = "{{quote .dayStyle}}"
{{- end}}
{{- if notEmpty .calendar}}
  calendar = "{{quote .calendar}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const triggerCronTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .tasks}}
  tasks = [{{stringList .tasks}}]
{{- end}}
{{- if notEmpty .minutes}}
  minutes = "{{quote .minutes}}"
{{- end}}
{{- if notEmpty .hours}}
  hours = "{{quote .hours}}"
{{- end}}
{{- if notEmpty .dayOfMonth}}
  day_of_month = "{{quote .dayOfMonth}}"
{{- end}}
{{- if notEmpty .month}}
  month = "{{quote .month}}"
{{- end}}
{{- if notEmpty .dayOfWeek}}
  day_of_week = "{{quote .dayOfWeek}}"
{{- end}}
{{- if notEmpty .dayLogic}}
  day_logic = "{{quote .dayLogic}}"
{{- end}}
{{- if notEmpty .timeZone}}
  time_zone = "{{quote .timeZone}}"
{{- end}}
{{- if notEmpty .calendar}}
  calendar = "{{quote .calendar}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const triggerFileMonitorTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .tasks}}
  tasks = [{{stringList .tasks}}]
{{- end}}
{{- if notEmpty .taskMonitor}}
  task_monitor = "{{quote .taskMonitor}}"
{{- end}}
{{- if notEmpty .timeZone}}
  time_zone = "{{quote .timeZone}}"
{{- end}}
{{- if notEmpty .calendar}}
  calendar = "{{quote .calendar}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

const triggerTaskMonitorTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .tasks}}
  tasks = [{{stringList .tasks}}]
{{- end}}
{{- if notEmpty .taskMonitor}}
  task_monitor = "{{quote .taskMonitor}}"
{{- end}}
{{- if notEmpty .timeZone}}
  time_zone = "{{quote .timeZone}}"
{{- end}}
{{- if notEmpty .calendar}}
  calendar = "{{quote .calendar}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
{{- if notEmpty .variables}}
{{template "task_variables_block" .variables}}
{{- end}}
}
`

// ============================================================================
// Other Resources
// ============================================================================

const customDayTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .comments}}
  comments = "{{quote .comments}}"
{{- end}}
  ctype = "{{quote .ctype}}"
{{- /* The provider's ValidateConfig rejects date/date_list/month/day/
       dayofweek/relfreq/nth_amount/nth_type combinations that don't match
       the selected ctype, so each field below is gated on ctype (not just
       notEmpty) even though the raw UAC export often populates all of them
       with leftover/default values regardless of ctype. */}}
{{- if and (notEmpty .date) (eq .ctype "Single Date")}}
  date = "{{quote .date}}"
{{- end}}
{{- if and (notEmpty .dateList) (eq .ctype "List of Dates")}}
  date_list = [{{stringList .dateList}}]
{{- end}}
{{- if and (notEmpty .month) (or (eq .ctype "Absolute Repeating Date") (eq .ctype "Relative Repeating Date"))}}
  month = "{{quote .month}}"
{{- end}}
{{- if and (notEmpty .day) (eq .ctype "Absolute Repeating Date")}}
  day = {{.day}}
{{- end}}
{{- if and (notEmpty .dayofweek) (eq .ctype "Relative Repeating Date")}}
  dayofweek = "{{quote .dayofweek}}"
{{- end}}
{{- if and (notEmpty .relfreq) (eq .ctype "Relative Repeating Date")}}
  relfreq = "{{quote .relfreq}}"
{{- end}}
{{- if and (notEmpty .nthAmount) (eq .ctype "Relative Repeating Date")}}
  nth_amount = {{.nthAmount}}
{{- end}}
{{- if and (notEmpty .nthType) (eq .ctype "Relative Repeating Date")}}
  nth_type = "{{quote .nthType}}"
{{- end}}
{{- if notEmpty .adjustment}}
  adjustment = "{{quote .adjustment}}"
{{- end}}
{{- if notEmpty .adjustmentAmount}}
  adjustment_amount = {{.adjustmentAmount}}
{{- end}}
{{- if notEmpty .adjustmentType}}
  adjustment_type = "{{quote .adjustmentType}}"
{{- end}}
{{- if isTrue .holiday}}
  holiday = true
{{- end}}
{{- if isTrue .period}}
  period = true
{{- end}}
{{- if notEmpty .observedRules}}
  observed_rules = [
{{- range .observedRules}}
    { actual_day_of_week = "{{quote .actualDayOfWeek}}", observed_day_of_week = "{{quote .observedDayOfWeek}}" },
{{- end}}
  ]
{{- end}}
}
`

// emailTemplateTemplate note: the resource's "name" attribute maps to the
// JSON field "templateName" (not "name"), and its "email_connection"
// attribute maps to JSON field "connection" (not "emailConnection").
const emailTemplateTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .templateName}}"
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
  email_connection = "{{quote .connection}}"
{{- if notEmpty .replyTo}}
  reply_to = "{{quote .replyTo}}"
{{- end}}
{{- if notEmpty .to}}
  to = "{{quote .to}}"
{{- end}}
{{- if notEmpty .cc}}
  cc = "{{quote .cc}}"
{{- end}}
{{- if notEmpty .bcc}}
  bcc = "{{quote .bcc}}"
{{- end}}
{{- if notEmpty .subject}}
  subject = "{{quote .subject}}"
{{- end}}
{{- if notEmpty .body}}
  body = "{{quote .body}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

// universalTemplateTemplate handles stonebranch_universal_template. This is
// the most complex template: the top-level "fields" list-nested attribute
// contains, per field, an optional "choices" list-nested attribute and an
// optional "array_field_value" list-nested attribute. List-of-objects HCL2
// literals tolerate trailing commas, so each nested range loop simply
// appends a trailing comma after every element instead of using an
// index-based separator.
const universalTemplateTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .description}}
  description = "{{quote .description}}"
{{- end}}
{{- if notEmpty .extension}}
  extension = "{{quote .extension}}"
{{- end}}
  variable_prefix = "{{quote .variablePrefix}}"
{{- if notEmpty .templateType}}
  template_type = "{{quote .templateType}}"
{{- end}}
  agent_type = "{{quote .agentType}}"
{{- if isTrue .useCommonScript}}
  use_common_script = true
{{- end}}
{{- if notEmpty .script}}
  script = "{{quote .script}}"
{{- end}}
{{- if notEmpty .scriptUnix}}
  script_unix = "{{quote .scriptUnix}}"
{{- end}}
{{- if notEmpty .scriptWindows}}
  script_windows = "{{quote .scriptWindows}}"
{{- end}}
{{- if notEmpty .scriptTypeWindows}}
  script_type_windows = "{{quote .scriptTypeWindows}}"
{{- end}}
{{- if isTrue .alwaysCancelOnFinish}}
  always_cancel_on_finish = true
{{- end}}
{{- if notEmpty .credentials}}
  credentials = "{{quote .credentials}}"
{{- end}}
{{- if notEmpty .credentialsVar}}
  credentials_var = "{{quote .credentialsVar}}"
{{- end}}
{{- if notEmpty .agent}}
  agent = "{{quote .agent}}"
{{- end}}
{{- if notEmpty .agentVar}}
  agent_var = "{{quote .agentVar}}"
{{- end}}
{{- if notEmpty .agentCluster}}
  agent_cluster = "{{quote .agentCluster}}"
{{- end}}
{{- if notEmpty .agentClusterVar}}
  agent_cluster_var = "{{quote .agentClusterVar}}"
{{- end}}
{{- if notEmpty .broadcastCluster}}
  broadcast_cluster = "{{quote .broadcastCluster}}"
{{- end}}
{{- if notEmpty .broadcastClusterVar}}
  broadcast_cluster_var = "{{quote .broadcastClusterVar}}"
{{- end}}
{{- if notEmpty .runtimeDir}}
  runtime_dir = "{{quote .runtimeDir}}"
{{- end}}
{{- if notEmpty .environment}}
  environment = [
{{- range .environment}}
    { name = "{{quote .name}}", value = "{{quote .value}}" },
{{- end}}
  ]
{{- end}}
{{- if notEmpty .sendEnvironment}}
  send_environment = "{{quote .sendEnvironment}}"
{{- end}}
{{- if notEmpty .sendVariables}}
  send_variables = "{{quote .sendVariables}}"
{{- end}}
  exit_codes = "{{quote .exitCodes}}"
{{- if notEmpty .exitCodeProcessing}}
  exit_code_processing = "{{quote .exitCodeProcessing}}"
{{- end}}
{{- if notEmpty .exitCodeText}}
  exit_code_text = "{{quote .exitCodeText}}"
{{- end}}
{{- if notEmpty .exitCodeOutput}}
  exit_code_output = "{{quote .exitCodeOutput}}"
{{- end}}
{{- if notEmpty .outputType}}
  output_type = "{{quote .outputType}}"
{{- end}}
{{- if notEmpty .outputContentType}}
  output_content_type = "{{quote .outputContentType}}"
{{- end}}
{{- if notEmpty .outputPathExpression}}
  output_path_expression = "{{quote .outputPathExpression}}"
{{- end}}
{{- if notEmpty .outputConditionOperator}}
  output_condition_operator = "{{quote .outputConditionOperator}}"
{{- end}}
{{- if notEmpty .outputConditionValue}}
  output_condition_value = "{{quote .outputConditionValue}}"
{{- end}}
{{- if notEmpty .outputConditionStrategy}}
  output_condition_strategy = "{{quote .outputConditionStrategy}}"
{{- end}}
{{- if isTrue .autoCleanup}}
  auto_cleanup = true
{{- end}}
{{- if notEmpty .outputReturnType}}
  output_return_type = "{{quote .outputReturnType}}"
{{- end}}
{{- if notEmpty .outputReturnFile}}
  output_return_file = "{{quote .outputReturnFile}}"
{{- end}}
{{- if notEmpty .outputReturnSline}}
  output_return_sline = "{{quote .outputReturnSline}}"
{{- end}}
{{- if notEmpty .outputReturnNline}}
  output_return_nline = "{{quote .outputReturnNline}}"
{{- end}}
{{- if notEmpty .outputReturnText}}
  output_return_text = "{{quote .outputReturnText}}"
{{- end}}
{{- if isTrue .waitForOutput}}
  wait_for_output = true
{{- end}}
{{- if isTrue .outputFailureOnly}}
  output_failure_only = true
{{- end}}
{{- if isTrue .elevateUser}}
  elevate_user = true
{{- end}}
{{- if isTrue .desktopInteract}}
  desktop_interact = true
{{- end}}
{{- if isTrue .createConsole}}
  create_console = true
{{- end}}
{{- if notEmpty .logLevel}}
  log_level = "{{quote .logLevel}}"
{{- end}}
{{- if notEmpty .fields}}
  fields = [
{{- range $f := .fields}}
    {
      name = "{{quote $f.name}}"
      label = "{{quote $f.label}}"
      field_mapping = "{{quote $f.fieldMapping}}"
{{- if notEmpty $f.fieldType}}
      field_type = "{{quote $f.fieldType}}"
{{- end}}
{{- if notEmpty $f.hint}}
      hint = "{{quote $f.hint}}"
{{- end}}
{{- if notEmpty $f.textType}}
      text_type = "{{quote $f.textType}}"
{{- end}}
{{- if notEmpty $f.fieldValue}}
      field_value = "{{quote $f.fieldValue}}"
{{- end}}
{{- if isTrue $f.defaultListView}}
      default_list_view = true
{{- end}}
{{- if isTrue $f.allowVariable}}
      allow_variable = true
{{- end}}
{{- if notEmpty $f.fieldRestriction}}
      field_restriction = "{{quote $f.fieldRestriction}}"
{{- end}}
{{- if isTrue $f.preserveOutputOnRerun}}
      preserve_output_on_rerun = true
{{- end}}
{{- if isTrue $f.extensionStatus}}
      extension_status = true
{{- end}}
{{- if notEmpty $f.booleanValueType}}
      boolean_value_type = "{{quote $f.booleanValueType}}"
{{- end}}
{{- if notEmpty $f.booleanYesValue}}
      boolean_yes_value = "{{quote $f.booleanYesValue}}"
{{- end}}
{{- if notEmpty $f.booleanNoValue}}
      boolean_no_value = "{{quote $f.booleanNoValue}}"
{{- end}}
{{- if notEmpty $f.choiceSortOption}}
      choice_sort_option = "{{quote $f.choiceSortOption}}"
{{- end}}
{{- if isTrue $f.choiceAllowEmpty}}
      choice_allow_empty = true
{{- end}}
{{- if isTrue $f.choiceAllowMultiple}}
      choice_allow_multiple = true
{{- end}}
{{- if isTrue $f.choiceDynamic}}
      choice_dynamic = true
{{- end}}
{{- if notEmpty $f.choiceFields}}
      choice_fields = [{{stringList $f.choiceFields}}]
{{- end}}
{{- if notEmpty $f.choices}}
      choices = [
{{- range $c := $f.choices}}
        {
          field_value = "{{quote $c.fieldValue}}"
{{- if notEmpty $c.fieldValueLabel}}
          field_value_label = "{{quote $c.fieldValueLabel}}"
{{- end}}
{{- if isTrue $c.useFieldValueForLabel}}
          use_field_value_for_label = true
{{- end}}
        },
{{- end}}
      ]
{{- end}}
{{- if isTrue $f.required}}
      required = true
{{- end}}
{{- if notEmpty $f.requireIfField}}
      require_if_field = "{{quote $f.requireIfField}}"
{{- end}}
{{- if notEmpty $f.requireIfFieldValue}}
      require_if_field_value = "{{quote $f.requireIfFieldValue}}"
{{- end}}
{{- if notEmpty $f.showIfField}}
      show_if_field = "{{quote $f.showIfField}}"
{{- end}}
{{- if notEmpty $f.showIfFieldValue}}
      show_if_field_value = "{{quote $f.showIfFieldValue}}"
{{- end}}
{{- if isTrue $f.requireIfVisible}}
      require_if_visible = true
{{- end}}
{{- if isTrue $f.preserveValueIfHidden}}
      preserve_value_if_hidden = true
{{- end}}
{{- if isTrue $f.noSpaceIfHidden}}
      no_space_if_hidden = true
{{- end}}
{{- if notEmpty $f.fieldLength}}
      field_length = {{$f.fieldLength}}
{{- end}}
{{- if notEmpty $f.intFieldMin}}
      int_field_min = {{$f.intFieldMin}}
{{- end}}
{{- if notEmpty $f.intFieldMax}}
      int_field_max = {{$f.intFieldMax}}
{{- end}}
{{- if notEmpty $f.fieldRegex}}
      field_regex = "{{quote $f.fieldRegex}}"
{{- end}}
{{- if notEmpty $f.formColumnSpan}}
      form_column_span = {{$f.formColumnSpan}}
{{- end}}
{{- if isTrue $f.formStartRow}}
      form_start_row = true
{{- end}}
{{- if isTrue $f.formEndRow}}
      form_end_row = true
{{- end}}
{{- if notEmpty $f.arrayNameTitle}}
      array_name_title = "{{quote $f.arrayNameTitle}}"
{{- end}}
{{- if notEmpty $f.arrayValueTitle}}
      array_value_title = "{{quote $f.arrayValueTitle}}"
{{- end}}
{{- if notEmpty $f.arrayFieldValue}}
      array_field_value = [
{{- range $a := $f.arrayFieldValue}}
        { name = "{{quote $a.name}}", value = "{{quote $a.value}}" },
{{- end}}
      ]
{{- end}}
    },
{{- end}}
  ]
{{- end}}
}
`

const virtualResourceTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  name = "{{quote .name}}"
{{- if notEmpty .limit}}
  limit = {{.limit}}
{{- end}}
{{- if notEmpty .summary}}
  summary = "{{quote .summary}}"
{{- end}}
{{- if notEmpty .type}}
  type = "{{quote .type}}"
{{- end}}
{{- if notEmpty .opswiseGroups}}
  opswise_groups = [{{stringList .opswiseGroups}}]
{{- end}}
}
`

// ============================================================================
// Workflow Components
// ============================================================================

const workflowVertexTemplate = `# Generated by sb2tf
resource "{{._terraformResource}}" "{{._resourceName}}" {
  workflow_name = "{{quote .workflowName}}"
  task_name     = "{{quote .taskName}}"
{{- if notEmpty .alias}}
  alias = "{{quote .alias}}"
{{- end}}
}
`

const workflowEdgeTemplate = `# Generated by sb2tf
# Edge: {{.sourceTask}} -> {{.targetTask}}
resource "{{._terraformResource}}" "{{._resourceName}}" {
  workflow_name = "{{quote .workflowName}}"
  source_id     = stonebranch_workflow_vertex.{{.sourceVertexTfName}}.vertex_id
  target_id     = stonebranch_workflow_vertex.{{.targetVertexTfName}}.vertex_id
}
`
