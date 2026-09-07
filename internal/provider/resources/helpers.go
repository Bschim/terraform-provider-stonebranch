package resources

import (
	"context"
	"reflect"

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

// TaskVariablesFromAPIOrdered converts API variable models to a Terraform list,
// re-ordering them (matched by name) to match priorOrder. The `variables`
// attribute is not Computed, so its final value must match the plan exactly;
// the UAC API does not preserve the order variables were submitted in, so
// blindly using its response order can trip Terraform's "inconsistent result
// after apply" check even though the actual name/value pairs are unchanged.
// Variables present in apiVars but not in priorOrder (e.g. added server-side)
// are appended at the end in API order.
func TaskVariablesFromAPIOrdered(ctx context.Context, apiVars []TaskVariableAPIModel, priorOrder types.List) types.List {
	if len(apiVars) == 0 || priorOrder.IsNull() || priorOrder.IsUnknown() {
		return TaskVariablesFromAPI(ctx, apiVars)
	}

	var priorVars []TaskVariableModel
	priorOrder.ElementsAs(ctx, &priorVars, false)

	byName := make(map[string]TaskVariableAPIModel, len(apiVars))
	for _, v := range apiVars {
		byName[v.Name] = v
	}

	ordered := make([]TaskVariableAPIModel, 0, len(apiVars))
	seen := make(map[string]bool, len(apiVars))
	for _, pv := range priorVars {
		name := pv.Name.ValueString()
		if v, ok := byName[name]; ok && !seen[name] {
			ordered = append(ordered, v)
			seen[name] = true
		}
	}
	for _, v := range apiVars {
		if !seen[v.Name] {
			ordered = append(ordered, v)
			seen[v.Name] = true
		}
	}

	return TaskVariablesFromAPI(ctx, ordered)
}

// countFieldMatches counts how many exported struct fields two values of the
// same type share equal values for (via reflect.DeepEqual), recursing into
// anonymous (embedded) struct fields — e.g. commonActionAPIModel — since
// those are themselves unexported at the top level and Field.Interface()
// would panic on them directly.
func countFieldMatches(va, vb reflect.Value) int {
	ta := va.Type()
	n := 0
	for i := 0; i < va.NumField(); i++ {
		f := ta.Field(i)
		fa, fb := va.Field(i), vb.Field(i)
		if f.Anonymous && fa.Kind() == reflect.Struct {
			n += countFieldMatches(fa, fb)
			continue
		}
		if f.PkgPath != "" {
			// Unexported, non-embedded field — skip, can't Interface() it.
			continue
		}
		if reflect.DeepEqual(fa.Interface(), fb.Interface()) {
			n++
		}
	}
	return n
}

// reorderToMatchPrior reorders apiItems to best match the element order of
// priorItems. Several UAC list-typed attributes (e.g. actions.*, variables)
// are not Computed, so Terraform requires the applied result to match the
// plan exactly — but the UAC API does not preserve the order elements were
// submitted in, which trips Terraform's "inconsistent result after apply"
// check even when the actual elements are unchanged, just reordered.
//
// Matches are chosen greedily by counting how many top-level struct fields
// two elements share (via reflect.DeepEqual per field), which tolerates
// Computed fields the server fills in/normalizes that weren't set in the
// prior config/state. Elements with no good match (e.g. genuinely new ones)
// are appended at the end in API order.
func reorderToMatchPrior[T any](apiItems []T, priorItems []T) []T {
	if len(apiItems) < 2 || len(priorItems) == 0 {
		return apiItems
	}

	score := func(a, b T) int {
		return countFieldMatches(reflect.ValueOf(a), reflect.ValueOf(b))
	}

	used := make([]bool, len(apiItems))
	ordered := make([]T, 0, len(apiItems))
	for _, p := range priorItems {
		best, bestScore := -1, -1
		for i, a := range apiItems {
			if used[i] {
				continue
			}
			if s := score(a, p); s > bestScore {
				best, bestScore = i, s
			}
		}
		if best >= 0 {
			used[best] = true
			ordered = append(ordered, apiItems[best])
		}
	}
	for i, a := range apiItems {
		if !used[i] {
			ordered = append(ordered, a)
		}
	}
	return ordered
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
