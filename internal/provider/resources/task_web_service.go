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
	_ resource.Resource                   = &TaskWebServiceResource{}
	_ resource.ResourceWithImportState    = &TaskWebServiceResource{}
	_ resource.ResourceWithValidateConfig = &TaskWebServiceResource{}
)

func NewTaskWebServiceResource() resource.Resource {
	return &TaskWebServiceResource{}
}

// TaskWebServiceResource defines the resource implementation.
type TaskWebServiceResource struct {
	client *client.Client
}

// TaskWebServiceResourceModel describes the resource data model.
type TaskWebServiceResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Protocol and method
	Protocol    types.String `tfsdk:"protocol"`
	HttpMethod  types.String `tfsdk:"http_method"`
	HttpVersion types.String `tfsdk:"http_version"`
	SoapVersion types.String `tfsdk:"soap_version"`

	// URL and parameters
	Url           types.String `tfsdk:"url"`
	UrlParameters types.List   `tfsdk:"url_parameters"`

	// Authentication
	HttpAuth       types.String `tfsdk:"http_auth"`
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

	// Request payload
	HttpPayloadType types.String `tfsdk:"http_payload_type"`
	SoapPayloadType types.String `tfsdk:"soap_payload_type"`
	MimeType        types.String `tfsdk:"mime_type"`
	PayloadSource   types.String `tfsdk:"payload_source"`
	Payload         types.String `tfsdk:"payload"`
	PayloadScript   types.String `tfsdk:"payload_script"`
	FormData        types.List   `tfsdk:"form_data"`
	SoapAction      types.String `tfsdk:"soap_action"`

	// Headers
	HttpHeaders types.List `tfsdk:"http_headers"`

	// Response processing
	SoapResponseOutput      types.String `tfsdk:"soap_response_output"`
	ResponseProcessingType  types.String `tfsdk:"response_processing_type"`
	StatusCodeRange         types.String `tfsdk:"status_code_range"`
	OutputType              types.String `tfsdk:"output_type"`
	OutputPathExpression    types.String `tfsdk:"output_path_expression"`
	OutputConditionOperator types.String `tfsdk:"output_condition_operator"`
	OutputConditionValue    types.String `tfsdk:"output_condition_value"`
	OutputConditionStrategy types.String `tfsdk:"output_condition_strategy"`

	// Timeout and options
	Timeout     types.Int64 `tfsdk:"timeout"`
	AutoCleanup types.Bool  `tfsdk:"auto_cleanup"`
	Insecure    types.Bool  `tfsdk:"insecure"`

	// Exit codes
	ExitCodes types.String `tfsdk:"exit_codes"`

	// Retry configuration
	RetryMaximum         types.Int64 `tfsdk:"retry_maximum"`
	RetryIndefinitely    types.Bool  `tfsdk:"retry_indefinitely"`
	RetryInterval        types.Int64 `tfsdk:"retry_interval"`
	RetrySuppressFailure types.Bool  `tfsdk:"retry_suppress_failure"`

	// Wait/Delay options
	WaitToStart       types.String `tfsdk:"wait_to_start"`
	WaitTime          types.String `tfsdk:"wait_time"`
	WaitDuration      types.String `tfsdk:"wait_duration"`
	WaitAmount        types.String `tfsdk:"wait_amount"`
	WaitDayConstraint types.String `tfsdk:"wait_day_constraint"`
	DelayOnStart      types.String `tfsdk:"delay_on_start"`
	DelayDuration     types.String `tfsdk:"delay_duration"`
	DelayAmount       types.String `tfsdk:"delay_amount"`
	WorkflowOnly      types.String `tfsdk:"workflow_only"`

	// Variables
	Variables types.List `tfsdk:"variables"`

	// Resource management
	HoldResources     types.Bool `tfsdk:"hold_resources"`
	ExclusiveTasks    types.List `tfsdk:"exclusive_tasks"`
	ExclusiveWithSelf types.Bool `tfsdk:"exclusive_with_self"`
	VirtualResources  types.List `tfsdk:"virtual_resources"`

	// Actions
	Actions types.Object `tfsdk:"actions"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// NameValueModel represents a name-value pair.
type NameValueModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// TaskWebServiceAPIModel represents the API request/response structure.
type TaskWebServiceAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	Protocol    string `json:"protocol,omitempty"`
	HttpMethod  string `json:"httpMethod,omitempty"`
	HttpVersion string `json:"httpVersion,omitempty"`
	SoapVersion string `json:"soapVersion,omitempty"`

	Url           string              `json:"url,omitempty"`
	UrlParameters []NameValueAPIModel `json:"urlParameters,omitempty"`

	HttpAuth       string `json:"httpAuth,omitempty"`
	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

	HttpPayloadType string              `json:"httpPayloadType,omitempty"`
	SoapPayloadType string              `json:"soapPayloadType,omitempty"`
	MimeType        string              `json:"mimeType,omitempty"`
	PayloadSource   string              `json:"payloadSource,omitempty"`
	Payload         string              `json:"payload,omitempty"`
	PayloadScript   string              `json:"payloadScript,omitempty"`
	FormData        []NameValueAPIModel `json:"formData,omitempty"`
	SoapAction      string              `json:"soapAction,omitempty"`

	HttpHeaders []NameValueAPIModel `json:"httpHeaders,omitempty"`

	SoapResponseOutput      string `json:"soapResponseOutput,omitempty"`
	ResponseProcessingType  string `json:"responseProcessingType,omitempty"`
	StatusCodeRange         string `json:"statusCodeRange,omitempty"`
	OutputType              string `json:"outputType,omitempty"`
	OutputPathExpression    string `json:"outputPathExpression,omitempty"`
	OutputConditionOperator string `json:"outputConditionOperator,omitempty"`
	OutputConditionValue    string `json:"outputConditionValue,omitempty"`
	OutputConditionStrategy string `json:"outputConditionStrategy,omitempty"`

	Timeout     int64 `json:"timeout,omitempty"`
	AutoCleanup bool  `json:"autoCleanup,omitempty"`
	Insecure    bool  `json:"insecure,omitempty"`

	ExitCodes string `json:"exitCodes,omitempty"`

	RetryMaximum         int64 `json:"retryMaximum,omitempty"`
	RetryIndefinitely    bool  `json:"retryIndefinitely,omitempty"`
	RetryInterval        int64 `json:"retryInterval,omitempty"`
	RetrySuppressFailure bool  `json:"retrySuppressFailure,omitempty"`

	// Wait/Delay options
	WaitToStart       string `json:"twWaitType,omitempty"`
	WaitTime          string `json:"twWaitTime,omitempty"`
	WaitDuration      string `json:"twWaitDuration,omitempty"`
	WaitAmount        string `json:"twWaitAmount,omitempty"`
	WaitDayConstraint string `json:"twWaitDayConstraint,omitempty"`
	DelayOnStart      string `json:"twDelayType,omitempty"`
	DelayDuration     string `json:"twDelayDuration,omitempty"`
	DelayAmount       string `json:"twDelayAmount,omitempty"`
	WorkflowOnly      string `json:"twWorkflowOnly,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources     bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks    []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	ExclusiveWithSelf bool                          `json:"exclusiveWithSelf,omitempty"`
	VirtualResources  []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

