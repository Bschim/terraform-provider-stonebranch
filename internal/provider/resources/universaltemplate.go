package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &UniversalTemplateResource{}
	_ resource.ResourceWithImportState    = &UniversalTemplateResource{}
	_ resource.ResourceWithValidateConfig = &UniversalTemplateResource{}
)

func NewUniversalTemplateResource() resource.Resource {
	return &UniversalTemplateResource{}
}

// UniversalTemplateResource defines the resource implementation.
type UniversalTemplateResource struct {
	client *client.Client
}

// UniversalTemplateEnvVarModel represents a single environment variable (name/value pair).
type UniversalTemplateEnvVarModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// UniversalTemplateEnvVarAPIModel is the API wire format for an environment variable.
type UniversalTemplateEnvVarAPIModel struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

func universalTemplateEnvVarAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}
}

// UniversalTemplateFieldChoiceModel represents a single choice option for a
// Choice-type custom UI field.
type UniversalTemplateFieldChoiceModel struct {
	FieldValue            types.String `tfsdk:"field_value"`
	FieldValueLabel       types.String `tfsdk:"field_value_label"`
	UseFieldValueForLabel types.Bool   `tfsdk:"use_field_value_for_label"`
}

// UniversalTemplateFieldChoiceAPIModel is the API wire format for a field choice.
type UniversalTemplateFieldChoiceAPIModel struct {
	FieldValue            string `json:"fieldValue,omitempty"`
	FieldValueLabel       string `json:"fieldValueLabel,omitempty"`
	UseFieldValueForLabel bool   `json:"useFieldValueForLabel,omitempty"`
}

func universalTemplateFieldChoiceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"field_value":               types.StringType,
		"field_value_label":         types.StringType,
		"use_field_value_for_label": types.BoolType,
	}
}

// UniversalTemplateFieldModel represents a single custom UI form field
// definition (the "fields" array of UniversalTemplateWsData). Each field maps
// to one of a fixed set of typed extension slots on the underlying template
// (see the `field_mapping` attribute) and its behavior is governed by
// `field_type` plus a number of type-specific attributes below.
type UniversalTemplateFieldModel struct {
	Name         types.String `tfsdk:"name"`
	Label        types.String `tfsdk:"label"`
	FieldMapping types.String `tfsdk:"field_mapping"`
	FieldType    types.String `tfsdk:"field_type"`
	Hint         types.String `tfsdk:"hint"`

	// Text-type fields
	TextType types.String `tfsdk:"text_type"`

	// Generic default value (used by most field types)
	FieldValue types.String `tfsdk:"field_value"`

	DefaultListView       types.Bool   `tfsdk:"default_list_view"`
	AllowVariable         types.Bool   `tfsdk:"allow_variable"`
	FieldRestriction      types.String `tfsdk:"field_restriction"`
	PreserveOutputOnRerun types.Bool   `tfsdk:"preserve_output_on_rerun"`
	ExtensionStatus       types.Bool   `tfsdk:"extension_status"`

	// Boolean-type fields
	BooleanValueType types.String `tfsdk:"boolean_value_type"`
	BooleanYesValue  types.String `tfsdk:"boolean_yes_value"`
	BooleanNoValue   types.String `tfsdk:"boolean_no_value"`

	// Choice-type fields
	ChoiceSortOption    types.String `tfsdk:"choice_sort_option"`
	ChoiceAllowEmpty    types.Bool   `tfsdk:"choice_allow_empty"`
	ChoiceAllowMultiple types.Bool   `tfsdk:"choice_allow_multiple"`
	ChoiceDynamic       types.Bool   `tfsdk:"choice_dynamic"`
	ChoiceFields        types.List   `tfsdk:"choice_fields"`
	Choices             types.List   `tfsdk:"choices"`

	// Conditional visibility / requirement
	Required              types.Bool   `tfsdk:"required"`
	RequireIfField        types.String `tfsdk:"require_if_field"`
	RequireIfFieldValue   types.String `tfsdk:"require_if_field_value"`
	ShowIfField           types.String `tfsdk:"show_if_field"`
	ShowIfFieldValue      types.String `tfsdk:"show_if_field_value"`
	RequireIfVisible      types.Bool   `tfsdk:"require_if_visible"`
	PreserveValueIfHidden types.Bool   `tfsdk:"preserve_value_if_hidden"`
	NoSpaceIfHidden       types.Bool   `tfsdk:"no_space_if_hidden"`

	// Text/Integer-type field constraints
	FieldLength types.Int64  `tfsdk:"field_length"`
	IntFieldMin types.Int64  `tfsdk:"int_field_min"`
	IntFieldMax types.Int64  `tfsdk:"int_field_max"`
	FieldRegex  types.String `tfsdk:"field_regex"`

	// Form layout
	FormColumnSpan types.Int64 `tfsdk:"form_column_span"`
	FormStartRow   types.Bool  `tfsdk:"form_start_row"`
	FormEndRow     types.Bool  `tfsdk:"form_end_row"`

	// Array-type fields
	ArrayNameTitle  types.String `tfsdk:"array_name_title"`
	ArrayValueTitle types.String `tfsdk:"array_value_title"`
	ArrayFieldValue types.List   `tfsdk:"array_field_value"`
}

// UniversalTemplateFieldAPIModel is the API wire format for a custom UI field.
type UniversalTemplateFieldAPIModel struct {
	Name         string `json:"name"`
	Label        string `json:"label"`
	FieldMapping string `json:"fieldMapping"`
	FieldType    string `json:"fieldType,omitempty"`
	Hint         string `json:"hint,omitempty"`

	TextType string `json:"textType,omitempty"`

	FieldValue string `json:"fieldValue,omitempty"`

	DefaultListView       bool   `json:"defaultListView,omitempty"`
	AllowVariable         bool   `json:"allowVariable,omitempty"`
	FieldRestriction      string `json:"fieldRestriction,omitempty"`
	PreserveOutputOnRerun bool   `json:"preserveOutputOnRerun,omitempty"`
	ExtensionStatus       bool   `json:"extensionStatus,omitempty"`

	BooleanValueType string `json:"booleanValueType,omitempty"`
	BooleanYesValue  string `json:"booleanYesValue,omitempty"`
	BooleanNoValue   string `json:"booleanNoValue,omitempty"`

	ChoiceSortOption    string                                 `json:"choiceSortOption,omitempty"`
	ChoiceAllowEmpty    bool                                   `json:"choiceAllowEmpty,omitempty"`
	ChoiceAllowMultiple bool                                   `json:"choiceAllowMultiple,omitempty"`
	ChoiceDynamic       bool                                   `json:"choiceDynamic,omitempty"`
	ChoiceFields        []string                               `json:"choiceFields,omitempty"`
	Choices             []UniversalTemplateFieldChoiceAPIModel `json:"choices,omitempty"`

	Required              bool   `json:"required,omitempty"`
	RequireIfField        string `json:"requireIfField,omitempty"`
	RequireIfFieldValue   string `json:"requireIfFieldValue,omitempty"`
	ShowIfField           string `json:"showIfField,omitempty"`
	ShowIfFieldValue      string `json:"showIfFieldValue,omitempty"`
	RequireIfVisible      bool   `json:"requireIfVisible,omitempty"`
	PreserveValueIfHidden bool   `json:"preserveValueIfHidden,omitempty"`
	NoSpaceIfHidden       bool   `json:"noSpaceIfHidden,omitempty"`

	FieldLength int64  `json:"fieldLength,omitempty"`
	IntFieldMin int64  `json:"intFieldMin,omitempty"`
	IntFieldMax int64  `json:"intFieldMax,omitempty"`
	FieldRegex  string `json:"fieldRegex,omitempty"`

	FormColumnSpan int64 `json:"formColumnSpan,omitempty"`
	FormStartRow   bool  `json:"formStartRow,omitempty"`
	FormEndRow     bool  `json:"formEndRow,omitempty"`

	ArrayNameTitle  string                            `json:"arrayNameTitle,omitempty"`
	ArrayValueTitle string                            `json:"arrayValueTitle,omitempty"`
	ArrayFieldValue []UniversalTemplateEnvVarAPIModel `json:"arrayFieldValue,omitempty"`
}

func universalTemplateFieldAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                     types.StringType,
		"label":                    types.StringType,
		"field_mapping":            types.StringType,
		"field_type":               types.StringType,
		"hint":                     types.StringType,
		"text_type":                types.StringType,
		"field_value":              types.StringType,
		"default_list_view":        types.BoolType,
		"allow_variable":           types.BoolType,
		"field_restriction":        types.StringType,
		"preserve_output_on_rerun": types.BoolType,
		"extension_status":         types.BoolType,
		"boolean_value_type":       types.StringType,
		"boolean_yes_value":        types.StringType,
		"boolean_no_value":         types.StringType,
		"choice_sort_option":       types.StringType,
		"choice_allow_empty":       types.BoolType,
		"choice_allow_multiple":    types.BoolType,
		"choice_dynamic":           types.BoolType,
		"choice_fields":            types.ListType{ElemType: types.StringType},
		"choices":                  types.ListType{ElemType: types.ObjectType{AttrTypes: universalTemplateFieldChoiceAttrTypes()}},
		"required":                 types.BoolType,
		"require_if_field":         types.StringType,
		"require_if_field_value":   types.StringType,
		"show_if_field":            types.StringType,
		"show_if_field_value":      types.StringType,
		"require_if_visible":       types.BoolType,
		"preserve_value_if_hidden": types.BoolType,
		"no_space_if_hidden":       types.BoolType,
		"field_length":             types.Int64Type,
		"int_field_min":            types.Int64Type,
		"int_field_max":            types.Int64Type,
		"field_regex":              types.StringType,
		"form_column_span":         types.Int64Type,
		"form_start_row":           types.BoolType,
		"form_end_row":             types.BoolType,
		"array_name_title":         types.StringType,
		"array_value_title":        types.StringType,
		"array_field_value":        types.ListType{ElemType: types.ObjectType{AttrTypes: universalTemplateEnvVarAttrTypes()}},
	}
}

// UniversalTemplateResourceModel describes the resource data model.
//
// This models a curated core subset of UniversalTemplateWsData. The
// following are intentionally deferred from v1 and are not represented here:
//   - The `commands` and `events` nested arrays (event-metric definitions
//     for the template).
//   - The five top-level `*FieldsRestriction` attributes, which only affect
//     how form fields are restricted in the UAC UI (the per-field
//     `field_restriction` attribute on each `fields` entry is modeled).
//   - Icon upload attributes (`iconFilename`/`iconFilesize`/`iconDateCreated`).
//   - Export/system bookkeeping attributes (`retainSysIds`, `minReleaseLevel`,
//     `excludeRelated`, `exportReleaseLevel`, `exportTable`), consistent with
//     every other resource in this provider.
type UniversalTemplateResourceModel struct {
	// Identity
	SysId       types.String `tfsdk:"sys_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`

	// Core
	Extension      types.String `tfsdk:"extension"`
	VariablePrefix types.String `tfsdk:"variable_prefix"`
	TemplateType   types.String `tfsdk:"template_type"`
	AgentType      types.String `tfsdk:"agent_type"`

	// Script
	UseCommonScript      types.Bool   `tfsdk:"use_common_script"`
	Script               types.String `tfsdk:"script"`
	ScriptUnix           types.String `tfsdk:"script_unix"`
	ScriptWindows        types.String `tfsdk:"script_windows"`
	ScriptTypeWindows    types.String `tfsdk:"script_type_windows"`
	AlwaysCancelOnFinish types.Bool   `tfsdk:"always_cancel_on_finish"`

	// Credentials
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

	// Agent targeting
	Agent               types.String `tfsdk:"agent"`
	AgentVar            types.String `tfsdk:"agent_var"`
	AgentCluster        types.String `tfsdk:"agent_cluster"`
	AgentClusterVar     types.String `tfsdk:"agent_cluster_var"`
	BroadcastCluster    types.String `tfsdk:"broadcast_cluster"`
	BroadcastClusterVar types.String `tfsdk:"broadcast_cluster_var"`

	// Runtime / environment / variables
	RuntimeDir      types.String `tfsdk:"runtime_dir"`
	Environment     types.List   `tfsdk:"environment"`
	SendEnvironment types.String `tfsdk:"send_environment"`
	SendVariables   types.String `tfsdk:"send_variables"`

	// Exit codes
	ExitCodes          types.String `tfsdk:"exit_codes"`
	ExitCodeProcessing types.String `tfsdk:"exit_code_processing"`
	ExitCodeText       types.String `tfsdk:"exit_code_text"`
	ExitCodeOutput     types.String `tfsdk:"exit_code_output"`

	// Output handling
	OutputType              types.String `tfsdk:"output_type"`
	OutputContentType       types.String `tfsdk:"output_content_type"`
	OutputPathExpression    types.String `tfsdk:"output_path_expression"`
	OutputConditionOperator types.String `tfsdk:"output_condition_operator"`
	OutputConditionValue    types.String `tfsdk:"output_condition_value"`
	OutputConditionStrategy types.String `tfsdk:"output_condition_strategy"`
	AutoCleanup             types.Bool   `tfsdk:"auto_cleanup"`
	OutputReturnType        types.String `tfsdk:"output_return_type"`
	OutputReturnFile        types.String `tfsdk:"output_return_file"`
	OutputReturnSline       types.String `tfsdk:"output_return_sline"`
	OutputReturnNline       types.String `tfsdk:"output_return_nline"`
	OutputReturnText        types.String `tfsdk:"output_return_text"`
	WaitForOutput           types.Bool   `tfsdk:"wait_for_output"`
	OutputFailureOnly       types.Bool   `tfsdk:"output_failure_only"`

	// Windows-specific execution options
	ElevateUser     types.Bool `tfsdk:"elevate_user"`
	DesktopInteract types.Bool `tfsdk:"desktop_interact"`
	CreateConsole   types.Bool `tfsdk:"create_console"`

	// Logging
	LogLevel types.String `tfsdk:"log_level"`

	// Custom UI form fields
	Fields types.List `tfsdk:"fields"`
}

