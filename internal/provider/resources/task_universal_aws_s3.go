package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

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
	_ resource.Resource                = &TaskUniversalAwsS3Resource{}
	_ resource.ResourceWithImportState = &TaskUniversalAwsS3Resource{}
)

func NewTaskUniversalAwsS3Resource() resource.Resource {
	return &TaskUniversalAwsS3Resource{}
}

// TaskUniversalAwsS3Resource defines the resource implementation.
type TaskUniversalAwsS3Resource struct {
	client *client.Client
}

// TaskUniversalAwsS3ResourceModel describes the resource data model.
type TaskUniversalAwsS3ResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Basic info
	Summary types.String `tfsdk:"summary"`

	// Agent configuration
	Agent           types.String `tfsdk:"agent"`
	AgentCluster    types.String `tfsdk:"agent_cluster"`
	AgentVar        types.String `tfsdk:"agent_var"`
	AgentClusterVar types.String `tfsdk:"agent_cluster_var"`

	// Credentials (for task execution, not AWS)
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

	// Exit code handling
	ExitCodes          types.String `tfsdk:"exit_codes"`
	ExitCodeProcessing types.String `tfsdk:"exit_code_processing"`

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

	// ===========================================
	// CS AWS S3 Template-specific fields
	// ===========================================

	// Action (Choice Field 3)
	// Values: list-buckets, list-objects, upload-file, download-file, delete-objects,
	//         delete-bucket, create-bucket, copy-object-to-bucket, monitor-object
	Action types.String `tfsdk:"action"`

	// Bucket/Object settings
	Bucket       types.String `tfsdk:"bucket"`        // Text Field 3 - object_store
	TargetBucket types.String `tfsdk:"target_bucket"` // Text Field 8 - target_object_store
	S3Key        types.String `tfsdk:"s3_key"`        // Text Field 4 - object
	TargetS3Key  types.String `tfsdk:"target_s3_key"` // Text Field 12 - target_object
	Prefix       types.String `tfsdk:"prefix"`        // Text Field 7
	SourcePrefix types.String `tfsdk:"source_prefix"` // Text Field 13
	TargetPrefix types.String `tfsdk:"target_prefix"` // Text Field 14

	// File settings
	Sourcefile      types.String `tfsdk:"sourcefile"`       // Text Field 6
	TargetDirectory types.String `tfsdk:"target_directory"` // Text Field 9 - targetdir

	// Operation (Choice Field 4): copy, move
	Operation types.String `tfsdk:"operation"`

	// Write options
	DownloadWriteOptions types.String `tfsdk:"download_write_options"` // Choice Field 5
	UploadWriteOptions   types.String `tfsdk:"upload_write_options"`   // Choice Field 6

	// AWS credentials
	AwsAccessKeyId     types.String `tfsdk:"aws_access_key_id"`     // Credential Field 3
	AwsSecretAccessKey types.String `tfsdk:"aws_secret_access_key"` // Credential Field 4
	AwsDefaultRegion   types.String `tfsdk:"aws_default_region"`    // Text Field 5

	// Role-based access
	RoleBasedAccess types.String `tfsdk:"role_based_access"` // Choice Field 9: no, yes
	ServiceName     types.String `tfsdk:"service_name"`      // Choice Field 8: sts, s3
	RoleArn         types.String `tfsdk:"role_arn"`          // Text Field 10

	// Proxy settings
	UseProxy  types.String `tfsdk:"use_proxy"`  // Choice Field 1: 0, 1
	ProxyType types.String `tfsdk:"proxy_type"` // Choice Field 2: http, https, https_with_password
	Proxy     types.String `tfsdk:"proxy"`      // Text Field 1
	ProxyCred types.String `tfsdk:"proxy_cred"` // Credential Field 1
	Port      types.String `tfsdk:"port"`       // Text Field 2

	// Other options
	ShowDetails types.Bool   `tfsdk:"show_details"` // Boolean Field 2
	LogLevel    types.String `tfsdk:"log_level"`    // Choice Field 7: INFO, DEBUG, WARNING, ERROR, CRITICAL
	EndpointUrl types.String `tfsdk:"endpoint_url"` // Text Field 11
	Interval    types.String `tfsdk:"interval"`     // Choice Field 10: 10, 60, 180
	Acl         types.String `tfsdk:"acl"`          // Choice Field 11
}

