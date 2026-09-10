package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TaskRunCriterionModel describes a single Run/Skip Criteria entry in Terraform.
type TaskRunCriterionModel struct {
	Type        types.String `tfsdk:"type"`
	Task        types.String `tfsdk:"task"`
	VertexId    types.String `tfsdk:"vertex_id"`
	Description types.String `tfsdk:"description"`

	BusinessDay types.Bool `tfsdk:"business_day"`
	Holiday     types.Bool `tfsdk:"holiday"`

	HolidayAdjustment       types.String `tfsdk:"holiday_adjustment"`
	HolidayAdjustmentAmount types.Int64  `tfsdk:"holiday_adjustment_amount"`
	HolidayAdjustmentType   types.String `tfsdk:"holiday_adjustment_type"`

	SpecificDay    types.Bool `tfsdk:"specific_day"`
	SpecificDaySun types.Bool `tfsdk:"specific_day_sun"`
	SpecificDayMon types.Bool `tfsdk:"specific_day_mon"`
	SpecificDayTue types.Bool `tfsdk:"specific_day_tue"`
	SpecificDayWed types.Bool `tfsdk:"specific_day_wed"`
	SpecificDayThu types.Bool `tfsdk:"specific_day_thu"`
	SpecificDayFri types.Bool `tfsdk:"specific_day_fri"`
	SpecificDaySat types.Bool `tfsdk:"specific_day_sat"`

	CustomDay       types.Bool   `tfsdk:"custom_day"`
	CustomDayChoice types.String `tfsdk:"custom_day_choice"`

	Variable      types.Bool   `tfsdk:"variable"`
	EvaluateAt    types.String `tfsdk:"evaluate_at"`
	VariableName  types.String `tfsdk:"variable_name"`
	VariableValue types.String `tfsdk:"variable_value"`
	VariableOp    types.String `tfsdk:"variable_op"`

	Complex                 types.Bool   `tfsdk:"complex"`
	ComplexAdjective        types.String `tfsdk:"complex_adjective"`
	ComplexNoun             types.String `tfsdk:"complex_noun"`
	ComplexQualifier        types.String `tfsdk:"complex_qualifier"`
	ComplexNthAmount        types.Int64  `tfsdk:"complex_nth_amount"`
	ComplexAdjustment       types.String `tfsdk:"complex_adjustment"`
	ComplexAdjustmentAmount types.Int64  `tfsdk:"complex_adjustment_amount"`
	ComplexAdjustmentType   types.String `tfsdk:"complex_adjustment_type"`
}

// complexDayWrapperAPIModel mirrors the API's ComplexDayWrapperWsData shape.
type complexDayWrapperAPIModel struct {
	Value string `json:"value,omitempty"`
}

// TaskRunCriterionAPIModel represents a single Run/Skip Criteria entry in the API.
type TaskRunCriterionAPIModel struct {
	Type        string `json:"type,omitempty"`
	Task        string `json:"task,omitempty"`
	VertexId    string `json:"vertexId,omitempty"`
	Description string `json:"description,omitempty"`

	BusinessDay bool `json:"businessDay,omitempty"`
	Holiday     bool `json:"holiday,omitempty"`

	HolidayAdjustment       string `json:"holidayAdjustment,omitempty"`
	HolidayAdjustmentAmount int64  `json:"holidayAdjustmentAmount,omitempty"`
	HolidayAdjustmentType   string `json:"holidayAdjustmentType,omitempty"`

	SpecificDay    bool `json:"specificDay,omitempty"`
	SpecificDaySun bool `json:"specificDaySun,omitempty"`
	SpecificDayMon bool `json:"specificDayMon,omitempty"`
	SpecificDayTue bool `json:"specificDayTue,omitempty"`
	SpecificDayWed bool `json:"specificDayWed,omitempty"`
	SpecificDayThu bool `json:"specificDayThu,omitempty"`
	SpecificDayFri bool `json:"specificDayFri,omitempty"`
	SpecificDaySat bool `json:"specificDaySat,omitempty"`

	CustomDay       bool   `json:"customDay,omitempty"`
	CustomDayChoice string `json:"customDayChoice,omitempty"`

	Variable      bool   `json:"variable,omitempty"`
	EvaluateAt    string `json:"evaluateAt,omitempty"`
	VariableName  string `json:"variableName,omitempty"`
	VariableValue string `json:"variableValue,omitempty"`
	VariableOp    string `json:"variableOp,omitempty"`

	Complex                 bool                        `json:"complex,omitempty"`
	ComplexAdjective        string                      `json:"complexAdjective,omitempty"`
	ComplexNoun             *complexDayWrapperAPIModel  `json:"complexNoun,omitempty"`
	ComplexNouns            []complexDayWrapperAPIModel `json:"complexNouns,omitempty"`
	ComplexQualifier        *complexDayWrapperAPIModel  `json:"complexQualifier,omitempty"`
	ComplexQualifiers       []complexDayWrapperAPIModel `json:"complexQualifiers,omitempty"`
	ComplexNthAmount        int64                       `json:"complexNthAmount,omitempty"`
	ComplexAdjustment       string                      `json:"complexAdjustment,omitempty"`
	ComplexAdjustmentAmount int64                       `json:"complexAdjustmentAmount,omitempty"`
	ComplexAdjustmentType   string                      `json:"complexAdjustmentType,omitempty"`
}