// NameValueAPIModel represents a name-value pair in the API.
type NameValueAPIModel struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

func (r *TaskWebServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_web_service"
}

func (r *TaskWebServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	nameValueNestedSchema := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the parameter/header.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value of the parameter/header.",
				Required:            true,
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Web Service Task. Web service tasks call REST or SOAP web services.",

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

			// Protocol and method
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol to use. Valid values: 'REST', 'SOAP'.",
				Optional:            true,
				Computed:            true,
			},
			"http_method": schema.StringAttribute{
				MarkdownDescription: "HTTP method. Valid values: 'GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'.",
				Optional:            true,
				Computed:            true,
			},
			"http_version": schema.StringAttribute{
				MarkdownDescription: "HTTP version. Valid values: 'HTTP/1.0', 'HTTP/1.1'.",
				Optional:            true,
				Computed:            true,
			},
			"soap_version": schema.StringAttribute{
				MarkdownDescription: "SOAP version (for SOAP protocol). Valid values: 'SOAP 1.1', 'SOAP 1.2'.",
				Optional:            true,
				Computed:            true,
			},

			// URL and parameters
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the web service endpoint.",
				Required:            true,
			},
			"url_parameters": schema.ListNestedAttribute{
				MarkdownDescription: "URL query parameters.",
				Optional:            true,
				NestedObject:        nameValueNestedSchema,
			},

			// Authentication
			"http_auth": schema.StringAttribute{
				MarkdownDescription: "HTTP authentication method. Valid values: 'None', 'Basic', 'OAuth'.",
				Optional:            true,
				Computed:            true,
			},
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use for authentication.",
				Optional:            true,
			},
			"credentials_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the credentials name.",
				Optional:            true,
			},

			// Request payload
			"http_payload_type": schema.StringAttribute{
				MarkdownDescription: "HTTP payload type. Valid values: 'None', 'Body', 'Form Data'.",
				Optional:            true,
				Computed:            true,
			},
			"soap_payload_type": schema.StringAttribute{
				MarkdownDescription: "SOAP payload type.",
				Optional:            true,
				Computed:            true,
			},
			"mime_type": schema.StringAttribute{
				MarkdownDescription: "MIME type for the request body. E.g., 'application/json', 'application/xml'.",
				Optional:            true,
				Computed:            true,
			},
			"payload_source": schema.StringAttribute{
				MarkdownDescription: "Source of the payload. Valid values: 'Direct', 'Script'.",
				Optional:            true,
				Computed:            true,
			},
			"payload": schema.StringAttribute{
				MarkdownDescription: "Request body content (when payload_source is 'Direct').",
				Optional:            true,
			},
			"payload_script": schema.StringAttribute{
				MarkdownDescription: "Name of the script to generate the payload (when payload_source is 'Script').",
				Optional:            true,
			},
			"form_data": schema.ListNestedAttribute{
				MarkdownDescription: "Form data parameters (when http_payload_type is 'Form Data').",
				Optional:            true,
				NestedObject:        nameValueNestedSchema,
			},
			"soap_action": schema.StringAttribute{
				MarkdownDescription: "SOAP action header value.",
				Optional:            true,
			},

			// Headers
			"http_headers": schema.ListNestedAttribute{
				MarkdownDescription: "HTTP headers to include in the request.",
				Optional:            true,
				NestedObject:        nameValueNestedSchema,
			},

			// Response processing
			"soap_response_output": schema.StringAttribute{
				MarkdownDescription: "SOAP response output handling.",
				Optional:            true,
				Computed:            true,
			},
			"response_processing_type": schema.StringAttribute{
				MarkdownDescription: "Response processing type. Valid values: 'None', 'Status Code', 'JSON Path', 'XPath'.",
				Optional:            true,
				Computed:            true,
			},
			"status_code_range": schema.StringAttribute{
				MarkdownDescription: "Expected status code range. E.g., '200-299', '200,201,204'.",
				Optional:            true,
				Computed:            true,
			},
			"output_type": schema.StringAttribute{
				MarkdownDescription: "Output type for response processing.",
				Optional:            true,
				Computed:            true,
			},
			"output_path_expression": schema.StringAttribute{
				MarkdownDescription: "JSON Path or XPath expression to extract response data.",
				Optional:            true,
			},
			"output_condition_operator": schema.StringAttribute{
				MarkdownDescription: "Comparison operator for response condition. Valid values: '=', '!=', '<', '>', '<=', '>='.",
				Optional:            true,
				Computed:            true,
			},
			"output_condition_value": schema.StringAttribute{
				MarkdownDescription: "Value to compare against for response condition.",
				Optional:            true,
			},
			"output_condition_strategy": schema.StringAttribute{
				MarkdownDescription: "Strategy for handling condition results.",
				Optional:            true,
				Computed:            true,
			},

			// Timeout and options
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Request timeout in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"auto_cleanup": schema.BoolAttribute{
				MarkdownDescription: "Enable automatic cleanup of temporary resources.",
				Optional:            true,
				Computed:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Allow insecure SSL/TLS connections (skip certificate verification).",
				Optional:            true,
				Computed:            true,
			},

			// Exit codes
			"exit_codes": schema.StringAttribute{
				MarkdownDescription: "Exit codes that indicate successful completion.",
				Optional:            true,
				Computed:            true,
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

			// Wait/Delay options
			"wait_to_start":       TaskWaitToStartSchema(),
			"wait_time":           TaskWaitTimeSchema(),
			"wait_duration":       TaskWaitDurationSchema(),
			"wait_amount":         TaskWaitAmountSchema(),
			"wait_day_constraint": TaskWaitDayConstraintSchema(),
			"delay_on_start":      TaskDelayOnStartSchema(),
			"delay_duration":      TaskDelayDurationSchema(),
			"delay_amount":        TaskDelayAmountSchema(),
			"workflow_only":       TaskWorkflowOnlySchema(),

			// Variables
			"variables": TaskVariablesSchema(),

			// Resource management
			"hold_resources": schema.BoolAttribute{
				MarkdownDescription: "Whether to hold the task's virtual resources for the duration of any retries.",
				Optional:            true,
				Computed:            true,
			},
			"exclusive_tasks":     TaskExclusiveTasksSchema(),
			"exclusive_with_self": TaskExclusiveWithSelfSchema(),
			"virtual_resources":   TaskVirtualResourcesSchema(),

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

func (r *TaskWebServiceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TaskWebServiceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(ValidateTaskWaitDelay(data.WaitToStart, data.WaitAmount, data.DelayOnStart, data.DelayAmount)...)
}

func (r *TaskWebServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskWebServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskWebServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating web service task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Web Service Task",
			fmt.Sprintf("Could not create web service task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Web Service Task",
			fmt.Sprintf("Could not read web service task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created web service task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskWebServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskWebServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Web service task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Web Service Task",
			fmt.Sprintf("Could not read web service task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskWebServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskWebServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskWebServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating web service task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Web Service Task",
			fmt.Sprintf("Could not update web service task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Web Service Task",
			fmt.Sprintf("Could not read web service task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskWebServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskWebServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting web service task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Web Service Task",
			fmt.Sprintf("Could not delete web service task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskWebServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskWebServiceResource) readTask(ctx context.Context, data *TaskWebServiceResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskWebServiceAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskWebServiceResource) toAPIModel(ctx context.Context, data *TaskWebServiceResourceModel) *TaskWebServiceAPIModel {
	model := &TaskWebServiceAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskWebService",
		Summary: data.Summary.ValueString(),

		Protocol:    data.Protocol.ValueString(),
		HttpMethod:  data.HttpMethod.ValueString(),
		HttpVersion: data.HttpVersion.ValueString(),
		SoapVersion: data.SoapVersion.ValueString(),

		Url: data.Url.ValueString(),

		HttpAuth:       data.HttpAuth.ValueString(),
		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),

		HttpPayloadType: data.HttpPayloadType.ValueString(),
		SoapPayloadType: data.SoapPayloadType.ValueString(),
		MimeType:        data.MimeType.ValueString(),
		PayloadSource:   data.PayloadSource.ValueString(),
		Payload:         data.Payload.ValueString(),
		PayloadScript:   data.PayloadScript.ValueString(),
		SoapAction:      data.SoapAction.ValueString(),

		SoapResponseOutput:      data.SoapResponseOutput.ValueString(),
		ResponseProcessingType:  data.ResponseProcessingType.ValueString(),
		StatusCodeRange:         data.StatusCodeRange.ValueString(),
		OutputType:              data.OutputType.ValueString(),
		OutputPathExpression:    data.OutputPathExpression.ValueString(),
		OutputConditionOperator: data.OutputConditionOperator.ValueString(),
		OutputConditionValue:    data.OutputConditionValue.ValueString(),
		OutputConditionStrategy: data.OutputConditionStrategy.ValueString(),

		Timeout:     data.Timeout.ValueInt64(),
		AutoCleanup: data.AutoCleanup.ValueBool(),
		Insecure:    data.Insecure.ValueBool(),

		ExitCodes: data.ExitCodes.ValueString(),

		RetryMaximum:         data.RetryMaximum.ValueInt64(),
		RetryIndefinitely:    data.RetryIndefinitely.ValueBool(),
		RetryInterval:        data.RetryInterval.ValueInt64(),
		RetrySuppressFailure: data.RetrySuppressFailure.ValueBool(),

		// Wait/Delay options
		WaitToStart:       data.WaitToStart.ValueString(),
		WaitTime:          data.WaitTime.ValueString(),
		WaitDuration:      data.WaitDuration.ValueString(),
		WaitAmount:        data.WaitAmount.ValueString(),
		WaitDayConstraint: data.WaitDayConstraint.ValueString(),
		DelayOnStart:      data.DelayOnStart.ValueString(),
		DelayDuration:     data.DelayDuration.ValueString(),
		DelayAmount:       data.DelayAmount.ValueString(),
		WorkflowOnly:      data.WorkflowOnly.ValueString(),
	}

	// Handle variables
	model.Variables = TaskVariablesToAPI(ctx, data.Variables)

	// Handle resource management fields
	if !data.HoldResources.IsNull() && !data.HoldResources.IsUnknown() {
		model.HoldResources = data.HoldResources.ValueBool()
	}
	model.ExclusiveTasks = TaskExclusiveTasksToAPI(ctx, data.ExclusiveTasks)
	model.ExclusiveWithSelf = data.ExclusiveWithSelf.ValueBool()
	model.VirtualResources = TaskVirtualResourcesToAPI(ctx, data.VirtualResources)

	// Handle actions
	model.Actions = TaskActionsToAPI(ctx, data.Actions)

	// Handle URL parameters
	if !data.UrlParameters.IsNull() && !data.UrlParameters.IsUnknown() {
		var params []NameValueModel
		data.UrlParameters.ElementsAs(ctx, &params, false)
		for _, p := range params {
			model.UrlParameters = append(model.UrlParameters, NameValueAPIModel{
				Name:  p.Name.ValueString(),
				Value: p.Value.ValueString(),
			})
		}
	}

	// Handle form data
	if !data.FormData.IsNull() && !data.FormData.IsUnknown() {
		var formData []NameValueModel
		data.FormData.ElementsAs(ctx, &formData, false)
		for _, f := range formData {
			model.FormData = append(model.FormData, NameValueAPIModel{
				Name:  f.Name.ValueString(),
				Value: f.Value.ValueString(),
			})
		}
	}

	// Handle HTTP headers
	if !data.HttpHeaders.IsNull() && !data.HttpHeaders.IsUnknown() {
		var headers []NameValueModel
		data.HttpHeaders.ElementsAs(ctx, &headers, false)
		for _, h := range headers {
			model.HttpHeaders = append(model.HttpHeaders, NameValueAPIModel{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
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
func (r *TaskWebServiceResource) fromAPIModel(ctx context.Context, apiModel *TaskWebServiceAPIModel, data *TaskWebServiceResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Protocol and method
	data.Protocol = StringValueOrNull(apiModel.Protocol)
	data.HttpMethod = StringValueOrNull(apiModel.HttpMethod)
	data.HttpVersion = StringValueOrNull(apiModel.HttpVersion)
	data.SoapVersion = StringValueOrNull(apiModel.SoapVersion)

	// URL
	data.Url = StringValueOrNull(apiModel.Url)

	// Authentication
	data.HttpAuth = StringValueOrNull(apiModel.HttpAuth)
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)

	// Request payload
	data.HttpPayloadType = StringValueOrNull(apiModel.HttpPayloadType)
	data.SoapPayloadType = StringValueOrNull(apiModel.SoapPayloadType)
	data.MimeType = StringValueOrNull(apiModel.MimeType)
	data.PayloadSource = StringValueOrNull(apiModel.PayloadSource)
	data.Payload = StringValueOrNull(apiModel.Payload)
	data.PayloadScript = StringValueOrNull(apiModel.PayloadScript)
	data.SoapAction = StringValueOrNull(apiModel.SoapAction)

	// Response processing
	data.SoapResponseOutput = StringValueOrNull(apiModel.SoapResponseOutput)
	data.ResponseProcessingType = StringValueOrNull(apiModel.ResponseProcessingType)
	data.StatusCodeRange = StringValueOrNull(apiModel.StatusCodeRange)
	data.OutputType = StringValueOrNull(apiModel.OutputType)
	data.OutputPathExpression = StringValueOrNull(apiModel.OutputPathExpression)
	data.OutputConditionOperator = StringValueOrNull(apiModel.OutputConditionOperator)
	data.OutputConditionValue = StringValueOrNull(apiModel.OutputConditionValue)
	data.OutputConditionStrategy = StringValueOrNull(apiModel.OutputConditionStrategy)

	// Timeout and options
	data.Timeout = types.Int64Value(apiModel.Timeout)
	data.AutoCleanup = types.BoolValue(apiModel.AutoCleanup)
	data.Insecure = types.BoolValue(apiModel.Insecure)

	// Exit codes
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)

	// Retry configuration
	data.RetryMaximum = types.Int64Value(apiModel.RetryMaximum)
	data.RetryIndefinitely = types.BoolValue(apiModel.RetryIndefinitely)
	data.RetryInterval = types.Int64Value(apiModel.RetryInterval)
	data.RetrySuppressFailure = types.BoolValue(apiModel.RetrySuppressFailure)

	// Wait/Delay options - always returned by API
	data.WaitToStart = types.StringValue(apiModel.WaitToStart)
	data.WaitTime = types.StringValue(apiModel.WaitTime)
	data.WaitDuration = types.StringValue(apiModel.WaitDuration)
	data.WaitAmount = types.StringValue(apiModel.WaitAmount)
	data.WaitDayConstraint = types.StringValue(apiModel.WaitDayConstraint)
	data.DelayOnStart = types.StringValue(apiModel.DelayOnStart)
	data.DelayDuration = types.StringValue(apiModel.DelayDuration)
	data.DelayAmount = types.StringValue(apiModel.DelayAmount)
	data.WorkflowOnly = types.StringValue(apiModel.WorkflowOnly)

	// Handle variables
	data.Variables = TaskVariablesFromAPIOrdered(ctx, apiModel.Variables, data.Variables)

	// Handle resource management fields
	data.HoldResources = types.BoolValue(apiModel.HoldResources)
	data.ExclusiveTasks = TaskExclusiveTasksFromAPI(apiModel.ExclusiveTasks)
	data.ExclusiveWithSelf = types.BoolValue(apiModel.ExclusiveWithSelf)
	data.VirtualResources = TaskVirtualResourcesFromAPI(apiModel.VirtualResources)

	// Handle actions
	data.Actions = TaskActionsFromAPI(ctx, apiModel.Actions, data.Actions)

	// URL parameters
	if len(apiModel.UrlParameters) > 0 {
		data.UrlParameters = r.nameValueListToTerraform(ctx, apiModel.UrlParameters)
	}

	// Form data
	if len(apiModel.FormData) > 0 {
		data.FormData = r.nameValueListToTerraform(ctx, apiModel.FormData)
	}

	// HTTP headers
	if len(apiModel.HttpHeaders) > 0 {
		data.HttpHeaders = r.nameValueListToTerraform(ctx, apiModel.HttpHeaders)
	}

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}

// nameValueListToTerraform converts API name-value pairs to Terraform list.
func (r *TaskWebServiceResource) nameValueListToTerraform(ctx context.Context, items []NameValueAPIModel) types.List {
	attrTypes := map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}

	values := make([]attr.Value, len(items))
	for i, item := range items {
		values[i], _ = types.ObjectValue(attrTypes, map[string]attr.Value{
			"name":  types.StringValue(item.Name),
			"value": types.StringValue(item.Value),
		})
	}

	list, _ := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, values)
	return list
}