// UniversalFieldWsData represents a field in the Universal Task API.
type UniversalFieldWsData struct {
	Name  string `json:"name,omitempty"`
	Label string `json:"label,omitempty"`
	Value any    `json:"value,omitempty"`
}

// TaskUniversalAwsS3APIModel represents the API request/response structure.
type TaskUniversalAwsS3APIModel struct {
	SysId    string `json:"sysId,omitempty"`
	Name     string `json:"name"`
	Version  int64  `json:"version,omitempty"`
	Type     string `json:"type"`
	Template string `json:"template"`
	Summary  string `json:"summary,omitempty"`

	Agent           string `json:"agent,omitempty"`
	AgentCluster    string `json:"agentCluster,omitempty"`
	AgentVar        string `json:"agentVar,omitempty"`
	AgentClusterVar string `json:"agentClusterVar,omitempty"`

	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

	ExitCodes          string `json:"exitCodes,omitempty"`
	ExitCodeProcessing string `json:"exitCodeProcessing,omitempty"`

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

	// Universal Task fields - mapped to slot names
	// Text fields
	TextField1  *UniversalFieldWsData `json:"textField1,omitempty"`  // proxy
	TextField2  *UniversalFieldWsData `json:"textField2,omitempty"`  // port
	TextField3  *UniversalFieldWsData `json:"textField3,omitempty"`  // object_store (bucket)
	TextField4  *UniversalFieldWsData `json:"textField4,omitempty"`  // object (s3_key)
	TextField5  *UniversalFieldWsData `json:"textField5,omitempty"`  // aws_default_region
	TextField6  *UniversalFieldWsData `json:"textField6,omitempty"`  // sourcefile
	TextField7  *UniversalFieldWsData `json:"textField7,omitempty"`  // prefix
	TextField8  *UniversalFieldWsData `json:"textField8,omitempty"`  // target_object_store (target_bucket)
	TextField9  *UniversalFieldWsData `json:"textField9,omitempty"`  // targetdir (target_directory)
	TextField10 *UniversalFieldWsData `json:"textField10,omitempty"` // rolearn (role_arn)
	TextField11 *UniversalFieldWsData `json:"textField11,omitempty"` // endpoint_url
	TextField12 *UniversalFieldWsData `json:"textField12,omitempty"` // target_object (target_s3_key)
	TextField13 *UniversalFieldWsData `json:"textField13,omitempty"` // source_prefix
	TextField14 *UniversalFieldWsData `json:"textField14,omitempty"` // target_prefix

	// Choice fields
	ChoiceField1  *UniversalFieldWsData `json:"choiceField1,omitempty"`  // useproxy
	ChoiceField2  *UniversalFieldWsData `json:"choiceField2,omitempty"`  // proxy_type
	ChoiceField3  *UniversalFieldWsData `json:"choiceField3,omitempty"`  // action
	ChoiceField4  *UniversalFieldWsData `json:"choiceField4,omitempty"`  // operation
	ChoiceField5  *UniversalFieldWsData `json:"choiceField5,omitempty"`  // writeoptions_download
	ChoiceField6  *UniversalFieldWsData `json:"choiceField6,omitempty"`  // writeoptions_upload
	ChoiceField7  *UniversalFieldWsData `json:"choiceField7,omitempty"`  // loglevel
	ChoiceField8  *UniversalFieldWsData `json:"choiceField8,omitempty"`  // service_name
	ChoiceField9  *UniversalFieldWsData `json:"choiceField9,omitempty"`  // rbca (role_based_access)
	ChoiceField10 *UniversalFieldWsData `json:"choiceField10,omitempty"` // interval
	ChoiceField11 *UniversalFieldWsData `json:"choiceField11,omitempty"` // acl

	// Boolean fields
	BooleanField2 *UniversalFieldWsData `json:"booleanField2,omitempty"` // show_details

	// Credential fields
	CredentialField1 *UniversalFieldWsData `json:"credentialField1,omitempty"` // proxycred
	CredentialField3 *UniversalFieldWsData `json:"credentialField3,omitempty"` // aws_access_key_id
	CredentialField4 *UniversalFieldWsData `json:"credentialField4,omitempty"` // aws_secret_access_key
}

func (r *TaskUniversalAwsS3Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_universal_aws_s3"
}

