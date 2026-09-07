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
	_ resource.Resource                = &TriggerTimeResource{}
	_ resource.ResourceWithImportState = &TriggerTimeResource{}
)

func NewTriggerTimeResource() resource.Resource {
	return &TriggerTimeResource{}
}

// TriggerTimeResource defines the resource implementation.
type TriggerTimeResource struct {
	client *client.Client
}

// TriggerTimeResourceModel describes the resource data model.
type TriggerTimeResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`

	// Tasks to trigger (required)
	Tasks types.List `tfsdk:"tasks"`

	// Time configuration
	Time              types.String `tfsdk:"time"`
	TimeZone          types.String `tfsdk:"time_zone"`
	TimeStyle         types.String `tfsdk:"time_style"`
	TimeInterval      types.Int64  `tfsdk:"time_interval"`
	TimeIntervalUnits types.String `tfsdk:"time_interval_units"`

	// Day configuration
	DayStyle    types.String `tfsdk:"day_style"`
	DayInterval types.Int64  `tfsdk:"day_interval"`

	// Day of week flags (for weekly schedules)
	Sunday    types.Bool `tfsdk:"sunday"`
	Monday    types.Bool `tfsdk:"monday"`
	Tuesday   types.Bool `tfsdk:"tuesday"`
	Wednesday types.Bool `tfsdk:"wednesday"`
	Thursday  types.Bool `tfsdk:"thursday"`
	Friday    types.Bool `tfsdk:"friday"`
	Saturday  types.Bool `tfsdk:"saturday"`

	// Calendar
	Calendar types.String `tfsdk:"calendar"`

	// Variables
	Variables types.List `tfsdk:"variables"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// TriggerTimeAPIModel represents the API request/response structure.
type TriggerTimeAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version int64  `json:"version,omitempty"`

	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`

	Tasks []string `json:"tasks,omitempty"`

	Time              string `json:"time,omitempty"`
	TimeZone          string `json:"timeZone,omitempty"`
	TimeStyle         string `json:"timeStyle,omitempty"`
	TimeInterval      int64  `json:"timeInterval,omitempty"`
	TimeIntervalUnits string `json:"timeIntervalUnits,omitempty"`

	DayStyle    string `json:"dayStyle,omitempty"`
	DayInterval int64  `json:"dayInterval,omitempty"`

	Sun bool `json:"sun,omitempty"`
	Mon bool `json:"mon,omitempty"`
	Tue bool `json:"tue,omitempty"`
	Wed bool `json:"wed,omitempty"`
	Thu bool `json:"thu,omitempty"`
	Fri bool `json:"fri,omitempty"`
	Sat bool `json:"sat,omitempty"`

	Calendar string `json:"calendar,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TriggerTimeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trigger_time"
}

