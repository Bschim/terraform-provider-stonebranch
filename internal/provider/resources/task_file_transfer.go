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
	_ resource.Resource                   = &TaskFileTransferResource{}
	_ resource.ResourceWithImportState    = &TaskFileTransferResource{}
	_ resource.ResourceWithValidateConfig = &TaskFileTransferResource{}
)

func NewTaskFileTransferResource() resource.Resource {
	return &TaskFileTransferResource{}
}

// TaskFileTransferResource defines the resource implementation.
type TaskFileTransferResource struct {
	client *client.Client
}

// TaskFileTransferResourceModel describes the resource data model.
type TaskFileTransferResourceModel struct {
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

	// Transfer configuration
	TransferDirection types.String `tfsdk:"transfer_direction"`
	TransferMode      types.String `tfsdk:"transfer_mode"`
	ServerType        types.String `tfsdk:"server_type"`

	// Remote configuration
	RemoteServer      types.String `tfsdk:"remote_server"`
	RemoteFilename    types.String `tfsdk:"remote_filename"`
	RemoteCredentials types.String `tfsdk:"remote_credentials"`
	RemoteCredVar     types.String `tfsdk:"remote_credentials_var"`

	// Local configuration
	LocalFilename types.String `tfsdk:"local_filename"`

	// Credentials
	Credentials    types.String `tfsdk:"credentials"`
	CredentialsVar types.String `tfsdk:"credentials_var"`

	// Exit code handling
	ExitCodes          types.String `tfsdk:"exit_codes"`
	ExitCodeProcessing types.String `tfsdk:"exit_code_processing"`

	// Options
	UseRegex types.Bool   `tfsdk:"use_regex"`
	Encrypt  types.String `tfsdk:"encrypt"`
	Compress types.String `tfsdk:"compress"`

	// UDM agent configuration
	PrimaryBrokerChoice   types.String `tfsdk:"primary_broker_choice"`
	PrimaryBroker         types.String `tfsdk:"primary_broker"`
	PrimaryCluster        types.String `tfsdk:"primary_cluster"`
	PrimaryClusterRef     types.String `tfsdk:"primary_cluster_ref"`
	PrimaryCredentials    types.String `tfsdk:"primary_credentials"`
	PrimaryCredVar        types.String `tfsdk:"primary_cred_var"`
	PrimaryFilesys        types.String `tfsdk:"primary_filesys"`
	PrimaryOpenOptions    types.String `tfsdk:"primary_open_options"`
	SecondaryBrokerChoice types.String `tfsdk:"secondary_broker_choice"`
	SecondaryBroker       types.String `tfsdk:"secondary_broker"`
	SecondaryCluster      types.String `tfsdk:"secondary_cluster"`
	SecondaryClusterRef   types.String `tfsdk:"secondary_cluster_ref"`
	SecondaryCredentials  types.String `tfsdk:"secondary_credentials"`
	SecondaryCredVar      types.String `tfsdk:"secondary_cred_var"`
	SecondaryFilesys      types.String `tfsdk:"secondary_filesys"`
	SecondaryOpenOptions  types.String `tfsdk:"secondary_open_options"`
	UdmOperation          types.String `tfsdk:"udm_operation"`
	UdmOptions            types.String `tfsdk:"udm_options"`
	Script                types.String `tfsdk:"script"`
	Format                types.String `tfsdk:"format"`
	FormOrScript          types.String `tfsdk:"form_or_script"`
	Command               types.String `tfsdk:"command"`

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

// TaskFileTransferAPIModel represents the API request/response structure.
type TaskFileTransferAPIModel struct {
	SysId   string `json:"sysId,omitempty"`
	Name    string `json:"name"`
	Version int64  `json:"version,omitempty"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`

	Agent           string `json:"agent,omitempty"`
	AgentCluster    string `json:"agentCluster,omitempty"`
	AgentVar        string `json:"agentVar,omitempty"`
	AgentClusterVar string `json:"agentClusterVar,omitempty"`

	TransferDirection string `json:"transferDirection,omitempty"`
	TransferMode      string `json:"transferMode,omitempty"`
	ServerType        string `json:"serverType,omitempty"`

	RemoteServer      string `json:"remoteServer,omitempty"`
	RemoteFilename    string `json:"remoteFilename,omitempty"`
	RemoteCredentials string `json:"remoteCredentials,omitempty"`
	RemoteCredVar     string `json:"remoteCredVar,omitempty"`

	LocalFilename string `json:"localFilename,omitempty"`

	Credentials    string `json:"credentials,omitempty"`
	CredentialsVar string `json:"credentialsVar,omitempty"`

	ExitCodes          string `json:"exitCodes,omitempty"`
	ExitCodeProcessing string `json:"exitCodeProcessing,omitempty"`

	UseRegex bool   `json:"useRegex,omitempty"`
	Encrypt  string `json:"encrypt,omitempty"`
	Compress string `json:"compress,omitempty"`

	PrimaryBrokerChoice   string `json:"primaryBrokerChoice,omitempty"`
	PrimaryBroker         string `json:"primaryBroker,omitempty"`
	PrimaryCluster        string `json:"primaryCluster,omitempty"`
	PrimaryClusterRef     string `json:"primaryClusterRef,omitempty"`
	PrimaryCredentials    string `json:"primaryCredentials,omitempty"`
	PrimaryCredVar        string `json:"primaryCredVar,omitempty"`
	PrimaryFilesys        string `json:"primaryFilesys,omitempty"`
	PrimaryOpenOptions    string `json:"primaryOpenOptions,omitempty"`
	SecondaryBrokerChoice string `json:"secondaryBrokerChoice,omitempty"`
	SecondaryBroker       string `json:"secondaryBroker,omitempty"`
	SecondaryCluster      string `json:"secondaryCluster,omitempty"`
	SecondaryClusterRef   string `json:"secondaryClusterRef,omitempty"`
	SecondaryCredentials  string `json:"secondaryCredentials,omitempty"`
	SecondaryCredVar      string `json:"secondaryCredVar,omitempty"`
	SecondaryFilesys      string `json:"secondaryFilesys,omitempty"`
	SecondaryOpenOptions  string `json:"secondaryOpenOptions,omitempty"`
	UdmOperation          string `json:"udmOperation,omitempty"`
	UdmOptions            string `json:"udmOptions,omitempty"`
	Script                string `json:"script,omitempty"`
	Format                string `json:"format,omitempty"`
	FormOrScript          string `json:"formOrScript,omitempty"`
	Command               string `json:"command,omitempty"`

	Variables []TaskVariableAPIModel `json:"variables,omitempty"`

	HoldResources    bool                          `json:"holdResources,omitempty"`
	ExclusiveTasks   []TaskExclusiveTaskAPIModel   `json:"exclusiveTasks,omitempty"`
	VirtualResources []TaskVirtualResourceAPIModel `json:"virtualResources,omitempty"`

	Actions *ActionsAPIModel `json:"actions,omitempty"`

	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *TaskFileTransferResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_file_transfer"
}

func (r *TaskFileTransferResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch File Transfer Task. Used for transferring files between systems using FTP, SFTP, or other protocols.",

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

			// Transfer configuration
			"transfer_direction": schema.StringAttribute{
				MarkdownDescription: "Direction of file transfer: 'GET' (remote to local) or 'PUT' (local to remote). Default is 'PUT'. Only applies to UDM agents.",
				Optional:            true,
				Computed:            true,
			},
			"transfer_mode": schema.StringAttribute{
				MarkdownDescription: "Transfer mode: 'ASCII' or 'Binary'.",
				Optional:            true,
				Computed:            true,
			},
			"server_type": schema.StringAttribute{
				MarkdownDescription: "Type of server: 'FTP', 'SFTP', 'FTPS', etc.",
				Optional:            true,
				Computed:            true,
			},

			// Remote configuration
			"remote_server": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address of the remote server.",
				Optional:            true,
			},
			"remote_filename": schema.StringAttribute{
				MarkdownDescription: "Path to the file on the remote server.",
				Optional:            true,
			},
			"remote_credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use for the remote server.",
				Optional:            true,
			},
			"remote_credentials_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the remote credentials name.",
				Optional:            true,
			},

			// Local configuration
			"local_filename": schema.StringAttribute{
				MarkdownDescription: "Path to the file on the local system.",
				Optional:            true,
			},

			// Credentials
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
				MarkdownDescription: "Exit codes that indicate success (e.g., '0' or '0,1,2').",
				Optional:            true,
				Computed:            true,
			},
			"exit_code_processing": schema.StringAttribute{
				MarkdownDescription: "How to process exit codes. Values: 'Success Exitcode Range', 'Failure Exitcode Range'.",
				Optional:            true,
				Computed:            true,
			},

			// Options
			"use_regex": schema.BoolAttribute{
				MarkdownDescription: "Whether to use regex for filename matching.",
				Optional:            true,
				Computed:            true,
			},
			"encrypt": schema.StringAttribute{
				MarkdownDescription: "Encryption setting for the transfer.",
				Optional:            true,
				Computed:            true,
			},
			"compress": schema.StringAttribute{
				MarkdownDescription: "Compression setting for the transfer.",
				Optional:            true,
				Computed:            true,
			},

			// UDM agent configuration (only applies when server_type = "UDM")
			"primary_broker_choice": schema.StringAttribute{
				MarkdownDescription: "Primary UDM agent option, e.g. `Agent` or `Agent Cluster`. Required when `server_type` is `UDM`.",
				Optional:            true,
				Computed:            true,
			},
			"primary_broker": schema.StringAttribute{
				MarkdownDescription: "Primary UDM agent name. Used when `primary_broker_choice` is `Agent`.",
				Optional:            true,
			},
			"primary_cluster": schema.StringAttribute{
				MarkdownDescription: "Primary UDM agent cluster name.",
				Optional:            true,
			},
			"primary_cluster_ref": schema.StringAttribute{
				MarkdownDescription: "Primary UDM agent cluster reference. Used when `primary_broker_choice` is `Agent Cluster`.",
				Optional:            true,
			},
			"primary_credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use for the primary UDM agent.",
				Optional:            true,
			},
			"primary_cred_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the primary UDM agent credentials name.",
				Optional:            true,
			},
			"primary_filesys": schema.StringAttribute{
				MarkdownDescription: "Primary UDM agent file system type.",
				Optional:            true,
				Computed:            true,
			},
			"primary_open_options": schema.StringAttribute{
				MarkdownDescription: "Additional open options for the primary UDM agent.",
				Optional:            true,
			},
			"secondary_broker_choice": schema.StringAttribute{
				MarkdownDescription: "Secondary UDM agent option, e.g. `Agent` or `Agent Cluster`. Required when `server_type` is `UDM`.",
				Optional:            true,
				Computed:            true,
			},
			"secondary_broker": schema.StringAttribute{
				MarkdownDescription: "Secondary UDM agent name. Used when `secondary_broker_choice` is `Agent`.",
				Optional:            true,
			},
			"secondary_cluster": schema.StringAttribute{
				MarkdownDescription: "Secondary UDM agent cluster name.",
				Optional:            true,
			},
			"secondary_cluster_ref": schema.StringAttribute{
				MarkdownDescription: "Secondary UDM agent cluster reference. Used when `secondary_broker_choice` is `Agent Cluster`.",
				Optional:            true,
			},
			"secondary_credentials": schema.StringAttribute{
				MarkdownDescription: "Name of the credentials to use for the secondary UDM agent.",
				Optional:            true,
			},
			"secondary_cred_var": schema.StringAttribute{
				MarkdownDescription: "Variable containing the secondary UDM agent credentials name.",
				Optional:            true,
			},
			"secondary_filesys": schema.StringAttribute{
				MarkdownDescription: "Secondary UDM agent file system type.",
				Optional:            true,
				Computed:            true,
			},
			"secondary_open_options": schema.StringAttribute{
				MarkdownDescription: "Additional open options for the secondary UDM agent.",
				Optional:            true,
			},
			"udm_operation": schema.StringAttribute{
				MarkdownDescription: "UDM operation to perform, e.g. `Copy`.",
				Optional:            true,
				Computed:            true,
			},
			"udm_options": schema.StringAttribute{
				MarkdownDescription: "Additional options for the UDM operation.",
				Optional:            true,
			},
			"script": schema.StringAttribute{
				MarkdownDescription: "Name of the UDM script to run. Used when `form_or_script` is `Script`.",
				Optional:            true,
			},
			"format": schema.StringAttribute{
				MarkdownDescription: "UDM transfer format, e.g. `Binary` or `ASCII`.",
				Optional:            true,
				Computed:            true,
			},
			"form_or_script": schema.StringAttribute{
				MarkdownDescription: "Whether the UDM transfer is configured via `Form` or `Script`.",
				Optional:            true,
				Computed:            true,
			},
			"command": schema.StringAttribute{
				MarkdownDescription: "UDM command to run, e.g. `GET` or `PUT`.",
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

func (r *TaskFileTransferResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TaskFileTransferResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If server_type is unknown (e.g. computed from another resource),
	// defer this check to the server rather than failing the plan early.
	if data.ServerType.IsUnknown() {
		return
	}

	if data.ServerType.ValueString() != "UDM" {
		return
	}

	if !isSet(data.PrimaryBrokerChoice) && !data.PrimaryBrokerChoice.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("primary_broker_choice"),
			"Missing Required Field",
			`"primary_broker_choice" must be set when "server_type" is "UDM".`,
		)
	}

	if !isSet(data.SecondaryBrokerChoice) && !data.SecondaryBrokerChoice.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("secondary_broker_choice"),
			"Missing Required Field",
			`"secondary_broker_choice" must be set when "server_type" is "UDM".`,
		)
	}
}

