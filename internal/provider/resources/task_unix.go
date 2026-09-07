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
	_ resource.Resource                = &TaskUnixResource{}
	_ resource.ResourceWithImportState = &TaskUnixResource{}
)

func NewTaskUnixResource() resource.Resource {
	return &TaskUnixResource{}
}

// TaskUnixResource defines the resource implementation.
type TaskUnixResource struct {
	client *client.Client
}

// TaskUnixResourceModel describes the resource data model.
type TaskUnixResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Agent configuration
	Agent           types.String `tfsdk:"agent"`
	AgentCluster    types.String `tfsdk:"agent_cluster"`
	AgentVar        types.String `tfsdk:"agent_var"`
	AgentClusterVar types.String `tfsdk:"agent_cluster_var"`

	// Command configuration
	Command         types.String `tfsdk:"command"`
	CommandOrScript types.String `tfsdk:"command_or_script"`
	Script          types.String `tfsdk:"script"`
	RuntimeDir      types.String `tfsdk:"runtime_dir"`
	Parameters      types.String `tfsdk:"parameters"`

	// Credentials
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

	// Exit code handling
	ExitCodes          types.String `tfsdk:"exit_codes"`
	ExitCodeProcessing types.String `tfsdk:"exit_code_processing"`

	// Output handling
	OutputType        types.String `tfsdk:"output_type"`
	WaitForOutput     types.Bool   `tfsdk:"wait_for_output"`
	OutputReturnFile  types.String `tfsdk:"output_return_file"`
	OutputReturnType  types.String `tfsdk:"output_return_type"`
	OutputReturnSline types.String `tfsdk:"output_return_sline"`
	OutputReturnNline types.String `tfsdk:"output_return_nline"`

	// Retry configuration
	RetryMaximum         types.Int64 `tfsdk:"retry_maximum"`
	RetryIndefinitely    types.Bool  `tfsdk:"retry_indefinitely"`
	RetryInterval        types.Int64 `tfsdk:"retry_interval"`
	RetrySuppressFailure types.Bool  `tfsdk:"retry_suppress_failure"`

	// Unix-specific
	RunAsSudo types.Bool `tfsdk:"run_as_sudo"`

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
}

