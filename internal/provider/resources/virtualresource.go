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
	_ resource.Resource                = &VirtualResourceResource{}
	_ resource.ResourceWithImportState = &VirtualResourceResource{}
)

func NewVirtualResourceResource() resource.Resource {
	return &VirtualResourceResource{}
}

// VirtualResourceResource defines the resource implementation.
//
// NOTE: Unlike most other UAC resource types, the wire endpoint for this
// resource is "/resources/virtual", NOT "/resources/virtualresource". This
// mirrors the same kind of naming divergence documented in script.go for the
// "scriptName" field.
type VirtualResourceResource struct {
	client *client.Client
}

// VirtualResourceResourceModel describes the resource data model.
type VirtualResourceResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Content
	Limit   types.Int64  `tfsdk:"limit"`
	Summary types.String `tfsdk:"summary"`
	Type    types.String `tfsdk:"type"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// VirtualResourceAPIModel represents the API request/response structure.
type VirtualResourceAPIModel struct {
	SysId         string   `json:"sysId,omitempty"`
	Name          string   `json:"name"`
	Version       int64    `json:"version,omitempty"`
	Limit         int64    `json:"limit,omitempty"`
	Summary       string   `json:"summary,omitempty"`
	Type          string   `json:"type,omitempty"`
	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *VirtualResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_resource"
}

func (r *VirtualResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Virtual Resource. Virtual Resources provide concurrency " +
			"control for tasks (e.g. limiting how many tasks may run against a shared resource at once). " +
			"Note: the underlying UAC API endpoint for this resource is `/resources/virtual`, not `/resources/virtualresource`.",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the virtual resource (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the virtual resource.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the virtual resource (for optimistic locking).",
				Computed:            true,
			},

			// Content
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum concurrent usage allowed for this virtual resource. " +
					"Defaults to server setting if not specified.",
				Optional: true,
				Computed: true,
			},
			"summary": schema.StringAttribute{
				MarkdownDescription: "Description/summary of the virtual resource.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of virtual resource. Valid values: `Renewable`, `Boundary`, `Depletable`. " +
					"Defaults to server setting if not specified.",
				Optional: true,
				Computed: true,
			},

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this virtual resource belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *VirtualResourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtualResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating virtual resource", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the virtual resource
	_, err := r.client.Post(ctx, "/resources/virtual", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Virtual Resource",
			fmt.Sprintf("Could not create virtual resource %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created virtual resource to get sysId and other computed fields
	err = r.readVirtualResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Virtual Resource",
			fmt.Sprintf("Could not read virtual resource %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created virtual resource", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtualResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readVirtualResource(ctx, &data)
	if err != nil {
		// Check if virtual resource was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Virtual resource not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Virtual Resource",
			fmt.Sprintf("Could not read virtual resource %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtualResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state VirtualResourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating virtual resource", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the virtual resource
	_, err := r.client.Put(ctx, "/resources/virtual", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Virtual Resource",
			fmt.Sprintf("Could not update virtual resource %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readVirtualResource(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Virtual Resource",
			fmt.Sprintf("Could not read virtual resource %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtualResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting virtual resource", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("resourceid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/virtual", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Virtual Resource",
			fmt.Sprintf("Could not delete virtual resource %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *VirtualResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readVirtualResource fetches the virtual resource from the API and updates the model.
func (r *VirtualResourceResource) readVirtualResource(ctx context.Context, data *VirtualResourceResourceModel) error {
	query := url.Values{}
	query.Set("resourcename", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/virtual", query)
	if err != nil {
		return err
	}

	var apiModel VirtualResourceAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse virtual resource response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *VirtualResourceResource) toAPIModel(ctx context.Context, data *VirtualResourceResourceModel) *VirtualResourceAPIModel {
	model := &VirtualResourceAPIModel{
		SysId:   data.SysId.ValueString(),
		Name:    data.Name.ValueString(),
		Limit:   data.Limit.ValueInt64(),
		Summary: data.Summary.ValueString(),
		Type:    data.Type.ValueString(),
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
func (r *VirtualResourceResource) fromAPIModel(ctx context.Context, apiModel *VirtualResourceAPIModel, data *VirtualResourceResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.Name)
	data.Version = types.Int64Value(apiModel.Version)

	// Content
	data.Limit = types.Int64Value(apiModel.Limit)
	data.Summary = StringValueOrNull(apiModel.Summary)
	data.Type = StringValueOrNull(apiModel.Type)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	} else {
		data.OpswiseGroups = types.ListNull(types.StringType)
	}
}
