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
	_ resource.Resource                = &TaskMonitorResource{}
	_ resource.ResourceWithImportState = &TaskMonitorResource{}
)

func NewTaskMonitorResource() resource.Resource {
	return &TaskMonitorResource{}
}

// TaskMonitorResource defines the resource implementation.
type TaskMonitorResource struct {
	client *client.Client
}

// TaskMonitorResourceModel describes the resource data model.
type TaskMonitorResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Task monitor specific
	MonType              types.String `tfsdk:"mon_type"`
	TaskMonName          types.String `tfsdk:"task_mon_name"`
	TaskNameLookup       types.String `tfsdk:"task_name_lookup"`
	TaskNameValue        types.String `tfsdk:"task_name_value"`
	TaskNameValueResolve types.Bool   `tfsdk:"task_name_value_resolve"`
	StatusText           types.String `tfsdk:"status_text"`
	TypeText             types.String `tfsdk:"type_text"`
	TimeScope            types.String `tfsdk:"time_scope"`
	RelativeTimeFrom     types.String `tfsdk:"relative_time_from"`
	RelativeTimeTo       types.String `tfsdk:"relative_time_to"`
	ExpirationAction     types.String `tfsdk:"expiration_action"`

	// Workflow condition
	WfConditionType  types.String `tfsdk:"wf_condition_type"`
	WfConditionValue types.String `tfsdk:"wf_condition_value"`

	// Monitoring options
	MonitorLateStart   types.Bool `tfsdk:"monitor_late_start"`
	MonitorLateFinish  types.Bool `tfsdk:"monitor_late_finish"`
	MonitorEarlyFinish types.Bool `tfsdk:"monitor_early_finish"`
	UseExitCode        types.Bool `tfsdk:"use_exit_code"`

	// Retry
	RetryMaximum         types.Int64 `tfsdk:"retry_maximum"`
	RetryInterval        types.Int64 `tfsdk:"retry_interval"`
	RetryIndefinitely    types.Bool  `tfsdk:"retry_indefinitely"`
	RetrySuppressFailure types.Bool  `tfsdk:"retry_suppress_failure"`

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