// TaskAPIModel represents the API request/response structure.
type TaskAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	Agent           string `json:"agent,omitempty"`
	AgentCluster    string `json:"agentCluster,omitempty"`
	AgentVar        string `json:"agentVar,omitempty"`
	AgentClusterVar string `json:"agentClusterVar,omitempty"`

	Command         string `json:"command,omitempty"`
	CommandOrScript string `json:"commandOrScript,omitempty"`
	Script          string `json:"script,omitempty"`
	RuntimeDir      string `json:"runtimeDir,omitempty"`
	Parameters      string `json:"parameters,omitempty"`

	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

	ExitCodes          string `json:"exitCodes,omitempty"`
	ExitCodeProcessing string `json:"exitCodeProcessing,omitempty"`

	OutputType        string `json:"outputType,omitempty"`
	WaitForOutput     bool   `json:"waitForOutput,omitempty"`
	OutputReturnFile  string `json:"outputReturnFile,omitempty"`
	OutputReturnType  string `json:"outputReturnType,omitempty"`
	OutputReturnSline string `json:"outputReturnSline,omitempty"`
	OutputReturnNline string `json:"outputReturnNline,omitempty"`

	RetryMaximum         int64 `json:"retryMaximum,omitempty"`
	RetryIndefinitely    bool  `json:"retryIndefinitely,omitempty"`
	RetryInterval        int64 `json:"retryInterval,omitempty"`
	RetrySuppressFailure bool  `json:"retrySuppressFailure,omitempty"`

	RunAsSudo bool `json:"runAsSudo,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources    bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks   []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	VirtualResources []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TaskUnixResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_unix"
}

func (r *TaskUnixResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Task (Unix/Linux task type).",

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

			// Command configuration
			"command": schema.StringAttribute{
				MarkdownDescription: "Command to execute (when command_or_script is 'Command').",
				Optional:            true,
			},
			"command_or_script": schema.StringAttribute{
				MarkdownDescription: "Whether to run a command or script. Values: 'Command', 'Script'.",
				Optional:            true,
				Computed:            true,
			},
			"script": schema.StringAttribute{
				MarkdownDescription: "Name of the Script resource to execute (when command_or_script is 'Script').",
				Optional:            true,
			},
			"runtime_dir": schema.StringAttribute{
				MarkdownDescription: "Working directory for the task execution.",
				Optional:            true,
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "Parameters to pass to the command or script.",
				Optional:            true,
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

			// Output handling
			"output_type": schema.StringAttribute{
				MarkdownDescription: "Type of output to capture. Values: 'STDOUT', 'STDERR', 'FILE', 'OUTERR'.",
				Optional:            true,
				Computed:            true,
			},
			"wait_for_output": schema.BoolAttribute{
				MarkdownDescription: "Whether to wait for output before completing.",
				Optional:            true,
				Computed:            true,
			},
			"output_return_file": schema.StringAttribute{
				MarkdownDescription: "File to capture output from.",
				Optional:            true,
				Computed:            true,
			},
			"output_return_type": schema.StringAttribute{
				MarkdownDescription: "Type of output to return.",
				Optional:            true,
				Computed:            true,
			},
			"output_return_sline": schema.StringAttribute{
				MarkdownDescription: "Starting line for output capture.",
				Optional:            true,
				Computed:            true,
			},
			"output_return_nline": schema.StringAttribute{
				MarkdownDescription: "Number of lines to capture.",
				Optional:            true,
				Computed:            true,
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

			// Unix-specific
			"run_as_sudo": schema.BoolAttribute{
				MarkdownDescription: "Whether to run the command with sudo.",
				Optional:            true,
				Computed:            true,
			},

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
		},
	}
}

func (r *TaskUnixResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskUnixResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskUnixResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Task",
			fmt.Sprintf("Could not create task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Task",
			fmt.Sprintf("Could not read task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskUnixResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskUnixResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
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

func (r *TaskUnixResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskUnixResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskUnixResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Task",
			fmt.Sprintf("Could not update task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
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

func (r *TaskUnixResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskUnixResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
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

func (r *TaskUnixResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskUnixResource) readTask(ctx context.Context, data *TaskUnixResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskUnixResource) toAPIModel(ctx context.Context, data *TaskUnixResourceModel) *TaskAPIModel {
	model := &TaskAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskUnix",
		Summary: data.Summary.ValueString(),

		Agent:           data.Agent.ValueString(),
		AgentCluster:    data.AgentCluster.ValueString(),
		AgentVar:        data.AgentVar.ValueString(),
		AgentClusterVar: data.AgentClusterVar.ValueString(),

		Command:         data.Command.ValueString(),
		CommandOrScript: data.CommandOrScript.ValueString(),
		Script:          data.Script.ValueString(),
		RuntimeDir:      data.RuntimeDir.ValueString(),
		Parameters:      data.Parameters.ValueString(),

		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),

		ExitCodes:          StringValueOrDefault(data.ExitCodes, "0"),
		ExitCodeProcessing: data.ExitCodeProcessing.ValueString(),

		OutputType:        data.OutputType.ValueString(),
		WaitForOutput:     data.WaitForOutput.ValueBool(),
		OutputReturnFile:  data.OutputReturnFile.ValueString(),
		OutputReturnType:  data.OutputReturnType.ValueString(),
		OutputReturnSline: data.OutputReturnSline.ValueString(),
		OutputReturnNline: data.OutputReturnNline.ValueString(),

		RetryMaximum:         data.RetryMaximum.ValueInt64(),
		RetryIndefinitely:    data.RetryIndefinitely.ValueBool(),
		RetryInterval:        data.RetryInterval.ValueInt64(),
		RetrySuppressFailure: data.RetrySuppressFailure.ValueBool(),

		RunAsSudo: data.RunAsSudo.ValueBool(),
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
func (r *TaskUnixResource) fromAPIModel(ctx context.Context, apiModel *TaskAPIModel, data *TaskUnixResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Optional fields - only set if non-empty (these are truly optional)
	data.Summary = StringValueOrNull(apiModel.Summary)
	data.Agent = StringValueOrNull(apiModel.Agent)
	data.AgentCluster = StringValueOrNull(apiModel.AgentCluster)
	data.AgentVar = StringValueOrNull(apiModel.AgentVar)
	data.AgentClusterVar = StringValueOrNull(apiModel.AgentClusterVar)
	data.Command = StringValueOrNull(apiModel.Command)
	data.Script = StringValueOrNull(apiModel.Script)
	data.RuntimeDir = StringValueOrNull(apiModel.RuntimeDir)
	data.Parameters = StringValueOrNull(apiModel.Parameters)
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)
	data.ExitCodeProcessing = StringValueOrNull(apiModel.ExitCodeProcessing)

	// Computed fields - always set from API (server provides defaults)
	data.CommandOrScript = types.StringValue(apiModel.CommandOrScript)
	data.OutputType = types.StringValue(apiModel.OutputType)
	data.WaitForOutput = types.BoolValue(apiModel.WaitForOutput)
	data.OutputReturnFile = StringValueOrNull(apiModel.OutputReturnFile)
	data.OutputReturnType = types.StringValue(apiModel.OutputReturnType)
	data.OutputReturnSline = types.StringValue(apiModel.OutputReturnSline)
	data.OutputReturnNline = types.StringValue(apiModel.OutputReturnNline)

	data.RetryMaximum = types.Int64Value(apiModel.RetryMaximum)
	data.RetryIndefinitely = types.BoolValue(apiModel.RetryIndefinitely)
	data.RetryInterval = types.Int64Value(apiModel.RetryInterval)
	data.RetrySuppressFailure = types.BoolValue(apiModel.RetrySuppressFailure)

	data.RunAsSudo = types.BoolValue(apiModel.RunAsSudo)

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
