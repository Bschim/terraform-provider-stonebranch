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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &TaskStoredProcedureResource{}
	_ resource.ResourceWithImportState = &TaskStoredProcedureResource{}
)

func NewTaskStoredProcedureResource() resource.Resource {
	return &TaskStoredProcedureResource{}
}

// TaskStoredProcedureResource defines the resource implementation.
type TaskStoredProcedureResource struct {
	client *client.Client
}

// TaskStoredProcedureResourceModel describes the resource data model.
type TaskStoredProcedureResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Stored procedure
	StoredProcName types.String `tfsdk:"stored_proc_name"`

	// Database connection
	DatabaseConnection types.String `tfsdk:"database_connection"`
	ConnectionVar      types.String `tfsdk:"connection_var"`

	// Result processing
	MaxRows           types.Int64  `tfsdk:"max_rows"`
	AutoCleanup       types.Bool   `tfsdk:"auto_cleanup"`
	ResultProcessing  types.String `tfsdk:"result_processing"`
	ColumnName        types.String `tfsdk:"column_name"`
	ResultOp          types.String `tfsdk:"result_op"`
	ResultValue       types.String `tfsdk:"result_value"`
	ParameterPosition types.Int64  `tfsdk:"parameter_position"`

	// Exit codes
	ExitCodes types.String `tfsdk:"exit_codes"`

	// Parameters
	Parameters types.List `tfsdk:"parameters"`

	// Retry configuration
	RetryMaximum         types.Int64 `tfsdk:"retry_maximum"`
	RetryIndefinitely    types.Bool  `tfsdk:"retry_indefinitely"`
	RetryInterval        types.Int64 `tfsdk:"retry_interval"`
	RetrySuppressFailure types.Bool  `tfsdk:"retry_suppress_failure"`

	// Variables
	Variables types.List `tfsdk:"variables"`

	// Resource management
	HoldResources    types.Bool `tfsdk:"hold_resources"`
	ExclusiveTasks   types.List `tfsdk:"exclusive_tasks"`
	VirtualResources types.List `tfsdk:"virtual_resources"`

	// Actions
	Actions types.Object `tfsdk:"actions"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// StoredProcParamModel describes a stored procedure parameter.
type StoredProcParamModel struct {
	Description   types.String `tfsdk:"description"`
	ParamMode     types.String `tfsdk:"param_mode"`
	ParamType     types.String `tfsdk:"param_type"`
	ParamVar      types.String `tfsdk:"param_var"`
	InputValue    types.String `tfsdk:"input_value"`
	OutputValue   types.String `tfsdk:"output_value"`
	IsNull        types.Bool   `tfsdk:"is_null"`
	VariableScope types.String `tfsdk:"variable_scope"`
	Position      types.Int64  `tfsdk:"position"`
}

// TaskStoredProcedureAPIModel represents the API request/response structure.
type TaskStoredProcedureAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	StoredProcName string `json:"storedProcName,omitempty"`

	Connection    string `json:"connection,omitempty"`
	ConnectionVar string `json:"connectionVar,omitempty"`

	MaxRows           int64  `json:"maxRows,omitempty"`
	AutoCleanup       bool   `json:"autoCleanup,omitempty"`
	ResultProcessing  string `json:"resultProcessing,omitempty"`
	ColumnName        string `json:"columnName,omitempty"`
	ResultOp          string `json:"resultOp,omitempty"`
	ResultValue       string `json:"resultValue,omitempty"`
	ParameterPosition int64  `json:"parameterPosition,omitempty"`

	ExitCodes string `json:"exitCodes,omitempty"`

	StoredProcParams []StoredProcParamAPIModel `json:"storedProcParams,omitempty"`

	RetryMaximum         int64 `json:"retryMaximum,omitempty"`
	RetryIndefinitely    bool  `json:"retryIndefinitely,omitempty"`
	RetryInterval        int64 `json:"retryInterval,omitempty"`
	RetrySuppressFailure bool  `json:"retrySuppressFailure,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources    bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks   []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	VirtualResources []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

// StoredProcParamAPIModel represents a stored procedure parameter in the API.
type StoredProcParamAPIModel struct {
	SysId         string `json:"sysId,omitempty"`
	Description   string `json:"description,omitempty"`
	ParamMode     string `json:"paramMode,omitempty"`
	ParamType     string `json:"paramType,omitempty"`
	ParamVar      string `json:"paramVar,omitempty"`
	Ivalue        string `json:"ivalue,omitempty"`
	Ovalue        string `json:"ovalue,omitempty"`
	IsNull        bool   `json:"isNull,omitempty"`
	VariableScope string `json:"variableScope,omitempty"`
	Pos           int64  `json:"pos,omitempty"`
}

func (r *TaskStoredProcedureResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_stored_procedure"
}