// TaskMonitorAPIModel represents the API request/response structure.
type TaskMonitorAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version int64  `json:"version,omitempty"`

	Summary string `json:"summary,omitempty"`

	// Task monitor specific
	MonType              string `json:"monType,omitempty"`
	TaskMonName          string `json:"taskMonName,omitempty"`
	TaskNameLookup       string `json:"taskNameLookup,omitempty"`
	TaskNameValue        string `json:"taskNameValue,omitempty"`
	TaskNameValueResolve bool   `json:"taskNameValueResolve,omitempty"`
	StatusText           string `json:"statusText,omitempty"`
	TypeText             string `json:"typeText,omitempty"`
	TimeScope            string `json:"timeScope,omitempty"`
	RelativeTimeFrom     string `json:"relativeTimeFrom,omitempty"`
	RelativeTimeTo       string `json:"relativeTimeTo,omitempty"`
	ExpirationAction     string `json:"expirationAction,omitempty"`

	// Workflow condition
	WfConditionType  string `json:"wfConditionType,omitempty"`
	WfConditionValue string `json:"wfConditionValue,omitempty"`

	// Monitoring options
	MonitorLateStart   bool `json:"monitorLateStart,omitempty"`
	MonitorLateFinish  bool `json:"monitorLateFinish,omitempty"`
	MonitorEarlyFinish bool `json:"monitorEarlyFinish,omitempty"`
	UseExitCode        bool `json:"useExitCode,omitempty"`

	// Retry
	RetryMaximum         int64 `json:"retryMaximum,omitempty"`
	RetryInterval        int64 `json:"retryInterval,omitempty"`
	RetryIndefinitely    bool  `json:"retryIndefinitely,omitempty"`
	RetrySuppressFailure bool  `json:"retrySuppressFailure,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources    bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks   []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	VirtualResources []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TaskMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_monitor"
}

func (r *TaskMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Task Monitor Task. A task monitor task watches for specific task status changes (such as completion, success, or failure) and can be used as a trigger source for task monitor triggers.",

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
				MarkdownDescription: "Unique name of the task monitor task.",
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

			// Task monitor specific
			"mon_type": schema.StringAttribute{
				MarkdownDescription: "Monitoring type. Valid values: 'Task Instance' (monitor specific task instances), 'Task' (monitor task definition).",
				Optional:            true,
				Computed:            true,
			},
			"task_mon_name": schema.StringAttribute{
				MarkdownDescription: "Name of the task to monitor. When this task reaches the specified status, the monitor detects it.",
				Required:            true,
			},
			"task_name_lookup": schema.StringAttribute{
				MarkdownDescription: "How to look up the task name. Valid values: 'Task', 'Variable'.",
				Optional:            true,
				Computed:            true,
			},
			"task_name_value": schema.StringAttribute{
				MarkdownDescription: "Value for task name lookup when using variable-based lookup.",
				Optional:            true,
			},
			"task_name_value_resolve": schema.BoolAttribute{
				MarkdownDescription: "Whether to resolve variables in the task name value.",
				Optional:            true,
				Computed:            true,
			},
			"status_text": schema.StringAttribute{
				MarkdownDescription: "Task status to monitor for. Examples: 'Success', 'Failed', 'Running Cancelled', 'Finished'.",
				Optional:            true,
				Computed:            true,
			},
			"type_text": schema.StringAttribute{
				MarkdownDescription: "Task type filter for monitoring.",
				Optional:            true,
				Computed:            true,
			},
			"time_scope": schema.StringAttribute{
				MarkdownDescription: "Time scope for monitoring. Examples: 'Any Time', 'Relative Time'.",
				Optional:            true,
				Computed:            true,
			},
			"relative_time_from": schema.StringAttribute{
				MarkdownDescription: "Start of relative time range (e.g., '-1h', '-30m').",
				Optional:            true,
				Computed:            true,
			},
			"relative_time_to": schema.StringAttribute{
				MarkdownDescription: "End of relative time range (e.g., 'now', '+1h').",
				Optional:            true,
				Computed:            true,
			},
			"expiration_action": schema.StringAttribute{
				MarkdownDescription: "Action to take when the monitor expires without detecting the condition.",
				Optional:            true,
				Computed:            true,
			},

			// Workflow condition
			"wf_condition_type": schema.StringAttribute{
				MarkdownDescription: "Workflow condition type.",
				Optional:            true,
				Computed:            true,
			},
			"wf_condition_value": schema.StringAttribute{
				MarkdownDescription: "Workflow condition value.",
				Optional:            true,
				Computed:            true,
			},

			// Monitoring options
			"monitor_late_start": schema.BoolAttribute{
				MarkdownDescription: "Monitor for late task starts.",
				Optional:            true,
				Computed:            true,
			},
			"monitor_late_finish": schema.BoolAttribute{
				MarkdownDescription: "Monitor for late task finishes.",
				Optional:            true,
				Computed:            true,
			},
			"monitor_early_finish": schema.BoolAttribute{
				MarkdownDescription: "Monitor for early task finishes.",
				Optional:            true,
				Computed:            true,
			},
			"use_exit_code": schema.BoolAttribute{
				MarkdownDescription: "Use exit code for status determination.",
				Optional:            true,
				Computed:            true,
			},

			// Retry
			"retry_maximum": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retries.",
				Optional:            true,
				Computed:            true,
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Interval between retries in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"retry_indefinitely": schema.BoolAttribute{
				MarkdownDescription: "Whether to retry indefinitely.",
				Optional:            true,
				Computed:            true,
			},
			"retry_suppress_failure": schema.BoolAttribute{
				MarkdownDescription: "Whether to suppress failure on retry exhaustion.",
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

func (r *TaskMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating task monitor task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Task Monitor Task",
			fmt.Sprintf("Could not create task monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Task Monitor Task",
			fmt.Sprintf("Could not read task monitor task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created task monitor task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Task monitor task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Task Monitor Task",
			fmt.Sprintf("Could not read task monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating task monitor task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Task Monitor Task",
			fmt.Sprintf("Could not update task monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Task Monitor Task",
			fmt.Sprintf("Could not read task monitor task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting task monitor task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Task Monitor Task",
			fmt.Sprintf("Could not delete task monitor task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskMonitorResource) readTask(ctx context.Context, data *TaskMonitorResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskMonitorAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskMonitorResource) toAPIModel(ctx context.Context, data *TaskMonitorResourceModel) *TaskMonitorAPIModel {
	model := &TaskMonitorAPIModel{
		SysId:                data.SysId.ValueString(),
		Name:                 data.Name.ValueString(),
		Type:                 "taskMonitor",
		Summary:              data.Summary.ValueString(),
		MonType:              data.MonType.ValueString(),
		TaskMonName:          data.TaskMonName.ValueString(),
		TaskNameLookup:       data.TaskNameLookup.ValueString(),
		TaskNameValue:        data.TaskNameValue.ValueString(),
		TaskNameValueResolve: data.TaskNameValueResolve.ValueBool(),
		StatusText:           data.StatusText.ValueString(),
		TypeText:             data.TypeText.ValueString(),
		TimeScope:            data.TimeScope.ValueString(),
		RelativeTimeFrom:     data.RelativeTimeFrom.ValueString(),
		RelativeTimeTo:       data.RelativeTimeTo.ValueString(),
		ExpirationAction:     data.ExpirationAction.ValueString(),
		WfConditionType:      data.WfConditionType.ValueString(),
		WfConditionValue:     data.WfConditionValue.ValueString(),
		MonitorLateStart:     data.MonitorLateStart.ValueBool(),
		MonitorLateFinish:    data.MonitorLateFinish.ValueBool(),
		MonitorEarlyFinish:   data.MonitorEarlyFinish.ValueBool(),
		UseExitCode:          data.UseExitCode.ValueBool(),
		RetryMaximum:         data.RetryMaximum.ValueInt64(),
		RetryInterval:        data.RetryInterval.ValueInt64(),
		RetryIndefinitely:    data.RetryIndefinitely.ValueBool(),
		RetrySuppressFailure: data.RetrySuppressFailure.ValueBool(),
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
func (r *TaskMonitorResource) fromAPIModel(ctx context.Context, apiModel *TaskMonitorAPIModel, data *TaskMonitorResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Task monitor specific
	data.MonType = StringValueOrNull(apiModel.MonType)
	data.TaskMonName = StringValueOrNull(apiModel.TaskMonName)
	data.TaskNameLookup = StringValueOrNull(apiModel.TaskNameLookup)
	data.TaskNameValue = StringValueOrNull(apiModel.TaskNameValue)
	data.TaskNameValueResolve = types.BoolValue(apiModel.TaskNameValueResolve)
	data.StatusText = StringValueOrNull(apiModel.StatusText)
	data.TypeText = StringValueOrNull(apiModel.TypeText)
	data.TimeScope = StringValueOrNull(apiModel.TimeScope)
	data.RelativeTimeFrom = StringValueOrNull(apiModel.RelativeTimeFrom)
	data.RelativeTimeTo = StringValueOrNull(apiModel.RelativeTimeTo)
	data.ExpirationAction = StringValueOrNull(apiModel.ExpirationAction)

	// Workflow condition
	data.WfConditionType = StringValueOrNull(apiModel.WfConditionType)
	data.WfConditionValue = StringValueOrNull(apiModel.WfConditionValue)

	// Monitoring options
	data.MonitorLateStart = types.BoolValue(apiModel.MonitorLateStart)
	data.MonitorLateFinish = types.BoolValue(apiModel.MonitorLateFinish)
	data.MonitorEarlyFinish = types.BoolValue(apiModel.MonitorEarlyFinish)
	data.UseExitCode = types.BoolValue(apiModel.UseExitCode)

	// Retry
	data.RetryMaximum = types.Int64Value(apiModel.RetryMaximum)
	data.RetryInterval = types.Int64Value(apiModel.RetryInterval)
	data.RetryIndefinitely = types.BoolValue(apiModel.RetryIndefinitely)
	data.RetrySuppressFailure = types.BoolValue(apiModel.RetrySuppressFailure)

	// Handle variables
	data.Variables = TaskVariablesFromAPI(ctx, apiModel.Variables)

	// Handle resource management fields
	data.HoldResources = types.BoolValue(apiModel.HoldResources)
	data.ExclusiveTasks = TaskExclusiveTasksFromAPI(apiModel.ExclusiveTasks)
	data.VirtualResources = TaskVirtualResourcesFromAPI(apiModel.VirtualResources)

	// Handle actions
	data.Actions = TaskActionsFromAPI(ctx, apiModel.Actions)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