func (r *TriggerTimeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Time Trigger. A time trigger executes associated tasks based on a time schedule.",

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

			// Time configuration
			"time": schema.StringAttribute{
				MarkdownDescription: "Time to trigger (e.g., '12:00', '09:30'). Format depends on time_style.",
				Optional:            true,
			},
			"time_zone": schema.StringAttribute{
				MarkdownDescription: "Time zone for the trigger schedule (e.g., 'America/New_York').",
				Optional:            true,
				Computed:            true,
			},
			"time_style": schema.StringAttribute{
				MarkdownDescription: "Time style: 'Once' for single time, 'Interval' for repeating intervals, 'Offset' for offset-based.",
				Optional:            true,
				Computed:            true,
			},
			"time_interval": schema.Int64Attribute{
				MarkdownDescription: "Interval between triggers (when time_style is 'Interval').",
				Optional:            true,
				Computed:            true,
			},
			"time_interval_units": schema.StringAttribute{
				MarkdownDescription: "Units for time_interval: 'Seconds', 'Minutes', 'Hours'.",
				Optional:            true,
				Computed:            true,
			},

			// Day configuration
			"day_style": schema.StringAttribute{
				MarkdownDescription: "Day style: 'Everyday', 'Interval', 'Specific Days', 'Specific Dates', 'Complex'.",
				Optional:            true,
				Computed:            true,
			},
			"day_interval": schema.Int64Attribute{
				MarkdownDescription: "Interval between days (when day_style is 'Interval').",
				Optional:            true,
				Computed:            true,
			},

			// Day of week flags
			"sunday": schema.BoolAttribute{
				MarkdownDescription: "Trigger on Sunday (when day_style is 'Specific Days').",
				Optional:            true,
				Computed:            true,
			},
			"monday": schema.BoolAttribute{
				MarkdownDescription: "Trigger on Monday (when day_style is 'Specific Days').",
				Optional:            true,
				Computed:            true,
			},
			"tuesday": schema.BoolAttribute{
				MarkdownDescription: "Trigger on Tuesday (when day_style is 'Specific Days').",
				Optional:            true,
				Computed:            true,
			},
			"wednesday": schema.BoolAttribute{
				MarkdownDescription: "Trigger on Wednesday (when day_style is 'Specific Days').",
				Optional:            true,
				Computed:            true,
			},
			"thursday": schema.BoolAttribute{
				MarkdownDescription: "Trigger on Thursday (when day_style is 'Specific Days').",
				Optional:            true,
				Computed:            true,
			},
			"friday": schema.BoolAttribute{
				MarkdownDescription: "Trigger on Friday (when day_style is 'Specific Days').",
				Optional:            true,
				Computed:            true,
			},
			"saturday": schema.BoolAttribute{
				MarkdownDescription: "Trigger on Saturday (when day_style is 'Specific Days').",
				Optional:            true,
				Computed:            true,
			},

			// Calendar
			"calendar": schema.StringAttribute{
				MarkdownDescription: "Name of the calendar to use for scheduling.",
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

func (r *TriggerTimeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TriggerTimeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TriggerTimeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating time trigger", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the trigger
	_, err := r.client.Post(ctx, "/resources/trigger", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Time Trigger",
			fmt.Sprintf("Could not create time trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created trigger to get sysId and other computed fields
	err = r.readTrigger(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Time Trigger",
			fmt.Sprintf("Could not read time trigger %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created time trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerTimeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TriggerTimeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTrigger(ctx, &data)
	if err != nil {
		// Check if trigger was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Time trigger not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Time Trigger",
			fmt.Sprintf("Could not read time trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerTimeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TriggerTimeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TriggerTimeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating time trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the trigger
	_, err := r.client.Put(ctx, "/resources/trigger", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Time Trigger",
			fmt.Sprintf("Could not update time trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTrigger(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Time Trigger",
			fmt.Sprintf("Could not read time trigger %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TriggerTimeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TriggerTimeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting time trigger", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("triggerid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/trigger", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Time Trigger",
			fmt.Sprintf("Could not delete time trigger %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TriggerTimeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTrigger fetches the trigger from the API and updates the model.
func (r *TriggerTimeResource) readTrigger(ctx context.Context, data *TriggerTimeResourceModel) error {
	query := url.Values{}
	query.Set("triggername", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/trigger", query)
	if err != nil {
		return err
	}

	var apiModel TriggerTimeAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse trigger response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TriggerTimeResource) toAPIModel(ctx context.Context, data *TriggerTimeResourceModel) *TriggerTimeAPIModel {
	model := &TriggerTimeAPIModel{
		SysId:             data.SysId.ValueString(),
		Name:              data.Name.ValueString(),
		Type:              "triggerTime",
		Description:       data.Description.ValueString(),
		Enabled:           data.Enabled.ValueBool(),
		Time:              data.Time.ValueString(),
		TimeZone:          data.TimeZone.ValueString(),
		TimeStyle:         data.TimeStyle.ValueString(),
		TimeInterval:      data.TimeInterval.ValueInt64(),
		TimeIntervalUnits: data.TimeIntervalUnits.ValueString(),
		DayStyle:          data.DayStyle.ValueString(),
		DayInterval:       data.DayInterval.ValueInt64(),
		Sun:               data.Sunday.ValueBool(),
		Mon:               data.Monday.ValueBool(),
		Tue:               data.Tuesday.ValueBool(),
		Wed:               data.Wednesday.ValueBool(),
		Thu:               data.Thursday.ValueBool(),
		Fri:               data.Friday.ValueBool(),
		Sat:               data.Saturday.ValueBool(),
		Calendar:          data.Calendar.ValueString(),
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
func (r *TriggerTimeResource) fromAPIModel(ctx context.Context, apiModel *TriggerTimeAPIModel, data *TriggerTimeResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Description = StringValueOrNull(apiModel.Description)
	data.Enabled = types.BoolValue(apiModel.Enabled)

	// Time configuration
	data.Time = StringValueOrNull(apiModel.Time)
	data.TimeZone = StringValueOrNull(apiModel.TimeZone)
	data.TimeStyle = StringValueOrNull(apiModel.TimeStyle)
	data.TimeInterval = types.Int64Value(apiModel.TimeInterval)
	data.TimeIntervalUnits = StringValueOrNull(apiModel.TimeIntervalUnits)

	// Day configuration
	data.DayStyle = StringValueOrNull(apiModel.DayStyle)
	data.DayInterval = types.Int64Value(apiModel.DayInterval)

	// Day of week flags
	data.Sunday = types.BoolValue(apiModel.Sun)
	data.Monday = types.BoolValue(apiModel.Mon)
	data.Tuesday = types.BoolValue(apiModel.Tue)
	data.Wednesday = types.BoolValue(apiModel.Wed)
	data.Thursday = types.BoolValue(apiModel.Thu)
	data.Friday = types.BoolValue(apiModel.Fri)
	data.Saturday = types.BoolValue(apiModel.Sat)

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
