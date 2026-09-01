package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &CustomDayResource{}
	_ resource.ResourceWithImportState    = &CustomDayResource{}
	_ resource.ResourceWithValidateConfig = &CustomDayResource{}
)

// Known enum values for the custom day "ctype" and "relfreq" fields.
// These are not documented in openapi.yaml; they were determined by
// empirically probing a live UAC instance (see PR description).
var (
	validCustomDayCtypes = []string{
		"Single Date",
		"List of Dates",
		"Absolute Repeating Date",
		"Relative Repeating Date",
	}
	validCustomDayRelFreqs = []string{
		"1st", "2nd", "3rd", "4th", "Last", "Every", "Nth", "Last Day", "Last Business Day",
	}
	validCustomDayMonths    = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	validCustomDayWeekdays  = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	validCustomDayAdjustTyp = []string{"Day", "Business Day"}
)

func NewCustomDayResource() resource.Resource {
	return &CustomDayResource{}
}

// CustomDayResource defines the resource implementation.
type CustomDayResource struct {
	client *client.Client
}

// CustomDayResourceModel describes the resource data model.
type CustomDayResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Content
	Comments types.String `tfsdk:"comments"`

	// Server-computed classification, derived from Holiday/Period.
	Category types.String `tfsdk:"category"`

	// Definition style - exactly one of the field groups below applies,
	// depending on the value of ctype. See ValidateConfig.
	Ctype types.String `tfsdk:"ctype"`

	// ctype = "Single Date"
	Date types.String `tfsdk:"date"`

	// ctype = "List of Dates"
	DateList types.List `tfsdk:"date_list"`

	// ctype = "Absolute Repeating Date" (month+day) or
	// ctype = "Relative Repeating Date" (month+dayofweek+relfreq, or
	// month+relfreq="Nth"+nth_amount+nth_type)
	Month     types.String `tfsdk:"month"`
	Day       types.Int64  `tfsdk:"day"`
	DayOfWeek types.String `tfsdk:"dayofweek"`
	RelFreq   types.String `tfsdk:"relfreq"`
	NthAmount types.Int64  `tfsdk:"nth_amount"`
	NthType   types.String `tfsdk:"nth_type"`

	// Offset adjustment applied to the resolved date.
	Adjustment       types.String `tfsdk:"adjustment"`
	AdjustmentAmount types.Int64  `tfsdk:"adjustment_amount"`
	AdjustmentType   types.String `tfsdk:"adjustment_type"`

	// Holiday / period flags and holiday weekend-observance rules.
	Holiday       types.Bool `tfsdk:"holiday"`
	Period        types.Bool `tfsdk:"period"`
	ObservedRules types.List `tfsdk:"observed_rules"`
}

// ObservedRuleModel describes a single holiday weekend-observance rule.
type ObservedRuleModel struct {
	ActualDayOfWeek   types.String `tfsdk:"actual_day_of_week"`
	ObservedDayOfWeek types.String `tfsdk:"observed_day_of_week"`
}

// ObservedRuleAttrTypes returns the attribute types for ObservedRuleModel.
func ObservedRuleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"actual_day_of_week":   types.StringType,
		"observed_day_of_week": types.StringType,
	}
}

// CustomDayAPIModel represents the API request/response structure.
type CustomDayAPIModel struct {
	SysId    string `json:"sysId,omitempty"`
	Name     string `json:"name"`
	Version  int64  `json:"version,omitempty"`
	Comments string `json:"comments,omitempty"`
	Category string `json:"category,omitempty"`

	Ctype string `json:"ctype,omitempty"`

	Date     string   `json:"date,omitempty"`
	DateList []string `json:"dateList,omitempty"`

	Month     string `json:"month,omitempty"`
	Day       int64  `json:"day,omitempty"`
	DayOfWeek string `json:"dayofweek,omitempty"`
	RelFreq   string `json:"relfreq,omitempty"`
	NthAmount int64  `json:"nthAmount,omitempty"`
	NthType   string `json:"nthType,omitempty"`

	Adjustment       string `json:"adjustment,omitempty"`
	AdjustmentAmount int64  `json:"adjustmentAmount,omitempty"`
	AdjustmentType   string `json:"adjustmentType,omitempty"`

	Holiday       bool                   `json:"holiday,omitempty"`
	Period        bool                   `json:"period,omitempty"`
	ObservedRules []ObservedRuleAPIModel `json:"observedRules,omitempty"`
}

