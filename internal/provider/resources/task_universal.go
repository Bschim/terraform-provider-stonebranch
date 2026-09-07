package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &TaskUniversalResource{}
	_ resource.ResourceWithImportState    = &TaskUniversalResource{}
	_ resource.ResourceWithValidateConfig = &TaskUniversalResource{}
)

func NewTaskUniversalResource() resource.Resource {
	return &TaskUniversalResource{}
}

// TaskUniversalResource defines the resource implementation.
//
// Unlike stonebranch_task_universal_aws_s3 (which hardcodes field names for one
// specific Universal Template), this resource is a generic pass-through for any
// Universal-Template-backed task: it exposes UAC's raw numbered field slots
// directly, the same way UAC itself represents a taskUniversal before applying
// template-specific field labels.
type TaskUniversalResource struct {
	client *client.Client
}

// TaskUniversalResourceModel describes the resource data model.
type TaskUniversalResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary  types.String `tfsdk:"summary"`
	Template types.String `tfsdk:"template"`

	// Agent configuration
	Agent           types.String `tfsdk:"agent"`
	AgentCluster    types.String `tfsdk:"agent_cluster"`
	AgentVar        types.String `tfsdk:"agent_var"`
	AgentClusterVar types.String `tfsdk:"agent_cluster_var"`

	// Credentials (for task execution)
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

	// Exit code handling
	ExitCodes          types.String `tfsdk:"exit_codes"`
	ExitCodeProcessing types.String `tfsdk:"exit_code_processing"`
	ExitCodeText       types.String `tfsdk:"exit_code_text"`

	// Retry configuration
	RetryMaximum         types.Int64 `tfsdk:"retry_maximum"`
	RetryIndefinitely    types.Bool  `tfsdk:"retry_indefinitely"`
	RetryInterval        types.Int64 `tfsdk:"retry_interval"`
	RetrySuppressFailure types.Bool  `tfsdk:"retry_suppress_failure"`

	// Wait/Delay options
	WaitToStart       types.String `tfsdk:"wait_to_start"`
	WaitTime          types.String `tfsdk:"wait_time"`
	WaitDuration      types.String `tfsdk:"wait_duration"`
	WaitAmount        types.String `tfsdk:"wait_amount"`
	WaitDayConstraint types.String `tfsdk:"wait_day_constraint"`
	DelayOnStart      types.String `tfsdk:"delay_on_start"`
	DelayDuration     types.String `tfsdk:"delay_duration"`
	DelayAmount       types.String `tfsdk:"delay_amount"`
	WorkflowOnly      types.String `tfsdk:"workflow_only"`

	// Variables
	Variables types.List `tfsdk:"variables"`

	// Resource management
	HoldResources    types.Bool `tfsdk:"hold_resources"`
	ExclusiveTasks   types.List `tfsdk:"exclusive_tasks"`
	VirtualResources types.List `tfsdk:"virtual_resources"`

	// Actions
	Actions types.Object `tfsdk:"actions"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`

	// ===========================================
	// Generic Universal Template field slots
	// ===========================================

	TextField1  types.String `tfsdk:"text_field_1"`
	TextField2  types.String `tfsdk:"text_field_2"`
	TextField3  types.String `tfsdk:"text_field_3"`
	TextField4  types.String `tfsdk:"text_field_4"`
	TextField5  types.String `tfsdk:"text_field_5"`
	TextField6  types.String `tfsdk:"text_field_6"`
	TextField7  types.String `tfsdk:"text_field_7"`
	TextField8  types.String `tfsdk:"text_field_8"`
	TextField9  types.String `tfsdk:"text_field_9"`
	TextField10 types.String `tfsdk:"text_field_10"`

	ChoiceField1  types.String `tfsdk:"choice_field_1"`
	ChoiceField2  types.String `tfsdk:"choice_field_2"`
	ChoiceField3  types.String `tfsdk:"choice_field_3"`
	ChoiceField4  types.String `tfsdk:"choice_field_4"`
	ChoiceField5  types.String `tfsdk:"choice_field_5"`
	ChoiceField6  types.String `tfsdk:"choice_field_6"`
	ChoiceField7  types.String `tfsdk:"choice_field_7"`
	ChoiceField8  types.String `tfsdk:"choice_field_8"`
	ChoiceField9  types.String `tfsdk:"choice_field_9"`
	ChoiceField10 types.String `tfsdk:"choice_field_10"`
	ChoiceField11 types.String `tfsdk:"choice_field_11"`

	BooleanField1 types.Bool `tfsdk:"boolean_field_1"`
	BooleanField2 types.Bool `tfsdk:"boolean_field_2"`
	BooleanField3 types.Bool `tfsdk:"boolean_field_3"`
	BooleanField4 types.Bool `tfsdk:"boolean_field_4"`
	BooleanField5 types.Bool `tfsdk:"boolean_field_5"`
	BooleanField6 types.Bool `tfsdk:"boolean_field_6"`
	BooleanField7 types.Bool `tfsdk:"boolean_field_7"`

	CredentialField1 types.String `tfsdk:"credential_field_1"`
	CredentialField2 types.String `tfsdk:"credential_field_2"`
	CredentialField3 types.String `tfsdk:"credential_field_3"`
	CredentialField4 types.String `tfsdk:"credential_field_4"`

	CredentialVarField1 types.String `tfsdk:"credential_var_field_1"`
	CredentialVarField4 types.String `tfsdk:"credential_var_field_4"`

	CustomField1 types.String `tfsdk:"custom_field_1"`
	CustomField2 types.String `tfsdk:"custom_field_2"`

	IntField1 types.Int64 `tfsdk:"int_field_1"`

	LargeTextField1 types.String `tfsdk:"large_text_field_1"`
	LargeTextField2 types.String `tfsdk:"large_text_field_2"`
	LargeTextField3 types.String `tfsdk:"large_text_field_3"`
	LargeTextField4 types.String `tfsdk:"large_text_field_4"`

	ScriptField1 types.String `tfsdk:"script_field_1"`
	ScriptField2 types.String `tfsdk:"script_field_2"`

	ScriptVarField2 types.String `tfsdk:"script_var_field_2"`
}

