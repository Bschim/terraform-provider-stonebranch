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
	_ resource.Resource                = &CalendarResource{}
	_ resource.ResourceWithImportState = &CalendarResource{}
)

func NewCalendarResource() resource.Resource {
	return &CalendarResource{}
}

// CalendarResource defines the resource implementation.
type CalendarResource struct {
	client *client.Client
}

// CalendarResourceModel describes the resource data model.
type CalendarResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Content
	Comments       types.String `tfsdk:"comments"`
	FirstDayOfWeek types.String `tfsdk:"first_day_of_week"`

	// Business days (comma-separated: "monday,tuesday,wednesday,thursday,friday")
	BusinessDays types.String `tfsdk:"business_days"`

	// Quarters (month/day format)
	FirstQuarterMonth  types.String `tfsdk:"first_quarter_month"`
	FirstQuarterDay    types.String `tfsdk:"first_quarter_day"`
	SecondQuarterMonth types.String `tfsdk:"second_quarter_month"`
	SecondQuarterDay   types.String `tfsdk:"second_quarter_day"`
	ThirdQuarterMonth  types.String `tfsdk:"third_quarter_month"`
	ThirdQuarterDay    types.String `tfsdk:"third_quarter_day"`
	FourthQuarterMonth types.String `tfsdk:"fourth_quarter_month"`
	FourthQuarterDay   types.String `tfsdk:"fourth_quarter_day"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`

	// Custom days (holiday/period exception dates) linked to this calendar
	CustomDays types.List `tfsdk:"custom_days"`
}

// CalendarAPIModel represents the API request/response structure.
type CalendarAPIModel struct {
	SysId              string            `json:"sysId,omitempty"`
	Name               string            `json:"name"`
	Version            int64             `json:"version,omitempty"`
	Comments           string            `json:"comments,omitempty"`
	FirstDayOfWeek     string            `json:"firstDayOfWeek,omitempty"`
	BusinessDays       *BusinessDaysAPI  `json:"businessDays,omitempty"`
	FirstQuarterStart  *QuarterAPI       `json:"firstQuarterStart,omitempty"`
	SecondQuarterStart *QuarterAPI       `json:"secondQuarterStart,omitempty"`
	ThirdQuarterStart  *QuarterAPI       `json:"thirdQuarterStart,omitempty"`
	FourthQuarterStart *QuarterAPI       `json:"fourthQuarterStart,omitempty"`
	OpswiseGroups      []string          `json:"opswiseGroups,omitempty"`
	CustomDays         []CustomDayRefAPI `json:"customDays,omitempty"`
}

// CustomDayRefAPI represents a reference to a custom day by name.
type CustomDayRefAPI struct {
	Name  string `json:"name,omitempty"`
	SysId string `json:"sysId,omitempty"`
}

// BusinessDaysAPI represents the business days wrapper.
type BusinessDaysAPI struct {
	Value string `json:"value,omitempty"`
}

// QuarterAPI represents a quarter start date.
type QuarterAPI struct {
	Month string `json:"month,omitempty"`
	Day   string `json:"day,omitempty"`
}

func (r *CalendarResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_calendar"
}

