package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

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
	_ resource.Resource                = &TaskWorkflowResource{}
	_ resource.ResourceWithImportState = &TaskWorkflowResource{}
)

func NewTaskWorkflowResource() resource.Resource {
	return &TaskWorkflowResource{}
}

// TaskWorkflowResource defines the resource implementation.
type TaskWorkflowResource struct {
	client *client.Client
}

// TaskWorkflowResourceModel describes the resource data model.
type TaskWorkflowResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Workflow configuration
	SkippedOption      types.String `tfsdk:"skipped_option"`
	LayoutOption       types.String `tfsdk:"layout_option"`
	Calendar           types.String `tfsdk:"calendar"`
	OverrideCalendar   types.Bool   `tfsdk:"override_calendar"`
	CalculateCp        types.Bool   `tfsdk:"calculate_critical_path"`
	InstanceWait       types.String `tfsdk:"instance_wait"`
	InstanceWaitLookup types.String `tfsdk:"instance_wait_lookup"`

	// Computed
	NumberOfTasks types.Int64 `tfsdk:"number_of_tasks"`

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

	// Workflow-specific run criteria and step actions/conditions
	RunCriteria    types.List `tfsdk:"run_criteria"`
	StepActions    types.List `tfsdk:"step_actions"`
	StepConditions types.List `tfsdk:"step_conditions"`

	// Actions
	Actions types.Object `tfsdk:"actions"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// TaskWorkflowAPIModel represents the API request/response structure.
type TaskWorkflowAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	SkippedOption      string `json:"skippedOption,omitempty"`
	LayoutOption       string `json:"layoutOption,omitempty"`
	Calendar           string `json:"calendar,omitempty"`
	OverrideCalendar   bool   `json:"overrideCalendar,omitempty"`
	CalculateCp        bool   `json:"calculateCp,omitempty"`
	InstanceWait       string `json:"instanceWait,omitempty"`
	InstanceWaitLookup string `json:"instanceWaitLookup,omitempty"`
	NumberOfTasks      int64  `json:"numberOfTasks,omitempty"`

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

	RunCriteria    []TaskRunCriterionAPIModel  `json:"runCriteria,omitempty"`
	StepActions    []TaskStepActionAPIModel    `json:"stepActions,omitempty"`
	StepConditions []TaskStepConditionAPIModel `json:"stepConditions,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`

	// WorkflowVertices is a pass-through only: the API validates runCriteria/
	// stepActions/stepConditions task references against the workflowVertices
	// array included in the SAME request (not persisted state), so the
	// current vertices must be fetched and echoed back whenever those fields
	// are set. This provider does not otherwise manage vertices here (see
	// the separate stonebranch_workflow_vertex resource).
	WorkflowVertices json.RawMessage `json:"workflowVertices,omitempty"`
}

func (r *TaskWorkflowResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_workflow"
}

