package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &TaskSQLResource{}
	_ resource.ResourceWithImportState = &TaskSQLResource{}
)

func NewTaskSQLResource() resource.Resource {
	return &TaskSQLResource{}
}

// TaskSQLResource defines the resource implementation.
type TaskSQLResource struct {
	client *client.Client
}

// TaskSQLResourceModel describes the resource data model.
type TaskSQLResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Database connection
	DatabaseConnection types.String `tfsdk:"database_connection"`
	ConnectionVar      types.String `tfsdk:"connection_var"`

	// SQL configuration
	SQLCommand types.String `tfsdk:"sql_command"`
	MaxRows    types.Int64  `tfsdk:"max_rows"`

	// Result processing
	ResultProcessing types.String `tfsdk:"result_processing"`
	ColumnName       types.String `tfsdk:"column_name"`
	ColumnOp         types.String `tfsdk:"column_op"`
	ColumnValue      types.String `tfsdk:"column_value"`

	// Exit codes
	ExitCodes types.String `tfsdk:"exit_codes"`

	// Auto cleanup
	AutoCleanup types.Bool `tfsdk:"auto_cleanup"`

	// Credentials
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

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

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// TaskSQLAPIModel represents the API request/response structure.
type TaskSQLAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	Connection    string `json:"connection,omitempty"`
	ConnectionVar string `json:"connectionVar,omitempty"`

	SQLCommand string `json:"sqlCommand,omitempty"`
	MaxRows    int64  `json:"maxRows,omitempty"`

	ResultProcessing string `json:"resultProcessing,omitempty"`
	ColumnName       string `json:"columnName,omitempty"`
	ColumnOp         string `json:"columnOp,omitempty"`
	ColumnValue      string `json:"columnValue,omitempty"`

	ExitCodes   string `json:"exitCodes,omitempty"`
	AutoCleanup bool   `json:"autoCleanup,omitempty"`

	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

	RetryMaximum         int64 `json:"retryMaximum,omitempty"`
	RetryIndefinitely    bool  `json:"retryIndefinitely,omitempty"`
	RetryInterval        int64 `json:"retryInterval,omitempty"`
	RetrySuppressFailure bool  `json:"retrySuppressFailure,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources    bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks   []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	VirtualResources []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TaskSQLResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_sql"
}

func (r *TaskSQLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch SQL Task. SQL tasks execute SQL queries against database connections.",

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

			// Database connection
			"database_connection": schema.StringAttribute{
				MarkdownDescription: "Name of the database connection to use.",
				Optional:            true,
			},
			"connection_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the database connection name.",
				Optional:            true,
			},

			// SQL configuration
			"sql_command": schema.StringAttribute{
				MarkdownDescription: "The SQL command to execute.",
				Optional:            true,
			},
			"max_rows": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of rows to return. Use 0 for unlimited.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			// Result processing
			"result_processing": schema.StringAttribute{
				MarkdownDescription: "How to process the query results. Valid values: 'None', 'First Column', 'Specific Column', 'Count'.",
				Optional:            true,
				Computed:            true,
			},
			"column_name": schema.StringAttribute{
				MarkdownDescription: "Column name for result processing (when result_processing is 'Specific Column').",
				Optional:            true,
			},
			"column_op": schema.StringAttribute{
				MarkdownDescription: "Comparison operator for result processing. Valid values: '=', '!=', '>', '<', '>=', '<='.",
				Optional:            true,
				Computed:            true,
			},
			"column_value": schema.StringAttribute{
				MarkdownDescription: "Value to compare against for result processing.",
				Optional:            true,
			},

			// Exit codes
			"exit_codes": schema.StringAttribute{
				MarkdownDescription: "Exit codes that indicate success (comma-separated).",
				Optional:            true,
				Computed:            true,
			},

			// Auto cleanup
			"auto_cleanup": schema.BoolAttribute{
				MarkdownDescription: "Whether to automatically clean up resources after execution.",
				Optional:            true,
				Computed:            true,
			},

			// Credentials
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credential to use (overrides database connection credentials).",
				Optional:            true,
			},
			"credentials_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the credential name.",
				Optional:            true,
			},

			// Retry configuration
			"retry_maximum": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retry attempts.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"retry_suppress_failure": schema.BoolAttribute{
				MarkdownDescription: "Whether to suppress failure after all retries are exhausted.",
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

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this task belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *TaskSQLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskSQLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskSQLResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating SQL task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating SQL Task",
			fmt.Sprintf("Could not create SQL task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created SQL Task",
			fmt.Sprintf("Could not read SQL task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created SQL task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskSQLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskSQLResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "SQL task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading SQL Task",
			fmt.Sprintf("Could not read SQL task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskSQLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskSQLResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskSQLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating SQL task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating SQL Task",
			fmt.Sprintf("Could not update SQL task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated SQL Task",
			fmt.Sprintf("Could not read SQL task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskSQLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskSQLResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting SQL task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting SQL Task",
			fmt.Sprintf("Could not delete SQL task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskSQLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskSQLResource) readTask(ctx context.Context, data *TaskSQLResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskSQLAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse SQL task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskSQLResource) toAPIModel(ctx context.Context, data *TaskSQLResourceModel) *TaskSQLAPIModel {
	model := &TaskSQLAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskSql",
		Summary: data.Summary.ValueString(),

		Connection:    data.DatabaseConnection.ValueString(),
		ConnectionVar: data.ConnectionVar.ValueString(),

		SQLCommand: data.SQLCommand.ValueString(),

		ResultProcessing: data.ResultProcessing.ValueString(),
		ColumnName:       data.ColumnName.ValueString(),
		ColumnOp:         data.ColumnOp.ValueString(),
		ColumnValue:      data.ColumnValue.ValueString(),

		ExitCodes: data.ExitCodes.ValueString(),

		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),
	}

	// Handle max_rows - only set if specified
	if !data.MaxRows.IsNull() && !data.MaxRows.IsUnknown() {
		model.MaxRows = data.MaxRows.ValueInt64()
	}

	// Handle auto_cleanup
	if !data.AutoCleanup.IsNull() && !data.AutoCleanup.IsUnknown() {
		model.AutoCleanup = data.AutoCleanup.ValueBool()
	}

	// Handle retry configuration
	if !data.RetryMaximum.IsNull() && !data.RetryMaximum.IsUnknown() {
		model.RetryMaximum = data.RetryMaximum.ValueInt64()
	}
	if !data.RetryIndefinitely.IsNull() && !data.RetryIndefinitely.IsUnknown() {
		model.RetryIndefinitely = data.RetryIndefinitely.ValueBool()
	}
	if !data.RetryInterval.IsNull() && !data.RetryInterval.IsUnknown() {
		model.RetryInterval = data.RetryInterval.ValueInt64()
	}
	if !data.RetrySuppressFailure.IsNull() && !data.RetrySuppressFailure.IsUnknown() {
		model.RetrySuppressFailure = data.RetrySuppressFailure.ValueBool()
	}

	// Handle variables
	model.Variables = TaskVariablesToAPI(ctx, data.Variables)

	// Handle resource management fields
	if !data.HoldResources.IsNull() && !data.HoldResources.IsUnknown() {
		model.HoldResources = data.HoldResources.ValueBool()
	}
	model.ExclusiveTasks = TaskExclusiveTasksToAPI(ctx, data.ExclusiveTasks)
	model.VirtualResources = TaskVirtualResourcesToAPI(ctx, data.VirtualResources)

	// Handle opswise_groups list
	if !data.OpswiseGroups.IsNull() && !data.OpswiseGroups.IsUnknown() {
		var groups []string
		data.OpswiseGroups.ElementsAs(ctx, &groups, false)
		model.OpswiseGroups = groups
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *TaskSQLResource) fromAPIModel(ctx context.Context, apiModel *TaskSQLAPIModel, data *TaskSQLResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Database connection
	data.DatabaseConnection = StringValueOrNull(apiModel.Connection)
	data.ConnectionVar = StringValueOrNull(apiModel.ConnectionVar)

	// SQL configuration
	data.SQLCommand = StringValueOrNull(apiModel.SQLCommand)
	data.MaxRows = types.Int64Value(apiModel.MaxRows)

	// Result processing
	data.ResultProcessing = StringValueOrNull(apiModel.ResultProcessing)
	data.ColumnName = StringValueOrNull(apiModel.ColumnName)
	data.ColumnOp = StringValueOrNull(apiModel.ColumnOp)
	data.ColumnValue = StringValueOrNull(apiModel.ColumnValue)

	// Exit codes
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)

	// Auto cleanup
	data.AutoCleanup = types.BoolValue(apiModel.AutoCleanup)

	// Credentials
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)

	// Retry configuration
	data.RetryMaximum = types.Int64Value(apiModel.RetryMaximum)
	data.RetryIndefinitely = types.BoolValue(apiModel.RetryIndefinitely)
	data.RetryInterval = types.Int64Value(apiModel.RetryInterval)
	data.RetrySuppressFailure = types.BoolValue(apiModel.RetrySuppressFailure)

	// Handle variables
	data.Variables = TaskVariablesFromAPI(ctx, apiModel.Variables)

	// Handle resource management fields
	data.HoldResources = types.BoolValue(apiModel.HoldResources)
	data.ExclusiveTasks = TaskExclusiveTasksFromAPI(apiModel.ExclusiveTasks)
	data.VirtualResources = TaskVirtualResourcesFromAPI(apiModel.VirtualResources)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