func (r *TaskStoredProcedureResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Stored Procedure Task. Stored procedure tasks execute database stored procedures with support for input and output parameters.",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the task (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the task.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the task (for optimistic locking).",
				Computed:            true,
			},

			// Basic info
			"summary": schema.StringAttribute{
				MarkdownDescription: "Summary/description of the task.",
				Optional:            true,
			},

			// Stored procedure
			"stored_proc_name": schema.StringAttribute{
				MarkdownDescription: "Name of the stored procedure to execute.",
				Required:            true,
			},

			// Database connection
			"database_connection": schema.StringAttribute{
				MarkdownDescription: "Name of the database connection to use.",
				Optional:            true,
			},
			"connection_var": schema.StringAttribute{
				MarkdownDescription: "Name of a variable containing the database connection name.",
				Optional:            true,
			},

			// Result processing
			"max_rows": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of rows to return.",
				Optional:            true,
				Computed:            true,
			},
			"auto_cleanup": schema.BoolAttribute{
				MarkdownDescription: "Enable automatic cleanup of temporary resources.",
				Optional:            true,
				Computed:            true,
			},
			"result_processing": schema.StringAttribute{
				MarkdownDescription: "How to process the results. Valid values: 'None', 'Row Count', 'Column Value'.",
				Optional:            true,
				Computed:            true,
			},
			"column_name": schema.StringAttribute{
				MarkdownDescription: "Column name for result processing (when result_processing is 'Column Value').",
				Optional:            true,
			},
			"result_op": schema.StringAttribute{
				MarkdownDescription: "Comparison operator for result processing. Valid values: '=', '!=', '<', '>', '<=', '>='.",
				Optional:            true,
				Computed:            true,
			},
			"result_value": schema.StringAttribute{
				MarkdownDescription: "Value to compare against for result processing.",
				Optional:            true,
			},
			"parameter_position": schema.Int64Attribute{
				MarkdownDescription: "Parameter position for output parameter result processing.",
				Optional:            true,
				Computed:            true,
			},

			// Exit codes
			"exit_codes": schema.StringAttribute{
				MarkdownDescription: "Exit codes that indicate successful completion.",
				Optional:            true,
				Computed:            true,
			},

			// Parameters
			"parameters": schema.ListNestedAttribute{
				MarkdownDescription: "List of stored procedure parameters.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the parameter.",
							Optional:            true,
						},
						"param_mode": schema.StringAttribute{
							MarkdownDescription: "Parameter mode. Valid values: 'IN', 'OUT', 'INOUT'.",
							Optional:            true,
						},
						"param_type": schema.StringAttribute{
							MarkdownDescription: "Parameter data type. Valid values: 'VARCHAR', 'INTEGER', 'DECIMAL', 'DATE', 'TIMESTAMP', etc.",
							Optional:            true,
						},
						"param_var": schema.StringAttribute{
							MarkdownDescription: "Name of the variable to use for the parameter value. Required for each parameter.",
							Required:            true,
						},
						"input_value": schema.StringAttribute{
							MarkdownDescription: "Input value for the parameter (for IN and INOUT parameters).",
							Optional:            true,
						},
						"output_value": schema.StringAttribute{
							MarkdownDescription: "Variable name to store the output value (for OUT and INOUT parameters).",
							Optional:            true,
						},
						"is_null": schema.BoolAttribute{
							MarkdownDescription: "Whether the parameter value is NULL.",
							Optional:            true,
							Computed:            true,
						},
						"variable_scope": schema.StringAttribute{
							MarkdownDescription: "Scope of the variable. Valid values: 'Self', 'Parent', 'Top Level'.",
							Optional:            true,
							Computed:            true,
						},
						"position": schema.Int64Attribute{
							MarkdownDescription: "Position of the parameter in the stored procedure call.",
							Optional:            true,
						},
					},
				},
			},

			// Retry configuration
			"retry_maximum": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retries.",
				Optional:            true,
				Computed:            true,
			},
			"retry_indefinitely": schema.BoolAttribute{
				MarkdownDescription: "Whether to retry indefinitely.",
				Optional:            true,
				Computed:            true,
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Interval between retries in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"retry_suppress_failure": schema.BoolAttribute{
				MarkdownDescription: "Whether to suppress failure on retry exhaustion.",
				Optional:            true,
				Computed:            true,
			},

			// Variables
			"variables": TaskVariablesSchema(),

			// Resource management
			"hold_resources": schema.BoolAttribute{
				MarkdownDescription: "Whether to hold the task's virtual resources for the duration of any retries.",
				Optional:            true,
				Computed:            true,
			},
			"exclusive_tasks":   TaskExclusiveTasksSchema(),
			"virtual_resources": TaskVirtualResourcesSchema(),

			// Actions
			"actions": TaskActionsSchema(),

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this task belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *TaskStoredProcedureResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskStoredProcedureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskStoredProcedureResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating stored procedure task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Stored Procedure Task",
			fmt.Sprintf("Could not create stored procedure task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Stored Procedure Task",
			fmt.Sprintf("Could not read stored procedure task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created stored procedure task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskStoredProcedureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskStoredProcedureResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Stored procedure task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Stored Procedure Task",
			fmt.Sprintf("Could not read stored procedure task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskStoredProcedureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskStoredProcedureResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskStoredProcedureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating stored procedure task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Stored Procedure Task",
			fmt.Sprintf("Could not update stored procedure task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Stored Procedure Task",
			fmt.Sprintf("Could not read stored procedure task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskStoredProcedureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskStoredProcedureResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting stored procedure task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Stored Procedure Task",
			fmt.Sprintf("Could not delete stored procedure task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskStoredProcedureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskStoredProcedureResource) readTask(ctx context.Context, data *TaskStoredProcedureResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskStoredProcedureAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskStoredProcedureResource) toAPIModel(ctx context.Context, data *TaskStoredProcedureResourceModel) *TaskStoredProcedureAPIModel {
	model := &TaskStoredProcedureAPIModel{
		SysId:                data.SysId.ValueString(),
		Name:                 data.Name.ValueString(),
		Type:                 "taskStoredProc",
		Summary:              data.Summary.ValueString(),
		StoredProcName:       data.StoredProcName.ValueString(),
		Connection:           data.DatabaseConnection.ValueString(),
		ConnectionVar:        data.ConnectionVar.ValueString(),
		MaxRows:              data.MaxRows.ValueInt64(),
		AutoCleanup:          data.AutoCleanup.ValueBool(),
		ResultProcessing:     data.ResultProcessing.ValueString(),
		ColumnName:           data.ColumnName.ValueString(),
		ResultOp:             data.ResultOp.ValueString(),
		ResultValue:          data.ResultValue.ValueString(),
		ParameterPosition:    data.ParameterPosition.ValueInt64(),
		ExitCodes:            data.ExitCodes.ValueString(),
		RetryMaximum:         data.RetryMaximum.ValueInt64(),
		RetryIndefinitely:    data.RetryIndefinitely.ValueBool(),
		RetryInterval:        data.RetryInterval.ValueInt64(),
		RetrySuppressFailure: data.RetrySuppressFailure.ValueBool(),
	}

	// Handle variables
	model.Variables = TaskVariablesToAPI(ctx, data.Variables)

	// Handle resource management fields
	if !data.HoldResources.IsNull() && !data.HoldResources.IsUnknown() {
		model.HoldResources = data.HoldResources.ValueBool()
	}
	model.ExclusiveTasks = TaskExclusiveTasksToAPI(ctx, data.ExclusiveTasks)
	model.VirtualResources = TaskVirtualResourcesToAPI(ctx, data.VirtualResources)

	// Handle actions
	model.Actions = TaskActionsToAPI(ctx, data.Actions)

	// Handle parameters
	if !data.Parameters.IsNull() && !data.Parameters.IsUnknown() {
		var params []StoredProcParamModel
		data.Parameters.ElementsAs(ctx, &params, false)
		for _, p := range params {
			model.StoredProcParams = append(model.StoredProcParams, StoredProcParamAPIModel{
				Description:   p.Description.ValueString(),
				ParamMode:     p.ParamMode.ValueString(),
				ParamType:     p.ParamType.ValueString(),
				ParamVar:      p.ParamVar.ValueString(),
				Ivalue:        p.InputValue.ValueString(),
				Ovalue:        p.OutputValue.ValueString(),
				IsNull:        p.IsNull.ValueBool(),
				VariableScope: p.VariableScope.ValueString(),
				Pos:           p.Position.ValueInt64(),
			})
		}
	}

	// Handle opswise_groups list
	if !data.OpswiseGroups.IsNull() && !data.OpswiseGroups.IsUnknown() {
		var groups []string
		data.OpswiseGroups.ElementsAs(ctx, &groups, false)
		model.OpswiseGroups = groups
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *TaskStoredProcedureResource) fromAPIModel(ctx context.Context, apiModel *TaskStoredProcedureAPIModel, data *TaskStoredProcedureResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Stored procedure
	data.StoredProcName = StringValueOrNull(apiModel.StoredProcName)

	// Database connection
	data.DatabaseConnection = StringValueOrNull(apiModel.Connection)
	data.ConnectionVar = StringValueOrNull(apiModel.ConnectionVar)

	// Result processing
	data.MaxRows = types.Int64Value(apiModel.MaxRows)
	data.AutoCleanup = types.BoolValue(apiModel.AutoCleanup)
	data.ResultProcessing = StringValueOrNull(apiModel.ResultProcessing)
	data.ColumnName = StringValueOrNull(apiModel.ColumnName)
	data.ResultOp = StringValueOrNull(apiModel.ResultOp)
	data.ResultValue = StringValueOrNull(apiModel.ResultValue)
	data.ParameterPosition = types.Int64Value(apiModel.ParameterPosition)

	// Exit codes
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)

	// Parameters - preserve from existing data since API doesn't return param_var
	// The parameters are defined by the user and should not be overwritten by API response
	// We only update computed fields (is_null, variable_scope) within existing parameters
	if len(apiModel.StoredProcParams) > 0 && !data.Parameters.IsNull() && !data.Parameters.IsUnknown() {
		// Get existing parameters to preserve param_var values
		var existingParams []StoredProcParamModel
		data.Parameters.ElementsAs(ctx, &existingParams, false)

		paramAttrTypes := map[string]attr.Type{
			"description":    types.StringType,
			"param_mode":     types.StringType,
			"param_type":     types.StringType,
			"param_var":      types.StringType,
			"input_value":    types.StringType,
			"output_value":   types.StringType,
			"is_null":        types.BoolType,
			"variable_scope": types.StringType,
			"position":       types.Int64Type,
		}

		paramValues := make([]attr.Value, len(apiModel.StoredProcParams))
		for i, p := range apiModel.StoredProcParams {
			// Use existing param_var if API returns empty
			paramVar := StringValueOrNull(p.ParamVar)
			if (paramVar.IsNull() || paramVar.ValueString() == "") && i < len(existingParams) {
				paramVar = existingParams[i].ParamVar
			}

			paramValues[i], _ = types.ObjectValue(paramAttrTypes, map[string]attr.Value{
				"description":    StringValueOrNull(p.Description),
				"param_mode":     StringValueOrNull(p.ParamMode),
				"param_type":     StringValueOrNull(p.ParamType),
				"param_var":      paramVar,
				"input_value":    StringValueOrNull(p.Ivalue),
				"output_value":   StringValueOrNull(p.Ovalue),
				"is_null":        types.BoolValue(p.IsNull),
				"variable_scope": StringValueOrNull(p.VariableScope),
				"position":       types.Int64Value(p.Pos),
			})
		}
		data.Parameters, _ = types.ListValue(types.ObjectType{AttrTypes: paramAttrTypes}, paramValues)
	} else if len(apiModel.StoredProcParams) > 0 {
		// No existing parameters, build from API response
		paramAttrTypes := map[string]attr.Type{
			"description":    types.StringType,
			"param_mode":     types.StringType,
			"param_type":     types.StringType,
			"param_var":      types.StringType,
			"input_value":    types.StringType,
			"output_value":   types.StringType,
			"is_null":        types.BoolType,
			"variable_scope": types.StringType,
			"position":       types.Int64Type,
		}

		paramValues := make([]attr.Value, len(apiModel.StoredProcParams))
		for i, p := range apiModel.StoredProcParams {
			paramValues[i], _ = types.ObjectValue(paramAttrTypes, map[string]attr.Value{
				"description":    StringValueOrNull(p.Description),
				"param_mode":     StringValueOrNull(p.ParamMode),
				"param_type":     StringValueOrNull(p.ParamType),
				"param_var":      StringValueOrNull(p.ParamVar),
				"input_value":    StringValueOrNull(p.Ivalue),
				"output_value":   StringValueOrNull(p.Ovalue),
				"is_null":        types.BoolValue(p.IsNull),
				"variable_scope": StringValueOrNull(p.VariableScope),
				"position":       types.Int64Value(p.Pos),
			})
		}
		data.Parameters, _ = types.ListValue(types.ObjectType{AttrTypes: paramAttrTypes}, paramValues)
	}

	// Retry configuration
	data.RetryMaximum = types.Int64Value(apiModel.RetryMaximum)
	data.RetryIndefinitely = types.BoolValue(apiModel.RetryIndefinitely)
	data.RetryInterval = types.Int64Value(apiModel.RetryInterval)
	data.RetrySuppressFailure = types.BoolValue(apiModel.RetrySuppressFailure)

	// Handle variables
	data.Variables = TaskVariablesFromAPIOrdered(ctx, apiModel.Variables, data.Variables)

	// Handle resource management fields
	data.HoldResources = types.BoolValue(apiModel.HoldResources)
	data.ExclusiveTasks = TaskExclusiveTasksFromAPI(apiModel.ExclusiveTasks)
	data.VirtualResources = TaskVirtualResourcesFromAPI(apiModel.VirtualResources)

	// Handle actions
	data.Actions = TaskActionsFromAPI(ctx, apiModel.Actions, data.Actions)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
