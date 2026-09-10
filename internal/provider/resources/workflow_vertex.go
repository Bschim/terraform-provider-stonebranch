package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

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
	_ resource.Resource                = &WorkflowVertexResource{}
	_ resource.ResourceWithImportState = &WorkflowVertexResource{}
)

func NewWorkflowVertexResource() resource.Resource {
	return &WorkflowVertexResource{}
}

// WorkflowVertexResource defines the resource implementation.
type WorkflowVertexResource struct {
	client *client.Client
}

// WorkflowVertexResourceModel describes the resource data model.
type WorkflowVertexResourceModel struct {
	// Identity - composite key
	WorkflowName types.String `tfsdk:"workflow_name"`
	TaskName     types.String `tfsdk:"task_name"`
	VertexId     types.String `tfsdk:"vertex_id"`

	// Optional
	Alias   types.String `tfsdk:"alias"`
	VertexX types.String `tfsdk:"vertex_x"`
	VertexY types.String `tfsdk:"vertex_y"`
}

// WorkflowVertexAPIModel represents the API request structure for creating a vertex.
type WorkflowVertexAPIModel struct {
	Task     *TaskRef `json:"task,omitempty"`
	Alias    string   `json:"alias,omitempty"`
	VertexId string   `json:"vertexId,omitempty"`
	VertexX  string   `json:"vertexX,omitempty"`
	VertexY  string   `json:"vertexY,omitempty"`
}

// TaskRef represents a task reference in the API.
type TaskRef struct {
	Value string `json:"value,omitempty"`
}

// WorkflowVertexResponseModel represents the API response structure.
type WorkflowVertexResponseModel struct {
	Task     *TaskRefResponse `json:"task,omitempty"`
	Alias    string           `json:"alias,omitempty"`
	VertexId string           `json:"vertexId,omitempty"`
	VertexX  string           `json:"vertexX,omitempty"`
	VertexY  string           `json:"vertexY,omitempty"`
}

type TaskRefResponse struct {
	Value string `json:"value,omitempty"`
	SysId string `json:"sysId,omitempty"`
}

func (r *WorkflowVertexResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_vertex"
}

