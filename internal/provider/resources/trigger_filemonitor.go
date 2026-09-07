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
	_ resource.Resource                = &TriggerFileMonitorResource{}
	_ resource.ResourceWithImportState = &TriggerFileMonitorResource{}
)

func NewTriggerFileMonitorResource() resource.Resource {
	return &TriggerFileMonitorResource{}
}

// TriggerFileMonitorResource defines the resource implementation.
type TriggerFileMonitorResource struct {
	client *client.Client
}

// TriggerFileMonitorResourceModel describes the resource data model.
type TriggerFileMonitorResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`

	// Tasks to trigger (required)
	Tasks types.List `tfsdk:"tasks"`

	// File monitor specific
	TaskMonitor types.String `tfsdk:"task_monitor"`

	// Time restrictions
	TimeZone        types.String `tfsdk:"time_zone"`
	Calendar        types.String `tfsdk:"calendar"`
	RestrictedTimes types.Bool   `tfsdk:"restricted_times"`
	EnabledStart    types.String `tfsdk:"enabled_start"`
	EnabledEnd      types.String `tfsdk:"enabled_end"`

	// Variables
	Variables types.List `tfsdk:"variables"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// TriggerFileMonitorAPIModel represents the API request/response structure.
type TriggerFileMonitorAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version int64  `json:"version,omitempty"`

	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`

	Tasks []string `json:"tasks,omitempty"`

	// File monitor specific
	TaskMonitor string `json:"taskMonitor,omitempty"`

	// Time restrictions
	TimeZone        string `json:"timeZone,omitempty"`
	Calendar        string `json:"calendar,omitempty"`
	RestrictedTimes bool   `json:"restrictedTimes,omitempty"`
	EnabledStart    string `json:"enabledStart,omitempty"`
	EnabledEnd      string `json:"enabledEnd,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TriggerFileMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trigger_file_monitor"
}