// TaskRunCriterionAttrTypes returns the attribute types for TaskRunCriterionModel.
func TaskRunCriterionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":        types.StringType,
		"task":        types.StringType,
		"vertex_id":   types.StringType,
		"description": types.StringType,

		"business_day": types.BoolType,
		"holiday":      types.BoolType,

		"holiday_adjustment":        types.StringType,
		"holiday_adjustment_amount": types.Int64Type,
		"holiday_adjustment_type":   types.StringType,

		"specific_day":     types.BoolType,
		"specific_day_sun": types.BoolType,
		"specific_day_mon": types.BoolType,
		"specific_day_tue": types.BoolType,
		"specific_day_wed": types.BoolType,
		"specific_day_thu": types.BoolType,
		"specific_day_fri": types.BoolType,
		"specific_day_sat": types.BoolType,

		"custom_day":        types.BoolType,
		"custom_day_choice": types.StringType,

		"variable":       types.BoolType,
		"evaluate_at":    types.StringType,
		"variable_name":  types.StringType,
		"variable_value": types.StringType,
		"variable_op":    types.StringType,

		"complex":                   types.BoolType,
		"complex_adjective":         types.StringType,
		"complex_noun":              types.StringType,
		"complex_qualifier":         types.StringType,
		"complex_nth_amount":        types.Int64Type,
		"complex_adjustment":        types.StringType,
		"complex_adjustment_amount": types.Int64Type,
		"complex_adjustment_type":   types.StringType,
	}
}

