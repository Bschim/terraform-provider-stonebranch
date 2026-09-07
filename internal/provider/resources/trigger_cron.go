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
	_ resource.Resource                = &TriggerCronResource{}
	_ resource.ResourceWithImportState = &TriggerCronResource{}
)

func NewTriggerCronResource() resource.Resource {
	return &TriggerCronResource{}
}

// TriggerCronResource defines the resource implementation.
type TriggerCronResource struct {
	client *client.Client
}

// TriggerCronResourceModel describes the resource data model.
type TriggerCronResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`

	// Tasks to trigger (required)
	Tasks types.List `tfsdk:"tasks"`

	// Cron expression fields
	Minutes    types.String `tfsdk:"minutes"`
	Hours      types.String `tfsdk:"hours"`
	DayOfMonth types.String `tfsdk:"day_of_month"`
	Month      types.String `tfsdk:"month"`
	DayOfWeek  types.String `tfsdk:"day_of_week"`
	DayLogic   types.String `tfsdk:"day_logic"`

	// Time zone
	TimeZone types.String `tfsdk:"time_zone"`

	// Calendar
	Calendar types.String `tfsdk:"calendar"`

	// Variables
	Variables types.List `tfsdk:"variables"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// TriggerCronAPIModel represents the API request/response structure.
type TriggerCronAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version int64  `json:"version,omitempty"`

	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`

	Tasks []string `json:"tasks,omitempty"`

	// Cron expression fields
	Minutes    string `json:"minutes,omitempty"`
	Hours      string `json:"hours,omitempty"`
	DayOfMonth string `json:"dayOfMonth,omitempty"`
	Month      string `json:"month,omitempty"`
	DayOfWeek  string `json:"dayOfWeek,omitempty"`
	DayLogic   string `json:"dayLogic,omitempty"`

	TimeZone string `json:"timeZone,omitempty"`
	Calendar string `json:"calendar,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TriggerCronResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trigger_cron"
}

func (r *TriggerCronResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Cron Trigger. A cron trigger executes associated tasks based on a cron expression schedule.",

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
				MarkdownDescription: "List of task names to execute when the trigger fires. At least one task is required.",
				Required:            true,
				ElementType:         types.StringType,
			},

			// Cron expression fields
			"minutes": schema.StringAttribute{
				MarkdownDescription: "Cron minutes field (0-59). Supports values, ranges, lists, and special characters (*, /, -).",
				Required:            true,
			},
			"hours": schema.StringAttribute{
				MarkdownDescription: "Cron hours field (0-23). Supports values, ranges, lists, and special characters (*, /, -).",
				Required:            true,
			},
			"day_of_month": schema.StringAttribute{
				MarkdownDescription: "Cron day of month field (1-31). Supports values, ranges, lists, and special characters (*, /, -, L, W).",
				Required:            true,
			},
			"month": schema.StringAttribute{
				MarkdownDescription: "Cron month field (1-12 or JAN-DEC). Supports values, ranges, lists, and special characters (*, /).",
				Required:            true,
			},
			"day_of_week": schema.StringAttribute{
				MarkdownDescription: "Cron day of week field (0-6 or SUN-SAT, where 0=Sunday). Supports values, ranges, lists, and special characters (*, /, -, L, #).",
				Required:            true,
			},
			"day_logic": schema.StringAttribute{
				MarkdownDescription: "Logic for combining day_of_month and day_of_week: 'And' (both must match) or 'Or' (either can match). Defaults to 'Or'.",
				Optional:            true,
				Computed:            true,
			},

			// Time zone
			"time_zone": schema.StringAttribute{
				MarkdownDescription: "Time zone for the trigger schedule (e.g., 'America/New_York', 'UTC').",
				Optional:            true,
				Computed:            true,
			},

			// Calendar
			"calendar": schema.StringAttribute{
				MarkdownDescription: "Name of the calendar to use for scheduling restrictions.",
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

func (r *TriggerCronResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TriggerCronResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TriggerCronResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating cron trigger", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the trigger
	_, err := r.client.Post(ctx, "/resources/trigger", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Cron Trigger",
			fmt.Sprintf("Could not create cron trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created trigger to get sysId and other computed fields
	err = r.readTrigger(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Cron Trigger",
			fmt.Sprintf("Could not read cron trigger %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created cron trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerCronResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TriggerCronResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTrigger(ctx, &data)
	if err != nil {
		// Check if trigger was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Cron trigger not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Cron Trigger",
			fmt.Sprintf("Could not read cron trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerCronResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TriggerCronResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TriggerCronResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating cron trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the trigger
	_, err := r.client.Put(ctx, "/resources/trigger", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Cron Trigger",
			fmt.Sprintf("Could not update cron trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTrigger(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Cron Trigger",
			fmt.Sprintf("Could not read cron trigger %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerCronResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TriggerCronResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting cron trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("triggerid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/trigger", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Cron Trigger",
			fmt.Sprintf("Could not delete cron trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TriggerCronResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTrigger fetches the trigger from the API and updates the model.
func (r *TriggerCronResource) readTrigger(ctx context.Context, data *TriggerCronResourceModel) error {
	query := url.Values{}
	query.Set("triggername", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/trigger", query)
	if err != nil {
		return err
	}

	var apiModel TriggerCronAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse trigger response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TriggerCronResource) toAPIModel(ctx context.Context, data *TriggerCronResourceModel) *TriggerCronAPIModel {
	model := &TriggerCronAPIModel{
		SysId:       data.SysId.ValueString(),
		Name:        data.Name.ValueString(),
		Type:        "triggerCron",
		Description: data.Description.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
		Minutes:     data.Minutes.ValueString(),
		Hours:       data.Hours.ValueString(),
		DayOfMonth:  data.DayOfMonth.ValueString(),
		Month:       data.Month.ValueString(),
		DayOfWeek:   data.DayOfWeek.ValueString(),
		DayLogic:    data.DayLogic.ValueString(),
		TimeZone:    data.TimeZone.ValueString(),
		Calendar:    data.Calendar.ValueString(),
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
func (r *TriggerCronResource) fromAPIModel(ctx context.Context, apiModel *TriggerCronAPIModel, data *TriggerCronResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Description = StringValueOrNull(apiModel.Description)
	data.Enabled = types.BoolValue(apiModel.Enabled)

	// Cron expression fields
	data.Minutes = StringValueOrNull(apiModel.Minutes)
	data.Hours = StringValueOrNull(apiModel.Hours)
	data.DayOfMonth = StringValueOrNull(apiModel.DayOfMonth)
	data.Month = StringValueOrNull(apiModel.Month)
	data.DayOfWeek = StringValueOrNull(apiModel.DayOfWeek)
	data.DayLogic = StringValueOrNull(apiModel.DayLogic)

	// Time zone
	data.TimeZone = StringValueOrNull(apiModel.TimeZone)

	// Calendar
	data.Calendar = StringValueOrNull(apiModel.Calendar)

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