func (r *TaskFileTransferResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaskFileTransferResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskFileTransferResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating file transfer task", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the task
	_, err := r.client.Post(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating File Transfer Task",
			fmt.Sprintf("Could not create file transfer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created task to get sysId and other computed fields
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created File Transfer Task",
			fmt.Sprintf("Could not read file transfer task %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created file transfer task", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskFileTransferResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskFileTransferResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readTask(ctx, &data)
	if err != nil {
		// Check if task was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "File transfer task not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading File Transfer Task",
			fmt.Sprintf("Could not read file transfer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskFileTransferResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskFileTransferResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state TaskFileTransferResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating file transfer task", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the task
	_, err := r.client.Put(ctx, "/resources/task", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating File Transfer Task",
			fmt.Sprintf("Could not update file transfer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readTask(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated File Transfer Task",
			fmt.Sprintf("Could not read file transfer task %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskFileTransferResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskFileTransferResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting file transfer task", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("taskid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/task", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting File Transfer Task",
			fmt.Sprintf("Could not delete file transfer task %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *TaskFileTransferResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readTask fetches the task from the API and updates the model.
func (r *TaskFileTransferResource) readTask(ctx context.Context, data *TaskFileTransferResourceModel) error {
	query := url.Values{}
	query.Set("taskname", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/task", query)
	if err != nil {
		return err
	}

	var apiModel TaskFileTransferAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse task response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *TaskFileTransferResource) toAPIModel(ctx context.Context, data *TaskFileTransferResourceModel) *TaskFileTransferAPIModel {
	model := &TaskFileTransferAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    "taskFtp",
		Summary: data.Summary.ValueString(),

		Agent:           data.Agent.ValueString(),
		AgentCluster:    data.AgentCluster.ValueString(),
		AgentVar:        data.AgentVar.ValueString(),
		AgentClusterVar: data.AgentClusterVar.ValueString(),

		TransferDirection: data.TransferDirection.ValueString(),
		TransferMode:      data.TransferMode.ValueString(),
		ServerType:        data.ServerType.ValueString(),

		RemoteServer:      data.RemoteServer.ValueString(),
		RemoteFilename:    data.RemoteFilename.ValueString(),
		RemoteCredentials: data.RemoteCredentials.ValueString(),
		RemoteCredVar:     data.RemoteCredVar.ValueString(),

		LocalFilename: data.LocalFilename.ValueString(),

		Credentials:    data.Credentials.ValueString(),
		CredentialsVar: data.CredentialsVar.ValueString(),

		ExitCodes:          data.ExitCodes.ValueString(),
		ExitCodeProcessing: data.ExitCodeProcessing.ValueString(),

		UseRegex: data.UseRegex.ValueBool(),
		Encrypt:  data.Encrypt.ValueString(),
		Compress: data.Compress.ValueString(),

		PrimaryBrokerChoice:   data.PrimaryBrokerChoice.ValueString(),
		PrimaryBroker:         data.PrimaryBroker.ValueString(),
		PrimaryCluster:        data.PrimaryCluster.ValueString(),
		PrimaryClusterRef:     data.PrimaryClusterRef.ValueString(),
		PrimaryCredentials:    data.PrimaryCredentials.ValueString(),
		PrimaryCredVar:        data.PrimaryCredVar.ValueString(),
		PrimaryFilesys:        data.PrimaryFilesys.ValueString(),
		PrimaryOpenOptions:    data.PrimaryOpenOptions.ValueString(),
		SecondaryBrokerChoice: data.SecondaryBrokerChoice.ValueString(),
		SecondaryBroker:       data.SecondaryBroker.ValueString(),
		SecondaryCluster:      data.SecondaryCluster.ValueString(),
		SecondaryClusterRef:   data.SecondaryClusterRef.ValueString(),
		SecondaryCredentials:  data.SecondaryCredentials.ValueString(),
		SecondaryCredVar:      data.SecondaryCredVar.ValueString(),
		SecondaryFilesys:      data.SecondaryFilesys.ValueString(),
		SecondaryOpenOptions:  data.SecondaryOpenOptions.ValueString(),
		UdmOperation:          data.UdmOperation.ValueString(),
		UdmOptions:            data.UdmOptions.ValueString(),
		Script:                data.Script.ValueString(),
		Format:                data.Format.ValueString(),
		FormOrScript:          data.FormOrScript.ValueString(),
		Command:               data.Command.ValueString(),
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
func (r *TaskFileTransferResource) fromAPIModel(ctx context.Context, apiModel *TaskFileTransferAPIModel, data *TaskFileTransferResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Basic info
	data.Summary = StringValueOrNull(apiModel.Summary)

	// Agent configuration
	data.Agent = StringValueOrNull(apiModel.Agent)
	data.AgentCluster = StringValueOrNull(apiModel.AgentCluster)
	data.AgentVar = StringValueOrNull(apiModel.AgentVar)
	data.AgentClusterVar = StringValueOrNull(apiModel.AgentClusterVar)

	// Transfer configuration
	data.TransferDirection = StringValueOrNull(apiModel.TransferDirection)
	data.TransferMode = StringValueOrNull(apiModel.TransferMode)
	data.ServerType = StringValueOrNull(apiModel.ServerType)

	// Remote configuration
	data.RemoteServer = StringValueOrNull(apiModel.RemoteServer)
	data.RemoteFilename = StringValueOrNull(apiModel.RemoteFilename)
	data.RemoteCredentials = StringValueOrNull(apiModel.RemoteCredentials)
	data.RemoteCredVar = StringValueOrNull(apiModel.RemoteCredVar)

	// Local configuration
	data.LocalFilename = StringValueOrNull(apiModel.LocalFilename)

	// Credentials
	data.Credentials = StringValueOrNull(apiModel.Credentials)
	data.CredentialsVar = StringValueOrNull(apiModel.CredentialsVar)

	// Exit code handling
	data.ExitCodes = StringValueOrNull(apiModel.ExitCodes)
	data.ExitCodeProcessing = StringValueOrNull(apiModel.ExitCodeProcessing)

	// Options
	data.UseRegex = types.BoolValue(apiModel.UseRegex)
	data.Encrypt = StringValueOrNull(apiModel.Encrypt)
	data.Compress = StringValueOrNull(apiModel.Compress)

	// UDM agent configuration
	data.PrimaryBrokerChoice = StringValueOrNull(apiModel.PrimaryBrokerChoice)
	data.PrimaryBroker = StringValueOrNull(apiModel.PrimaryBroker)
	data.PrimaryCluster = StringValueOrNull(apiModel.PrimaryCluster)
	data.PrimaryClusterRef = StringValueOrNull(apiModel.PrimaryClusterRef)
	data.PrimaryCredentials = StringValueOrNull(apiModel.PrimaryCredentials)
	data.PrimaryCredVar = StringValueOrNull(apiModel.PrimaryCredVar)
	data.PrimaryFilesys = StringValueOrNull(apiModel.PrimaryFilesys)
	data.PrimaryOpenOptions = StringValueOrNull(apiModel.PrimaryOpenOptions)
	data.SecondaryBrokerChoice = StringValueOrNull(apiModel.SecondaryBrokerChoice)
	data.SecondaryBroker = StringValueOrNull(apiModel.SecondaryBroker)
	data.SecondaryCluster = StringValueOrNull(apiModel.SecondaryCluster)
	data.SecondaryClusterRef = StringValueOrNull(apiModel.SecondaryClusterRef)
	data.SecondaryCredentials = StringValueOrNull(apiModel.SecondaryCredentials)
	data.SecondaryCredVar = StringValueOrNull(apiModel.SecondaryCredVar)
	data.SecondaryFilesys = StringValueOrNull(apiModel.SecondaryFilesys)
	data.SecondaryOpenOptions = StringValueOrNull(apiModel.SecondaryOpenOptions)
	data.UdmOperation = StringValueOrNull(apiModel.UdmOperation)
	data.UdmOptions = StringValueOrNull(apiModel.UdmOptions)
	data.Script = StringValueOrNull(apiModel.Script)
	data.Format = StringValueOrNull(apiModel.Format)
	data.FormOrScript = StringValueOrNull(apiModel.FormOrScript)
	data.Command = StringValueOrNull(apiModel.Command)

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
