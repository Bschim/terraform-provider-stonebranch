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
	_ resource.Resource                   = &TaskTimerResource{}
	_ resource.ResourceWithImportState    = &TaskTimerResource{}
	_ resource.ResourceWithValidateConfig = &TaskTimerResource{}
)

func NewTaskTimerResource() resource.Resource {
	return &TaskTimerResource{}
}

// TaskTimerResource defines the resource implementation.
type TaskTimerResource struct {
	client *client.Client
}

// TaskTimerResourceModel describes the resource data model.
type TaskTimerResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Timer configuration
	SleepType          types.String `tfsdk:"sleep_type"`
	SleepAmount        types.String `tfsdk:"sleep_amount"`
	SleepDuration      types.String `tfsdk:"sleep_duration"`
	SleepTime          types.String `tfsdk:"sleep_time"`
	SleepDayConstraint types.String `tfsdk:"sleep_day_constraint"`

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

// TaskTimerAPIModel represents the API request/response structure.
type TaskTimerAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	// Timer configuration
	SleepType          string `json:"sleepType,omitempty"`
	SleepAmount        string `json:"sleepAmount,omitempty"`
	SleepDuration      string `json:"sleepDuration,omitempty"`
	SleepTime          string `json:"sleepTime,omitempty"`
	SleepDayConstraint string `json:"sleepDayConstraint,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources    bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks   []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	VirtualResources []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TaskTimerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_timer"
}

func (r *TaskTimerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Timer Task (Sleep Task). Timer tasks pause workflow execution for a specified duration or until a specific time.",

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

			// Timer configuration
			"sleep_type": schema.StringAttribute{
				MarkdownDescription: "Type of timer delay. Valid values: `Duration` (wait for a duration), `Seconds` (wait for a number of seconds), `Time` (wait until a specific time), `Relative Time` (wait until a time relative to task start). Defaults to `Duration`.",
				Optional:            true,
				Computed:            true,
			},
			"sleep_amount": schema.StringAttribute{
				MarkdownDescription: "Number of seconds to wait. Required when `sleep_type` is `Seconds`. Example: `30`.",
				Optional:            true,
				Computed:            true,
			},
			"sleep_duration": schema.StringAttribute{
				MarkdownDescription: "Duration to wait in `DD:HH:MM:SS` format (days:hours:minutes:seconds). Required when `sleep_type` is `Duration`. Example: `00:00:00:30` for 30 seconds, `00:01:30:00` for 1 hour 30 minutes.",
				Optional:            true,
				Computed:            true,
			},
			"sleep_time": schema.StringAttribute{
				MarkdownDescription: "Time to wait until in `HH:MM` format (24-hour). Used when `sleep_type` is `Time` or `Relative Time`. Example: `14:30` for 2:30 PM.",
				Optional:            true,
				Computed:            true,
			},
			"sleep_day_constraint": schema.StringAttribute{
				MarkdownDescription: "Day constraint for time-based delays. Valid values: `None`, `Same Day`, `Next Day`, `Next Business Day`, `Next Non-business Day`, `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, `Saturday`. Defaults to `None`.",
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

func (r *TaskTimerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TaskTimerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If sleep_type is unknown (e.g. computed from another resource),
	// defer this check to the server rather than failing the plan early.
	if data.SleepType.IsUnknown() {
		return
	}

	if data.SleepType.ValueString() == "Seconds" {
		if isSet(data.SleepAmount) {
			return
		}
		// Deferred if sleep_amount is unknown.
		if data.SleepAmount.IsUnknown() {
			return
		}

		resp.Diagnostics.AddAttributeError(
			path.Root("sleep_amount"),
			"Missing Required Field",
			`"sleep_amount" must be set when "sleep_type" is "Seconds".`,
		)
	}
}

func (r *TaskTimerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskTimerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskTimerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating timer task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Timer Task",
			fmt.Sprintf("Could not create timer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Timer Task",
			fmt.Sprintf("Could not read timer task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created timer task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskTimerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskTimerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Timer task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Timer Task",
			fmt.Sprintf("Could not read timer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskTimerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskTimerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskTimerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating timer task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Timer Task",
			fmt.Sprintf("Could not update timer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Timer Task",
			fmt.Sprintf("Could not read timer task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskTimerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskTimerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting timer task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Timer Task",
			fmt.Sprintf("Could not delete timer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskTimerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskTimerResource) readTask(ctx context.Context, data *TaskTimerResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskTimerAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse timer task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskTimerResource) toAPIModel(ctx context.Context, data *TaskTimerResourceModel) *TaskTimerAPIModel {
	model := &TaskTimerAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskSleep",
		Summary: data.Summary.ValueString(),

		// Timer configuration
		SleepType:          data.SleepType.ValueString(),
		SleepAmount:        data.SleepAmount.ValueString(),
		SleepDuration:      data.SleepDuration.ValueString(),
		SleepTime:          data.SleepTime.ValueString(),
		SleepDayConstraint: data.SleepDayConstraint.ValueString(),
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
func (r *TaskTimerResource) fromAPIModel(ctx context.Context, apiModel *TaskTimerAPIModel, data *TaskTimerResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Timer configuration - always returned by API
	data.SleepType = types.StringValue(apiModel.SleepType)
	data.SleepAmount = types.StringValue(apiModel.SleepAmount)
	data.SleepDuration = types.StringValue(apiModel.SleepDuration)
	data.SleepTime = types.StringValue(apiModel.SleepTime)
	data.SleepDayConstraint = types.StringValue(apiModel.SleepDayConstraint)

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