// TaskUniversalAPIModel represents the API request/response structure.
type TaskUniversalAPIModel struct {
	SysId    string `json:"sysId,omitempty"`
	Name     string `json:"name"`
	Version  int64  `json:"version,omitempty"`
	Type     string `json:"type"`
	Template string `json:"template"`
	Summary  string `json:"summary,omitempty"`

	Agent           string `json:"agent,omitempty"`
	AgentCluster    string `json:"agentCluster,omitempty"`
	AgentVar        string `json:"agentVar,omitempty"`
	AgentClusterVar string `json:"agentClusterVar,omitempty"`

	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

	ExitCodes          string `json:"exitCodes,omitempty"`
	ExitCodeProcessing string `json:"exitCodeProcessing,omitempty"`
	ExitCodeText       string `json:"exitCodeText,omitempty"`

	RetryMaximum         int64 `json:"retryMaximum,omitempty"`
	RetryIndefinitely    bool  `json:"retryIndefinitely,omitempty"`
	RetryInterval        int64 `json:"retryInterval,omitempty"`
	RetrySuppressFailure bool  `json:"retrySuppressFailure,omitempty"`

	// Wait/Delay options
	WaitToStart       string `json:"twWaitType,omitempty"`
	WaitTime          string `json:"twWaitTime,omitempty"`
	WaitDuration      string `json:"twWaitDuration,omitempty"`
	WaitAmount        string `json:"twWaitAmount,omitempty"`
	WaitDayConstraint string `json:"twWaitDayConstraint,omitempty"`
	DelayOnStart      string `json:"twDelayType,omitempty"`
	DelayDuration     string `json:"twDelayDuration,omitempty"`
	DelayAmount       string `json:"twDelayAmount,omitempty"`
	WorkflowOnly      string `json:"twWorkflowOnly,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources    bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks   []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	VirtualResources []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`

	// Generic Universal Template field slots
	TextField1  *UniversalFieldWsData `json:"textField1,omitempty"`
	TextField2  *UniversalFieldWsData `json:"textField2,omitempty"`
	TextField3  *UniversalFieldWsData `json:"textField3,omitempty"`
	TextField4  *UniversalFieldWsData `json:"textField4,omitempty"`
	TextField5  *UniversalFieldWsData `json:"textField5,omitempty"`
	TextField6  *UniversalFieldWsData `json:"textField6,omitempty"`
	TextField7  *UniversalFieldWsData `json:"textField7,omitempty"`
	TextField8  *UniversalFieldWsData `json:"textField8,omitempty"`
	TextField9  *UniversalFieldWsData `json:"textField9,omitempty"`
	TextField10 *UniversalFieldWsData `json:"textField10,omitempty"`

	ChoiceField1  *UniversalFieldWsData `json:"choiceField1,omitempty"`
	ChoiceField2  *UniversalFieldWsData `json:"choiceField2,omitempty"`
	ChoiceField3  *UniversalFieldWsData `json:"choiceField3,omitempty"`
	ChoiceField4  *UniversalFieldWsData `json:"choiceField4,omitempty"`
	ChoiceField5  *UniversalFieldWsData `json:"choiceField5,omitempty"`
	ChoiceField6  *UniversalFieldWsData `json:"choiceField6,omitempty"`
	ChoiceField7  *UniversalFieldWsData `json:"choiceField7,omitempty"`
	ChoiceField8  *UniversalFieldWsData `json:"choiceField8,omitempty"`
	ChoiceField9  *UniversalFieldWsData `json:"choiceField9,omitempty"`
	ChoiceField10 *UniversalFieldWsData `json:"choiceField10,omitempty"`
	ChoiceField11 *UniversalFieldWsData `json:"choiceField11,omitempty"`

	BooleanField1 *UniversalFieldWsData `json:"booleanField1,omitempty"`
	BooleanField2 *UniversalFieldWsData `json:"booleanField2,omitempty"`
	BooleanField3 *UniversalFieldWsData `json:"booleanField3,omitempty"`
	BooleanField4 *UniversalFieldWsData `json:"booleanField4,omitempty"`
	BooleanField5 *UniversalFieldWsData `json:"booleanField5,omitempty"`
	BooleanField6 *UniversalFieldWsData `json:"booleanField6,omitempty"`
	BooleanField7 *UniversalFieldWsData `json:"booleanField7,omitempty"`

	CredentialField1 *UniversalFieldWsData `json:"credentialField1,omitempty"`
	CredentialField2 *UniversalFieldWsData `json:"credentialField2,omitempty"`
	CredentialField3 *UniversalFieldWsData `json:"credentialField3,omitempty"`
	CredentialField4 *UniversalFieldWsData `json:"credentialField4,omitempty"`

	CredentialVarField1 *UniversalFieldWsData `json:"credentialVarField1,omitempty"`
	CredentialVarField4 *UniversalFieldWsData `json:"credentialVarField4,omitempty"`

	CustomField1 *UniversalFieldWsData `json:"customField1,omitempty"`
	CustomField2 *UniversalFieldWsData `json:"customField2,omitempty"`

	IntField1 *UniversalFieldWsData `json:"intField1,omitempty"`

	LargeTextField1 *UniversalFieldWsData `json:"largeTextField1,omitempty"`
	LargeTextField2 *UniversalFieldWsData `json:"largeTextField2,omitempty"`
	LargeTextField3 *UniversalFieldWsData `json:"largeTextField3,omitempty"`
	LargeTextField4 *UniversalFieldWsData `json:"largeTextField4,omitempty"`

	ScriptField1 *UniversalFieldWsData `json:"scriptField1,omitempty"`
	ScriptField2 *UniversalFieldWsData `json:"scriptField2,omitempty"`

	ScriptVarField2 *UniversalFieldWsData `json:"scriptVarField2,omitempty"`
}

