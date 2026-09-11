package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TaskExclusiveTaskModel describes a single exclusive-task relationship in Terraform.
type TaskExclusiveTaskModel struct {
	Task types.String `tfsdk:"task"`
	Type types.String `tfsdk:"type"`
}

// TaskExclusiveTaskAPIModel represents a single exclusive-task relationship in the API.
type TaskExclusiveTaskAPIModel struct {
	Task string `json:"task,omitempty"`
	Type string `json:"type,omitempty"`
}

// TaskExclusiveTaskAttrTypes returns the attribute types for TaskExclusiveTaskModel.
func TaskExclusiveTaskAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"task": types.StringType,
		"type": types.StringType,
	}
}

// TaskExclusiveTasksSchema returns the schema for the exclusive_tasks attribute.
func TaskExclusiveTasksSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of tasks that cannot run concurrently with this task.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"task": schema.StringAttribute{
					MarkdownDescription: "Name of the task that this task is mutually exclusive with.",
					Required:            true,
				},
				"type": schema.StringAttribute{
					MarkdownDescription: "Type of the exclusive-task relationship. Defaults to 'Direct' when not specified.",
					Optional:            true,
					Computed:            true,
				},
			},
		},
	}
}

// TaskExclusiveTasksToAPI converts a Terraform exclusive_tasks list to API models.
func TaskExclusiveTasksToAPI(ctx context.Context, list types.List) []TaskExclusiveTaskAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var items []TaskExclusiveTaskModel
	list.ElementsAs(ctx, &items, false)

	result := make([]TaskExclusiveTaskAPIModel, len(items))
	for i, item := range items {
		result[i] = TaskExclusiveTaskAPIModel{
			Task: item.Task.ValueString(),
			Type: item.Type.ValueString(),
		}
	}
	return result
}

// TaskExclusiveTasksFromAPI converts API exclusive-task models to a Terraform list.
func TaskExclusiveTasksFromAPI(apiItems []TaskExclusiveTaskAPIModel) types.List {
	if len(apiItems) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: TaskExclusiveTaskAttrTypes()})
	}

	values := make([]attr.Value, len(apiItems))
	for i, item := range apiItems {
		values[i], _ = types.ObjectValue(TaskExclusiveTaskAttrTypes(), map[string]attr.Value{
			"task": types.StringValue(item.Task),
			"type": StringValueOrNull(item.Type),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: TaskExclusiveTaskAttrTypes()}, values)
	return result
}

// TaskExclusiveWithSelfSchema returns the schema for the exclusive_with_self attribute.
func TaskExclusiveWithSelfSchema() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Whether this task cannot run concurrently with another instance of itself.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskVirtualResourceModel describes a single virtual-resource consumption entry in Terraform.
type TaskVirtualResourceModel struct {
	Resource    types.String `tfsdk:"resource"`
	ResourceVar types.String `tfsdk:"resource_var"`
	Amount      types.Int64  `tfsdk:"amount"`
}

// TaskVirtualResourceAPIModel represents a single virtual-resource consumption entry in the API.
type TaskVirtualResourceAPIModel struct {
	Resource    string `json:"resource,omitempty"`
	ResourceVar string `json:"resourceVar,omitempty"`
	Amount      int64  `json:"amount,omitempty"`
}

// TaskVirtualResourceAttrTypes returns the attribute types for TaskVirtualResourceModel.
func TaskVirtualResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"resource":     types.StringType,
		"resource_var": types.StringType,
		"amount":       types.Int64Type,
	}
}

// TaskVirtualResourcesSchema returns the schema for the virtual_resources attribute.
func TaskVirtualResourcesSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of virtual resources consumed by this task during execution.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"resource": schema.StringAttribute{
					MarkdownDescription: "Name of the virtual resource consumed (see `stonebranch_virtual_resource`).",
					Optional:            true,
				},
				"resource_var": schema.StringAttribute{
					MarkdownDescription: "Variable containing the virtual resource name.",
					Optional:            true,
				},
				"amount": schema.Int64Attribute{
					MarkdownDescription: "Amount of the virtual resource consumed by this task.",
					Optional:            true,
				},
			},
		},
	}
}

// TaskVirtualResourcesToAPI converts a Terraform virtual_resources list to API models.
func TaskVirtualResourcesToAPI(ctx context.Context, list types.List) []TaskVirtualResourceAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var items []TaskVirtualResourceModel
	list.ElementsAs(ctx, &items, false)

	result := make([]TaskVirtualResourceAPIModel, len(items))
	for i, item := range items {
		result[i] = TaskVirtualResourceAPIModel{
			Resource:    item.Resource.ValueString(),
			ResourceVar: item.ResourceVar.ValueString(),
			Amount:      item.Amount.ValueInt64(),
		}
	}
	return result
}

