package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