func (r *TaskUniversalResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_universal"
}

func genericTextFieldSchema(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Optional:            true,
	}
}

func genericIntFieldSchema(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: desc,
		Optional:            true,
	}
}

func genericBoolFieldSchema(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Optional:            true,
	}
}

func (r *TaskUniversalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a generic StoneBranch Universal Task. Unlike template-specific resources " +
			"(e.g. `stonebranch_task_universal_aws_s3`), this resource exposes UAC's raw numbered Universal " +
			"Template field slots directly (`text_field_1`, `choice_field_1`, etc.), the same way UAC itself " +
			"represents any Universal-Template-backed task before applying template-specific field labels. " +
			"Use this resource for Universal Templates that don't have a dedicated typed resource.",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the task (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the task.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the task (for optimistic locking).",
				Computed:            true,
			},

			// Basic info
			"summary": schema.StringAttribute{
				MarkdownDescription: "Description/summary of the task.",
				Optional:            true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Name of the Universal Template this task is based on (see `stonebranch_universal_template`).",
				Required:            true,
			},

			// Agent configuration
			"agent": schema.StringAttribute{
				MarkdownDescription: "Name of the agent to run the task on.",
				Optional:            true,
			},
			"agent_cluster": schema.StringAttribute{
				MarkdownDescription: "Name of the agent cluster to run the task on.",
				Optional:            true,
			},
			"agent_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the agent name.",
				Optional:            true,
			},
			"agent_cluster_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the agent cluster name.",
				Optional:            true,
			},

			// Credentials (for task execution)
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use for task execution.",
				Optional:            true,
			},
			"credentials_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the credentials name.",
				Optional:            true,
			},

			// Exit code handling
			"exit_codes": schema.StringAttribute{
				MarkdownDescription: "Exit codes that indicate success (e.g., '0' or '0,1,2'). Defaults to '0'.",
				Optional:            true,
				Computed:            true,
			},
			"exit_code_processing": schema.StringAttribute{
				MarkdownDescription: "How to process exit codes. Values: 'Success Exitcode Range', 'Failure Exitcode Range'.",
				Optional:            true,
				Computed:            true,
			},
			"exit_code_text": schema.StringAttribute{
				MarkdownDescription: "Text/pattern to scan output for (UAC's 'Scan Output For' field). Required by the API when exit_code_processing is an 'Output Contains' mode.",
				Optional:            true,
			},

			// Retry configuration
			"retry_maximum": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retry attempts.",
				Optional:            true,
				Computed:            true,
			},
			"retry_indefinitely": schema.BoolAttribute{
				MarkdownDescription: "Whether to retry indefinitely on failure.",
				Optional:            true,
				Computed:            true,
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Interval between retry attempts (in seconds).",
				Optional:            true,
				Computed:            true,
			},
			"retry_suppress_failure": schema.BoolAttribute{
				MarkdownDescription: "Whether to suppress failure notifications during retries.",
				Optional:            true,
				Computed:            true,
			},

			// Wait/Delay options
			"wait_to_start":       TaskWaitToStartSchema(),
			"wait_time":           TaskWaitTimeSchema(),
			"wait_duration":       TaskWaitDurationSchema(),
			"wait_amount":         TaskWaitAmountSchema(),
			"wait_day_constraint": TaskWaitDayConstraintSchema(),
			"delay_on_start":      TaskDelayOnStartSchema(),
			"delay_duration":      TaskDelayDurationSchema(),
			"delay_amount":        TaskDelayAmountSchema(),
			"workflow_only":       TaskWorkflowOnlySchema(),

			// Variables
			"variables": TaskVariablesSchema(),

			// Resource management
			"hold_resources": schema.BoolAttribute{
				MarkdownDescription: "Whether to hold the task's virtual resources for the duration of any retries.",
				Optional:            true,
				Computed:            true,
			},
			"exclusive_tasks":   TaskExclusiveTasksSchema(),
			"virtual_resources": TaskVirtualResourcesSchema(),

			// Actions
			"actions": TaskActionsSchema(),

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this task belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},

			// ===========================================
			// Generic Universal Template field slots
			// ===========================================

			"text_field_1":  genericTextFieldSchema("Value of the Universal Template's Text Field 1 slot."),
			"text_field_2":  genericTextFieldSchema("Value of the Universal Template's Text Field 2 slot."),
			"text_field_3":  genericTextFieldSchema("Value of the Universal Template's Text Field 3 slot."),
			"text_field_4":  genericTextFieldSchema("Value of the Universal Template's Text Field 4 slot."),
			"text_field_5":  genericTextFieldSchema("Value of the Universal Template's Text Field 5 slot."),
			"text_field_6":  genericTextFieldSchema("Value of the Universal Template's Text Field 6 slot."),
			"text_field_7":  genericTextFieldSchema("Value of the Universal Template's Text Field 7 slot."),
			"text_field_8":  genericTextFieldSchema("Value of the Universal Template's Text Field 8 slot."),
			"text_field_9":  genericTextFieldSchema("Value of the Universal Template's Text Field 9 slot."),
			"text_field_10": genericTextFieldSchema("Value of the Universal Template's Text Field 10 slot."),

			"choice_field_1":  genericTextFieldSchema("Value of the Universal Template's Choice Field 1 slot."),
			"choice_field_2":  genericTextFieldSchema("Value of the Universal Template's Choice Field 2 slot."),
			"choice_field_3":  genericTextFieldSchema("Value of the Universal Template's Choice Field 3 slot."),
			"choice_field_4":  genericTextFieldSchema("Value of the Universal Template's Choice Field 4 slot."),
			"choice_field_5":  genericTextFieldSchema("Value of the Universal Template's Choice Field 5 slot."),
			"choice_field_6":  genericTextFieldSchema("Value of the Universal Template's Choice Field 6 slot."),
			"choice_field_7":  genericTextFieldSchema("Value of the Universal Template's Choice Field 7 slot."),
			"choice_field_8":  genericTextFieldSchema("Value of the Universal Template's Choice Field 8 slot."),
			"choice_field_9":  genericTextFieldSchema("Value of the Universal Template's Choice Field 9 slot."),
			"choice_field_10": genericTextFieldSchema("Value of the Universal Template's Choice Field 10 slot."),
			"choice_field_11": genericTextFieldSchema("Value of the Universal Template's Choice Field 11 slot."),

			"boolean_field_1": genericBoolFieldSchema("Value of the Universal Template's Boolean Field 1 slot."),
			"boolean_field_2": genericBoolFieldSchema("Value of the Universal Template's Boolean Field 2 slot."),
			"boolean_field_3": genericBoolFieldSchema("Value of the Universal Template's Boolean Field 3 slot."),
			"boolean_field_4": genericBoolFieldSchema("Value of the Universal Template's Boolean Field 4 slot."),
			"boolean_field_5": genericBoolFieldSchema("Value of the Universal Template's Boolean Field 5 slot."),
			"boolean_field_6": genericBoolFieldSchema("Value of the Universal Template's Boolean Field 6 slot."),
			"boolean_field_7": genericBoolFieldSchema("Value of the Universal Template's Boolean Field 7 slot."),

			"credential_field_1": genericTextFieldSchema("Name of the credential in the Universal Template's Credential Field 1 slot."),
			"credential_field_2": genericTextFieldSchema("Name of the credential in the Universal Template's Credential Field 2 slot."),
			"credential_field_3": genericTextFieldSchema("Name of the credential in the Universal Template's Credential Field 3 slot."),
			"credential_field_4": genericTextFieldSchema("Name of the credential in the Universal Template's Credential Field 4 slot."),

			"credential_var_field_1": genericTextFieldSchema("Variable containing the credential name for the Universal Template's Credential Field 1 slot."),
			"credential_var_field_4": genericTextFieldSchema("Variable containing the credential name for the Universal Template's Credential Field 4 slot."),

			"custom_field_1": genericTextFieldSchema("Value of the Universal Template's Custom Field 1 slot."),
			"custom_field_2": genericTextFieldSchema("Value of the Universal Template's Custom Field 2 slot."),

			"int_field_1": genericIntFieldSchema("Value of the Universal Template's Integer Field 1 slot."),

			"large_text_field_1": genericTextFieldSchema("Value of the Universal Template's Large Text Field 1 slot."),
			"large_text_field_2": genericTextFieldSchema("Value of the Universal Template's Large Text Field 2 slot."),
			"large_text_field_3": genericTextFieldSchema("Value of the Universal Template's Large Text Field 3 slot."),
			"large_text_field_4": genericTextFieldSchema("Value of the Universal Template's Large Text Field 4 slot."),

			"script_field_1": genericTextFieldSchema("Name of the script in the Universal Template's Script Field 1 slot."),
			"script_field_2": genericTextFieldSchema("Name of the script in the Universal Template's Script Field 2 slot."),

			"script_var_field_2": genericTextFieldSchema("Variable containing the script name for the Universal Template's Script Field 2 slot."),
		},
	}
}

