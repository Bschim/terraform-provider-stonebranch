// Package generator provides HCL code generation for StoneBranch resources.
package generator

// ResourceType defines metadata about a StoneBranch resource type.
type ResourceType struct {
	// CLIName is the name used in CLI commands (e.g., "task_unix")
	CLIName string

	// TerraformResource is the full Terraform resource name (e.g., "stonebranch_task_unix")
	TerraformResource string

	// APIEndpoint is the base API endpoint for CRUD operations
	APIEndpoint string

	// ListEndpoint is the API endpoint for listing resources
	ListEndpoint string

	// APITypeValue is the "type" field value in API responses (e.g., "taskUnix")
	// Used for local filtering since the API may not support type query params
	APITypeValue string

	// NameQueryParam is the query parameter name for filtering by resource name
	NameQueryParam string

	// NameField is the JSON field name containing the resource name in API responses
	// Defaults to "name" if not specified
	NameField string

	// HasTypeField indicates if this resource type has a "type" field in API responses
	HasTypeField bool

	// Category is the resource category (tasks, triggers, connections, etc.)
	Category string

	// Dependencies lists the fields that reference other resources
	Dependencies []Dependency
}

// Dependency describes a field that references another resource.
type Dependency struct {
	// Field is the JSON field name in the API response
	Field string

	// ResourceType is the CLI name of the referenced resource type
	ResourceType string

	// Condition is an optional condition (e.g., "command_or_script=Script")
	Condition string
}

// ResourceCategory groups resource types for display.
type ResourceCategory struct {
	Name  string
	Types []*ResourceType
}