func (r *TriggerFileMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch File Monitor Trigger. A file monitor trigger executes associated tasks when a file monitor task detects file events (file created, modified, deleted, etc.).",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the trigger (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the trigger.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the trigger (for optimistic locking).",
				Computed:            true,
			},

			// Basic info
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the trigger.",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the trigger is enabled. Note: Triggers are created disabled by default.",
				Optional:            true,
				Computed:            true,
			},

			// Tasks to trigger
			"tasks": schema.ListAttribute{
				MarkdownDescription: "List of task names to execute when the file monitor detects a file event. At least one task is required.",
				Required:            true,
				ElementType:         types.StringType,
			},

			// File monitor specific
			"task_monitor": schema.StringAttribute{
				MarkdownDescription: "Name of the file monitor task that triggers this trigger when it detects file events.",
				Required:            true,
			},

			// Time restrictions
			"time_zone": schema.StringAttribute{
				MarkdownDescription: "Time zone for the trigger schedule (e.g., 'America/New_York', 'UTC').",
				Optional:            true,
				Computed:            true,
			},
			"calendar": schema.StringAttribute{
				MarkdownDescription: "Name of the calendar to use for scheduling restrictions.",
				Optional:            true,
				Computed:            true,
			},
			"restricted_times": schema.BoolAttribute{
				MarkdownDescription: "Whether time restrictions are enabled for this trigger.",
				Optional:            true,
				Computed:            true,
			},
			"enabled_start": schema.StringAttribute{
				MarkdownDescription: "Start time for the enabled window (when restricted_times is true). Format: HH:MM.",
				Optional:            true,
				Computed:            true,
			},
			"enabled_end": schema.StringAttribute{
				MarkdownDescription: "End time for the enabled window (when restricted_times is true). Format: HH:MM.",
				Optional:            true,
				Computed:            true,
			},

			// Variables
			"variables": TaskVariablesSchema(),

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this trigger belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *TriggerFileMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TriggerFileMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TriggerFileMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating file monitor trigger", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the trigger
	_, err := r.client.Post(ctx, "/resources/trigger", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating File Monitor Trigger",
			fmt.Sprintf("Could not create file monitor trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created trigger to get sysId and other computed fields
	err = r.readTrigger(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created File Monitor Trigger",
			fmt.Sprintf("Could not read file monitor trigger %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created file monitor trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerFileMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TriggerFileMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTrigger(ctx, &data)
	if err != nil {
		// Check if trigger was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "File monitor trigger not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading File Monitor Trigger",
			fmt.Sprintf("Could not read file monitor trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerFileMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TriggerFileMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TriggerFileMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating file monitor trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the trigger
	_, err := r.client.Put(ctx, "/resources/trigger", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating File Monitor Trigger",
			fmt.Sprintf("Could not update file monitor trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTrigger(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated File Monitor Trigger",
			fmt.Sprintf("Could not read file monitor trigger %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerFileMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TriggerFileMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting file monitor trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("triggerid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/trigger", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting File Monitor Trigger",
			fmt.Sprintf("Could not delete file monitor trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TriggerFileMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTrigger fetches the trigger from the API and updates the model.
func (r *TriggerFileMonitorResource) readTrigger(ctx context.Context, data *TriggerFileMonitorResourceModel) error {
	query := url.Values{}
	query.Set("triggername", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/trigger", query)
	if err != nil {
		return err
	}

	var apiModel TriggerFileMonitorAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse trigger response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TriggerFileMonitorResource) toAPIModel(ctx context.Context, data *TriggerFileMonitorResourceModel) *TriggerFileMonitorAPIModel {
	model := &TriggerFileMonitorAPIModel{
		SysId:           data.SysId.ValueString(),
		Name:            data.Name.ValueString(),
		Type:            "triggerFm",
		Description:     data.Description.ValueString(),
		Enabled:         data.Enabled.ValueBool(),
		TaskMonitor:     data.TaskMonitor.ValueString(),
		TimeZone:        data.TimeZone.ValueString(),
		Calendar:        data.Calendar.ValueString(),
		RestrictedTimes: data.RestrictedTimes.ValueBool(),
		EnabledStart:    data.EnabledStart.ValueString(),
		EnabledEnd:      data.EnabledEnd.ValueString(),
	}

	// Handle tasks list
	if !data.Tasks.IsNull() && !data.Tasks.IsUnknown() {
		var tasks []string
		data.Tasks.ElementsAs(ctx, &tasks, false)
		model.Tasks = tasks
	}

	// Handle variables
	model.Variables = TaskVariablesToAPI(ctx, data.Variables)

	// Handle opswise_groups list
	if !data.OpswiseGroups.IsNull() && !data.OpswiseGroups.IsUnknown() {
		var groups []string
		data.OpswiseGroups.ElementsAs(ctx, &groups, false)
		model.OpswiseGroups = groups
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *TriggerFileMonitorResource) fromAPIModel(ctx context.Context, apiModel *TriggerFileMonitorAPIModel, data *TriggerFileMonitorResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Description = StringValueOrNull(apiModel.Description)
	data.Enabled = types.BoolValue(apiModel.Enabled)

	// File monitor specific
	data.TaskMonitor = StringValueOrNull(apiModel.TaskMonitor)

	// Time restrictions
	data.TimeZone = StringValueOrNull(apiModel.TimeZone)
	data.Calendar = StringValueOrNull(apiModel.Calendar)
	data.RestrictedTimes = types.BoolValue(apiModel.RestrictedTimes)
	data.EnabledStart = StringValueOrNull(apiModel.EnabledStart)
	data.EnabledEnd = StringValueOrNull(apiModel.EnabledEnd)

	// Handle tasks list
	if len(apiModel.Tasks) > 0 {
		tasks, _ := types.ListValueFrom(ctx, types.StringType, apiModel.Tasks)
		data.Tasks = tasks
	}

	// Handle variables
	data.Variables = TaskVariablesFromAPIOrdered(ctx, apiModel.Variables, data.Variables)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