// TaskRunCriteriaSchema returns the schema for the run_criteria attribute (workflow tasks only).
func TaskRunCriteriaSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of Run/Skip Criteria that control whether a task within the workflow executes.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					MarkdownDescription: "Criteria type. Valid values: 'Run Criteria', 'Skip Criteria'.",
					Required:            true,
				},
				"task": schema.StringAttribute{
					MarkdownDescription: "Name of the task within the workflow that this criterion applies to.",
					Required:            true,
				},
				"vertex_id": schema.StringAttribute{
					MarkdownDescription: "Identifier of the workflow vertex this criterion applies to. Defaults to 'Any'.",
					Optional:            true,
					Computed:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Description of the criterion.",
					Optional:            true,
				},
				"business_day": schema.BoolAttribute{
					MarkdownDescription: "Whether this criterion is based on a business day.",
					Optional:            true,
					Computed:            true,
				},
				"holiday": schema.BoolAttribute{
					MarkdownDescription: "Whether this criterion is based on a holiday.",
					Optional:            true,
					Computed:            true,
				},
				"holiday_adjustment": schema.StringAttribute{
					MarkdownDescription: "Holiday adjustment type. Valid values: 'None', 'Before', 'After'.",
					Optional:            true,
					Computed:            true,
				},
				"holiday_adjustment_amount": schema.Int64Attribute{
					MarkdownDescription: "Number of adjustment units to apply for the holiday adjustment.",
					Optional:            true,
					Computed:            true,
				},
				"holiday_adjustment_type": schema.StringAttribute{
					MarkdownDescription: "Unit of the holiday adjustment (e.g. 'Day').",
					Optional:            true,
					Computed:            true,
				},
				"specific_day": schema.BoolAttribute{
					MarkdownDescription: "Whether this criterion is based on specific days of the week.",
					Optional:            true,
					Computed:            true,
				},
				"specific_day_sun": schema.BoolAttribute{
					MarkdownDescription: "Whether Sunday is included when specific_day is true.",
					Optional:            true,
					Computed:            true,
				},
				"specific_day_mon": schema.BoolAttribute{
					MarkdownDescription: "Whether Monday is included when specific_day is true.",
					Optional:            true,
					Computed:            true,
				},
				"specific_day_tue": schema.BoolAttribute{
					MarkdownDescription: "Whether Tuesday is included when specific_day is true.",
					Optional:            true,
					Computed:            true,
				},
				"specific_day_wed": schema.BoolAttribute{
					MarkdownDescription: "Whether Wednesday is included when specific_day is true.",
					Optional:            true,
					Computed:            true,
				},
				"specific_day_thu": schema.BoolAttribute{
					MarkdownDescription: "Whether Thursday is included when specific_day is true.",
					Optional:            true,
					Computed:            true,
				},
				"specific_day_fri": schema.BoolAttribute{
					MarkdownDescription: "Whether Friday is included when specific_day is true.",
					Optional:            true,
					Computed:            true,
				},
				"specific_day_sat": schema.BoolAttribute{
					MarkdownDescription: "Whether Saturday is included when specific_day is true.",
					Optional:            true,
					Computed:            true,
				},
				"custom_day": schema.BoolAttribute{
					MarkdownDescription: "Whether this criterion is based on a custom day.",
					Optional:            true,
					Computed:            true,
				},
				"custom_day_choice": schema.StringAttribute{
					MarkdownDescription: "Name of the custom day (see `stonebranch_custom_day`) when custom_day is true.",
					Optional:            true,
				},
				"variable": schema.BoolAttribute{
					MarkdownDescription: "Whether this criterion is based on a variable comparison.",
					Optional:            true,
					Computed:            true,
				},
				"evaluate_at": schema.StringAttribute{
					MarkdownDescription: "When the variable comparison is evaluated. Valid values: 'Trigger Time', 'Vertex Time'.",
					Optional:            true,
					Computed:            true,
				},
				"variable_name": schema.StringAttribute{
					MarkdownDescription: "Name of the variable to compare when variable is true.",
					Optional:            true,
				},
				"variable_value": schema.StringAttribute{
					MarkdownDescription: "Value to compare the variable against when variable is true.",
					Optional:            true,
				},
				"variable_op": schema.StringAttribute{
					MarkdownDescription: "Comparison operator used when variable is true (e.g. '=', '!=').",
					Optional:            true,
					Computed:            true,
				},
				"complex": schema.BoolAttribute{
					MarkdownDescription: "Whether this criterion uses a complex (e.g. 'every Nth day of period') day expression.",
					Optional:            true,
					Computed:            true,
				},
				"complex_adjective": schema.StringAttribute{
					MarkdownDescription: "Complex expression adjective (e.g. 'Every', 'First', 'Last').",
					Optional:            true,
					Computed:            true,
				},
				"complex_noun": schema.StringAttribute{
					MarkdownDescription: "Complex expression noun/unit (e.g. 'Day').",
					Optional:            true,
					Computed:            true,
				},
				"complex_qualifier": schema.StringAttribute{
					MarkdownDescription: "Complex expression qualifier/period (e.g. 'Year').",
					Optional:            true,
					Computed:            true,
				},
				"complex_nth_amount": schema.Int64Attribute{
					MarkdownDescription: "The 'Nth' amount used in the complex expression.",
					Optional:            true,
					Computed:            true,
				},
				"complex_adjustment": schema.StringAttribute{
					MarkdownDescription: "Complex expression adjustment type. Valid values: 'None', 'Before', 'After'.",
					Optional:            true,
					Computed:            true,
				},
				"complex_adjustment_amount": schema.Int64Attribute{
					MarkdownDescription: "Number of adjustment units to apply for the complex adjustment.",
					Optional:            true,
					Computed:            true,
				},
				"complex_adjustment_type": schema.StringAttribute{
					MarkdownDescription: "Unit of the complex adjustment (e.g. 'Day').",
					Optional:            true,
					Computed:            true,
				},
			},
		},
	}
}