func (r *TaskWorkflowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Workflow Task. Workflow tasks orchestrate the execution of multiple tasks in a defined sequence with dependencies.",

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
				MarkdownDescription: "Unique name of the workflow task.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the task (for optimistic locking).",
				Computed:            true,
			},

			// Basic info
			"summary": schema.StringAttribute{
				MarkdownDescription: "Summary/description of the workflow task.",
				Optional:            true,
			},

			// Workflow configuration
			"skipped_option": schema.StringAttribute{
				MarkdownDescription: "Action to take when a task in the workflow is skipped. Valid values: 'Run Successors On Skip', 'Skip Successors On Skip', 'No Action'.",
				Optional:            true,
				Computed:            true,
			},
			"layout_option": schema.StringAttribute{
				MarkdownDescription: "Layout option for the workflow diagram. Valid values: 'Vertical', 'Horizontal'.",
				Optional:            true,
				Computed:            true,
			},
			"calendar": schema.StringAttribute{
				MarkdownDescription: "Name of the calendar to use for scheduling.",
				Optional:            true,
			},
			"override_calendar": schema.BoolAttribute{
				MarkdownDescription: "Whether to override the calendar settings of child tasks.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"calculate_critical_path": schema.BoolAttribute{
				MarkdownDescription: "Whether to calculate the critical path for the workflow.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_wait": schema.StringAttribute{
				MarkdownDescription: "Instance wait behavior. Valid values: 'None', 'Wait For Any', 'Wait For All'.",
				Optional:            true,
				Computed:            true,
			},
			"instance_wait_lookup": schema.StringAttribute{
				MarkdownDescription: "How to look up instances for waiting. Valid values: 'Oldest Instance', 'Latest Instance'.",
				Optional:            true,
				Computed:            true,
			},

			// Computed
			"number_of_tasks": schema.Int64Attribute{
				MarkdownDescription: "Number of tasks in the workflow (computed by server).",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			// Credentials
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credential to use for the workflow.",
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
				MarkdownDescription: "Whether to hold the task's virtual resources for the duration of any retries. Note: the Stonebranch API does not appear to persist this setting for workflow tasks; it is included here for schema consistency with other task types.",
				Optional:            true,
				Computed:            true,
			},
			"exclusive_tasks":   TaskExclusiveTasksSchema(),
			"virtual_resources": TaskVirtualResourcesSchema(),

			// Workflow-specific run criteria and step actions/conditions
			"run_criteria":    TaskRunCriteriaSchema(),
			"step_actions":    TaskStepActionsSchema(),
			"step_conditions": TaskStepConditionsSchema(),

			// Actions
			"actions": TaskActionsSchema(),

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this workflow belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *TaskWorkflowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskWorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskWorkflowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating workflow task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Workflow Task",
			fmt.Sprintf("Could not create workflow task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Workflow Task",
			fmt.Sprintf("Could not read workflow task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created workflow task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskWorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskWorkflowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Workflow task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Workflow Task",
			fmt.Sprintf("Could not read workflow task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskWorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskWorkflowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskWorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating workflow task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// The API validates runCriteria/stepActions/stepConditions task references
	// against the workflowVertices array included in the same request rather
	// than against persisted state, so fetch and echo back the task's current
	// vertices whenever any of those fields are set.
	if taskWorkflowHasCriteriaFields(&data) {
		vertices, err := r.fetchWorkflowVertices(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Fetching Workflow Vertices",
				fmt.Sprintf("Could not fetch current vertices for workflow task %s: %s", data.Name.ValueString(), err),
			)
			return
		}
		apiModel.WorkflowVertices = vertices
	}

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Workflow Task",
			fmt.Sprintf("Could not update workflow task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Workflow Task",
			fmt.Sprintf("Could not read workflow task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskWorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskWorkflowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting workflow task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Workflow Task",
			fmt.Sprintf("Could not delete workflow task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskWorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskWorkflowResource) readTask(ctx context.Context, data *TaskWorkflowResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskWorkflowAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse workflow task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// taskWorkflowHasCriteriaFields reports whether any of the run_criteria,
// step_actions, or step_conditions fields are set with at least one entry.
func taskWorkflowHasCriteriaFields(data *TaskWorkflowResourceModel) bool {
	hasElements := func(list types.List) bool {
		return !list.IsNull() && !list.IsUnknown() && len(list.Elements()) > 0
	}
	return hasElements(data.RunCriteria) || hasElements(data.StepActions) || hasElements(data.StepConditions)
}

// fetchWorkflowVertices fetches the raw workflowVertices JSON for the named
// workflow task, for pass-through into an update payload that also sets
// run_criteria/step_actions/step_conditions.
func (r *TaskWorkflowResource) fetchWorkflowVertices(ctx context.Context, name string) (json.RawMessage, error) {
	query := url.Values{}
	query.Set("taskname", name)

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		WorkflowVertices json.RawMessage `json:"workflowVertices,omitempty"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse workflow vertices: %w", err)
	}
	return wrapper.WorkflowVertices, nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskWorkflowResource) toAPIModel(ctx context.Context, data *TaskWorkflowResourceModel) *TaskWorkflowAPIModel {
	model := &TaskWorkflowAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskWorkflow",
		Summary: data.Summary.ValueString(),

		SkippedOption:      data.SkippedOption.ValueString(),
		LayoutOption:       data.LayoutOption.ValueString(),
		Calendar:           data.Calendar.ValueString(),
		InstanceWait:       data.InstanceWait.ValueString(),
		InstanceWaitLookup: data.InstanceWaitLookup.ValueString(),

		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),
	}

	// Handle boolean fields
	if !data.OverrideCalendar.IsNull() && !data.OverrideCalendar.IsUnknown() {
		model.OverrideCalendar = data.OverrideCalendar.ValueBool()
	}
	if !data.CalculateCp.IsNull() && !data.CalculateCp.IsUnknown() {
		model.CalculateCp = data.CalculateCp.ValueBool()
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

	// Handle workflow-specific run criteria and step actions/conditions
	model.RunCriteria = TaskRunCriteriaToAPI(ctx, data.RunCriteria)
	model.StepActions = TaskStepActionsToAPI(ctx, data.StepActions)
	model.StepConditions = TaskStepConditionsToAPI(ctx, data.StepConditions)

	// Handle actions
	model.Actions = TaskActionsToAPI(ctx, data.Actions)

	// Handle opswise_groups list
	if !data.OpswiseGroups.IsNull() && !data.OpswiseGroups.IsUnknown() {
		var groups []string
		data.OpswiseGroups.ElementsAs(ctx, &groups, false)
		model.OpswiseGroups = groups
	}

	return model
}

// fromAPIModel converts an API model to the Terraform model.
func (r *TaskWorkflowResource) fromAPIModel(ctx context.Context, apiModel *TaskWorkflowAPIModel, data *TaskWorkflowResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Workflow configuration
	data.SkippedOption = StringValueOrNull(apiModel.SkippedOption)
	data.LayoutOption = StringValueOrNull(apiModel.LayoutOption)
	data.Calendar = StringValueOrNull(apiModel.Calendar)
	data.OverrideCalendar = types.BoolValue(apiModel.OverrideCalendar)
	data.CalculateCp = types.BoolValue(apiModel.CalculateCp)
	data.InstanceWait = StringValueOrNull(apiModel.InstanceWait)
	data.InstanceWaitLookup = StringValueOrNull(apiModel.InstanceWaitLookup)

	// Computed
	data.NumberOfTasks = types.Int64Value(apiModel.NumberOfTasks)

	// Credentials
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)

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

	// Handle workflow-specific run criteria and step actions/conditions
	data.RunCriteria = TaskRunCriteriaFromAPI(apiModel.RunCriteria)
	data.StepActions = TaskStepActionsFromAPI(apiModel.StepActions)
	data.StepConditions = TaskStepConditionsFromAPI(apiModel.StepConditions)

	// Handle actions
	data.Actions = TaskActionsFromAPI(ctx, apiModel.Actions, data.Actions)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