func (r *TaskUniversalAwsS3Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Universal Task based on the 'CS AWS S3' template for AWS S3 operations.",

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
				MarkdownDescription: "Description/summary of the task.",
				Optional:            true,
			},

			// Agent configuration
			"agent": schema.StringAttribute{
				MarkdownDescription: "Name of the agent to run the task on.",
				Optional:            true,
			},
			"agent_cluster": schema.StringAttribute{
				MarkdownDescription: "Name of the agent cluster to run the task on.",
				Optional:            true,
			},
			"agent_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the agent name.",
				Optional:            true,
			},
			"agent_cluster_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the agent cluster name.",
				Optional:            true,
			},

			// Credentials (for task execution)
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use for task execution.",
				Optional:            true,
			},
			"credentials_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the credentials name.",
				Optional:            true,
			},

			// Exit code handling
			"exit_codes": schema.StringAttribute{
				MarkdownDescription: "Exit codes that indicate success (e.g., '0' or '0,1,2'). Defaults to '0'.",
				Optional:            true,
				Computed:            true,
			},
			"exit_code_processing": schema.StringAttribute{
				MarkdownDescription: "How to process exit codes. Values: 'Success Exitcode Range', 'Failure Exitcode Range'.",
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
				MarkdownDescription: "Whether to retry indefinitely on failure.",
				Optional:            true,
				Computed:            true,
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Interval between retry attempts (in seconds).",
				Optional:            true,
				Computed:            true,
			},
			"retry_suppress_failure": schema.BoolAttribute{
				MarkdownDescription: "Whether to suppress failure notifications during retries.",
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

			// ===========================================
			// CS AWS S3 Template-specific fields
			// ===========================================

			"action": schema.StringAttribute{
				MarkdownDescription: "S3 action to perform. Values: `list-buckets`, `list-objects`, `upload-file`, `download-file`, `delete-objects`, `delete-bucket`, `create-bucket`, `copy-object-to-bucket`, `monitor-object`.",
				Optional:            true,
				Computed:            true,
			},

			// Bucket/Object settings
			"bucket": schema.StringAttribute{
				MarkdownDescription: "S3 bucket name.",
				Optional:            true,
			},
			"target_bucket": schema.StringAttribute{
				MarkdownDescription: "Target S3 bucket name (for copy operations).",
				Optional:            true,
			},
			"s3_key": schema.StringAttribute{
				MarkdownDescription: "S3 object key (path to object in bucket).",
				Optional:            true,
			},
			"target_s3_key": schema.StringAttribute{
				MarkdownDescription: "Target S3 object key (for copy operations).",
				Optional:            true,
			},
			"prefix": schema.StringAttribute{
				MarkdownDescription: "S3 key prefix for filtering objects.",
				Optional:            true,
			},
			"source_prefix": schema.StringAttribute{
				MarkdownDescription: "Source S3 key prefix (for copy operations).",
				Optional:            true,
			},
			"target_prefix": schema.StringAttribute{
				MarkdownDescription: "Target S3 key prefix (for copy operations).",
				Optional:            true,
			},

			// File settings
			"sourcefile": schema.StringAttribute{
				MarkdownDescription: "Local file path for upload operations.",
				Optional:            true,
			},
			"target_directory": schema.StringAttribute{
				MarkdownDescription: "Local directory path for download operations.",
				Optional:            true,
			},

			// Operation
			"operation": schema.StringAttribute{
				MarkdownDescription: "Operation type for copy-object-to-bucket action. Values: `copy`, `move`.",
				Optional:            true,
				Computed:            true,
			},

			// Write options
			"download_write_options": schema.StringAttribute{
				MarkdownDescription: "How to handle existing files on download. Values: `True` (overwrite), `False` (skip), `Timestamp`, `AlwaysTimestamp`, `Rename`.",
				Optional:            true,
				Computed:            true,
			},
			"upload_write_options": schema.StringAttribute{
				MarkdownDescription: "How to handle existing objects on upload. Values: `True` (overwrite), `False` (skip), `Timestamp`, `AlwaysTimestamp`.",
				Optional:            true,
				Computed:            true,
			},

			// AWS credentials
			"aws_access_key_id": schema.StringAttribute{
				MarkdownDescription: "Name of the credential containing the AWS Access Key ID.",
				Optional:            true,
			},
			"aws_secret_access_key": schema.StringAttribute{
				MarkdownDescription: "Name of the credential containing the AWS Secret Access Key.",
				Optional:            true,
			},
			"aws_default_region": schema.StringAttribute{
				MarkdownDescription: "AWS region (e.g., `us-east-1`).",
				Optional:            true,
			},

			// Role-based access
			"role_based_access": schema.StringAttribute{
				MarkdownDescription: "Whether to use IAM role-based access. Values: `no`, `yes`.",
				Optional:            true,
				Computed:            true,
			},
			"service_name": schema.StringAttribute{
				MarkdownDescription: "AWS service name for role assumption. Values: `sts`, `s3`.",
				Optional:            true,
				Computed:            true,
			},
			"role_arn": schema.StringAttribute{
				MarkdownDescription: "IAM Role ARN to assume.",
				Optional:            true,
			},

			// Proxy settings
			"use_proxy": schema.StringAttribute{
				MarkdownDescription: "Whether to use a proxy. Values: `0` (no), `1` (yes).",
				Optional:            true,
				Computed:            true,
			},
			"proxy_type": schema.StringAttribute{
				MarkdownDescription: "Proxy type. Values: `http`, `https`, `https_with_password`.",
				Optional:            true,
				Computed:            true,
			},
			"proxy": schema.StringAttribute{
				MarkdownDescription: "Proxy server address.",
				Optional:            true,
			},
			"proxy_cred": schema.StringAttribute{
				MarkdownDescription: "Name of the credential for proxy authentication.",
				Optional:            true,
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Proxy port.",
				Optional:            true,
			},

			// Other options
			"show_details": schema.BoolAttribute{
				MarkdownDescription: "Whether to show detailed output.",
				Optional:            true,
				Computed:            true,
			},
			"log_level": schema.StringAttribute{
				MarkdownDescription: "Log level. Values: `INFO`, `DEBUG`, `WARNING`, `ERROR`, `CRITICAL`.",
				Optional:            true,
				Computed:            true,
			},
			"endpoint_url": schema.StringAttribute{
				MarkdownDescription: "Custom S3 endpoint URL (for S3-compatible services).",
				Optional:            true,
			},
			"interval": schema.StringAttribute{
				MarkdownDescription: "Monitoring interval in seconds (for monitor-object action). Values: `10`, `60`, `180`.",
				Optional:            true,
				Computed:            true,
			},
			"acl": schema.StringAttribute{
				MarkdownDescription: "Access control list for uploaded objects. Values: `bucket-owner-full-control`, `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, `bucket-owner-read`, `log-delivery-write`.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *TaskUniversalAwsS3Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskUniversalAwsS3Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskUniversalAwsS3ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating universal AWS S3 task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Task",
			fmt.Sprintf("Could not create task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Task",
			fmt.Sprintf("Could not read task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created universal AWS S3 task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskUniversalAwsS3Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskUniversalAwsS3ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Task",
			fmt.Sprintf("Could not read task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskUniversalAwsS3Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskUniversalAwsS3ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskUniversalAwsS3ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating universal AWS S3 task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Task",
			fmt.Sprintf("Could not update task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Task",
			fmt.Sprintf("Could not read task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskUniversalAwsS3Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskUniversalAwsS3ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting universal AWS S3 task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Task",
			fmt.Sprintf("Could not delete task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskUniversalAwsS3Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskUniversalAwsS3Resource) readTask(ctx context.Context, data *TaskUniversalAwsS3ResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskUniversalAwsS3APIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// Helper to create a text field
func textField(name, value string) *UniversalFieldWsData {
	if value == "" {
		return nil
	}
	return &UniversalFieldWsData{
		Name:  name,
		Value: value,
	}
}

// Helper to create a choice field
func choiceField(name, value string) *UniversalFieldWsData {
	if value == "" {
		return nil
	}
	return &UniversalFieldWsData{
		Name:  name,
		Value: value,
	}
}

// Helper to create a choice field with default value (for required choice fields)
func choiceFieldWithDefault(name, value, defaultValue string) *UniversalFieldWsData {
	if value == "" {
		value = defaultValue
	}
	return &UniversalFieldWsData{
		Name:  name,
		Value: value,
	}
}

// Helper to create a boolean field
func booleanField(name string, value types.Bool) *UniversalFieldWsData {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &UniversalFieldWsData{
		Name:  name,
		Value: value.ValueBool(),
	}
}

// Helper to create a credential field
func credentialField(name, value string) *UniversalFieldWsData {
	if value == "" {
		return nil
	}
	return &UniversalFieldWsData{
		Name:  name,
		Value: value,
	}
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskUniversalAwsS3Resource) toAPIModel(ctx context.Context, data *TaskUniversalAwsS3ResourceModel) *TaskUniversalAwsS3APIModel {
	model := &TaskUniversalAwsS3APIModel{
		SysId:    data.SysId.ValueString(),
		Name:     data.Name.ValueString(),
		Type:     "taskUniversal",
		Template: "CS AWS S3",
		Summary:  data.Summary.ValueString(),

		Agent:           data.Agent.ValueString(),
		AgentCluster:    data.AgentCluster.ValueString(),
		AgentVar:        data.AgentVar.ValueString(),
		AgentClusterVar: data.AgentClusterVar.ValueString(),

		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),

		ExitCodes:          StringValueOrDefault(data.ExitCodes, "0"),
		ExitCodeProcessing: data.ExitCodeProcessing.ValueString(),

		RetryMaximum:         data.RetryMaximum.ValueInt64(),
		RetryIndefinitely:    data.RetryIndefinitely.ValueBool(),
		RetryInterval:        data.RetryInterval.ValueInt64(),
		RetrySuppressFailure: data.RetrySuppressFailure.ValueBool(),

		// Text fields
		TextField1:  textField("proxy", data.Proxy.ValueString()),
		TextField2:  textField("port", data.Port.ValueString()),
		TextField3:  textField("object_store", data.Bucket.ValueString()),
		TextField4:  textField("object", data.S3Key.ValueString()),
		TextField5:  textField("aws_default_region", data.AwsDefaultRegion.ValueString()),
		TextField6:  textField("sourcefile", data.Sourcefile.ValueString()),
		TextField7:  textField("prefix", data.Prefix.ValueString()),
		TextField8:  textField("target_object_store", data.TargetBucket.ValueString()),
		TextField9:  textField("targetdir", data.TargetDirectory.ValueString()),
		TextField10: textField("rolearn", data.RoleArn.ValueString()),
		TextField11: textField("endpoint_url", data.EndpointUrl.ValueString()),
		TextField12: textField("target_object", data.TargetS3Key.ValueString()),
		TextField13: textField("source_prefix", data.SourcePrefix.ValueString()),
		TextField14: textField("target_prefix", data.TargetPrefix.ValueString()),

		// Choice fields - all have required defaults per the API
		ChoiceField1:  choiceFieldWithDefault("useproxy", data.UseProxy.ValueString(), "0"),
		ChoiceField2:  choiceFieldWithDefault("proxy_type", data.ProxyType.ValueString(), "http"),
		ChoiceField3:  choiceFieldWithDefault("action", data.Action.ValueString(), "list-buckets"),
		ChoiceField4:  choiceFieldWithDefault("operation", data.Operation.ValueString(), "copy"),
		ChoiceField5:  choiceFieldWithDefault("writeoptions_download", data.DownloadWriteOptions.ValueString(), "False"),
		ChoiceField6:  choiceFieldWithDefault("writeoptions_upload", data.UploadWriteOptions.ValueString(), "False"),
		ChoiceField7:  choiceFieldWithDefault("loglevel", data.LogLevel.ValueString(), "INFO"),
		ChoiceField8:  choiceFieldWithDefault("service_name", data.ServiceName.ValueString(), "sts"),
		ChoiceField9:  choiceFieldWithDefault("rbca", data.RoleBasedAccess.ValueString(), "no"),
		ChoiceField10: choiceFieldWithDefault("interval", data.Interval.ValueString(), "10"),
		ChoiceField11: choiceFieldWithDefault("acl", data.Acl.ValueString(), "bucket-owner-full-control"),

		// Boolean fields
		BooleanField2: booleanField("show_details", data.ShowDetails),

		// Credential fields
		CredentialField1: credentialField("proxycred", data.ProxyCred.ValueString()),
		CredentialField3: credentialField("aws_access_key_id", data.AwsAccessKeyId.ValueString()),
		CredentialField4: credentialField("aws_secret_access_key", data.AwsSecretAccessKey.ValueString()),
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

// Helper to extract string value from universal field
func getFieldStringValue(field *UniversalFieldWsData) string {
	if field == nil {
		return ""
	}
	if v, ok := field.Value.(string); ok {
		return v
	}
	return ""
}

// Helper to extract bool value from universal field
func getFieldBoolValue(field *UniversalFieldWsData) (bool, bool) {
	if field == nil {
		return false, false
	}
	if v, ok := field.Value.(bool); ok {
		return v, true
	}
	// Handle string "true"/"false"
	if v, ok := field.Value.(string); ok {
		return v == "true", true
	}
	return false, false
}

// fromAPIModel converts an API model to the Terraform model.
func (r *TaskUniversalAwsS3Resource) fromAPIModel(ctx context.Context, apiModel *TaskUniversalAwsS3APIModel, data *TaskUniversalAwsS3ResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Optional fields
	data.Summary = StringValueOrNull(apiModel.Summary)
	data.Agent = StringValueOrNull(apiModel.Agent)
	data.AgentCluster = StringValueOrNull(apiModel.AgentCluster)
	data.AgentVar = StringValueOrNull(apiModel.AgentVar)
	data.AgentClusterVar = StringValueOrNull(apiModel.AgentClusterVar)
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)
	data.ExitCodeProcessing = StringValueOrNull(apiModel.ExitCodeProcessing)

	// Computed fields
	data.RetryMaximum = types.Int64Value(apiModel.RetryMaximum)
	data.RetryIndefinitely = types.BoolValue(apiModel.RetryIndefinitely)
	data.RetryInterval = types.Int64Value(apiModel.RetryInterval)
	data.RetrySuppressFailure = types.BoolValue(apiModel.RetrySuppressFailure)

	// Text fields
	data.Proxy = StringValueOrNull(getFieldStringValue(apiModel.TextField1))
	data.Port = StringValueOrNull(getFieldStringValue(apiModel.TextField2))
	data.Bucket = StringValueOrNull(getFieldStringValue(apiModel.TextField3))
	data.S3Key = StringValueOrNull(getFieldStringValue(apiModel.TextField4))
	data.AwsDefaultRegion = StringValueOrNull(getFieldStringValue(apiModel.TextField5))
	data.Sourcefile = StringValueOrNull(getFieldStringValue(apiModel.TextField6))
	data.Prefix = StringValueOrNull(getFieldStringValue(apiModel.TextField7))
	data.TargetBucket = StringValueOrNull(getFieldStringValue(apiModel.TextField8))
	data.TargetDirectory = StringValueOrNull(getFieldStringValue(apiModel.TextField9))
	data.RoleArn = StringValueOrNull(getFieldStringValue(apiModel.TextField10))
	data.EndpointUrl = StringValueOrNull(getFieldStringValue(apiModel.TextField11))
	data.TargetS3Key = StringValueOrNull(getFieldStringValue(apiModel.TextField12))
	data.SourcePrefix = StringValueOrNull(getFieldStringValue(apiModel.TextField13))
	data.TargetPrefix = StringValueOrNull(getFieldStringValue(apiModel.TextField14))

	// Choice fields
	data.UseProxy = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField1))
	data.ProxyType = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField2))
	data.Action = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField3))
	data.Operation = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField4))
	data.DownloadWriteOptions = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField5))
	data.UploadWriteOptions = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField6))
	data.LogLevel = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField7))
	data.ServiceName = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField8))
	data.RoleBasedAccess = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField9))
	data.Interval = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField10))
	data.Acl = StringValueOrNull(getFieldStringValue(apiModel.ChoiceField11))

	// Boolean fields
	if v, ok := getFieldBoolValue(apiModel.BooleanField2); ok {
		data.ShowDetails = types.BoolValue(v)
	} else {
		data.ShowDetails = types.BoolNull()
	}

	// Credential fields
	data.ProxyCred = StringValueOrNull(getFieldStringValue(apiModel.CredentialField1))
	data.AwsAccessKeyId = StringValueOrNull(getFieldStringValue(apiModel.CredentialField3))
	data.AwsSecretAccessKey = StringValueOrNull(getFieldStringValue(apiModel.CredentialField4))

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
