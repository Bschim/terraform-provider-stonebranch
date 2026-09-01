package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TaskVariableModel describes a task variable in Terraform.
type TaskVariableModel struct {
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
}

// TaskVariableAPIModel represents a task variable in the API.
type TaskVariableAPIModel struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
}

// TaskVariableAttrTypes returns the attribute types for TaskVariableModel.
func TaskVariableAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"value":       types.StringType,
		"description": types.StringType,
	}
}

// TaskVariablesSchema returns the schema for the variables attribute.
func TaskVariablesSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of task variables. These variables are scoped to the task and can be referenced using `${variable_name}` syntax.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name of the variable.",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value of the variable.",
					Optional:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Description of the variable.",
					Optional:            true,
				},
			},
		},
	}
}

// TaskVariablesToAPI converts Terraform variables list to API models.
func TaskVariablesToAPI(ctx context.Context, variables types.List) []TaskVariableAPIModel {
	if variables.IsNull() || variables.IsUnknown() {
		return nil
	}

	var vars []TaskVariableModel
	variables.ElementsAs(ctx, &vars, false)

	result := make([]TaskVariableAPIModel, len(vars))
	for i, v := range vars {
		result[i] = TaskVariableAPIModel{
			Name:        v.Name.ValueString(),
			Value:       v.Value.ValueString(),
			Description: v.Description.ValueString(),
		}
	}
	return result
}

// TaskVariablesFromAPI converts API variable models to Terraform list.
func TaskVariablesFromAPI(ctx context.Context, apiVars []TaskVariableAPIModel) types.List {
	if len(apiVars) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: TaskVariableAttrTypes()})
	}

	varValues := make([]attr.Value, len(apiVars))
	for i, v := range apiVars {
		varValues[i], _ = types.ObjectValue(TaskVariableAttrTypes(), map[string]attr.Value{
			"name":        types.StringValue(v.Name),
			"value":       StringValueOrNull(v.Value),
			"description": StringValueOrNull(v.Description),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: TaskVariableAttrTypes()}, varValues)
	return result
}

// StringValueOrNull returns a StringValue if s is non-empty, otherwise StringNull.
func StringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// StringValueOrDefault returns the string value or a default if null/unknown/empty.
func StringValueOrDefault(s types.String, defaultValue string) string {
	if s.IsNull() || s.IsUnknown() || s.ValueString() == "" {
		return defaultValue
	}
	return s.ValueString()
}

// isSet reports whether a string attribute has a non-empty known value.
func isSet(s types.String) bool {
	return !s.IsNull() && !s.IsUnknown() && s.ValueString() != ""
}