// resourceTypes is the registry of all supported resource types.
var resourceTypes = map[string]*ResourceType{
	// Tasks
	"task_unix": {
		CLIName:           "task_unix",
		TerraformResource: "stonebranch_task_unix",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskUnix",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
			{Field: "script", ResourceType: "script", Condition: "commandOrScript=Script"},
			{Field: "agentCluster", ResourceType: "agent_cluster"},
		},
	},
	"task_windows": {
		CLIName:           "task_windows",
		TerraformResource: "stonebranch_task_windows",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskWindows",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
			{Field: "script", ResourceType: "script", Condition: "commandOrScript=Script"},
			{Field: "agentCluster", ResourceType: "agent_cluster"},
		},
	},
	"task_sql": {
		CLIName:           "task_sql",
		TerraformResource: "stonebranch_task_sql",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskSql",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "databaseConnection", ResourceType: "database_connection"},
		},
	},
	"task_email": {
		CLIName:           "task_email",
		TerraformResource: "stonebranch_task_email",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskEmail",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "connection", ResourceType: "email_connection"},
		},
	},
	"task_workflow": {
		CLIName:           "task_workflow",
		TerraformResource: "stonebranch_task_workflow",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskWorkflow",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		// Workflow dependencies are handled specially (vertices and edges)
	},
	"task_file_monitor": {
		CLIName:           "task_file_monitor",
		TerraformResource: "stonebranch_task_file_monitor",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskFileMonitor",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
			{Field: "agentCluster", ResourceType: "agent_cluster"},
		},
	},
	"task_file_transfer": {
		CLIName:           "task_file_transfer",
		TerraformResource: "stonebranch_task_file_transfer",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskFtp",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
			{Field: "agentCluster", ResourceType: "agent_cluster"},
		},
	},
	"task_timer": {
		CLIName:           "task_timer",
		TerraformResource: "stonebranch_task_timer",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskSleep",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
	},
	"task_recurring": {
		CLIName:           "task_recurring",
		TerraformResource: "stonebranch_task_recurring",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskRecurring",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
	},
	"task_monitor": {
		CLIName:           "task_monitor",
		TerraformResource: "stonebranch_task_monitor",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskMonitor",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
	},
	"task_stored_procedure": {
		CLIName:           "task_stored_procedure",
		TerraformResource: "stonebranch_task_stored_procedure",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskStoredProc",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "databaseConnection", ResourceType: "database_connection"},
		},
	},
	"task_web_service": {
		CLIName:           "task_web_service",
		TerraformResource: "stonebranch_task_web_service",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskWebService",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
		},
	},
	"task_universal_aws_s3": {
		CLIName:           "task_universal_aws_s3",
		TerraformResource: "stonebranch_task_universal_aws_s3",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskUniversal",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
			{Field: "agentCluster", ResourceType: "agent_cluster"},
		},
	},
	"task_universal": {
		CLIName:           "task_universal",
		TerraformResource: "stonebranch_task_universal",
		APIEndpoint:       "/resources/task",
		ListEndpoint:      "/resources/task/listadv",
		APITypeValue:      "taskUniversal",
		NameQueryParam:    "taskname",
		HasTypeField:      true,
		Category:          "Tasks",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
			{Field: "agentCluster", ResourceType: "agent_cluster"},
		},
	},

	// Triggers
	"trigger_time": {
		CLIName:           "trigger_time",
		TerraformResource: "stonebranch_trigger_time",
		APIEndpoint:       "/resources/trigger",
		ListEndpoint:      "/resources/trigger/listadv",
		APITypeValue:      "triggerTime",
		NameQueryParam:    "triggername",
		HasTypeField:      true,
		Category:          "Triggers",
		Dependencies: []Dependency{
			{Field: "calendar", ResourceType: "calendar"},
		},
	},
	"trigger_cron": {
		CLIName:           "trigger_cron",
		TerraformResource: "stonebranch_trigger_cron",
		APIEndpoint:       "/resources/trigger",
		ListEndpoint:      "/resources/trigger/listadv",
		APITypeValue:      "triggerCron",
		NameQueryParam:    "triggername",
		HasTypeField:      true,
		Category:          "Triggers",
		Dependencies: []Dependency{
			{Field: "calendar", ResourceType: "calendar"},
		},
	},
	"trigger_file_monitor": {
		CLIName:           "trigger_file_monitor",
		TerraformResource: "stonebranch_trigger_file_monitor",
		APIEndpoint:       "/resources/trigger",
		ListEndpoint:      "/resources/trigger/listadv",
		APITypeValue:      "triggerFm",
		NameQueryParam:    "triggername",
		HasTypeField:      true,
		Category:          "Triggers",
		Dependencies: []Dependency{
			{Field: "taskMonitor", ResourceType: "task_file_monitor"},
			{Field: "calendar", ResourceType: "calendar"},
		},
	},
	"trigger_task_monitor": {
		CLIName:           "trigger_task_monitor",
		TerraformResource: "stonebranch_trigger_task_monitor",
		APIEndpoint:       "/resources/trigger",
		ListEndpoint:      "/resources/trigger/listadv",
		APITypeValue:      "triggerTm",
		NameQueryParam:    "triggername",
		HasTypeField:      true,
		Category:          "Triggers",
		Dependencies: []Dependency{
			{Field: "taskMonitor", ResourceType: "task_monitor"},
			{Field: "calendar", ResourceType: "calendar"},
		},
	},

	// Connections
	"database_connection": {
		CLIName:           "database_connection",
		TerraformResource: "stonebranch_database_connection",
		APIEndpoint:       "/resources/databaseconnection",
		ListEndpoint:      "/resources/databaseconnection/list",
		NameQueryParam:    "connectionname",
		HasTypeField:      false,
		Category:          "Connections",
		Dependencies: []Dependency{
			{Field: "credentials", ResourceType: "credential"},
		},
	},
	"email_connection": {
		CLIName:           "email_connection",
		TerraformResource: "stonebranch_email_connection",
		APIEndpoint:       "/resources/emailconnection",
		ListEndpoint:      "/resources/emailconnection/list",
		NameQueryParam:    "connectionname",
		HasTypeField:      false,
		Category:          "Connections",
	},

	// Other
	"script": {
		CLIName:           "script",
		TerraformResource: "stonebranch_script",
		APIEndpoint:       "/resources/script",
		ListEndpoint:      "/resources/script/list",
		NameQueryParam:    "scriptname",
		NameField:         "scriptName",
		HasTypeField:      false,
		Category:          "Other",
	},
	"variable": {
		CLIName:           "variable",
		TerraformResource: "stonebranch_variable",
		APIEndpoint:       "/resources/variable",
		ListEndpoint:      "/resources/variable/listadv",
		NameQueryParam:    "variablename",
		HasTypeField:      false,
		Category:          "Other",
	},
	"credential": {
		CLIName:           "credential",
		TerraformResource: "stonebranch_credential",
		APIEndpoint:       "/resources/credential",
		ListEndpoint:      "/resources/credential/list",
		NameQueryParam:    "credentialname",
		HasTypeField:      false,
		Category:          "Other",
	},
	"business_service": {
		CLIName:           "business_service",
		TerraformResource: "stonebranch_business_service",
		APIEndpoint:       "/resources/businessservice",
		ListEndpoint:      "/resources/businessservice/list",
		NameQueryParam:    "busservicename",
		HasTypeField:      false,
		Category:          "Other",
	},
	"agent_cluster": {
		CLIName:           "agent_cluster",
		TerraformResource: "stonebranch_agent_cluster",
		APIEndpoint:       "/resources/agentcluster",
		ListEndpoint:      "/resources/agentcluster/list",
		NameQueryParam:    "agentclustername",
		HasTypeField:      false,
		Category:          "Other",
	},
	"calendar": {
		CLIName:           "calendar",
		TerraformResource: "stonebranch_calendar",
		APIEndpoint:       "/resources/calendar",
		ListEndpoint:      "/resources/calendar/list",
		NameQueryParam:    "calendarname",
		HasTypeField:      false,
		Category:          "Other",
	},
	"customday": {
		CLIName:           "customday",
		TerraformResource: "stonebranch_custom_day",
		APIEndpoint:       "/resources/customday",
		ListEndpoint:      "/resources/customday/list",
		NameQueryParam:    "customdayname",
		NameField:         "name",
		HasTypeField:      false,
		Category:          "Other",
	},
	"emailtemplate": {
		CLIName:           "emailtemplate",
		TerraformResource: "stonebranch_email_template",
		APIEndpoint:       "/resources/emailtemplate",
		ListEndpoint:      "/resources/emailtemplate/list",
		NameQueryParam:    "templatename",
		NameField:         "templateName",
		HasTypeField:      false,
		Category:          "Other",
	},
	"universaltemplate": {
		CLIName:           "universaltemplate",
		TerraformResource: "stonebranch_universal_template",
		APIEndpoint:       "/resources/universaltemplate",
		ListEndpoint:      "/resources/universaltemplate/list",
		NameQueryParam:    "templatename",
		NameField:         "name",
		HasTypeField:      false,
		Category:          "Other",
	},
	"virtualresource": {
		CLIName:           "virtualresource",
		TerraformResource: "stonebranch_virtual_resource",
		APIEndpoint:       "/resources/virtual",
		ListEndpoint:      "/resources/virtual/list",
		NameQueryParam:    "resourcename",
		NameField:         "name",
		HasTypeField:      false,
		Category:          "Other",
	},

	// Workflow components (special handling)
	"workflow_vertex": {
		CLIName:           "workflow_vertex",
		TerraformResource: "stonebranch_workflow_vertex",
		APIEndpoint:       "/resources/workflow/vertices",
		HasTypeField:      false,
		Category:          "Workflow",
	},
	"workflow_edge": {
		CLIName:           "workflow_edge",
		TerraformResource: "stonebranch_workflow_edge",
		APIEndpoint:       "/resources/workflow/edges",
		HasTypeField:      false,
		Category:          "Workflow",
	},
}