// TaskRunCriteriaToAPI converts a Terraform run_criteria list to API models.
func TaskRunCriteriaToAPI(ctx context.Context, list types.List) []TaskRunCriterionAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var items []TaskRunCriterionModel
	list.ElementsAs(ctx, &items, false)

	result := make([]TaskRunCriterionAPIModel, len(items))
	for i, item := range items {
		api := TaskRunCriterionAPIModel{
			Type:        item.Type.ValueString(),
			Task:        item.Task.ValueString(),
			VertexId:    StringValueOrDefault(item.VertexId, "Any"),
			Description: item.Description.ValueString(),

			BusinessDay: item.BusinessDay.ValueBool(),
			Holiday:     item.Holiday.ValueBool(),

			HolidayAdjustment:       item.HolidayAdjustment.ValueString(),
			HolidayAdjustmentAmount: item.HolidayAdjustmentAmount.ValueInt64(),
			HolidayAdjustmentType:   item.HolidayAdjustmentType.ValueString(),

			SpecificDay:    item.SpecificDay.ValueBool(),
			SpecificDaySun: item.SpecificDaySun.ValueBool(),
			SpecificDayMon: item.SpecificDayMon.ValueBool(),
			SpecificDayTue: item.SpecificDayTue.ValueBool(),
			SpecificDayWed: item.SpecificDayWed.ValueBool(),
			SpecificDayThu: item.SpecificDayThu.ValueBool(),
			SpecificDayFri: item.SpecificDayFri.ValueBool(),
			SpecificDaySat: item.SpecificDaySat.ValueBool(),

			CustomDay:       item.CustomDay.ValueBool(),
			CustomDayChoice: item.CustomDayChoice.ValueString(),

			Variable:      item.Variable.ValueBool(),
			EvaluateAt:    item.EvaluateAt.ValueString(),
			VariableName:  item.VariableName.ValueString(),
			VariableValue: item.VariableValue.ValueString(),
			VariableOp:    item.VariableOp.ValueString(),

			Complex:                 item.Complex.ValueBool(),
			ComplexAdjective:        item.ComplexAdjective.ValueString(),
			ComplexNthAmount:        item.ComplexNthAmount.ValueInt64(),
			ComplexAdjustment:       item.ComplexAdjustment.ValueString(),
			ComplexAdjustmentAmount: item.ComplexAdjustmentAmount.ValueInt64(),
			ComplexAdjustmentType:   item.ComplexAdjustmentType.ValueString(),
		}

		// The API stores the complex noun/qualifier as both a singular object
		// and a single-element array; populate both for a faithful round-trip.
		if isSet(item.ComplexNoun) {
			w := complexDayWrapperAPIModel{Value: item.ComplexNoun.ValueString()}
			api.ComplexNoun = &w
			api.ComplexNouns = []complexDayWrapperAPIModel{w}
		}
		if isSet(item.ComplexQualifier) {
			w := complexDayWrapperAPIModel{Value: item.ComplexQualifier.ValueString()}
			api.ComplexQualifier = &w
			api.ComplexQualifiers = []complexDayWrapperAPIModel{w}
		}

		result[i] = api
	}
	return result
}