// UniversalTemplateAPIModel represents the API request/response structure.
type UniversalTemplateAPIModel struct {
	SysId       string `json:"sysId,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Extension      string `json:"extension,omitempty"`
	VariablePrefix string `json:"variablePrefix"`
	TemplateType   string `json:"templateType,omitempty"`
	AgentType      string `json:"agentType"`

	UseCommonScript      bool   `json:"useCommonScript,omitempty"`
	Script               string `json:"script,omitempty"`
	ScriptUnix           string `json:"scriptUnix,omitempty"`
	ScriptWindows        string `json:"scriptWindows,omitempty"`
	ScriptTypeWindows    string `json:"scriptTypeWindows,omitempty"`
	AlwaysCancelOnFinish bool   `json:"alwaysCancelOnFinish,omitempty"`

	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

	Agent               string `json:"agent,omitempty"`
	AgentVar            string `json:"agentVar,omitempty"`
	AgentCluster        string `json:"agentCluster,omitempty"`
	AgentClusterVar     string `json:"agentClusterVar,omitempty"`
	BroadcastCluster    string `json:"broadcastCluster,omitempty"`
	BroadcastClusterVar string `json:"broadcastClusterVar,omitempty"`

	RuntimeDir      string                            `json:"runtimeDir,omitempty"`
	Environment     []UniversalTemplateEnvVarAPIModel `json:"environment,omitempty"`
	SendEnvironment string                            `json:"sendEnvironment,omitempty"`
	SendVariables   string                            `json:"sendVariables,omitempty"`

	ExitCodes          string `json:"exitCodes,omitempty"`
	ExitCodeProcessing string `json:"exitCodeProcessing,omitempty"`
	ExitCodeText       string `json:"exitCodeText,omitempty"`
	ExitCodeOutput     string `json:"exitCodeOutput,omitempty"`

	OutputType              string `json:"outputType,omitempty"`
	OutputContentType       string `json:"outputContentType,omitempty"`
	OutputPathExpression    string `json:"outputPathExpression,omitempty"`
	OutputConditionOperator string `json:"outputConditionOperator,omitempty"`
	OutputConditionValue    string `json:"outputConditionValue,omitempty"`
	OutputConditionStrategy string `json:"outputConditionStrategy,omitempty"`
	AutoCleanup             bool   `json:"autoCleanup,omitempty"`
	OutputReturnType        string `json:"outputReturnType,omitempty"`
	OutputReturnFile        string `json:"outputReturnFile,omitempty"`
	OutputReturnSline       string `json:"outputReturnSline,omitempty"`
	OutputReturnNline       string `json:"outputReturnNline,omitempty"`
	OutputReturnText        string `json:"outputReturnText,omitempty"`
	WaitForOutput           bool   `json:"waitForOutput,omitempty"`
	OutputFailureOnly       bool   `json:"outputFailureOnly,omitempty"`

	ElevateUser     bool `json:"elevateUser,omitempty"`
	DesktopInteract bool `json:"desktopInteract,omitempty"`
	CreateConsole   bool `json:"createConsole,omitempty"`

	LogLevel string `json:"logLevel,omitempty"`

	Fields []UniversalTemplateFieldAPIModel `json:"fields,omitempty"`
}

func (r *UniversalTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_universal_template"
}

func (r *UniversalTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Universal Template. Universal templates define a reusable custom task type (script-based) that can be executed on Unix/Windows/Any agents via a Universal Agent plugin.\n\n" +
			"**Deferred fields (not yet supported by this resource):** the `commands` and `events` (event-metric) definitions, and the five top-level per-attribute UI restriction settings (`*FieldsRestriction`). These must be managed outside Terraform (e.g. via the UAC UI) for now.",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the universal template (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the universal template.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the universal template.",
				Optional:            true,
			},

			// Core
			"extension": schema.StringAttribute{
				MarkdownDescription: "File extension associated with the template's script.",
				Optional:            true,
			},
			"variable_prefix": schema.StringAttribute{
				MarkdownDescription: "Prefix used for variables exposed by this template (e.g. tasks based on this template expose `<prefix>_<field>` variables).",
				Required:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: "Type of universal template. Confirmed value: `Script`. Other values may exist but were not discovered empirically.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agent_type": schema.StringAttribute{
				MarkdownDescription: "Agent platform this template targets. Confirmed values: `Windows`, `Any`.",
				Required:            true,
			},

			// Script
			"use_common_script": schema.BoolAttribute{
				MarkdownDescription: "Whether to use a single common script (`script`) for all agent types instead of platform-specific scripts (`script_unix`/`script_windows`).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"script": schema.StringAttribute{
				MarkdownDescription: "Common script content (used when `use_common_script` is `true`).",
				Optional:            true,
			},
			"script_unix": schema.StringAttribute{
				MarkdownDescription: "Unix/Linux-specific script content.",
				Optional:            true,
			},
			"script_windows": schema.StringAttribute{
				MarkdownDescription: "Windows-specific script content.",
				Optional:            true,
			},
			"script_type_windows": schema.StringAttribute{
				MarkdownDescription: "Script type/interpreter for the Windows script (e.g. batch, PowerShell). Exact enum values not confirmed empirically.",
				Optional:            true,
			},
			"always_cancel_on_finish": schema.BoolAttribute{
				MarkdownDescription: "Whether to always cancel the task on finish.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// Credentials
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use for task execution.",
				Optional:            true,
			},
			"credentials_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the credentials name.",
				Optional:            true,
			},

			// Agent targeting
			"agent": schema.StringAttribute{
				MarkdownDescription: "Name of the agent to run tasks based on this template.",
				Optional:            true,
			},
			"agent_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the agent name.",
				Optional:            true,
			},
			"agent_cluster": schema.StringAttribute{
				MarkdownDescription: "Name of the agent cluster to run tasks based on this template.",
				Optional:            true,
			},
			"agent_cluster_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the agent cluster name.",
				Optional:            true,
			},
			"broadcast_cluster": schema.StringAttribute{
				MarkdownDescription: "Name of the broadcast cluster.",
				Optional:            true,
			},
			"broadcast_cluster_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the broadcast cluster name.",
				Optional:            true,
			},

			// Runtime / environment / variables
			"runtime_dir": schema.StringAttribute{
				MarkdownDescription: "Working directory for the task execution.",
				Optional:            true,
			},
			"environment": schema.ListNestedAttribute{
				MarkdownDescription: "Environment variables (name/value pairs) passed to the script. Full-replace on update: omitting this attribute clears any previously-set environment variables.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the environment variable.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Value of the environment variable.",
							Optional:            true,
						},
					},
				},
			},
			"send_environment": schema.StringAttribute{
				MarkdownDescription: "Controls which environment variables are sent to the agent. Confirmed default: `Launch`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"send_variables": schema.StringAttribute{
				MarkdownDescription: "Controls which task variables are sent to the agent. Confirmed values: `None`. Default: `None`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Exit codes
			"exit_codes": schema.StringAttribute{
				MarkdownDescription: "Exit codes that indicate success (e.g. '0' or '0,1,2'). The server has no default for this field; it must be set whenever exit_code_processing is a range-based mode (the only modes discovered).",
				Required:            true,
			},
			"exit_code_processing": schema.StringAttribute{
				MarkdownDescription: "How to process exit codes. Confirmed values: `Success Exitcode Range`, `Failure Exitcode Range`. Default: `Success Exitcode Range`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"exit_code_text": schema.StringAttribute{
				MarkdownDescription: "Text used for exit code processing when applicable.",
				Optional:            true,
			},
			"exit_code_output": schema.StringAttribute{
				MarkdownDescription: "Output source used for exit code processing when applicable.",
				Optional:            true,
			},

			// Output handling
			"output_type": schema.StringAttribute{
				MarkdownDescription: "Type of output to capture. Default: `STDOUT`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"output_content_type": schema.StringAttribute{
				MarkdownDescription: "Content type of the captured output.",
				Optional:            true,
			},
			"output_path_expression": schema.StringAttribute{
				MarkdownDescription: "Path expression used to extract output.",
				Optional:            true,
			},
			"output_condition_operator": schema.StringAttribute{
				MarkdownDescription: "Operator used to evaluate output conditions. Default: `=`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"output_condition_value": schema.StringAttribute{
				MarkdownDescription: "Value used to evaluate output conditions.",
				Optional:            true,
			},
			"output_condition_strategy": schema.StringAttribute{
				MarkdownDescription: "Strategy used to combine output conditions. Default: `Match Any`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_cleanup": schema.BoolAttribute{
				MarkdownDescription: "Whether to automatically clean up output artifacts.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"output_return_type": schema.StringAttribute{
				MarkdownDescription: "How output should be returned. Default: `NONE`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"output_return_file": schema.StringAttribute{
				MarkdownDescription: "File to capture output from.",
				Optional:            true,
			},
			"output_return_sline": schema.StringAttribute{
				MarkdownDescription: "Starting line number for output return. Default: `1`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"output_return_nline": schema.StringAttribute{
				MarkdownDescription: "Number of lines to return for output. Default: `100`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"output_return_text": schema.StringAttribute{
				MarkdownDescription: "Text used to filter returned output.",
				Optional:            true,
			},
			"wait_for_output": schema.BoolAttribute{
				MarkdownDescription: "Whether to wait for output before completing.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"output_failure_only": schema.BoolAttribute{
				MarkdownDescription: "Whether to only capture output on failure.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// Windows-specific execution options
			"elevate_user": schema.BoolAttribute{
				MarkdownDescription: "Whether to run the task with elevated (administrator) privileges. Windows agent types only.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"desktop_interact": schema.BoolAttribute{
				MarkdownDescription: "Whether the task can interact with the desktop. Windows agent types only.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"create_console": schema.BoolAttribute{
				MarkdownDescription: "Whether to create a console window for the task. Windows agent types only.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// Logging
			"log_level": schema.StringAttribute{
				MarkdownDescription: "Logging level for tasks based on this template. Default: `Inherited`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Custom UI form fields
			"fields": schema.ListNestedAttribute{
				MarkdownDescription: "Custom UI form fields exposed on tasks based on this template. Each field maps to one of a fixed set of typed extension slots on the template via `field_mapping` (e.g. `Text Field 1`..`Text Field 20`, `Integer Field 1`..`Integer Field 10`, `Boolean Field 1`..`Boolean Field 15`, `Choice Field 1`..`Choice Field 15`, `Credential Field 1`..`Credential Field 6`, `Script Field 1`..`Script Field 4`, `Array Field 1`..`Array Field 4`, `Float Field 1`..`Float Field 4`, `Large Text Field 1`..`Large Text Field 4`, `SAP Connection Field 1`, `Database Connection Field 1`). Full-replace on update: omitting this attribute clears any previously-set fields. **Known API limitation:** once a template has fields, the server rejects updates that would reduce the list to empty (`fields = []` or omitting the attribute) — you can only ever replace fields with a different non-empty set.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Internal name of the field (used in `<variable_prefix>_<name>` task variables).",
							Required:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "Display label of the field in the UAC UI.",
							Required:            true,
						},
						"field_mapping": schema.StringAttribute{
							MarkdownDescription: "The typed extension slot this field is backed by. Must match `field_type` (e.g. `field_type = \"Text\"` requires a `Text Field N` or `Large Text Field N` mapping). No server default; always required.",
							Required:            true,
						},
						"field_type": schema.StringAttribute{
							MarkdownDescription: "Type of the field. Confirmed values: `Text`, `Integer`, `Boolean`, `Choice`, `Credential`, `Script`, `Array`, `Float`, `SAP Connection`, `Database Connection`. Default: `Text`.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"hint": schema.StringAttribute{
							MarkdownDescription: "Help text displayed for the field in the UAC UI.",
							Optional:            true,
						},
						"text_type": schema.StringAttribute{
							MarkdownDescription: "Text formatting for Text-type fields. Confirmed values: `Plain`, `XML`, `JSON`, `YAML`. Default: `Plain`.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"field_value": schema.StringAttribute{
							MarkdownDescription: "Default value for the field.",
							Optional:            true,
						},
						"default_list_view": schema.BoolAttribute{
							MarkdownDescription: "Whether this field is shown by default in list views.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"allow_variable": schema.BoolAttribute{
							MarkdownDescription: "Whether the field's value may reference a variable.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"field_restriction": schema.StringAttribute{
							MarkdownDescription: "UI restriction applied to this field. Confirmed value: `No Restriction`. Other values may exist but were not discovered empirically. Default: `No Restriction`.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"preserve_output_on_rerun": schema.BoolAttribute{
							MarkdownDescription: "Whether to preserve the field's output value when the task is rerun.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"extension_status": schema.BoolAttribute{
							MarkdownDescription: "Whether this field reflects an extension status.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"boolean_value_type": schema.StringAttribute{
							MarkdownDescription: "For Boolean-type fields, how the boolean is represented. Confirmed values: `true/false` (default), `1/0`, `Custom` (requires `boolean_yes_value`/`boolean_no_value` to also be set).",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"boolean_yes_value": schema.StringAttribute{
							MarkdownDescription: "Custom \"true\" display value. Required when `boolean_value_type` is `Custom`.",
							Optional:            true,
						},
						"boolean_no_value": schema.StringAttribute{
							MarkdownDescription: "Custom \"false\" display value. Required when `boolean_value_type` is `Custom`.",
							Optional:            true,
						},
						"choice_sort_option": schema.StringAttribute{
							MarkdownDescription: "For Choice-type fields, how `choices` are sorted in the UI. Confirmed values: `Sequence` (default), `Label`.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"choice_allow_empty": schema.BoolAttribute{
							MarkdownDescription: "For Choice-type fields, whether an empty selection is allowed.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"choice_allow_multiple": schema.BoolAttribute{
							MarkdownDescription: "For Choice-type fields, whether multiple selections are allowed.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"choice_dynamic": schema.BoolAttribute{
							MarkdownDescription: "For Choice-type fields, whether choices are populated dynamically (see `choice_fields`) instead of from the static `choices` list.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"choice_fields": schema.ListAttribute{
							MarkdownDescription: "For dynamic Choice-type fields, the field names used to populate choices.",
							Optional:            true,
							ElementType:         types.StringType,
						},
						"choices": schema.ListNestedAttribute{
							MarkdownDescription: "Static choice options for a Choice-type field. Full-replace on update.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"field_value": schema.StringAttribute{
										MarkdownDescription: "The stored value of this choice.",
										Required:            true,
									},
									"field_value_label": schema.StringAttribute{
										MarkdownDescription: "The display label of this choice.",
										Optional:            true,
									},
									"use_field_value_for_label": schema.BoolAttribute{
										MarkdownDescription: "Whether to display `field_value` instead of `field_value_label` in the UI.",
										Optional:            true,
										Computed:            true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"required": schema.BoolAttribute{
							MarkdownDescription: "Whether the field is required. **Known API limitation:** the server silently ignores this attribute (always resets it to `false`) for `field_type` values `Choice`, `Boolean`, and `Array`; it is honored for `Text`, `Integer`, `Credential`, `Script`, `Float`, `SAP Connection`, and `Database Connection`.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"require_if_field": schema.StringAttribute{
							MarkdownDescription: "Name of another field whose value conditionally makes this field required (see `require_if_field_value`).",
							Optional:            true,
						},
						"require_if_field_value": schema.StringAttribute{
							MarkdownDescription: "Value of `require_if_field` that makes this field required.",
							Optional:            true,
						},
						"show_if_field": schema.StringAttribute{
							MarkdownDescription: "Name of another field whose value conditionally shows this field (see `show_if_field_value`).",
							Optional:            true,
						},
						"show_if_field_value": schema.StringAttribute{
							MarkdownDescription: "Value of `show_if_field` that shows this field.",
							Optional:            true,
						},
						"require_if_visible": schema.BoolAttribute{
							MarkdownDescription: "Whether this field becomes required whenever it is visible.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"preserve_value_if_hidden": schema.BoolAttribute{
							MarkdownDescription: "Whether to preserve the field's value when it is hidden.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"no_space_if_hidden": schema.BoolAttribute{
							MarkdownDescription: "Whether to remove the field's form space when it is hidden.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"field_length": schema.Int64Attribute{
							MarkdownDescription: "Maximum length for Text-type fields.",
							Optional:            true,
						},
						"int_field_min": schema.Int64Attribute{
							MarkdownDescription: "Minimum value for Integer-type fields.",
							Optional:            true,
						},
						"int_field_max": schema.Int64Attribute{
							MarkdownDescription: "Maximum value for Integer-type fields.",
							Optional:            true,
						},
						"field_regex": schema.StringAttribute{
							MarkdownDescription: "Regular expression used to validate the field's value.",
							Optional:            true,
						},
						"form_column_span": schema.Int64Attribute{
							MarkdownDescription: "Number of form columns this field spans in the UAC UI.",
							Optional:            true,
						},
						"form_start_row": schema.BoolAttribute{
							MarkdownDescription: "Whether this field starts a new row in the UAC UI form layout.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"form_end_row": schema.BoolAttribute{
							MarkdownDescription: "Whether this field ends its row in the UAC UI form layout.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"array_name_title": schema.StringAttribute{
							MarkdownDescription: "For Array-type fields, the column header used for the name/key column.",
							Optional:            true,
						},
						"array_value_title": schema.StringAttribute{
							MarkdownDescription: "For Array-type fields, the column header used for the value column.",
							Optional:            true,
						},
						"array_field_value": schema.ListNestedAttribute{
							MarkdownDescription: "For Array-type fields, the default name/value pairs. Full-replace on update.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "Name/key of the array entry.",
										Required:            true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "Value of the array entry.",
										Optional:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *UniversalTemplateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data UniversalTemplateResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(ValidateTaskOutputReturn(data.OutputReturnType, data.OutputReturnSline)...)
}

func (r *UniversalTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *UniversalTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UniversalTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating universal template", map[string]any{"name": data.Name.ValueString()})

	apiModel := r.toAPIModel(ctx, &data)

	_, err := r.client.Post(ctx, "/resources/universaltemplate", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Universal Template",
			fmt.Sprintf("Could not create universal template %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	err = r.readUniversalTemplate(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Universal Template",
			fmt.Sprintf("Could not read universal template %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created universal template", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UniversalTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UniversalTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readUniversalTemplate(ctx, &data)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Universal template not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Universal Template",
			fmt.Sprintf("Could not read universal template %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UniversalTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UniversalTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UniversalTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating universal template", map[string]any{"sys_id": data.SysId.ValueString()})

	apiModel := r.toAPIModel(ctx, &data)

	_, err := r.client.Put(ctx, "/resources/universaltemplate", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Universal Template",
			fmt.Sprintf("Could not update universal template %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	err = r.readUniversalTemplate(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Universal Template",
			fmt.Sprintf("Could not read universal template %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UniversalTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UniversalTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting universal template", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("templateid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/universaltemplate", query)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Universal Template",
			fmt.Sprintf("Could not delete universal template %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *UniversalTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readUniversalTemplate fetches the universal template from the API and updates the model.
func (r *UniversalTemplateResource) readUniversalTemplate(ctx context.Context, data *UniversalTemplateResourceModel) error {
	query := url.Values{}
	query.Set("templatename", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/universaltemplate", query)
	if err != nil {
		return err
	}

	var apiModel UniversalTemplateAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse universal template response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *UniversalTemplateResource) toAPIModel(ctx context.Context, data *UniversalTemplateResourceModel) *UniversalTemplateAPIModel {
	model := &UniversalTemplateAPIModel{
		SysId:             data.SysId.ValueString(),
		Name:              data.Name.ValueString(),
		Description:       data.Description.ValueString(),
		Extension:         data.Extension.ValueString(),
		VariablePrefix:    data.VariablePrefix.ValueString(),
		AgentType:         data.AgentType.ValueString(),
		Script:            data.Script.ValueString(),
		ScriptUnix:        data.ScriptUnix.ValueString(),
		ScriptWindows:     data.ScriptWindows.ValueString(),
		ScriptTypeWindows: data.ScriptTypeWindows.ValueString(),
		Credentials:       data.Credentials.ValueString(),
		CredentialsVar:    data.CredentialsVar.ValueString(),

		Agent:               data.Agent.ValueString(),
		AgentVar:            data.AgentVar.ValueString(),
		AgentCluster:        data.AgentCluster.ValueString(),
		AgentClusterVar:     data.AgentClusterVar.ValueString(),
		BroadcastCluster:    data.BroadcastCluster.ValueString(),
		BroadcastClusterVar: data.BroadcastClusterVar.ValueString(),

		RuntimeDir: data.RuntimeDir.ValueString(),

		ExitCodes:      data.ExitCodes.ValueString(),
		ExitCodeText:   data.ExitCodeText.ValueString(),
		ExitCodeOutput: data.ExitCodeOutput.ValueString(),

		OutputContentType:    data.OutputContentType.ValueString(),
		OutputPathExpression: data.OutputPathExpression.ValueString(),
		OutputConditionValue: data.OutputConditionValue.ValueString(),
		OutputReturnFile:     data.OutputReturnFile.ValueString(),
		OutputReturnText:     data.OutputReturnText.ValueString(),
	}

	if !data.TemplateType.IsNull() && !data.TemplateType.IsUnknown() {
		model.TemplateType = data.TemplateType.ValueString()
	}
	if !data.UseCommonScript.IsNull() && !data.UseCommonScript.IsUnknown() {
		model.UseCommonScript = data.UseCommonScript.ValueBool()
	}
	if !data.AlwaysCancelOnFinish.IsNull() && !data.AlwaysCancelOnFinish.IsUnknown() {
		model.AlwaysCancelOnFinish = data.AlwaysCancelOnFinish.ValueBool()
	}
	if !data.SendEnvironment.IsNull() && !data.SendEnvironment.IsUnknown() {
		model.SendEnvironment = data.SendEnvironment.ValueString()
	}
	if !data.SendVariables.IsNull() && !data.SendVariables.IsUnknown() {
		model.SendVariables = data.SendVariables.ValueString()
	}
	if !data.ExitCodeProcessing.IsNull() && !data.ExitCodeProcessing.IsUnknown() {
		model.ExitCodeProcessing = data.ExitCodeProcessing.ValueString()
	}
	if !data.OutputType.IsNull() && !data.OutputType.IsUnknown() {
		model.OutputType = data.OutputType.ValueString()
	}
	if !data.OutputConditionOperator.IsNull() && !data.OutputConditionOperator.IsUnknown() {
		model.OutputConditionOperator = data.OutputConditionOperator.ValueString()
	}
	if !data.OutputConditionStrategy.IsNull() && !data.OutputConditionStrategy.IsUnknown() {
		model.OutputConditionStrategy = data.OutputConditionStrategy.ValueString()
	}
	if !data.AutoCleanup.IsNull() && !data.AutoCleanup.IsUnknown() {
		model.AutoCleanup = data.AutoCleanup.ValueBool()
	}
	if !data.OutputReturnType.IsNull() && !data.OutputReturnType.IsUnknown() {
		model.OutputReturnType = data.OutputReturnType.ValueString()
	}
	if !data.OutputReturnSline.IsNull() && !data.OutputReturnSline.IsUnknown() {
		model.OutputReturnSline = data.OutputReturnSline.ValueString()
	}
	if !data.OutputReturnNline.IsNull() && !data.OutputReturnNline.IsUnknown() {
		model.OutputReturnNline = data.OutputReturnNline.ValueString()
	}
	if !data.WaitForOutput.IsNull() && !data.WaitForOutput.IsUnknown() {
		model.WaitForOutput = data.WaitForOutput.ValueBool()
	}
	if !data.OutputFailureOnly.IsNull() && !data.OutputFailureOnly.IsUnknown() {
		model.OutputFailureOnly = data.OutputFailureOnly.ValueBool()
	}
	if !data.ElevateUser.IsNull() && !data.ElevateUser.IsUnknown() {
		model.ElevateUser = data.ElevateUser.ValueBool()
	}
	if !data.DesktopInteract.IsNull() && !data.DesktopInteract.IsUnknown() {
		model.DesktopInteract = data.DesktopInteract.ValueBool()
	}
	if !data.CreateConsole.IsNull() && !data.CreateConsole.IsUnknown() {
		model.CreateConsole = data.CreateConsole.ValueBool()
	}
	if !data.LogLevel.IsNull() && !data.LogLevel.IsUnknown() {
		model.LogLevel = data.LogLevel.ValueString()
	}

	// Environment variables (full-replace on the wire).
	if !data.Environment.IsNull() && !data.Environment.IsUnknown() {
		var envVars []UniversalTemplateEnvVarModel
		data.Environment.ElementsAs(ctx, &envVars, false)
		for _, e := range envVars {
			model.Environment = append(model.Environment, UniversalTemplateEnvVarAPIModel{
				Name:  e.Name.ValueString(),
				Value: e.Value.ValueString(),
			})
		}
	}

	// Custom UI form fields (full-replace on the wire).
	if !data.Fields.IsNull() && !data.Fields.IsUnknown() {
		var fields []UniversalTemplateFieldModel
		data.Fields.ElementsAs(ctx, &fields, false)
		for _, f := range fields {
			model.Fields = append(model.Fields, universalTemplateFieldToAPIModel(ctx, &f))
		}
	}

	return model
}

// universalTemplateFieldToAPIModel converts a single field model to its API wire format.
func universalTemplateFieldToAPIModel(ctx context.Context, f *UniversalTemplateFieldModel) UniversalTemplateFieldAPIModel {
	apiField := UniversalTemplateFieldAPIModel{
		Name:                f.Name.ValueString(),
		Label:               f.Label.ValueString(),
		FieldMapping:        f.FieldMapping.ValueString(),
		Hint:                f.Hint.ValueString(),
		FieldValue:          f.FieldValue.ValueString(),
		BooleanYesValue:     f.BooleanYesValue.ValueString(),
		BooleanNoValue:      f.BooleanNoValue.ValueString(),
		RequireIfField:      f.RequireIfField.ValueString(),
		RequireIfFieldValue: f.RequireIfFieldValue.ValueString(),
		ShowIfField:         f.ShowIfField.ValueString(),
		ShowIfFieldValue:    f.ShowIfFieldValue.ValueString(),
		FieldRegex:          f.FieldRegex.ValueString(),
		ArrayNameTitle:      f.ArrayNameTitle.ValueString(),
		ArrayValueTitle:     f.ArrayValueTitle.ValueString(),
	}

	if !f.FieldType.IsNull() && !f.FieldType.IsUnknown() {
		apiField.FieldType = f.FieldType.ValueString()
	}
	if !f.TextType.IsNull() && !f.TextType.IsUnknown() {
		apiField.TextType = f.TextType.ValueString()
	}
	if !f.DefaultListView.IsNull() && !f.DefaultListView.IsUnknown() {
		apiField.DefaultListView = f.DefaultListView.ValueBool()
	}
	if !f.AllowVariable.IsNull() && !f.AllowVariable.IsUnknown() {
		apiField.AllowVariable = f.AllowVariable.ValueBool()
	}
	if !f.FieldRestriction.IsNull() && !f.FieldRestriction.IsUnknown() {
		apiField.FieldRestriction = f.FieldRestriction.ValueString()
	}
	if !f.PreserveOutputOnRerun.IsNull() && !f.PreserveOutputOnRerun.IsUnknown() {
		apiField.PreserveOutputOnRerun = f.PreserveOutputOnRerun.ValueBool()
	}
	if !f.ExtensionStatus.IsNull() && !f.ExtensionStatus.IsUnknown() {
		apiField.ExtensionStatus = f.ExtensionStatus.ValueBool()
	}
	if !f.BooleanValueType.IsNull() && !f.BooleanValueType.IsUnknown() {
		apiField.BooleanValueType = f.BooleanValueType.ValueString()
	}
	if !f.ChoiceSortOption.IsNull() && !f.ChoiceSortOption.IsUnknown() {
		apiField.ChoiceSortOption = f.ChoiceSortOption.ValueString()
	}
	if !f.ChoiceAllowEmpty.IsNull() && !f.ChoiceAllowEmpty.IsUnknown() {
		apiField.ChoiceAllowEmpty = f.ChoiceAllowEmpty.ValueBool()
	}
	if !f.ChoiceAllowMultiple.IsNull() && !f.ChoiceAllowMultiple.IsUnknown() {
		apiField.ChoiceAllowMultiple = f.ChoiceAllowMultiple.ValueBool()
	}
	if !f.ChoiceDynamic.IsNull() && !f.ChoiceDynamic.IsUnknown() {
		apiField.ChoiceDynamic = f.ChoiceDynamic.ValueBool()
	}
	if !f.Required.IsNull() && !f.Required.IsUnknown() {
		apiField.Required = f.Required.ValueBool()
	}
	if !f.RequireIfVisible.IsNull() && !f.RequireIfVisible.IsUnknown() {
		apiField.RequireIfVisible = f.RequireIfVisible.ValueBool()
	}
	if !f.PreserveValueIfHidden.IsNull() && !f.PreserveValueIfHidden.IsUnknown() {
		apiField.PreserveValueIfHidden = f.PreserveValueIfHidden.ValueBool()
	}
	if !f.NoSpaceIfHidden.IsNull() && !f.NoSpaceIfHidden.IsUnknown() {
		apiField.NoSpaceIfHidden = f.NoSpaceIfHidden.ValueBool()
	}
	if !f.FieldLength.IsNull() && !f.FieldLength.IsUnknown() {
		apiField.FieldLength = f.FieldLength.ValueInt64()
	}
	if !f.IntFieldMin.IsNull() && !f.IntFieldMin.IsUnknown() {
		apiField.IntFieldMin = f.IntFieldMin.ValueInt64()
	}
	if !f.IntFieldMax.IsNull() && !f.IntFieldMax.IsUnknown() {
		apiField.IntFieldMax = f.IntFieldMax.ValueInt64()
	}
	if !f.FormColumnSpan.IsNull() && !f.FormColumnSpan.IsUnknown() {
		apiField.FormColumnSpan = f.FormColumnSpan.ValueInt64()
	}
	if !f.FormStartRow.IsNull() && !f.FormStartRow.IsUnknown() {
		apiField.FormStartRow = f.FormStartRow.ValueBool()
	}
	if !f.FormEndRow.IsNull() && !f.FormEndRow.IsUnknown() {
		apiField.FormEndRow = f.FormEndRow.ValueBool()
	}

	if !f.ChoiceFields.IsNull() && !f.ChoiceFields.IsUnknown() {
		var choiceFields []string
		f.ChoiceFields.ElementsAs(ctx, &choiceFields, false)
		apiField.ChoiceFields = choiceFields
	}

	if !f.Choices.IsNull() && !f.Choices.IsUnknown() {
		var choices []UniversalTemplateFieldChoiceModel
		f.Choices.ElementsAs(ctx, &choices, false)
		for _, c := range choices {
			choice := UniversalTemplateFieldChoiceAPIModel{
				FieldValue:      c.FieldValue.ValueString(),
				FieldValueLabel: c.FieldValueLabel.ValueString(),
			}
			if !c.UseFieldValueForLabel.IsNull() && !c.UseFieldValueForLabel.IsUnknown() {
				choice.UseFieldValueForLabel = c.UseFieldValueForLabel.ValueBool()
			}
			apiField.Choices = append(apiField.Choices, choice)
		}
	}

	if !f.ArrayFieldValue.IsNull() && !f.ArrayFieldValue.IsUnknown() {
		var arrayValues []UniversalTemplateEnvVarModel
		f.ArrayFieldValue.ElementsAs(ctx, &arrayValues, false)
		for _, v := range arrayValues {
			apiField.ArrayFieldValue = append(apiField.ArrayFieldValue, UniversalTemplateEnvVarAPIModel{
				Name:  v.Name.ValueString(),
				Value: v.Value.ValueString(),
			})
		}
	}

	return apiField
}

// fromAPIModel converts an API model to the Terraform model.
func (r *UniversalTemplateResource) fromAPIModel(ctx context.Context, apiModel *UniversalTemplateAPIModel, data *UniversalTemplateResourceModel) {
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Description = StringValueOrNull(apiModel.Description)

	data.Extension = StringValueOrNull(apiModel.Extension)
	data.VariablePrefix = types.StringValue(apiModel.VariablePrefix)
	data.TemplateType = StringValueOrNull(apiModel.TemplateType)
	data.AgentType = types.StringValue(apiModel.AgentType)

	data.UseCommonScript = types.BoolValue(apiModel.UseCommonScript)
	data.Script = StringValueOrNull(apiModel.Script)
	data.ScriptUnix = StringValueOrNull(apiModel.ScriptUnix)
	data.ScriptWindows = StringValueOrNull(apiModel.ScriptWindows)
	data.ScriptTypeWindows = StringValueOrNull(apiModel.ScriptTypeWindows)
	data.AlwaysCancelOnFinish = types.BoolValue(apiModel.AlwaysCancelOnFinish)

	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)

	data.Agent = StringValueOrNull(apiModel.Agent)
	data.AgentVar = StringValueOrNull(apiModel.AgentVar)
	data.AgentCluster = StringValueOrNull(apiModel.AgentCluster)
	data.AgentClusterVar = StringValueOrNull(apiModel.AgentClusterVar)
	data.BroadcastCluster = StringValueOrNull(apiModel.BroadcastCluster)
	data.BroadcastClusterVar = StringValueOrNull(apiModel.BroadcastClusterVar)

	data.RuntimeDir = StringValueOrNull(apiModel.RuntimeDir)
	data.SendEnvironment = StringValueOrNull(apiModel.SendEnvironment)
	data.SendVariables = StringValueOrNull(apiModel.SendVariables)

	data.ExitCodes = types.StringValue(apiModel.ExitCodes)
	data.ExitCodeProcessing = StringValueOrNull(apiModel.ExitCodeProcessing)
	data.ExitCodeText = StringValueOrNull(apiModel.ExitCodeText)
	data.ExitCodeOutput = StringValueOrNull(apiModel.ExitCodeOutput)

	data.OutputType = StringValueOrNull(apiModel.OutputType)
	data.OutputContentType = StringValueOrNull(apiModel.OutputContentType)
	data.OutputPathExpression = StringValueOrNull(apiModel.OutputPathExpression)
	data.OutputConditionOperator = StringValueOrNull(apiModel.OutputConditionOperator)
	data.OutputConditionValue = StringValueOrNull(apiModel.OutputConditionValue)
	data.OutputConditionStrategy = StringValueOrNull(apiModel.OutputConditionStrategy)
	data.AutoCleanup = types.BoolValue(apiModel.AutoCleanup)
	data.OutputReturnType = StringValueOrNull(apiModel.OutputReturnType)
	data.OutputReturnFile = StringValueOrNull(apiModel.OutputReturnFile)
	data.OutputReturnSline = StringValueOrNull(apiModel.OutputReturnSline)
	data.OutputReturnNline = StringValueOrNull(apiModel.OutputReturnNline)
	data.OutputReturnText = StringValueOrNull(apiModel.OutputReturnText)
	data.WaitForOutput = types.BoolValue(apiModel.WaitForOutput)
	data.OutputFailureOnly = types.BoolValue(apiModel.OutputFailureOnly)

	data.ElevateUser = types.BoolValue(apiModel.ElevateUser)
	data.DesktopInteract = types.BoolValue(apiModel.DesktopInteract)
	data.CreateConsole = types.BoolValue(apiModel.CreateConsole)

	data.LogLevel = StringValueOrNull(apiModel.LogLevel)

	if len(apiModel.Environment) > 0 {
		values := make([]attr.Value, len(apiModel.Environment))
		for i, e := range apiModel.Environment {
			values[i], _ = types.ObjectValue(universalTemplateEnvVarAttrTypes(), map[string]attr.Value{
				"name":  types.StringValue(e.Name),
				"value": StringValueOrNull(e.Value),
			})
		}
		list, _ := types.ListValue(types.ObjectType{AttrTypes: universalTemplateEnvVarAttrTypes()}, values)
		data.Environment = list
	} else {
		data.Environment = types.ListNull(types.ObjectType{AttrTypes: universalTemplateEnvVarAttrTypes()})
	}

	if len(apiModel.Fields) > 0 {
		values := make([]attr.Value, len(apiModel.Fields))
		for i, f := range apiModel.Fields {
			values[i] = universalTemplateFieldFromAPIModel(&f)
		}
		list, _ := types.ListValue(types.ObjectType{AttrTypes: universalTemplateFieldAttrTypes()}, values)
		data.Fields = list
	} else {
		data.Fields = types.ListNull(types.ObjectType{AttrTypes: universalTemplateFieldAttrTypes()})
	}
}

// universalTemplateFieldFromAPIModel converts a single API field to its Terraform object representation.
func universalTemplateFieldFromAPIModel(f *UniversalTemplateFieldAPIModel) attr.Value {
	var choiceFieldsList types.List
	if len(f.ChoiceFields) > 0 {
		values := make([]attr.Value, len(f.ChoiceFields))
		for i, cf := range f.ChoiceFields {
			values[i] = types.StringValue(cf)
		}
		choiceFieldsList, _ = types.ListValue(types.StringType, values)
	} else {
		choiceFieldsList = types.ListNull(types.StringType)
	}

	var choicesList types.List
	if len(f.Choices) > 0 {
		values := make([]attr.Value, len(f.Choices))
		for i, c := range f.Choices {
			values[i], _ = types.ObjectValue(universalTemplateFieldChoiceAttrTypes(), map[string]attr.Value{
				"field_value":               types.StringValue(c.FieldValue),
				"field_value_label":         StringValueOrNull(c.FieldValueLabel),
				"use_field_value_for_label": types.BoolValue(c.UseFieldValueForLabel),
			})
		}
		choicesList, _ = types.ListValue(types.ObjectType{AttrTypes: universalTemplateFieldChoiceAttrTypes()}, values)
	} else {
		choicesList = types.ListNull(types.ObjectType{AttrTypes: universalTemplateFieldChoiceAttrTypes()})
	}

	var arrayFieldValueList types.List
	if len(f.ArrayFieldValue) > 0 {
		values := make([]attr.Value, len(f.ArrayFieldValue))
		for i, v := range f.ArrayFieldValue {
			values[i], _ = types.ObjectValue(universalTemplateEnvVarAttrTypes(), map[string]attr.Value{
				"name":  types.StringValue(v.Name),
				"value": StringValueOrNull(v.Value),
			})
		}
		arrayFieldValueList, _ = types.ListValue(types.ObjectType{AttrTypes: universalTemplateEnvVarAttrTypes()}, values)
	} else {
		arrayFieldValueList = types.ListNull(types.ObjectType{AttrTypes: universalTemplateEnvVarAttrTypes()})
	}

	value, _ := types.ObjectValue(universalTemplateFieldAttrTypes(), map[string]attr.Value{
		"name":                     types.StringValue(f.Name),
		"label":                    types.StringValue(f.Label),
		"field_mapping":            types.StringValue(f.FieldMapping),
		"field_type":               StringValueOrNull(f.FieldType),
		"hint":                     StringValueOrNull(f.Hint),
		"text_type":                StringValueOrNull(f.TextType),
		"field_value":              StringValueOrNull(f.FieldValue),
		"default_list_view":        types.BoolValue(f.DefaultListView),
		"allow_variable":           types.BoolValue(f.AllowVariable),
		"field_restriction":        StringValueOrNull(f.FieldRestriction),
		"preserve_output_on_rerun": types.BoolValue(f.PreserveOutputOnRerun),
		"extension_status":         types.BoolValue(f.ExtensionStatus),
		"boolean_value_type":       StringValueOrNull(f.BooleanValueType),
		"boolean_yes_value":        StringValueOrNull(f.BooleanYesValue),
		"boolean_no_value":         StringValueOrNull(f.BooleanNoValue),
		"choice_sort_option":       StringValueOrNull(f.ChoiceSortOption),
		"choice_allow_empty":       types.BoolValue(f.ChoiceAllowEmpty),
		"choice_allow_multiple":    types.BoolValue(f.ChoiceAllowMultiple),
		"choice_dynamic":           types.BoolValue(f.ChoiceDynamic),
		"choice_fields":            choiceFieldsList,
		"choices":                  choicesList,
		"required":                 types.BoolValue(f.Required),
		"require_if_field":         StringValueOrNull(f.RequireIfField),
		"require_if_field_value":   StringValueOrNull(f.RequireIfFieldValue),
		"show_if_field":            StringValueOrNull(f.ShowIfField),
		"show_if_field_value":      StringValueOrNull(f.ShowIfFieldValue),
		"require_if_visible":       types.BoolValue(f.RequireIfVisible),
		"preserve_value_if_hidden": types.BoolValue(f.PreserveValueIfHidden),
		"no_space_if_hidden":       types.BoolValue(f.NoSpaceIfHidden),
		"field_length":             int64ValueOrNull(f.FieldLength),
		"int_field_min":            int64ValueOrNull(f.IntFieldMin),
		"int_field_max":            int64ValueOrNull(f.IntFieldMax),
		"field_regex":              StringValueOrNull(f.FieldRegex),
		"form_column_span":         int64ValueOrNull(f.FormColumnSpan),
		"form_start_row":           types.BoolValue(f.FormStartRow),
		"form_end_row":             types.BoolValue(f.FormEndRow),
		"array_name_title":         StringValueOrNull(f.ArrayNameTitle),
		"array_value_title":        StringValueOrNull(f.ArrayValueTitle),
		"array_field_value":        arrayFieldValueList,
	})
	return value
}

// int64ValueOrNull returns a null Int64 for a zero value, otherwise the value itself.
// Used for optional numeric field attributes where the API has no way to
// distinguish "not set" from zero.
func int64ValueOrNull(v int64) types.Int64 {
	if v == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(v)
}