// ObservedRuleAPIModel represents a holiday weekend-observance rule in the API.
type ObservedRuleAPIModel struct {
	ActualDayOfWeek   string `json:"actualDayOfWeek,omitempty"`
	ObservedDayOfWeek string `json:"observedDayOfWeek,omitempty"`
}

func (r *CustomDayResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_day"
}

func (r *CustomDayResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Custom Day. Custom days define calendar exception dates (single dates, date lists, or yearly repeating dates) which can be referenced by calendars and holiday-observance rules.",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the custom day (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the custom day.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the custom day (for optimistic locking).",
				Computed:            true,
			},

			// Content
			"comments": schema.StringAttribute{
				MarkdownDescription: "Comments or description for the custom day.",
				Optional:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Server-computed classification of the custom day: 'Day', 'Holiday', or 'Period'. Derived from `holiday`/`period`.",
				Computed:            true,
			},

			// Definition style
			"ctype": schema.StringAttribute{
				MarkdownDescription: "Definition style for the custom day. One of: 'Single Date' (requires `date`), 'List of Dates' (requires `date_list`), 'Absolute Repeating Date' (requires `month`+`day`), 'Relative Repeating Date' (requires `month`+`relfreq`, plus either `dayofweek` or, when `relfreq` is 'Nth', `nth_amount`+`nth_type`).",
				Required:            true,
			},

			"date": schema.StringAttribute{
				MarkdownDescription: "Specific date for ctype = 'Single Date'. Format: `yyyy-MM-dd`.",
				Optional:            true,
			},
			"date_list": schema.ListAttribute{
				MarkdownDescription: "List of specific dates for ctype = 'List of Dates'. Format: `yyyy-MM-dd`.",
				Optional:            true,
				ElementType:         types.StringType,
			},

			"month": schema.StringAttribute{
				MarkdownDescription: "Month for ctype = 'Absolute Repeating Date' or 'Relative Repeating Date'. Values: 'Jan'-'Dec'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"day": schema.Int64Attribute{
				MarkdownDescription: "Day of the month for ctype = 'Absolute Repeating Date'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dayofweek": schema.StringAttribute{
				MarkdownDescription: "Day of week for ctype = 'Relative Repeating Date' (not used when `relfreq` is 'Nth', 'Last Day', or 'Last Business Day'). Values: 'Sun'-'Sat'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"relfreq": schema.StringAttribute{
				MarkdownDescription: "Relative frequency for ctype = 'Relative Repeating Date'. One of: '1st', '2nd', '3rd', '4th', 'Last', 'Every', 'Nth', 'Last Day', 'Last Business Day'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"nth_amount": schema.Int64Attribute{
				MarkdownDescription: "Nth day-of-month value, used when `relfreq` is 'Nth'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nth_type": schema.StringAttribute{
				MarkdownDescription: "Type of day counted by `nth_amount`, used when `relfreq` is 'Nth'. Values: 'Day' (default), 'Business Day'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Adjustment
			"adjustment": schema.StringAttribute{
				MarkdownDescription: "Offset direction applied to the resolved date. Values: 'None' (default), 'Less', 'Plus'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"adjustment_amount": schema.Int64Attribute{
				MarkdownDescription: "Number of units to offset the resolved date by, used when `adjustment` is not 'None'. Defaults to 1.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"adjustment_type": schema.StringAttribute{
				MarkdownDescription: "Unit of the adjustment offset. Values: 'Day' (default), 'Business Day'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Holiday / period
			"holiday": schema.BoolAttribute{
				MarkdownDescription: "Marks this custom day as a holiday, enabling `observed_rules` for weekend-observance handling. Defaults to false.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"period": schema.BoolAttribute{
				MarkdownDescription: "Marks this custom day as a period (a span rather than discrete days). Not allowed when `ctype` is 'Single Date'. Defaults to false.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"observed_rules": schema.ListNestedAttribute{
				MarkdownDescription: "Weekend-observance rules, used when `holiday` is true. Maps an actual day of week to the day of week it should be observed on instead.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"actual_day_of_week": schema.StringAttribute{
							MarkdownDescription: "Actual day of week the holiday falls on. Values: 'Sun'-'Sat'.",
							Required:            true,
						},
						"observed_day_of_week": schema.StringAttribute{
							MarkdownDescription: "Day of week the holiday should be observed on instead. Values: 'Sun'-'Sat'.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *CustomDayResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig enforces the mutually-exclusive field groups implied by
// ctype/relfreq, since the UAC API validates these server-side but with
// generic error messages. Catching them at plan time gives clearer feedback.
func (r *CustomDayResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data CustomDayResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Ctype.IsUnknown() || data.Ctype.IsNull() {
		return
	}
	ctype := data.Ctype.ValueString()
	if !contains(validCustomDayCtypes, ctype) {
		resp.Diagnostics.AddAttributeError(
			path.Root("ctype"),
			"Invalid Attribute Value",
			fmt.Sprintf("ctype must be one of %v, got: %q.", validCustomDayCtypes, ctype),
		)
		return
	}

	hasDate := isSet(data.Date)
	hasDateList := !data.DateList.IsNull() && !data.DateList.IsUnknown()
	hasMonth := isSet(data.Month)
	hasDay := !data.Day.IsNull() && !data.Day.IsUnknown()
	hasDayOfWeek := isSet(data.DayOfWeek)
	hasRelFreq := isSet(data.RelFreq)
	hasNthAmount := !data.NthAmount.IsNull() && !data.NthAmount.IsUnknown()

	conflict := func(rootField string, msg string) {
		resp.Diagnostics.AddAttributeError(path.Root(rootField), "Conflicting Fields", msg)
	}
	missing := func(rootField string, msg string) {
		resp.Diagnostics.AddAttributeError(path.Root(rootField), "Missing Required Field", msg)
	}

	switch ctype {
	case "Single Date":
		if !hasDate {
			missing("date", `"date" is required when ctype is "Single Date".`)
		}
		if hasDateList || hasMonth || hasDay || hasDayOfWeek || hasRelFreq {
			conflict("ctype", `"date_list", "month", "day", "dayofweek", and "relfreq" must not be set when ctype is "Single Date".`)
		}
		if !data.Period.IsNull() && !data.Period.IsUnknown() && data.Period.ValueBool() {
			conflict("period", `"period" cannot be true when ctype is "Single Date".`)
		}
	case "List of Dates":
		if !hasDateList {
			missing("date_list", `"date_list" is required when ctype is "List of Dates".`)
		}
		if hasDate || hasMonth || hasDay || hasDayOfWeek || hasRelFreq {
			conflict("ctype", `"date", "month", "day", "dayofweek", and "relfreq" must not be set when ctype is "List of Dates".`)
		}
	case "Absolute Repeating Date":
		if !hasMonth || !hasDay {
			missing("ctype", `"month" and "day" are required when ctype is "Absolute Repeating Date".`)
		}
		if hasDate || hasDateList || hasDayOfWeek || hasRelFreq {
			conflict("ctype", `"date", "date_list", "dayofweek", and "relfreq" must not be set when ctype is "Absolute Repeating Date".`)
		}
		if hasMonth && !contains(validCustomDayMonths, data.Month.ValueString()) {
			resp.Diagnostics.AddAttributeError(path.Root("month"), "Invalid Attribute Value", fmt.Sprintf("month must be one of %v, got: %q.", validCustomDayMonths, data.Month.ValueString()))
		}
	case "Relative Repeating Date":
		if !hasMonth || !hasRelFreq {
			missing("ctype", `"month" and "relfreq" are required when ctype is "Relative Repeating Date".`)
		}
		if hasDate || hasDateList || hasDay {
			conflict("ctype", `"date", "date_list", and "day" must not be set when ctype is "Relative Repeating Date".`)
		}
		if hasMonth && !contains(validCustomDayMonths, data.Month.ValueString()) {
			resp.Diagnostics.AddAttributeError(path.Root("month"), "Invalid Attribute Value", fmt.Sprintf("month must be one of %v, got: %q.", validCustomDayMonths, data.Month.ValueString()))
		}
		if hasRelFreq {
			relfreq := data.RelFreq.ValueString()
			if !contains(validCustomDayRelFreqs, relfreq) {
				resp.Diagnostics.AddAttributeError(path.Root("relfreq"), "Invalid Attribute Value", fmt.Sprintf("relfreq must be one of %v, got: %q.", validCustomDayRelFreqs, relfreq))
			}
			switch relfreq {
			case "Nth":
				if !hasNthAmount {
					missing("nth_amount", `"nth_amount" is required when relfreq is "Nth".`)
				}
				if hasDayOfWeek {
					conflict("dayofweek", `"dayofweek" must not be set when relfreq is "Nth" (use "nth_amount" and "nth_type" instead).`)
				}
			case "Last Day", "Last Business Day":
				if hasDayOfWeek {
					conflict("dayofweek", fmt.Sprintf(`"dayofweek" must not be set when relfreq is %q.`, relfreq))
				}
				if hasNthAmount {
					conflict("nth_amount", fmt.Sprintf(`"nth_amount" must not be set when relfreq is %q.`, relfreq))
				}
			default:
				// 1st, 2nd, 3rd, 4th, Last, Every
				if !hasDayOfWeek {
					missing("dayofweek", fmt.Sprintf(`"dayofweek" is required when relfreq is %q.`, relfreq))
				}
				if hasNthAmount {
					conflict("nth_amount", fmt.Sprintf(`"nth_amount" must not be set when relfreq is %q.`, relfreq))
				}
			}
		}
		if hasDayOfWeek && !contains(validCustomDayWeekdays, data.DayOfWeek.ValueString()) {
			resp.Diagnostics.AddAttributeError(path.Root("dayofweek"), "Invalid Attribute Value", fmt.Sprintf("dayofweek must be one of %v, got: %q.", validCustomDayWeekdays, data.DayOfWeek.ValueString()))
		}
	}

	if adj := data.Adjustment; !adj.IsNull() && !adj.IsUnknown() && adj.ValueString() != "" && adj.ValueString() != "None" {
		if !contains([]string{"None", "Less", "Plus"}, adj.ValueString()) {
			resp.Diagnostics.AddAttributeError(path.Root("adjustment"), "Invalid Attribute Value", fmt.Sprintf(`adjustment must be one of ["None" "Less" "Plus"], got: %q.`, adj.ValueString()))
		}
	}
	if at := data.AdjustmentType; !at.IsNull() && !at.IsUnknown() && at.ValueString() != "" && !contains(validCustomDayAdjustTyp, at.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("adjustment_type"), "Invalid Attribute Value", fmt.Sprintf("adjustment_type must be one of %v, got: %q.", validCustomDayAdjustTyp, at.ValueString()))
	}
	if nt := data.NthType; !nt.IsNull() && !nt.IsUnknown() && nt.ValueString() != "" && !contains(validCustomDayAdjustTyp, nt.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("nth_type"), "Invalid Attribute Value", fmt.Sprintf("nth_type must be one of %v, got: %q.", validCustomDayAdjustTyp, nt.ValueString()))
	}
}

func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func (r *CustomDayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomDayResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating custom day", map[string]any{"name": data.Name.ValueString()})

	apiModel := r.toAPIModel(ctx, &data)

	_, err := r.client.Post(ctx, "/resources/customday", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Custom Day",
			fmt.Sprintf("Could not create custom day %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	err = r.readCustomDay(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Custom Day",
			fmt.Sprintf("Could not read custom day %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created custom day", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomDayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomDayResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readCustomDay(ctx, &data)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Custom day not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Custom Day",
			fmt.Sprintf("Could not read custom day %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomDayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomDayResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CustomDayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating custom day", map[string]any{"sys_id": data.SysId.ValueString()})

	apiModel := r.toAPIModel(ctx, &data)

	_, err := r.client.Put(ctx, "/resources/customday", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Custom Day",
			fmt.Sprintf("Could not update custom day %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	err = r.readCustomDay(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Custom Day",
			fmt.Sprintf("Could not read custom day %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomDayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomDayResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting custom day", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("customdayid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/customday", query)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Custom Day",
			fmt.Sprintf("Could not delete custom day %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *CustomDayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readCustomDay fetches the custom day from the API and updates the model.
func (r *CustomDayResource) readCustomDay(ctx context.Context, data *CustomDayResourceModel) error {
	query := url.Values{}
	query.Set("customdayname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/customday", query)
	if err != nil {
		return err
	}

	var apiModel CustomDayAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse custom day response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *CustomDayResource) toAPIModel(ctx context.Context, data *CustomDayResourceModel) *CustomDayAPIModel {
	model := &CustomDayAPIModel{
		SysId:            data.SysId.ValueString(),
		Name:             data.Name.ValueString(),
		Comments:         data.Comments.ValueString(),
		Ctype:            data.Ctype.ValueString(),
		Date:             data.Date.ValueString(),
		Month:            data.Month.ValueString(),
		Day:              data.Day.ValueInt64(),
		DayOfWeek:        data.DayOfWeek.ValueString(),
		RelFreq:          data.RelFreq.ValueString(),
		NthAmount:        data.NthAmount.ValueInt64(),
		NthType:          data.NthType.ValueString(),
		Adjustment:       data.Adjustment.ValueString(),
		AdjustmentAmount: data.AdjustmentAmount.ValueInt64(),
		AdjustmentType:   data.AdjustmentType.ValueString(),
		Holiday:          data.Holiday.ValueBool(),
		Period:           data.Period.ValueBool(),
	}

	if !data.DateList.IsNull() && !data.DateList.IsUnknown() {
		var dates []string
		data.DateList.ElementsAs(ctx, &dates, false)
		model.DateList = dates
	}

	if !data.ObservedRules.IsNull() && !data.ObservedRules.IsUnknown() {
		var rules []ObservedRuleModel
		data.ObservedRules.ElementsAs(ctx, &rules, false)
		for _, rule := range rules {
			model.ObservedRules = append(model.ObservedRules, ObservedRuleAPIModel{
				ActualDayOfWeek:   rule.ActualDayOfWeek.ValueString(),
				ObservedDayOfWeek: rule.ObservedDayOfWeek.ValueString(),
			})
		}
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *CustomDayResource) fromAPIModel(ctx context.Context, apiModel *CustomDayAPIModel, data *CustomDayResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Content
	data.Comments = StringValueOrNull(apiModel.Comments)
	data.Category = StringValueOrNull(apiModel.Category)
	data.Ctype = StringValueOrNull(apiModel.Ctype)

	data.Date = StringValueOrNull(apiModel.Date)
	data.Month = StringValueOrNull(apiModel.Month)
	data.DayOfWeek = StringValueOrNull(apiModel.DayOfWeek)
	data.RelFreq = StringValueOrNull(apiModel.RelFreq)
	data.NthType = StringValueOrNull(apiModel.NthType)
	data.Adjustment = StringValueOrNull(apiModel.Adjustment)
	data.AdjustmentType = StringValueOrNull(apiModel.AdjustmentType)

	data.Day = types.Int64Value(apiModel.Day)
	data.NthAmount = types.Int64Value(apiModel.NthAmount)
	data.AdjustmentAmount = types.Int64Value(apiModel.AdjustmentAmount)

	data.Holiday = types.BoolValue(apiModel.Holiday)
	data.Period = types.BoolValue(apiModel.Period)

	// Date list
	if len(apiModel.DateList) > 0 {
		dateList, _ := types.ListValueFrom(ctx, types.StringType, apiModel.DateList)
		data.DateList = dateList
	} else {
		data.DateList = types.ListNull(types.StringType)
	}

	// Observed rules
	if len(apiModel.ObservedRules) > 0 {
		ruleValues := make([]attr.Value, len(apiModel.ObservedRules))
		for i, rule := range apiModel.ObservedRules {
			ruleValues[i], _ = types.ObjectValue(ObservedRuleAttrTypes(), map[string]attr.Value{
				"actual_day_of_week":   types.StringValue(rule.ActualDayOfWeek),
				"observed_day_of_week": types.StringValue(rule.ObservedDayOfWeek),
			})
		}
		observedRules, _ := types.ListValue(types.ObjectType{AttrTypes: ObservedRuleAttrTypes()}, ruleValues)
		data.ObservedRules = observedRules
	} else {
		data.ObservedRules = types.ListNull(types.ObjectType{AttrTypes: ObservedRuleAttrTypes()})
	}
}