func (r *CalendarResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Calendar. Calendars define business days, quarters, and can be used by triggers to control scheduling.",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the calendar (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the calendar.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the calendar (for optimistic locking).",
				Computed:            true,
			},

			// Content
			"comments": schema.StringAttribute{
				MarkdownDescription: "Comments or description for the calendar.",
				Optional:            true,
			},
			"first_day_of_week": schema.StringAttribute{
				MarkdownDescription: "First day of the week. Values: 'Sunday', 'Monday'. Defaults to server setting if not specified.",
				Optional:            true,
				Computed:            true,
			},
			"business_days": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of business days with capitalized names. Example: 'Monday,Tuesday,Wednesday,Thursday,Friday'. Defaults to server setting if not specified.",
				Optional:            true,
				Computed:            true,
			},

			// First Quarter
			"first_quarter_month": schema.StringAttribute{
				MarkdownDescription: "Month when first quarter starts (Jan, Feb, Mar, etc.).",
				Optional:            true,
				Computed:            true,
			},
			"first_quarter_day": schema.StringAttribute{
				MarkdownDescription: "Day when first quarter starts (1-31).",
				Optional:            true,
				Computed:            true,
			},

			// Second Quarter
			"second_quarter_month": schema.StringAttribute{
				MarkdownDescription: "Month when second quarter starts (Jan, Feb, Mar, etc.).",
				Optional:            true,
				Computed:            true,
			},
			"second_quarter_day": schema.StringAttribute{
				MarkdownDescription: "Day when second quarter starts (1-31).",
				Optional:            true,
				Computed:            true,
			},

			// Third Quarter
			"third_quarter_month": schema.StringAttribute{
				MarkdownDescription: "Month when third quarter starts (Jan, Feb, Mar, etc.).",
				Optional:            true,
				Computed:            true,
			},
			"third_quarter_day": schema.StringAttribute{
				MarkdownDescription: "Day when third quarter starts (1-31).",
				Optional:            true,
				Computed:            true,
			},

			// Fourth Quarter
			"fourth_quarter_month": schema.StringAttribute{
				MarkdownDescription: "Month when fourth quarter starts (Jan, Feb, Mar, etc.).",
				Optional:            true,
				Computed:            true,
			},
			"fourth_quarter_day": schema.StringAttribute{
				MarkdownDescription: "Day when fourth quarter starts (1-31).",
				Optional:            true,
				Computed:            true,
			},

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this calendar belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},

			// Custom days
			"custom_days": schema.ListAttribute{
				MarkdownDescription: "List of custom day names (see `stonebranch_custom_day`) that define exception dates (holidays/periods) for this calendar.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *CalendarResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CalendarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CalendarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating calendar", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the calendar
	_, err := r.client.Post(ctx, "/resources/calendar", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Calendar",
			fmt.Sprintf("Could not create calendar %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created calendar to get sysId and other computed fields
	err = r.readCalendar(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Calendar",
			fmt.Sprintf("Could not read calendar %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created calendar", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CalendarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CalendarResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readCalendar(ctx, &data)
	if err != nil {
		// Check if calendar was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Calendar not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Calendar",
			fmt.Sprintf("Could not read calendar %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CalendarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CalendarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state CalendarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating calendar", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the calendar
	_, err := r.client.Put(ctx, "/resources/calendar", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Calendar",
			fmt.Sprintf("Could not update calendar %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readCalendar(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Calendar",
			fmt.Sprintf("Could not read calendar %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CalendarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CalendarResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting calendar", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("calendarid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/calendar", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Calendar",
			fmt.Sprintf("Could not delete calendar %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *CalendarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readCalendar fetches the calendar from the API and updates the model.
func (r *CalendarResource) readCalendar(ctx context.Context, data *CalendarResourceModel) error {
	query := url.Values{}
	query.Set("calendarname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/calendar", query)
	if err != nil {
		return err
	}

	var apiModel CalendarAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse calendar response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *CalendarResource) toAPIModel(ctx context.Context, data *CalendarResourceModel) *CalendarAPIModel {
	model := &CalendarAPIModel{
		SysId:          data.SysId.ValueString(),
		Name:           data.Name.ValueString(),
		Comments:       data.Comments.ValueString(),
		FirstDayOfWeek: data.FirstDayOfWeek.ValueString(),
	}

	// Handle business days
	if !data.BusinessDays.IsNull() && !data.BusinessDays.IsUnknown() && data.BusinessDays.ValueString() != "" {
		model.BusinessDays = &BusinessDaysAPI{
			Value: data.BusinessDays.ValueString(),
		}
	}

	// Handle quarters
	if !data.FirstQuarterMonth.IsNull() || !data.FirstQuarterDay.IsNull() {
		model.FirstQuarterStart = &QuarterAPI{
			Month: data.FirstQuarterMonth.ValueString(),
			Day:   data.FirstQuarterDay.ValueString(),
		}
	}
	if !data.SecondQuarterMonth.IsNull() || !data.SecondQuarterDay.IsNull() {
		model.SecondQuarterStart = &QuarterAPI{
			Month: data.SecondQuarterMonth.ValueString(),
			Day:   data.SecondQuarterDay.ValueString(),
		}
	}
	if !data.ThirdQuarterMonth.IsNull() || !data.ThirdQuarterDay.IsNull() {
		model.ThirdQuarterStart = &QuarterAPI{
			Month: data.ThirdQuarterMonth.ValueString(),
			Day:   data.ThirdQuarterDay.ValueString(),
		}
	}
	if !data.FourthQuarterMonth.IsNull() || !data.FourthQuarterDay.IsNull() {
		model.FourthQuarterStart = &QuarterAPI{
			Month: data.FourthQuarterMonth.ValueString(),
			Day:   data.FourthQuarterDay.ValueString(),
		}
	}

	// Handle opswise_groups list
	if !data.OpswiseGroups.IsNull() && !data.OpswiseGroups.IsUnknown() {
		var groups []string
		data.OpswiseGroups.ElementsAs(ctx, &groups, false)
		model.OpswiseGroups = groups
	}

	// Handle custom_days list
	if !data.CustomDays.IsNull() && !data.CustomDays.IsUnknown() {
		var names []string
		data.CustomDays.ElementsAs(ctx, &names, false)
		for _, name := range names {
			model.CustomDays = append(model.CustomDays, CustomDayRefAPI{Name: name})
		}
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *CalendarResource) fromAPIModel(ctx context.Context, apiModel *CalendarAPIModel, data *CalendarResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Content
	data.Comments = StringValueOrNull(apiModel.Comments)
	data.FirstDayOfWeek = StringValueOrNull(apiModel.FirstDayOfWeek)

	// Business days
	if apiModel.BusinessDays != nil {
		data.BusinessDays = StringValueOrNull(apiModel.BusinessDays.Value)
	} else {
		data.BusinessDays = types.StringNull()
	}

	// Quarters
	if apiModel.FirstQuarterStart != nil {
		data.FirstQuarterMonth = StringValueOrNull(apiModel.FirstQuarterStart.Month)
		data.FirstQuarterDay = StringValueOrNull(apiModel.FirstQuarterStart.Day)
	} else {
		data.FirstQuarterMonth = types.StringNull()
		data.FirstQuarterDay = types.StringNull()
	}
	if apiModel.SecondQuarterStart != nil {
		data.SecondQuarterMonth = StringValueOrNull(apiModel.SecondQuarterStart.Month)
		data.SecondQuarterDay = StringValueOrNull(apiModel.SecondQuarterStart.Day)
	} else {
		data.SecondQuarterMonth = types.StringNull()
		data.SecondQuarterDay = types.StringNull()
	}
	if apiModel.ThirdQuarterStart != nil {
		data.ThirdQuarterMonth = StringValueOrNull(apiModel.ThirdQuarterStart.Month)
		data.ThirdQuarterDay = StringValueOrNull(apiModel.ThirdQuarterStart.Day)
	} else {
		data.ThirdQuarterMonth = types.StringNull()
		data.ThirdQuarterDay = types.StringNull()
	}
	if apiModel.FourthQuarterStart != nil {
		data.FourthQuarterMonth = StringValueOrNull(apiModel.FourthQuarterStart.Month)
		data.FourthQuarterDay = StringValueOrNull(apiModel.FourthQuarterStart.Day)
	} else {
		data.FourthQuarterMonth = types.StringNull()
		data.FourthQuarterDay = types.StringNull()
	}

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	} else {
		data.OpswiseGroups = types.ListNull(types.StringType)
	}

	// Handle custom_days
	if len(apiModel.CustomDays) > 0 {
		names := make([]string, len(apiModel.CustomDays))
		for i, cd := range apiModel.CustomDays {
			names[i] = cd.Name
		}
		customDays, _ := types.ListValueFrom(ctx, types.StringType, names)
		data.CustomDays = customDays
	} else {
		data.CustomDays = types.ListNull(types.StringType)
	}
}