// TaskRunCriteriaFromAPI converts API run-criteria models to a Terraform list.
func TaskRunCriteriaFromAPI(apiItems []TaskRunCriterionAPIModel) types.List {
	if len(apiItems) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: TaskRunCriterionAttrTypes()})
	}

	values := make([]attr.Value, len(apiItems))
	for i, item := range apiItems {
		complexNoun := ""
		if item.ComplexNoun != nil {
			complexNoun = item.ComplexNoun.Value
		} else if len(item.ComplexNouns) > 0 {
			complexNoun = item.ComplexNouns[0].Value
		}
		complexQualifier := ""
		if item.ComplexQualifier != nil {
			complexQualifier = item.ComplexQualifier.Value
		} else if len(item.ComplexQualifiers) > 0 {
			complexQualifier = item.ComplexQualifiers[0].Value
		}

		values[i], _ = types.ObjectValue(TaskRunCriterionAttrTypes(), map[string]attr.Value{
			"type":        types.StringValue(item.Type),
			"task":        types.StringValue(item.Task),
			"vertex_id":   StringValueOrNull(item.VertexId),
			"description": StringValueOrNull(item.Description),

			"business_day": types.BoolValue(item.BusinessDay),
			"holiday":      types.BoolValue(item.Holiday),

			"holiday_adjustment":        StringValueOrNull(item.HolidayAdjustment),
			"holiday_adjustment_amount": types.Int64Value(item.HolidayAdjustmentAmount),
			"holiday_adjustment_type":   StringValueOrNull(item.HolidayAdjustmentType),

			"specific_day":     types.BoolValue(item.SpecificDay),
			"specific_day_sun": types.BoolValue(item.SpecificDaySun),
			"specific_day_mon": types.BoolValue(item.SpecificDayMon),
			"specific_day_tue": types.BoolValue(item.SpecificDayTue),
			"specific_day_wed": types.BoolValue(item.SpecificDayWed),
			"specific_day_thu": types.BoolValue(item.SpecificDayThu),
			"specific_day_fri": types.BoolValue(item.SpecificDayFri),
			"specific_day_sat": types.BoolValue(item.SpecificDaySat),

			"custom_day":        types.BoolValue(item.CustomDay),
			"custom_day_choice": StringValueOrNull(item.CustomDayChoice),

			"variable":       types.BoolValue(item.Variable),
			"evaluate_at":    StringValueOrNull(item.EvaluateAt),
			"variable_name":  StringValueOrNull(item.VariableName),
			"variable_value": StringValueOrNull(item.VariableValue),
			"variable_op":    StringValueOrNull(item.VariableOp),

			"complex":                   types.BoolValue(item.Complex),
			"complex_adjective":         StringValueOrNull(item.ComplexAdjective),
			"complex_noun":              StringValueOrNull(complexNoun),
			"complex_qualifier":         StringValueOrNull(complexQualifier),
			"complex_nth_amount":        types.Int64Value(item.ComplexNthAmount),
			"complex_adjustment":        StringValueOrNull(item.ComplexAdjustment),
			"complex_adjustment_amount": types.Int64Value(item.ComplexAdjustmentAmount),
			"complex_adjustment_type":   StringValueOrNull(item.ComplexAdjustmentType),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: TaskRunCriterionAttrTypes()}, values)
	return result
}

// TaskRunCriteriaFromAPIOrdered converts API run-criteria models to a Terraform
// list, reordered to match priorOrder's element order (see reorderToMatchPrior).
// run_criteria is not Computed, so Terraform requires the applied result to match
// the plan exactly, including order; UAC does not guarantee it returns entries in
// the order they were submitted.
func TaskRunCriteriaFromAPIOrdered(ctx context.Context, api []TaskRunCriterionAPIModel, priorOrder types.List) types.List {
	if len(api) == 0 || priorOrder.IsNull() || priorOrder.IsUnknown() {
		return TaskRunCriteriaFromAPI(api)
	}
	return TaskRunCriteriaFromAPI(reorderToMatchPrior(api, TaskRunCriteriaToAPI(ctx, priorOrder)))
}

// TaskStepConditionModel describes a single z/OS step condition in Terraform.
type TaskStepConditionModel struct {
	StepName    types.String `tfsdk:"step_name"`
	PstepName   types.String `tfsdk:"pstep_name"`
	ProgramName types.String `tfsdk:"program_name"`
	StepCodes   types.String `tfsdk:"step_codes"`
	StepAction  types.String `tfsdk:"step_action"`
	StepOrder   types.Int64  `tfsdk:"step_order"`
	WorkflowId  types.String `tfsdk:"workflow_id"`
}