// TaskVirtualResourcesFromAPI converts API virtual-resource models to a Terraform list.
func TaskVirtualResourcesFromAPI(apiItems []TaskVirtualResourceAPIModel) types.List {
	if len(apiItems) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: TaskVirtualResourceAttrTypes()})
	}

	values := make([]attr.Value, len(apiItems))
	for i, item := range apiItems {
		amount := types.Int64Null()
		if item.Amount != 0 {
			amount = types.Int64Value(item.Amount)
		}
		values[i], _ = types.ObjectValue(TaskVirtualResourceAttrTypes(), map[string]attr.Value{
			"resource":     StringValueOrNull(item.Resource),
			"resource_var": StringValueOrNull(item.ResourceVar),
			"amount":       amount,
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: TaskVirtualResourceAttrTypes()}, values)
	return result
}

// TaskWaitToStartSchema returns the schema for the wait_to_start attribute.
func TaskWaitToStartSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Determines whether the task must wait before it is eligible to start. Valid values: `None`, `Time` (wait until a specific time), `Duration` (wait for a fixed duration), `Seconds` (wait for a number of seconds), `Relative Time` (wait until a time relative to the task becoming eligible to run). Defaults to `None`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskWaitTimeSchema returns the schema for the wait_time attribute.
func TaskWaitTimeSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Time to wait until, in `HH:MM` format (24-hour). Used when `wait_to_start` is `Time` or `Relative Time`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskWaitDurationSchema returns the schema for the wait_duration attribute.
func TaskWaitDurationSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Duration to wait before starting, in `DD:HH:MM:SS` format (days:hours:minutes:seconds). Used when `wait_to_start` is `Duration`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskWaitAmountSchema returns the schema for the wait_amount attribute.
func TaskWaitAmountSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Number of seconds to wait before starting. Required when `wait_to_start` is `Seconds`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskWaitDayConstraintSchema returns the schema for the wait_day_constraint attribute.
func TaskWaitDayConstraintSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Day constraint applied to the time-based wait. Valid values: `None`, `Same Day`, `Next Day`, `Next Business Day`, `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, `Saturday`. Defaults to `None`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskDelayOnStartSchema returns the schema for the delay_on_start attribute.
func TaskDelayOnStartSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Determines whether the task delays once it becomes eligible to start. Valid values: `None`, `Duration` (delay for a fixed duration), `Seconds` (delay for a number of seconds). Defaults to `None`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskDelayDurationSchema returns the schema for the delay_duration attribute.
func TaskDelayDurationSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Duration to delay before starting, in `DD:HH:MM:SS` format (days:hours:minutes:seconds). Used when `delay_on_start` is `Duration`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskDelayAmountSchema returns the schema for the delay_amount attribute.
func TaskDelayAmountSchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Number of seconds to delay before starting. Required when `delay_on_start` is `Seconds`.",
		Optional:            true,
		Computed:            true,
	}
}

// TaskWorkflowOnlySchema returns the schema for the workflow_only attribute.
func TaskWorkflowOnlySchema() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Controls whether the wait/delay options only apply when this task runs within a workflow. Valid values: `System Default`, `Yes`, `No`. Defaults to `System Default`.",
		Optional:            true,
		Computed:            true,
	}
}

// ValidateTaskWaitDelay validates the cross-field requirements for the shared
// wait/delay options: when wait_to_start or delay_on_start is "Seconds", the
// corresponding *_amount field must be set. Deferred to the server if the
// relevant fields are unknown (e.g. computed from another resource).
func ValidateTaskWaitDelay(waitToStart, waitAmount, delayOnStart, delayAmount types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if !waitToStart.IsUnknown() && waitToStart.ValueString() == "Seconds" {
		if !isSet(waitAmount) && !waitAmount.IsUnknown() {
			diags.AddAttributeError(
				path.Root("wait_amount"),
				"Missing Required Field",
				`"wait_amount" must be set when "wait_to_start" is "Seconds".`,
			)
		}
	}

	if !delayOnStart.IsUnknown() && delayOnStart.ValueString() == "Seconds" {
		if !isSet(delayAmount) && !delayAmount.IsUnknown() {
			diags.AddAttributeError(
				path.Root("delay_amount"),
				"Missing Required Field",
				`"delay_amount" must be set when "delay_on_start" is "Seconds".`,
			)
		}
	}

	return diags
}

// ValidateTaskOutputReturn validates that output_return_sline is set whenever
// output_return_type requests a stream-based capture ("OUTERR", "STDOUT",
// "STDERR"). Confirmed against a real UAC apply: setting output_return_type =
// "OUTERR" without output_return_sline fails at apply time with "Automatic
// Output Retrieval Start Line: Field is required" (status 400) rather than at
// plan time. "FILE" is excluded since it captures from output_return_file
// instead and was not part of the confirmed failure. Deferred to the server if
// either field is unknown (e.g. computed from another resource).
func ValidateTaskOutputReturn(outputReturnType, outputReturnSline types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if outputReturnType.IsUnknown() {
		return diags
	}

	switch outputReturnType.ValueString() {
	case "OUTERR", "STDOUT", "STDERR":
		if !isSet(outputReturnSline) && !outputReturnSline.IsUnknown() {
			diags.AddAttributeError(
				path.Root("output_return_sline"),
				"Missing Required Field",
				`"output_return_sline" must be set when "output_return_type" is "OUTERR", "STDOUT", or "STDERR".`,
			)
		}
	}

	return diags
}