func (r *TaskUniversalResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TaskUniversalResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(ValidateTaskWaitDelay(data.WaitToStart, data.WaitAmount, data.DelayOnStart, data.DelayAmount)...)
}

func (r *TaskUniversalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskUniversalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskUniversalResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating universal task", map[string]any{"name": data.Name.ValueString()})

	apiModel := r.toAPIModel(ctx, &data)

	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Task",
			fmt.Sprintf("Could not create task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Task",
			fmt.Sprintf("Could not read task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created universal task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskUniversalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskUniversalResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Task",
			fmt.Sprintf("Could not read task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskUniversalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskUniversalResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TaskUniversalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating universal task", map[string]any{"sys_id": data.SysId.ValueString()})

	apiModel := r.toAPIModel(ctx, &data)

	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Task",
			fmt.Sprintf("Could not update task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Task",
			fmt.Sprintf("Could not read task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskUniversalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskUniversalResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting universal task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Task",
			fmt.Sprintf("Could not delete task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskUniversalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskUniversalResource) readTask(ctx context.Context, data *TaskUniversalResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskUniversalAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// genericStringField builds a *UniversalFieldWsData from a Terraform string
// attribute, or nil if the attribute is unset. The slot's "name"/"label" are
// intentionally left blank: they are template-defined metadata that UAC
// derives from the bound Universal Template, not something this generic
// resource can know in advance.
func genericStringField(value types.String) *UniversalFieldWsData {
	if !isSet(value) {
		return nil
	}
	return &UniversalFieldWsData{Value: value.ValueString()}
}

// genericBoolField builds a *UniversalFieldWsData from a Terraform bool attribute.
func genericBoolField(value types.Bool) *UniversalFieldWsData {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &UniversalFieldWsData{Value: value.ValueBool()}
}

// genericIntField builds a *UniversalFieldWsData from a Terraform int64 attribute.
func genericIntField(value types.Int64) *UniversalFieldWsData {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &UniversalFieldWsData{Value: value.ValueInt64()}
}

// getFieldInt64Value extracts an int64 value from a universal field.
func getFieldInt64Value(field *UniversalFieldWsData) (int64, bool) {
	if field == nil {
		return 0, false
	}
	switch v := field.Value.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	}
	return 0, false
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskUniversalResource) toAPIModel(ctx context.Context, data *TaskUniversalResourceModel) *TaskUniversalAPIModel {
	model := &TaskUniversalAPIModel{
		SysId:    data.SysId.ValueString(),
		Name:     data.Name.ValueString(),
		Type:     "taskUniversal",
		Template: data.Template.ValueString(),
		Summary:  data.Summary.ValueString(),

		Agent:           data.Agent.ValueString(),
		AgentCluster:    data.AgentCluster.ValueString(),
		AgentVar:        data.AgentVar.ValueString(),
		AgentClusterVar: data.AgentClusterVar.ValueString(),

		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),

		ExitCodes:          StringValueOrDefault(data.ExitCodes, "0"),
		ExitCodeProcessing: data.ExitCodeProcessing.ValueString(),
		ExitCodeText:       data.ExitCodeText.ValueString(),

		RetryMaximum:         data.RetryMaximum.ValueInt64(),
		RetryIndefinitely:    data.RetryIndefinitely.ValueBool(),
		RetryInterval:        data.RetryInterval.ValueInt64(),
		RetrySuppressFailure: data.RetrySuppressFailure.ValueBool(),

		// Wait/Delay options
		WaitToStart:       data.WaitToStart.ValueString(),
		WaitTime:          data.WaitTime.ValueString(),
		WaitDuration:      data.WaitDuration.ValueString(),
		WaitAmount:        data.WaitAmount.ValueString(),
		WaitDayConstraint: data.WaitDayConstraint.ValueString(),
		DelayOnStart:      data.DelayOnStart.ValueString(),
		DelayDuration:     data.DelayDuration.ValueString(),
		DelayAmount:       data.DelayAmount.ValueString(),
		WorkflowOnly:      data.WorkflowOnly.ValueString(),

		TextField1:  genericStringField(data.TextField1),
		TextField2:  genericStringField(data.TextField2),
		TextField3:  genericStringField(data.TextField3),
		TextField4:  genericStringField(data.TextField4),
		TextField5:  genericStringField(data.TextField5),
		TextField6:  genericStringField(data.TextField6),
		TextField7:  genericStringField(data.TextField7),
		TextField8:  genericStringField(data.TextField8),
		TextField9:  genericStringField(data.TextField9),
		TextField10: genericStringField(data.TextField10),

		ChoiceField1:  genericStringField(data.ChoiceField1),
		ChoiceField2:  genericStringField(data.ChoiceField2),
		ChoiceField3:  genericStringField(data.ChoiceField3),
		ChoiceField4:  genericStringField(data.ChoiceField4),
		ChoiceField5:  genericStringField(data.ChoiceField5),
		ChoiceField6:  genericStringField(data.ChoiceField6),
		ChoiceField7:  genericStringField(data.ChoiceField7),
		ChoiceField8:  genericStringField(data.ChoiceField8),
		ChoiceField9:  genericStringField(data.ChoiceField9),
		ChoiceField10: genericStringField(data.ChoiceField10),
		ChoiceField11: genericStringField(data.ChoiceField11),

		BooleanField1: genericBoolField(data.BooleanField1),
		BooleanField2: genericBoolField(data.BooleanField2),
		BooleanField3: genericBoolField(data.BooleanField3),
		BooleanField4: genericBoolField(data.BooleanField4),
		BooleanField5: genericBoolField(data.BooleanField5),
		BooleanField6: genericBoolField(data.BooleanField6),
		BooleanField7: genericBoolField(data.BooleanField7),

		CredentialField1: genericStringField(data.CredentialField1),
		CredentialField2: genericStringField(data.CredentialField2),
		CredentialField3: genericStringField(data.CredentialField3),
		CredentialField4: genericStringField(data.CredentialField4),

		CredentialVarField1: genericStringField(data.CredentialVarField1),
		CredentialVarField4: genericStringField(data.CredentialVarField4),

		CustomField1: genericStringField(data.CustomField1),
		CustomField2: genericStringField(data.CustomField2),

		IntField1: genericIntField(data.IntField1),

		LargeTextField1: genericStringField(data.LargeTextField1),
		LargeTextField2: genericStringField(data.LargeTextField2),
		LargeTextField3: genericStringField(data.LargeTextField3),
		LargeTextField4: genericStringField(data.LargeTextField4),

		ScriptField1: genericStringField(data.ScriptField1),
		ScriptField2: genericStringField(data.ScriptField2),

		ScriptVarField2: genericStringField(data.ScriptVarField2),
	}

	// Handle variables
	model.Variables = TaskVariablesToAPI(ctx, data.Variables)

	// Handle resource management fields
	if !data.HoldResources.IsNull() && !data.HoldResources.IsUnknown() {
		model.HoldResources = data.HoldResources.ValueBool()
	}
	model.ExclusiveTasks = TaskExclusiveTasksToAPI(ctx, data.ExclusiveTasks)
	model.VirtualResources = TaskVirtualResourcesToAPI(ctx, data.VirtualResources)

	// Handle actions
	model.Actions = TaskActionsToAPI(ctx, data.Actions)

	// Handle opswise_groups list
	if !data.OpswiseGroups.IsNull() && !data.OpswiseGroups.IsUnknown() {
		var groups []string
		data.OpswiseGroups.ElementsAs(ctx, &groups, false)
		model.OpswiseGroups = groups
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *TaskUniversalResource) fromAPIModel(ctx context.Context, apiModel *TaskUniversalAPIModel, data *TaskUniversalResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Optional fields
	data.Summary = StringValueOrNull(apiModel.Summary)
	data.Template = types.StringValue(apiModel.Template)
	data.Agent = StringValueOrNull(apiModel.Agent)
	data.AgentCluster = StringValueOrNull(apiModel.AgentCluster)
	data.AgentVar = StringValueOrNull(apiModel.AgentVar)
	data.AgentClusterVar = StringValueOrNull(apiModel.AgentClusterVar)
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)
	data.ExitCodeProcessing = StringValueOrNull(apiModel.ExitCodeProcessing)
	data.ExitCodeText = StringValueOrNull(apiModel.ExitCodeText)

	// Computed fields
	data.RetryMaximum = types.Int64Value(apiModel.RetryMaximum)
	data.RetryIndefinitely = types.BoolValue(apiModel.RetryIndefinitely)
	data.RetryInterval = types.Int64Value(apiModel.RetryInterval)
	data.RetrySuppressFailure = types.BoolValue(apiModel.RetrySuppressFailure)

	// Wait/Delay options - always returned by API
	data.WaitToStart = types.StringValue(apiModel.WaitToStart)
	data.WaitTime = types.StringValue(apiModel.WaitTime)
	data.WaitDuration = types.StringValue(apiModel.WaitDuration)
	data.WaitAmount = types.StringValue(apiModel.WaitAmount)
	data.WaitDayConstraint = types.StringValue(apiModel.WaitDayConstraint)
	data.DelayOnStart = types.StringValue(apiModel.DelayOnStart)
	data.DelayDuration = types.StringValue(apiModel.DelayDuration)
	data.DelayAmount = types.StringValue(apiModel.DelayAmount)
	data.WorkflowOnly = types.StringValue(apiModel.WorkflowOnly)

	// Text fields
	data.TextField1 = StringValueOrNull(getFieldStringValue(apiModel.TextField1))
	data.TextField2 = StringValueOrNull(getFieldStringValue(apiModel.TextField2))
	data.TextField3 = StringValueOrNull(getFieldStringValue(apiModel.TextField3))
	data.TextField4 = StringValueOrNull(getFieldStringValue(apiModel.TextField4))
	data.TextField5 = StringValueOrNull(getFieldStringValue(apiModel.TextField5))
	data.TextField6 = StringValueOrNull(getFieldStringValue(apiModel.TextField6))
	data.TextField7 = StringValueOrNull(getFieldStringValue(apiModel.TextField7))
	data.TextField8 = StringValueOrNull(getFieldStringValue(apiModel.TextField8))
	data.TextField9 = StringValueOrNull(getFieldStringValue(apiModel.TextField9))
	data.TextField10 = StringValueOrNull(getFieldStringValue(apiModel.TextField10))

	// Choice fields
	data.ChoiceField1 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField1))
	data.ChoiceField2 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField2))
	data.ChoiceField3 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField3))
	data.ChoiceField4 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField4))
	data.ChoiceField5 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField5))
	data.ChoiceField6 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField6))
	data.ChoiceField7 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField7))
	data.ChoiceField8 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField8))
	data.ChoiceField9 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField9))
	data.ChoiceField10 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField10))
	data.ChoiceField11 = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField11))

	// Boolean fields
	data.BooleanField1 = boolFromAPIField(apiModel.BooleanField1)
	data.BooleanField2 = boolFromAPIField(apiModel.BooleanField2)
	data.BooleanField3 = boolFromAPIField(apiModel.BooleanField3)
	data.BooleanField4 = boolFromAPIField(apiModel.BooleanField4)
	data.BooleanField5 = boolFromAPIField(apiModel.BooleanField5)
	data.BooleanField6 = boolFromAPIField(apiModel.BooleanField6)
	data.BooleanField7 = boolFromAPIField(apiModel.BooleanField7)

	// Credential fields
	data.CredentialField1 = StringValueOrNull(getFieldStringValue(apiModel.CredentialField1))
	data.CredentialField2 = StringValueOrNull(getFieldStringValue(apiModel.CredentialField2))
	data.CredentialField3 = StringValueOrNull(getFieldStringValue(apiModel.CredentialField3))
	data.CredentialField4 = StringValueOrNull(getFieldStringValue(apiModel.CredentialField4))

	data.CredentialVarField1 = StringValueOrNull(getFieldStringValue(apiModel.CredentialVarField1))
	data.CredentialVarField4 = StringValueOrNull(getFieldStringValue(apiModel.CredentialVarField4))

	// Custom fields
	data.CustomField1 = StringValueOrNull(getFieldStringValue(apiModel.CustomField1))
	data.CustomField2 = StringValueOrNull(getFieldStringValue(apiModel.CustomField2))

	// Integer field
	if v, ok := getFieldInt64Value(apiModel.IntField1); ok {
		data.IntField1 = types.Int64Value(v)
	} else {
		data.IntField1 = types.Int64Null()
	}

	// Large text fields
	data.LargeTextField1 = StringValueOrNull(getFieldStringValue(apiModel.LargeTextField1))
	data.LargeTextField2 = StringValueOrNull(getFieldStringValue(apiModel.LargeTextField2))
	data.LargeTextField3 = StringValueOrNull(getFieldStringValue(apiModel.LargeTextField3))
	data.LargeTextField4 = StringValueOrNull(getFieldStringValue(apiModel.LargeTextField4))

	// Script fields
	data.ScriptField1 = StringValueOrNull(getFieldStringValue(apiModel.ScriptField1))
	data.ScriptField2 = StringValueOrNull(getFieldStringValue(apiModel.ScriptField2))

	data.ScriptVarField2 = StringValueOrNull(getFieldStringValue(apiModel.ScriptVarField2))

	// Handle variables
	data.Variables = TaskVariablesFromAPIOrdered(ctx, apiModel.Variables, data.Variables)

	// Handle resource management fields
	data.HoldResources = types.BoolValue(apiModel.HoldResources)
	data.ExclusiveTasks = TaskExclusiveTasksFromAPI(apiModel.ExclusiveTasks)
	data.VirtualResources = TaskVirtualResourcesFromAPI(apiModel.VirtualResources)

	// Handle actions
	data.Actions = TaskActionsFromAPI(ctx, apiModel.Actions, data.Actions)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}

// boolFromAPIField converts a universal field to a Terraform bool, defaulting to null if unset.
func boolFromAPIField(field *UniversalFieldWsData) types.Bool {
	if v, ok := getFieldBoolValue(field); ok {
		return types.BoolValue(v)
	}
	return types.BoolNull()
}
