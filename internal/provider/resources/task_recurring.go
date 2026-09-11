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
	_ resource.Resource                   = &TaskRecurringResource{}
	_ resource.ResourceWithImportState    = &TaskRecurringResource{}
	_ resource.ResourceWithValidateConfig = &TaskRecurringResource{}
)

func NewTaskRecurringResource() resource.Resource {
	return &TaskRecurringResource{}
}

// TaskRecurringResource defines the resource implementation.
type TaskRecurringResource struct {
	client *client.Client
}

// TaskRecurringResourceModel describes the resource data model.
type TaskRecurringResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Target
	TargetTask                 types.String `tfsdk:"target_task"`
	TargetTaskMonitorCondition types.String `tfsdk:"target_task_monitor_condition"`
	TargetTaskStatusText       types.String `tfsdk:"target_task_status_text"`

	// Recurrence
	RecurrenceType         types.String `tfsdk:"recurrence_type"`
	RecurrenceInterval     types.String `tfsdk:"recurrence_interval"`
	RecurrenceIntervalUnit types.String `tfsdk:"recurrence_interval_unit"`

	// Interval window
	IntervalStartTime          types.String `tfsdk:"interval_start_time"`
	IntervalEndTime            types.String `tfsdk:"interval_end_time"`
	IntervalStartDayConstraint types.String `tfsdk:"interval_start_day_constraint"`
	IntervalEndDayConstraint   types.String `tfsdk:"interval_end_day_constraint"`
	TimeWindow                 types.Bool   `tfsdk:"time_window"`

	// Recurrence count
	IndefiniteRecurrences types.Bool   `tfsdk:"indefinite_recurrences"`
	NumberOfRecurrences   types.String `tfsdk:"number_of_recurrences"`

	// Skip condition
	SkipCondition types.String `tfsdk:"skip_condition"`

	// Recurring instance retention
	RetentionDurationRt      types.Int64  `tfsdk:"retention_duration_rt"`
	RetentionDurationUnitRt  types.String `tfsdk:"retention_duration_unit_rt"`
	RetentionDurationPurgeRt types.Bool   `tfsdk:"retention_duration_purge_rt"`
	RdExcludeBackupRt        types.Bool   `tfsdk:"rd_exclude_backup_rt"`

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

// TaskRecurringAPIModel represents the API request/response structure.
type TaskRecurringAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	TargetTask                 string `json:"targetTask,omitempty"`
	TargetTaskMonitorCondition string `json:"targetTaskMonitorCondition,omitempty"`
	TargetTaskStatusText       string `json:"targetTaskStatusText,omitempty"`

	RecurrenceType         string `json:"recurrenceType,omitempty"`
	RecurrenceInterval     string `json:"recurrenceInterval,omitempty"`
	RecurrenceIntervalUnit string `json:"recurrenceIntervalUnit,omitempty"`

	IntervalStartTime          string `json:"intervalStartTime,omitempty"`
	IntervalEndTime            string `json:"intervalEndTime,omitempty"`
	IntervalStartDayConstraint string `json:"intervalStartDayConstraint,omitempty"`
	IntervalEndDayConstraint   string `json:"intervalEndDayConstraint,omitempty"`
	TimeWindow                 bool   `json:"timeWindow,omitempty"`

	IndefiniteRecurrences bool   `json:"indefiniteRecurrences,omitempty"`
	NumberOfRecurrences   string `json:"numberOfRecurrences,omitempty"`

	SkipCondition string `json:"skipCondition,omitempty"`

	RetentionDurationRt      int64  `json:"retentionDurationRt,omitempty"`
	RetentionDurationUnitRt  string `json:"retentionDurationUnitRt,omitempty"`
	RetentionDurationPurgeRt bool   `json:"retentionDurationPurgeRt,omitempty"`
	RdExcludeBackupRt        bool   `json:"rdExcludeBackupRt,omitempty"`

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

func (r *TaskRecurringResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_recurring"
}

