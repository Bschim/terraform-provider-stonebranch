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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &TaskEmailResource{}
	_ resource.ResourceWithImportState    = &TaskEmailResource{}
	_ resource.ResourceWithValidateConfig = &TaskEmailResource{}
)

func NewTaskEmailResource() resource.Resource {
	return &TaskEmailResource{}
}

// TaskEmailResource defines the resource implementation.
type TaskEmailResource struct {
	client *client.Client
}

// TaskEmailResourceModel describes the resource data model.
type TaskEmailResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Email connection
	EmailConnection    types.String `tfsdk:"email_connection"`
	EmailConnectionVar types.String `tfsdk:"email_connection_var"`

	// Email template
	Template    types.String `tfsdk:"template"`
	TemplateVar types.String `tfsdk:"template_var"`

	// Email content
	ReplyTo       types.String `tfsdk:"reply_to"`
	ToRecipients  types.String `tfsdk:"to_recipients"`
	CCRecipients  types.String `tfsdk:"cc_recipients"`
	BCCRecipients types.String `tfsdk:"bcc_recipients"`
	Subject       types.String `tfsdk:"subject"`
	Body          types.String `tfsdk:"body"`

	// Attachments
	AttachLocalFile      types.Bool   `tfsdk:"attach_local_file"`
	LocalAttachmentsPath types.String `tfsdk:"local_attachments_path"`
	LocalAttachment      types.String `tfsdk:"local_attachment"`

	// Report
	ReportVar        types.String `tfsdk:"report_var"`
	ListReportFormat types.String `tfsdk:"list_report_format"`

	// Exit codes
	ExitCodes types.String `tfsdk:"exit_codes"`

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

// TaskEmailAPIModel represents the API request/response structure.
type TaskEmailAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	// Email connection
	Connection    string `json:"connection,omitempty"`
	ConnectionVar string `json:"connectionVar,omitempty"`

	// Email template
	Template    string `json:"template,omitempty"`
	TemplateVar string `json:"templateVar,omitempty"`

	// Email content
	ReplyTo       string `json:"replyTo,omitempty"`
	ToRecipients  string `json:"toRecipients,omitempty"`
	CCRecipients  string `json:"ccRecipients,omitempty"`
	BCCRecipients string `json:"bccRecipients,omitempty"`
	Subject       string `json:"subject,omitempty"`
	Body          string `json:"body,omitempty"`

	// Attachments
	AttachLocalFile      bool   `json:"attachLocalFile,omitempty"`
	LocalAttachmentsPath string `json:"localAttachmentsPath,omitempty"`
	LocalAttachment      string `json:"localAttachment,omitempty"`

	// Report
	ReportVar        string `json:"reportVar,omitempty"`
	ListReportFormat string `json:"listReportFormat,omitempty"`

	// Exit codes
	ExitCodes string `json:"exitCodes,omitempty"`

	// Retry configuration
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

func (r *TaskEmailResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_email"
}

func (r *TaskEmailResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Email Task. Email tasks send emails via configured email connections.",

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

			// Email connection
			"email_connection": schema.StringAttribute{
				MarkdownDescription: "Name of the email connection to use.",
				Optional:            true,
			},
			"email_connection_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the email connection name.",
				Optional:            true,
			},

			// Email template
			"template": schema.StringAttribute{
				MarkdownDescription: "Name of the email template to use.",
				Optional:            true,
			},
			"template_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the email template name.",
				Optional:            true,
			},

			// Email content
			"reply_to": schema.StringAttribute{
				MarkdownDescription: "Reply-to email address.",
				Optional:            true,
			},
			"to_recipients": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of To recipients.",
				Optional:            true,
			},
			"cc_recipients": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of CC recipients.",
				Optional:            true,
			},
			"bcc_recipients": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of BCC recipients.",
				Optional:            true,
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "Email subject line.",
				Optional:            true,
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "Email body content.",
				Optional:            true,
			},

			// Attachments
			"attach_local_file": schema.BoolAttribute{
				MarkdownDescription: "Whether to attach a local file to the email.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"local_attachments_path": schema.StringAttribute{
				MarkdownDescription: "Path to the directory containing local attachments.",
				Optional:            true,
			},
			"local_attachment": schema.StringAttribute{
				MarkdownDescription: "Name of the local file to attach.",
				Optional:            true,
			},

			// Report
			"report_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the report name to attach.",
				Optional:            true,
			},
			"list_report_format": schema.StringAttribute{
				MarkdownDescription: "Format for list reports (e.g., 'CSV', 'PDF').",
				Optional:            true,
				Computed:            true,
			},

			// Exit codes
			"exit_codes": schema.StringAttribute{
				MarkdownDescription: "Exit codes that indicate success (comma-separated).",
				Optional:            true,
				Computed:            true,
			},

			// Retry configuration
			"retry_maximum": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retry attempts.",
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