// TaskStepConditionAPIModel represents a single z/OS step condition in the API.
type TaskStepConditionAPIModel struct {
	StepName    string `json:"stepName,omitempty"`
	PstepName   string `json:"pstepName,omitempty"`
	ProgramName string `json:"programName,omitempty"`
	StepCodes   string `json:"stepCodes,omitempty"`
	StepAction  string `json:"stepAction,omitempty"`
	StepOrder   int64  `json:"stepOrder,omitempty"`
	WorkflowId  string `json:"workflowId,omitempty"`
}

// TaskStepConditionAttrTypes returns the attribute types for TaskStepConditionModel.
func TaskStepConditionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"step_name":    types.StringType,
		"pstep_name":   types.StringType,
		"program_name": types.StringType,
		"step_codes":   types.StringType,
		"step_action":  types.StringType,
		"step_order":   types.Int64Type,
		"workflow_id":  types.StringType,
	}
}

// TaskStepConditionsSchema returns the schema for the step_conditions attribute (z/OS workflow tasks only).
func TaskStepConditionsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of z/OS step conditions for the workflow. Rarely used outside z/OS environments.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"step_name": schema.StringAttribute{
					MarkdownDescription: "Name of the z/OS job step this condition applies to.",
					Optional:            true,
				},
				"pstep_name": schema.StringAttribute{
					MarkdownDescription: "Name of the procedure step (if any) this condition applies to.",
					Optional:            true,
				},
				"program_name": schema.StringAttribute{
					MarkdownDescription: "Name of the program executed by the step.",
					Optional:            true,
				},
				"step_codes": schema.StringAttribute{
					MarkdownDescription: "Condition/return codes that this step condition checks for.",
					Optional:            true,
				},
				"step_action": schema.StringAttribute{
					MarkdownDescription: "Action taken when the step condition is met.",
					Optional:            true,
				},
				"step_order": schema.Int64Attribute{
					MarkdownDescription: "Order of this step condition relative to others.",
					Optional:            true,
				},
				"workflow_id": schema.StringAttribute{
					MarkdownDescription: "Identifier of the workflow vertex this step condition applies to.",
					Optional:            true,
				},
			},
		},
	}
}

// TaskStepConditionsToAPI converts a Terraform step_conditions list to API models.
func TaskStepConditionsToAPI(ctx context.Context, list types.List) []TaskStepConditionAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var items []TaskStepConditionModel
	list.ElementsAs(ctx, &items, false)

	result := make([]TaskStepConditionAPIModel, len(items))
	for i, item := range items {
		result[i] = TaskStepConditionAPIModel{
			StepName:    item.StepName.ValueString(),
			PstepName:   item.PstepName.ValueString(),
			ProgramName: item.ProgramName.ValueString(),
			StepCodes:   item.StepCodes.ValueString(),
			StepAction:  item.StepAction.ValueString(),
			StepOrder:   item.StepOrder.ValueInt64(),
			WorkflowId:  item.WorkflowId.ValueString(),
		}
	}
	return result
}

// TaskStepConditionsFromAPI converts API step-condition models to a Terraform list.
func TaskStepConditionsFromAPI(apiItems []TaskStepConditionAPIModel) types.List {
	if len(apiItems) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: TaskStepConditionAttrTypes()})
	}

	values := make([]attr.Value, len(apiItems))
	for i, item := range apiItems {
		values[i], _ = types.ObjectValue(TaskStepConditionAttrTypes(), map[string]attr.Value{
			"step_name":    StringValueOrNull(item.StepName),
			"pstep_name":   StringValueOrNull(item.PstepName),
			"program_name": StringValueOrNull(item.ProgramName),
			"step_codes":   StringValueOrNull(item.StepCodes),
			"step_action":  StringValueOrNull(item.StepAction),
			"step_order":   types.Int64Value(item.StepOrder),
			"workflow_id":  StringValueOrNull(item.WorkflowId),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: TaskStepConditionAttrTypes()}, values)
	return result
}