// GetResourceType returns the resource type definition for the given CLI name.
func GetResourceType(cliName string) *ResourceType {
	return resourceTypes[cliName]
}

// GetAllResourceTypes returns all registered resource types.
func GetAllResourceTypes() map[string]*ResourceType {
	return resourceTypes
}

// GetResourceCategories returns resource types grouped by category.
func GetResourceCategories() []ResourceCategory {
	categories := []ResourceCategory{
		{Name: "Tasks", Types: make([]*ResourceType, 0)},
		{Name: "Triggers", Types: make([]*ResourceType, 0)},
		{Name: "Connections", Types: make([]*ResourceType, 0)},
		{Name: "Other", Types: make([]*ResourceType, 0)},
		{Name: "Workflow", Types: make([]*ResourceType, 0)},
	}

	categoryMap := make(map[string]*ResourceCategory)
	for i := range categories {
		categoryMap[categories[i].Name] = &categories[i]
	}

	for _, rt := range resourceTypes {
		if cat, ok := categoryMap[rt.Category]; ok {
			cat.Types = append(cat.Types, rt)
		}
	}

	return categories
}

// APITypeToResourceType maps API type values to CLI resource type names.
var APITypeToResourceType = map[string]string{
	"taskUnix":        "task_unix",
	"taskWindows":     "task_windows",
	"taskSql":         "task_sql",
	"taskEmail":       "task_email",
	"taskWorkflow":    "task_workflow",
	"taskFileMonitor": "task_file_monitor",
	"taskFtp":         "task_file_transfer",
	"taskSleep":       "task_timer",
	"taskMonitor":     "task_monitor",
	"taskStoredProc":  "task_stored_procedure",
	"taskWebService":  "task_web_service",
	"taskUniversal":   "task_universal",
	"taskRecurring":   "task_recurring",
	"triggerTime":     "trigger_time",
	"triggerCron":     "trigger_cron",
	"triggerFm":       "trigger_file_monitor",
	"triggerTm":       "trigger_task_monitor",
}