func (r *TaskEmailResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TaskEmailResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isSet(data.EmailConnection) || isSet(data.EmailConnectionVar) || isSet(data.Template) || isSet(data.TemplateVar) {
		return
	}
	// If any of the four are unknown (e.g. computed from another resource),
	// defer this check to the server rather than failing the plan early.
	if data.EmailConnection.IsUnknown() || data.EmailConnectionVar.IsUnknown() ||
		data.Template.IsUnknown() || data.TemplateVar.IsUnknown() {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("email_connection"),
		"Missing Required Field",
		`At least one of "email_connection", "email_connection_var", "template", or "template_var" must be set.`,
	)
}

func (r *TaskEmailResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskEmailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskEmailResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating email task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Email Task",
			fmt.Sprintf("Could not create email task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Email Task",
			fmt.Sprintf("Could not read email task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created email task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskEmailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskEmailResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Email task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Email Task",
			fmt.Sprintf("Could not read email task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskEmailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskEmailResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskEmailResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating email task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Email Task",
			fmt.Sprintf("Could not update email task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Email Task",
			fmt.Sprintf("Could not read email task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskEmailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskEmailResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting email task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Email Task",
			fmt.Sprintf("Could not delete email task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskEmailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskEmailResource) readTask(ctx context.Context, data *TaskEmailResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskEmailAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse email task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskEmailResource) toAPIModel(ctx context.Context, data *TaskEmailResourceModel) *TaskEmailAPIModel {
	model := &TaskEmailAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskEmail",
		Summary: data.Summary.ValueString(),

		// Email connection
		Connection:    data.EmailConnection.ValueString(),
		ConnectionVar: data.EmailConnectionVar.ValueString(),

		// Email template
		Template:    data.Template.ValueString(),
		TemplateVar: data.TemplateVar.ValueString(),

		// Email content
		ReplyTo:       data.ReplyTo.ValueString(),
		ToRecipients:  data.ToRecipients.ValueString(),
		CCRecipients:  data.CCRecipients.ValueString(),
		BCCRecipients: data.BCCRecipients.ValueString(),
		Subject:       data.Subject.ValueString(),
		Body:          data.Body.ValueString(),

		// Attachments
		LocalAttachmentsPath: data.LocalAttachmentsPath.ValueString(),
		LocalAttachment:      data.LocalAttachment.ValueString(),

		// Report
		ReportVar:        data.ReportVar.ValueString(),
		ListReportFormat: data.ListReportFormat.ValueString(),

		// Exit codes
		ExitCodes: data.ExitCodes.ValueString(),
	}

	// Handle attach_local_file
	if !data.AttachLocalFile.IsNull() && !data.AttachLocalFile.IsUnknown() {
		model.AttachLocalFile = data.AttachLocalFile.ValueBool()
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
func (r *TaskEmailResource) fromAPIModel(ctx context.Context, apiModel *TaskEmailAPIModel, data *TaskEmailResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Email connection
	data.EmailConnection = StringValueOrNull(apiModel.Connection)
	data.EmailConnectionVar = StringValueOrNull(apiModel.ConnectionVar)

	// Email template
	data.Template = StringValueOrNull(apiModel.Template)
	data.TemplateVar = StringValueOrNull(apiModel.TemplateVar)

	// Email content
	data.ReplyTo = StringValueOrNull(apiModel.ReplyTo)
	data.ToRecipients = StringValueOrNull(apiModel.ToRecipients)
	data.CCRecipients = StringValueOrNull(apiModel.CCRecipients)
	data.BCCRecipients = StringValueOrNull(apiModel.BCCRecipients)
	data.Subject = StringValueOrNull(apiModel.Subject)
	data.Body = StringValueOrNull(apiModel.Body)

	// Attachments
	data.AttachLocalFile = types.BoolValue(apiModel.AttachLocalFile)
	data.LocalAttachmentsPath = StringValueOrNull(apiModel.LocalAttachmentsPath)
	data.LocalAttachment = StringValueOrNull(apiModel.LocalAttachment)

	// Report
	data.ReportVar = StringValueOrNull(apiModel.ReportVar)
	data.ListReportFormat = StringValueOrNull(apiModel.ListReportFormat)

	// Exit codes
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)

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

	// Handle actions
	data.Actions = TaskActionsFromAPI(ctx, apiModel.Actions)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