// TaskStepActionModel describes a single z/OS step action (system operation) in Terraform.
// This models the core identification/target fields only; the full API object has
// additional mainframe-specific fields (vertex overrides, trigger date overrides,
// etc.) that are always empty in practice and are omitted here for simplicity.
type TaskStepActionModel struct {
	VertexId           types.String `tfsdk:"vertex_id"`
	Description        types.String `tfsdk:"description"`
	StepName           types.String `tfsdk:"step_name"`
	PstepName          types.String `tfsdk:"pstep_name"`
	ProgramName        types.String `tfsdk:"program_name"`
	StepCodes          types.String `tfsdk:"step_codes"`
	Operation          types.String `tfsdk:"operation"`
	Task               types.String `tfsdk:"task"`
	Agent              types.String `tfsdk:"agent"`
	AgentVar           types.String `tfsdk:"agent_var"`
	AgentCluster       types.String `tfsdk:"agent_cluster"`
	AgentClusterVar    types.String `tfsdk:"agent_cluster_var"`
	Trigger            types.String `tfsdk:"trigger"`
	ExecCriteria       types.String `tfsdk:"exec_criteria"`
	NotificationOption types.String `tfsdk:"notification_option"`
	Limit              types.Int64  `tfsdk:"limit"`
}

// TaskStepActionAPIModel represents a single z/OS step action in the API.
type TaskStepActionAPIModel struct {
	VertexId           string `json:"vertexId,omitempty"`
	Description        string `json:"description,omitempty"`
	StepName           string `json:"stepName,omitempty"`
	PstepName          string `json:"pstepName,omitempty"`
	ProgramName        string `json:"programName,omitempty"`
	StepCodes          string `json:"stepCodes,omitempty"`
	Operation          string `json:"operation,omitempty"`
	Task               string `json:"task,omitempty"`
	Agent              string `json:"agent,omitempty"`
	AgentVar           string `json:"agentVar,omitempty"`
	AgentCluster       string `json:"agentCluster,omitempty"`
	AgentClusterVar    string `json:"agentClusterVar,omitempty"`
	Trigger            string `json:"trigger,omitempty"`
	ExecCriteria       string `json:"execCriteria,omitempty"`
	NotificationOption string `json:"notificationOption,omitempty"`
	Limit              int64  `json:"limit,omitempty"`
}

// TaskStepActionAttrTypes returns the attribute types for TaskStepActionModel.
func TaskStepActionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vertex_id":           types.StringType,
		"description":         types.StringType,
		"step_name":           types.StringType,
		"pstep_name":          types.StringType,
		"program_name":        types.StringType,
		"step_codes":          types.StringType,
		"operation":           types.StringType,
		"task":                types.StringType,
		"agent":               types.StringType,
		"agent_var":           types.StringType,
		"agent_cluster":       types.StringType,
		"agent_cluster_var":   types.StringType,
		"trigger":             types.StringType,
		"exec_criteria":       types.StringType,
		"notification_option": types.StringType,
		"limit":               types.Int64Type,
	}
}