func (r *WorkflowVertexResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a task (vertex) within a StoneBranch Workflow. Use this resource to add existing tasks to a workflow.",

		Attributes: map[string]schema.Attribute{
			"workflow_name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow to add the task to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"task_name": schema.StringAttribute{
				MarkdownDescription: "Name of the task to add to the workflow.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vertex_id": schema.StringAttribute{
				MarkdownDescription: "Unique vertex ID assigned by StoneBranch. Used to identify this task instance within the workflow.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Alias for this task instance in the workflow. Useful when the same task appears multiple times.",
				Optional:            true,
			},
			"vertex_x": schema.StringAttribute{
				MarkdownDescription: "X coordinate for the task position in the workflow diagram.",
				Optional:            true,
				Computed:            true,
			},
			"vertex_y": schema.StringAttribute{
				MarkdownDescription: "Y coordinate for the task position in the workflow diagram.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *WorkflowVertexResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowVertexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowVertexResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Adding task to workflow", map[string]any{
		"workflow": data.WorkflowName.ValueString(),
		"task":     data.TaskName.ValueString(),
	})

	// Build API model
	apiModel := &WorkflowVertexAPIModel{
		Task: &TaskRef{
			Value: data.TaskName.ValueString(),
		},
		Alias:   data.Alias.ValueString(),
		VertexX: data.VertexX.ValueString(),
		VertexY: data.VertexY.ValueString(),
	}

	// Add the vertex to the workflow
	query := url.Values{}
	query.Set("workflowname", data.WorkflowName.ValueString())

	respBody, err := r.client.Post(ctx, "/resources/workflow/vertices?"+query.Encode(), apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Task to Workflow",
			fmt.Sprintf("Could not add task %s to workflow %s: %s",
				data.TaskName.ValueString(), data.WorkflowName.ValueString(), err),
		)
		return
	}

	// Parse response to get the vertexId
	var respModel WorkflowVertexResponseModel
	if err := json.Unmarshal(respBody, &respModel); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Response",
			fmt.Sprintf("Could not parse vertex response: %s", err),
		)
		return
	}

	// Update model with response data
	data.VertexId = types.StringValue(respModel.VertexId)
	if respModel.VertexX != "" {
		data.VertexX = types.StringValue(respModel.VertexX)
	}
	if respModel.VertexY != "" {
		data.VertexY = types.StringValue(respModel.VertexY)
	}

	tflog.Debug(ctx, "Added task to workflow", map[string]any{
		"vertex_id": data.VertexId.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowVertexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowVertexResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query for the vertex
	query := url.Values{}
	query.Set("workflowname", data.WorkflowName.ValueString())
	query.Set("vertexid", data.VertexId.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/workflow/vertices", query)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Workflow vertex not found, removing from state", map[string]any{
				"vertex_id": data.VertexId.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Workflow Vertex",
			fmt.Sprintf("Could not read vertex %s: %s", data.VertexId.ValueString(), err),
		)
		return
	}

	// Parse response - API returns an array
	var vertices []WorkflowVertexResponseModel
	if err := json.Unmarshal(respBody, &vertices); err != nil {
		// Try single object
		var vertex WorkflowVertexResponseModel
		if err := json.Unmarshal(respBody, &vertex); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Response",
				fmt.Sprintf("Could not parse vertex response: %s", err),
			)
			return
		}
		vertices = []WorkflowVertexResponseModel{vertex}
	}

	if len(vertices) == 0 {
		tflog.Debug(ctx, "Workflow vertex not found, removing from state", map[string]any{
			"vertex_id": data.VertexId.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model with response data
	vertex := vertices[0]
	if vertex.Task != nil && vertex.Task.Value != "" {
		data.TaskName = types.StringValue(vertex.Task.Value)
	}
	data.Alias = StringValueOrNull(vertex.Alias)
	if vertex.VertexX != "" {
		data.VertexX = types.StringValue(vertex.VertexX)
	}
	if vertex.VertexY != "" {
		data.VertexY = types.StringValue(vertex.VertexY)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowVertexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowVertexResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for vertexId
	var state WorkflowVertexResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.VertexId = state.VertexId

	tflog.Debug(ctx, "Updating workflow vertex", map[string]any{
		"vertex_id": data.VertexId.ValueString(),
	})

	// Build API model for update
	apiModel := &WorkflowVertexAPIModel{
		VertexId: data.VertexId.ValueString(),
		Alias:    data.Alias.ValueString(),
		VertexX:  data.VertexX.ValueString(),
		VertexY:  data.VertexY.ValueString(),
	}

	query := url.Values{}
	query.Set("workflowname", data.WorkflowName.ValueString())

	_, err := r.client.Put(ctx, "/resources/workflow/vertices?"+query.Encode(), apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Workflow Vertex",
			fmt.Sprintf("Could not update vertex %s: %s", data.VertexId.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowVertexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowVertexResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Removing task from workflow", map[string]any{
		"vertex_id": data.VertexId.ValueString(),
		"workflow":  data.WorkflowName.ValueString(),
	})

	query := url.Values{}
	query.Set("workflowname", data.WorkflowName.ValueString())
	query.Set("vertexid", data.VertexId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/workflow/vertices", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Removing Task from Workflow",
			fmt.Sprintf("Could not remove vertex %s from workflow %s: %s",
				data.VertexId.ValueString(), data.WorkflowName.ValueString(), err),
		)
		return
	}
}

// ImportState imports an existing workflow vertex using an ID of the form
// "workflow_name/vertex_id". The vertex ID (assigned by UAC) disambiguates
// cases where the same task appears more than once in a workflow's graph.
func (r *WorkflowVertexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf(`Expected import ID in the form "workflow_name/vertex_id", got: %q.`, req.ID),
		)
		return
	}
	workflowName, vertexId := parts[0], parts[1]

	query := url.Values{}
	query.Set("workflowname", workflowName)
	query.Set("vertexid", vertexId)

	respBody, err := r.client.Get(ctx, "/resources/workflow/vertices", query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Workflow Vertex",
			fmt.Sprintf("Could not read vertex %s in workflow %s: %s", vertexId, workflowName, err),
		)
		return
	}

	var vertices []WorkflowVertexResponseModel
	if err := json.Unmarshal(respBody, &vertices); err != nil {
		var vertex WorkflowVertexResponseModel
		if err := json.Unmarshal(respBody, &vertex); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Response",
				fmt.Sprintf("Could not parse vertex response: %s", err),
			)
			return
		}
		vertices = []WorkflowVertexResponseModel{vertex}
	}

	if len(vertices) == 0 || vertices[0].Task == nil {
		resp.Diagnostics.AddError(
			"Vertex Not Found",
			fmt.Sprintf("No vertex %s found in workflow %q.", vertexId, workflowName),
		)
		return
	}

	vertex := vertices[0]
	data := WorkflowVertexResourceModel{
		WorkflowName: types.StringValue(workflowName),
		TaskName:     types.StringValue(vertex.Task.Value),
		VertexId:     types.StringValue(vertexId),
		Alias:        StringValueOrNull(vertex.Alias),
		VertexX:      StringValueOrNull(vertex.VertexX),
		VertexY:      StringValueOrNull(vertex.VertexY),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