func (r *TaskRecurringResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Task (Recurring task type). A recurring task repeatedly launches a target task/workflow on an interval, within an optional time window.",

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

			// Target
			"target_task": schema.StringAttribute{
				MarkdownDescription: "Name of the task or workflow to launch on each recurrence.",
				Required:            true,
			},
			"target_task_monitor_condition": schema.StringAttribute{
				MarkdownDescription: "When to monitor the target task's status. Values: 'None', 'First Recurrence', 'Last Recurrence', 'All Recurrences'.",
				Optional:            true,
				Computed:            true,
			},
			"target_task_status_text": schema.StringAttribute{
				MarkdownDescription: "Status text to display for the target task.",
				Optional:            true,
			},

			// Recurrence
			"recurrence_type": schema.StringAttribute{
				MarkdownDescription: "How recurrences are scheduled. Values: 'Interval' (fixed interval within a time window), 'On' (explicit time list).",
				Optional:            true,
				Computed:            true,
			},
			"recurrence_interval": schema.StringAttribute{
				MarkdownDescription: "Interval amount between recurrences (used when recurrence_type is 'Interval').",
				Optional:            true,
				Computed:            true,
			},
			"recurrence_interval_unit": schema.StringAttribute{
				MarkdownDescription: "Unit for recurrence_interval. Values: 'Seconds', 'Minutes', 'Hours', 'Days'.",
				Optional:            true,
				Computed:            true,
			},

			// Interval window
			"interval_start_time": schema.StringAttribute{
				MarkdownDescription: "Time of day (HH:MM) the recurrence interval window starts.",
				Optional:            true,
			},
			"interval_end_time": schema.StringAttribute{
				MarkdownDescription: "Time of day (HH:MM) the recurrence interval window ends.",
				Optional:            true,
			},
			"interval_start_day_constraint": schema.StringAttribute{
				MarkdownDescription: "Day constraint applied to interval_start_time (e.g. 'None', 'Same Day').",
				Optional:            true,
				Computed:            true,
			},
			"interval_end_day_constraint": schema.StringAttribute{
				MarkdownDescription: "Day constraint applied to interval_end_time (e.g. 'None', 'Same Day').",
				Optional:            true,
				Computed:            true,
			},
			"time_window": schema.BoolAttribute{
				MarkdownDescription: "Whether recurrences are restricted to the interval_start_time/interval_end_time window.",
				Optional:            true,
				Computed:            true,
			},

			// Recurrence count
			"indefinite_recurrences": schema.BoolAttribute{
				MarkdownDescription: "Whether the task recurs indefinitely. If false, number_of_recurrences is required by the server.",
				Optional:            true,
				Computed:            true,
			},
			"number_of_recurrences": schema.StringAttribute{
				MarkdownDescription: "Number of times to recur before stopping (required by the server when indefinite_recurrences is false).",
				Optional:            true,
			},

			// Skip condition
			"skip_condition": schema.StringAttribute{
				MarkdownDescription: "Condition under which a recurrence is skipped. Values: 'None', 'Active', 'Active By Recurring Task Instance'.",
				Optional:            true,
				Computed:            true,
			},

			// Recurring instance retention
			"retention_duration_rt": schema.Int64Attribute{
				MarkdownDescription: "How long to retain completed recurrence instances.",
				Optional:            true,
				Computed:            true,
			},
			"retention_duration_unit_rt": schema.StringAttribute{
				MarkdownDescription: "Unit for retention_duration_rt (e.g. 'Days').",
				Optional:            true,
				Computed:            true,
			},
			"retention_duration_purge_rt": schema.BoolAttribute{
				MarkdownDescription: "Whether to purge retained recurrence instances after retention_duration_rt elapses.",
				Optional:            true,
				Computed:            true,
			},
			"rd_exclude_backup_rt": schema.BoolAttribute{
				MarkdownDescription: "Whether to exclude retained recurrence instances from backup.",
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

func (r *TaskRecurringResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig enforces the server's mutually-exclusive relationship between
// time_window, indefinite_recurrences, and number_of_recurrences:
//   - time_window and indefinite_recurrences cannot both be true.
//   - number_of_recurrences is only meaningful (and only accepted by the
//     server) when time_window is false and indefinite_recurrences is false;
//     the server silently discards it otherwise.
func (r *TaskRecurringResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TaskRecurringResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeWindowUnknown := data.TimeWindow.IsUnknown()
	indefiniteUnknown := data.IndefiniteRecurrences.IsUnknown()
	if timeWindowUnknown || indefiniteUnknown {
		return
	}

	timeWindow := !data.TimeWindow.IsNull() && data.TimeWindow.ValueBool()
	indefinite := !data.IndefiniteRecurrences.IsNull() && data.IndefiniteRecurrences.ValueBool()
	hasNumberOfRecurrences := !data.NumberOfRecurrences.IsNull() && !data.NumberOfRecurrences.IsUnknown() && data.NumberOfRecurrences.ValueString() != ""

	if timeWindow && indefinite {
		resp.Diagnostics.AddAttributeError(
			path.Root("time_window"),
			"Conflicting Fields",
			`"time_window" and "indefinite_recurrences" cannot both be true.`,
		)
	}

	if timeWindow && hasNumberOfRecurrences {
		resp.Diagnostics.AddAttributeError(
			path.Root("number_of_recurrences"),
			"Conflicting Fields",
			`"number_of_recurrences" must not be set when "time_window" is true; the recurrence count is bounded by the interval window instead.`,
		)
	}

	if !timeWindow && !indefinite && !hasNumberOfRecurrences {
		resp.Diagnostics.AddAttributeError(
			path.Root("number_of_recurrences"),
			"Missing Required Field",
			`"number_of_recurrences" is required when "time_window" is false and "indefinite_recurrences" is false.`,
		)
	}

	resp.Diagnostics.Append(ValidateTaskWaitDelay(data.WaitToStart, data.WaitAmount, data.DelayOnStart, data.DelayAmount)...)
}

func (r *TaskRecurringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskRecurringResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating task", map[string]any{"name": data.Name.ValueString()})

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

	tflog.Debug(ctx, "Created task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskRecurringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskRecurringResourceModel

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

func (r *TaskRecurringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskRecurringResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TaskRecurringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating task", map[string]any{"sys_id": data.SysId.ValueString()})

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

func (r *TaskRecurringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskRecurringResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting task", map[string]any{"sys_id": data.SysId.ValueString()})

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

func (r *TaskRecurringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskRecurringResource) readTask(ctx context.Context, data *TaskRecurringResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskRecurringAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskRecurringResource) toAPIModel(ctx context.Context, data *TaskRecurringResourceModel) *TaskRecurringAPIModel {
	model := &TaskRecurringAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskRecurring",
		Summary: data.Summary.ValueString(),

		TargetTask:                 data.TargetTask.ValueString(),
		TargetTaskMonitorCondition: StringValueOrDefault(data.TargetTaskMonitorCondition, "None"),
		TargetTaskStatusText:       data.TargetTaskStatusText.ValueString(),

		RecurrenceType:         StringValueOrDefault(data.RecurrenceType, "Interval"),
		RecurrenceInterval:     data.RecurrenceInterval.ValueString(),
		RecurrenceIntervalUnit: StringValueOrDefault(data.RecurrenceIntervalUnit, "Minutes"),

		IntervalStartTime:          data.IntervalStartTime.ValueString(),
		IntervalEndTime:            data.IntervalEndTime.ValueString(),
		IntervalStartDayConstraint: StringValueOrDefault(data.IntervalStartDayConstraint, "None"),
		IntervalEndDayConstraint:   StringValueOrDefault(data.IntervalEndDayConstraint, "None"),
		TimeWindow:                 data.TimeWindow.ValueBool(),

		IndefiniteRecurrences: data.IndefiniteRecurrences.ValueBool(),
		NumberOfRecurrences:   data.NumberOfRecurrences.ValueString(),

		SkipCondition: StringValueOrDefault(data.SkipCondition, "None"),

		RetentionDurationRt:      data.RetentionDurationRt.ValueInt64(),
		RetentionDurationUnitRt:  StringValueOrDefault(data.RetentionDurationUnitRt, "Days"),
		RetentionDurationPurgeRt: data.RetentionDurationPurgeRt.ValueBool(),
		RdExcludeBackupRt:        data.RdExcludeBackupRt.ValueBool(),

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

	if !data.OpswiseGroups.IsNull() && !data.OpswiseGroups.IsUnknown() {
		var groups []string
		data.OpswiseGroups.ElementsAs(ctx, &groups, false)
		model.OpswiseGroups = groups
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *TaskRecurringResource) fromAPIModel(ctx context.Context, apiModel *TaskRecurringAPIModel, data *TaskRecurringResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Optional fields - only set if non-empty
	data.Summary = StringValueOrNull(apiModel.Summary)
	data.TargetTask = types.StringValue(apiModel.TargetTask)
	data.TargetTaskStatusText = StringValueOrNull(apiModel.TargetTaskStatusText)
	data.IntervalStartTime = StringValueOrNull(apiModel.IntervalStartTime)
	data.IntervalEndTime = StringValueOrNull(apiModel.IntervalEndTime)
	data.NumberOfRecurrences = StringValueOrNull(apiModel.NumberOfRecurrences)

	// Computed fields - always set from API (server provides defaults)
	data.TargetTaskMonitorCondition = types.StringValue(apiModel.TargetTaskMonitorCondition)
	data.RecurrenceType = types.StringValue(apiModel.RecurrenceType)
	data.RecurrenceInterval = types.StringValue(apiModel.RecurrenceInterval)
	data.RecurrenceIntervalUnit = types.StringValue(apiModel.RecurrenceIntervalUnit)
	data.IntervalStartDayConstraint = types.StringValue(apiModel.IntervalStartDayConstraint)
	data.IntervalEndDayConstraint = types.StringValue(apiModel.IntervalEndDayConstraint)
	data.TimeWindow = types.BoolValue(apiModel.TimeWindow)
	data.IndefiniteRecurrences = types.BoolValue(apiModel.IndefiniteRecurrences)
	data.SkipCondition = types.StringValue(apiModel.SkipCondition)
	data.RetentionDurationRt = types.Int64Value(apiModel.RetentionDurationRt)
	data.RetentionDurationUnitRt = types.StringValue(apiModel.RetentionDurationUnitRt)
	data.RetentionDurationPurgeRt = types.BoolValue(apiModel.RetentionDurationPurgeRt)
	data.RdExcludeBackupRt = types.BoolValue(apiModel.RdExcludeBackupRt)

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