// TaskStepActionsSchema returns the schema for the step_actions attribute (z/OS workflow tasks only).
func TaskStepActionsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of z/OS step actions (system operations) for the workflow. Rarely used outside z/OS environments. Only the core identification/target fields are modeled; additional mainframe-specific override fields are not currently supported.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"vertex_id": schema.StringAttribute{
					MarkdownDescription: "Identifier of the workflow vertex this step action applies to.",
					Optional:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Description of the step action.",
					Optional:            true,
				},
				"step_name": schema.StringAttribute{
					MarkdownDescription: "Name of the z/OS job step this action applies to.",
					Optional:            true,
				},
				"pstep_name": schema.StringAttribute{
					MarkdownDescription: "Name of the procedure step (if any) this action applies to.",
					Optional:            true,
				},
				"program_name": schema.StringAttribute{
					MarkdownDescription: "Name of the program executed by the step.",
					Optional:            true,
				},
				"step_codes": schema.StringAttribute{
					MarkdownDescription: "Condition/return codes that trigger this action.",
					Optional:            true,
				},
				"operation": schema.StringAttribute{
					MarkdownDescription: "Operation to perform (e.g. 'Launch Task').",
					Optional:            true,
				},
				"task": schema.StringAttribute{
					MarkdownDescription: "Name of the task to operate on.",
					Optional:            true,
				},
				"agent": schema.StringAttribute{
					MarkdownDescription: "Name of the agent to run the operation on.",
					Optional:            true,
				},
				"agent_var": schema.StringAttribute{
					MarkdownDescription: "Variable containing the agent name.",
					Optional:            true,
				},
				"agent_cluster": schema.StringAttribute{
					MarkdownDescription: "Name of the agent cluster to run the operation on.",
					Optional:            true,
				},
				"agent_cluster_var": schema.StringAttribute{
					MarkdownDescription: "Variable containing the agent cluster name.",
					Optional:            true,
				},
				"trigger": schema.StringAttribute{
					MarkdownDescription: "Name of the trigger to operate on.",
					Optional:            true,
				},
				"exec_criteria": schema.StringAttribute{
					MarkdownDescription: "Execution criteria for the operation.",
					Optional:            true,
				},
				"notification_option": schema.StringAttribute{
					MarkdownDescription: "Notification option for the operation.",
					Optional:            true,
				},
				"limit": schema.Int64Attribute{
					MarkdownDescription: "Limit value associated with the operation.",
					Optional:            true,
				},
			},
		},
	}
}

// TaskStepActionsToAPI converts a Terraform step_actions list to API models.
func TaskStepActionsToAPI(ctx context.Context, list types.List) []TaskStepActionAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var items []TaskStepActionModel
	list.ElementsAs(ctx, &items, false)

	result := make([]TaskStepActionAPIModel, len(items))
	for i, item := range items {
		result[i] = TaskStepActionAPIModel{
			VertexId:           item.VertexId.ValueString(),
			Description:        item.Description.ValueString(),
			StepName:           item.StepName.ValueString(),
			PstepName:          item.PstepName.ValueString(),
			ProgramName:        item.ProgramName.ValueString(),
			StepCodes:          item.StepCodes.ValueString(),
			Operation:          item.Operation.ValueString(),
			Task:               item.Task.ValueString(),
			Agent:              item.Agent.ValueString(),
			AgentVar:           item.AgentVar.ValueString(),
			AgentCluster:       item.AgentCluster.ValueString(),
			AgentClusterVar:    item.AgentClusterVar.ValueString(),
			Trigger:            item.Trigger.ValueString(),
			ExecCriteria:       item.ExecCriteria.ValueString(),
			NotificationOption: item.NotificationOption.ValueString(),
			Limit:              item.Limit.ValueInt64(),
		}
	}
	return result
}

// TaskStepActionsFromAPI converts API step-action models to a Terraform list.
func TaskStepActionsFromAPI(apiItems []TaskStepActionAPIModel) types.List {
	if len(apiItems) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: TaskStepActionAttrTypes()})
	}

	values := make([]attr.Value, len(apiItems))
	for i, item := range apiItems {
		values[i], _ = types.ObjectValue(TaskStepActionAttrTypes(), map[string]attr.Value{
			"vertex_id":           StringValueOrNull(item.VertexId),
			"description":         StringValueOrNull(item.Description),
			"step_name":           StringValueOrNull(item.StepName),
			"pstep_name":          StringValueOrNull(item.PstepName),
			"program_name":        StringValueOrNull(item.ProgramName),
			"step_codes":          StringValueOrNull(item.StepCodes),
			"operation":           StringValueOrNull(item.Operation),
			"task":                StringValueOrNull(item.Task),
			"agent":               StringValueOrNull(item.Agent),
			"agent_var":           StringValueOrNull(item.AgentVar),
			"agent_cluster":       StringValueOrNull(item.AgentCluster),
			"agent_cluster_var":   StringValueOrNull(item.AgentClusterVar),
			"trigger":             StringValueOrNull(item.Trigger),
			"exec_criteria":       StringValueOrNull(item.ExecCriteria),
			"notification_option": StringValueOrNull(item.NotificationOption),
			"limit":               types.Int64Value(item.Limit),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: TaskStepActionAttrTypes()}, values)
	return result
}
