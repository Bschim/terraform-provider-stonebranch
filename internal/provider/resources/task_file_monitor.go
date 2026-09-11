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
	_ resource.Resource                   = &TaskFileMonitorResource{}
	_ resource.ResourceWithImportState    = &TaskFileMonitorResource{}
	_ resource.ResourceWithValidateConfig = &TaskFileMonitorResource{}
)

func NewTaskFileMonitorResource() resource.Resource {
	return &TaskFileMonitorResource{}
}

// TaskFileMonitorResource defines the resource implementation.
type TaskFileMonitorResource struct {
	client *client.Client
}

// TaskFileMonitorResourceModel describes the resource data model.
type TaskFileMonitorResourceModel struct {
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

	// File monitor configuration
	FileName        types.String `tfsdk:"file_name"`
	UseRegex        types.Bool   `tfsdk:"use_regex"`
	StableSeconds   types.Int64  `tfsdk:"stable_seconds"`
	FmType          types.String `tfsdk:"fm_type"`
	Recursive       types.Bool   `tfsdk:"recursive"`
	FileOwner       types.String `tfsdk:"file_owner"`
	FileGroup       types.String `tfsdk:"file_group"`
	ScanText        types.String `tfsdk:"scan_text"`
	ScanForward     types.Bool   `tfsdk:"scan_forward"`
	MaxFiles        types.Int64  `tfsdk:"max_files"`
	TriggerOnExist  types.Bool   `tfsdk:"trigger_on_exist"`
	TriggerOnCreate types.Bool   `tfsdk:"trigger_on_create"`
	MinFileSize     types.String `tfsdk:"min_file_size"`
	MinFileScale    types.String `tfsdk:"min_file_scale"`

	// Credentials
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

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
	HoldResources     types.Bool `tfsdk:"hold_resources"`
	ExclusiveTasks    types.List `tfsdk:"exclusive_tasks"`
	ExclusiveWithSelf types.Bool `tfsdk:"exclusive_with_self"`
	VirtualResources  types.List `tfsdk:"virtual_resources"`

	// Actions
	Actions types.Object `tfsdk:"actions"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// TaskFileMonitorAPIModel represents the API request/response structure.
type TaskFileMonitorAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	Agent           string `json:"agent,omitempty"`
	AgentCluster    string `json:"agentCluster,omitempty"`
	AgentVar        string `json:"agentVar,omitempty"`
	AgentClusterVar string `json:"agentClusterVar,omitempty"`

	// File monitor specific
	FileName        string `json:"fileName,omitempty"`
	UseRegex        bool   `json:"useRegex,omitempty"`
	StableSeconds   int64  `json:"stableSeconds,omitempty"`
	FmType          string `json:"fmtype,omitempty"`
	Recursive       bool   `json:"recursive,omitempty"`
	FileOwner       string `json:"fileOwner,omitempty"`
	FileGroup       string `json:"fileGroup,omitempty"`
	ScanText        string `json:"scanText,omitempty"`
	ScanForward     bool   `json:"scanForward,omitempty"`
	MaxFiles        int64  `json:"maxFiles,omitempty"`
	TriggerOnExist  bool   `json:"triggerOnExist,omitempty"`
	TriggerOnCreate bool   `json:"triggerOnCreate,omitempty"`
	MinFileSize     string `json:"minFileSize,omitempty"`
	MinFileScale    string `json:"minFileScale,omitempty"`

	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

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

	HoldResources     bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks    []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	ExclusiveWithSelf bool                          `json:"exclusiveWithSelf,omitempty"`
	VirtualResources  []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TaskFileMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_file_monitor"
}

func (r *TaskFileMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch File Monitor Task. File monitor tasks watch for file system events such as file creation, modification, or deletion.",

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
				MarkdownDescription: "Summary/description of the task.",
				Optional:            true,
			},

			// Agent configuration
			"agent": schema.StringAttribute{
				MarkdownDescription: "Name of the agent to run the task on. One of agent, agent_cluster, agent_var, or agent_cluster_var is required.",
				Optional:            true,
			},
			"agent_cluster": schema.StringAttribute{
				MarkdownDescription: "Name of the agent cluster to run the task on.",
				Optional:            true,
			},
			"agent_var": schema.StringAttribute{
				MarkdownDescription: "Name of a variable containing the agent name.",
				Optional:            true,
			},
			"agent_cluster_var": schema.StringAttribute{
				MarkdownDescription: "Name of a variable containing the agent cluster name.",
				Optional:            true,
			},

			// File monitor configuration
			"file_name": schema.StringAttribute{
				MarkdownDescription: "File path or pattern to monitor. Can include wildcards.",
				Required:            true,
			},
			"use_regex": schema.BoolAttribute{
				MarkdownDescription: "Whether to use regular expression pattern matching for the file name.",
				Optional:            true,
				Computed:            true,
			},
			"stable_seconds": schema.Int64Attribute{
				MarkdownDescription: "Number of seconds the file must be stable (unchanged) before triggering.",
				Optional:            true,
				Computed:            true,
			},
			"fm_type": schema.StringAttribute{
				MarkdownDescription: "File monitor type. Valid values: Created, Deleted, Changed, Exist, Missing. Default is Created.",
				Optional:            true,
				Computed:            true,
			},
			"recursive": schema.BoolAttribute{
				MarkdownDescription: "Whether to recursively monitor subdirectories.",
				Optional:            true,
				Computed:            true,
			},
			"file_owner": schema.StringAttribute{
				MarkdownDescription: "Filter files by owner.",
				Optional:            true,
			},
			"file_group": schema.StringAttribute{
				MarkdownDescription: "Filter files by group.",
				Optional:            true,
			},
			"scan_text": schema.StringAttribute{
				MarkdownDescription: "Text pattern to scan for within the file.",
				Optional:            true,
			},
			"scan_forward": schema.BoolAttribute{
				MarkdownDescription: "Whether to scan forward in the file (from beginning).",
				Optional:            true,
				Computed:            true,
			},
			"max_files": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of files to monitor.",
				Optional:            true,
				Computed:            true,
			},
			"trigger_on_exist": schema.BoolAttribute{
				MarkdownDescription: "Whether to trigger when the file already exists.",
				Optional:            true,
				Computed:            true,
			},
			"trigger_on_create": schema.BoolAttribute{
				MarkdownDescription: "Whether to trigger on file creation.",
				Optional:            true,
				Computed:            true,
			},
			"min_file_size": schema.StringAttribute{
				MarkdownDescription: "Minimum file size to trigger.",
				Optional:            true,
				Computed:            true,
			},
			"min_file_scale": schema.StringAttribute{
				MarkdownDescription: "Units for minimum file size (B, KB, MB, GB).",
				Optional:            true,
				Computed:            true,
			},

			// Credentials
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use.",
				Optional:            true,
			},
			"credentials_var": schema.StringAttribute{
				MarkdownDescription: "Name of a variable containing the credentials name.",
				Optional:            true,
			},

			// Retry configuration
			"retry_maximum": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of times to retry on failure.",
				Optional:            true,
				Computed:            true,
			},
			"retry_indefinitely": schema.BoolAttribute{
				MarkdownDescription: "Whether to retry indefinitely on failure.",
				Optional:            true,
				Computed:            true,
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Number of seconds to wait between retries.",
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
			"exclusive_tasks":     TaskExclusiveTasksSchema(),
			"exclusive_with_self": TaskExclusiveWithSelfSchema(),
			"virtual_resources":   TaskVirtualResourcesSchema(),

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

func (r *TaskFileMonitorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TaskFileMonitorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(ValidateTaskWaitDelay(data.WaitToStart, data.WaitAmount, data.DelayOnStart, data.DelayAmount)...)
}

func (r *TaskFileMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskFileMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskFileMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating file monitor task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating File Monitor Task",
			fmt.Sprintf("Could not create file monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created File Monitor Task",
			fmt.Sprintf("Could not read file monitor task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created file monitor task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskFileMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskFileMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "File monitor task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading File Monitor Task",
			fmt.Sprintf("Could not read file monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskFileMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskFileMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskFileMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating file monitor task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating File Monitor Task",
			fmt.Sprintf("Could not update file monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated File Monitor Task",
			fmt.Sprintf("Could not read file monitor task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskFileMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskFileMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting file monitor task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting File Monitor Task",
			fmt.Sprintf("Could not delete file monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskFileMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskFileMonitorResource) readTask(ctx context.Context, data *TaskFileMonitorResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskFileMonitorAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskFileMonitorResource) toAPIModel(ctx context.Context, data *TaskFileMonitorResourceModel) *TaskFileMonitorAPIModel {
	model := &TaskFileMonitorAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskFileMonitor",
		Summary: data.Summary.ValueString(),

		Agent:           data.Agent.ValueString(),
		AgentCluster:    data.AgentCluster.ValueString(),
		AgentVar:        data.AgentVar.ValueString(),
		AgentClusterVar: data.AgentClusterVar.ValueString(),

		FileName:        data.FileName.ValueString(),
		UseRegex:        data.UseRegex.ValueBool(),
		StableSeconds:   data.StableSeconds.ValueInt64(),
		FmType:          data.FmType.ValueString(),
		Recursive:       data.Recursive.ValueBool(),
		FileOwner:       data.FileOwner.ValueString(),
		FileGroup:       data.FileGroup.ValueString(),
		ScanText:        data.ScanText.ValueString(),
		ScanForward:     data.ScanForward.ValueBool(),
		MaxFiles:        data.MaxFiles.ValueInt64(),
		TriggerOnExist:  data.TriggerOnExist.ValueBool(),
		TriggerOnCreate: data.TriggerOnCreate.ValueBool(),
		MinFileSize:     data.MinFileSize.ValueString(),
		MinFileScale:    data.MinFileScale.ValueString(),

		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),

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
	}

	// Handle variables
	model.Variables = TaskVariablesToAPI(ctx, data.Variables)

	// Handle resource management fields
	if !data.HoldResources.IsNull() && !data.HoldResources.IsUnknown() {
		model.HoldResources = data.HoldResources.ValueBool()
	}
	model.ExclusiveTasks = TaskExclusiveTasksToAPI(ctx, data.ExclusiveTasks)
	model.ExclusiveWithSelf = data.ExclusiveWithSelf.ValueBool()
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
func (r *TaskFileMonitorResource) fromAPIModel(ctx context.Context, apiModel *TaskFileMonitorAPIModel, data *TaskFileMonitorResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Agent configuration
	data.Agent = StringValueOrNull(apiModel.Agent)
	data.AgentCluster = StringValueOrNull(apiModel.AgentCluster)
	data.AgentVar = StringValueOrNull(apiModel.AgentVar)
	data.AgentClusterVar = StringValueOrNull(apiModel.AgentClusterVar)

	// File monitor configuration
	data.FileName = StringValueOrNull(apiModel.FileName)
	data.UseRegex = types.BoolValue(apiModel.UseRegex)
	data.StableSeconds = types.Int64Value(apiModel.StableSeconds)
	data.FmType = StringValueOrNull(apiModel.FmType)
	data.Recursive = types.BoolValue(apiModel.Recursive)
	data.FileOwner = StringValueOrNull(apiModel.FileOwner)
	data.FileGroup = StringValueOrNull(apiModel.FileGroup)
	data.ScanText = StringValueOrNull(apiModel.ScanText)
	data.ScanForward = types.BoolValue(apiModel.ScanForward)
	data.MaxFiles = types.Int64Value(apiModel.MaxFiles)
	data.TriggerOnExist = types.BoolValue(apiModel.TriggerOnExist)
	data.TriggerOnCreate = types.BoolValue(apiModel.TriggerOnCreate)
	data.MinFileSize = StringValueOrNull(apiModel.MinFileSize)
	data.MinFileScale = StringValueOrNull(apiModel.MinFileScale)

	// Credentials
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)

	// Retry configuration
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

	// Handle variables
	data.Variables = TaskVariablesFromAPIOrdered(ctx, apiModel.Variables, data.Variables)

	// Handle resource management fields
	data.HoldResources = types.BoolValue(apiModel.HoldResources)
	data.ExclusiveTasks = TaskExclusiveTasksFromAPI(apiModel.ExclusiveTasks)
	data.ExclusiveWithSelf = types.BoolValue(apiModel.ExclusiveWithSelf)
	data.VirtualResources = TaskVirtualResourcesFromAPI(apiModel.VirtualResources)

	// Handle actions
	data.Actions = TaskActionsFromAPI(ctx, apiModel.Actions, data.Actions)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
