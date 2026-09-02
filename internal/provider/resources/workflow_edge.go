package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &WorkflowEdgeResource{}
	_ resource.ResourceWithValidateConfig = &WorkflowEdgeResource{}
)

// Known enum values for the edge condition "type" and "status" fields.
var (
	validEdgeConditionTypes    = []string{"Status", "Exit Code", "Variable"}
	validEdgeConditionStatuses = []string{"Success", "Failure", "Success/Failure"}
)

func NewWorkflowEdgeResource() resource.Resource {
	return &WorkflowEdgeResource{}
}

// WorkflowEdgeResource defines the resource implementation.
type WorkflowEdgeResource struct {
	client *client.Client
}

// WorkflowEdgeResourceModel describes the resource data model.
type WorkflowEdgeResourceModel struct {
	// Identity - composite key
	WorkflowName types.String `tfsdk:"workflow_name"`
	SourceId     types.String `tfsdk:"source_id"` // vertex ID of source task
	TargetId     types.String `tfsdk:"target_id"` // vertex ID of target task

	// Optional
	StraightEdge types.Bool   `tfsdk:"straight_edge"`
	Condition    types.Object `tfsdk:"condition"`
}

// EdgeConditionModel describes the branch condition for an edge in Terraform.
// Exactly one shape applies, selected by Type. See ValidateConfig.
type EdgeConditionModel struct {
	// Type: "Status" (default), "Exit Code", or "Variable".
	Type types.String `tfsdk:"type"`

	// Type = "Status"
	Status types.String `tfsdk:"status"`

	// Type = "Exit Code"
	ExitCode types.String `tfsdk:"exit_code"`

	// Type = "Variable"
	FirstValue  types.String `tfsdk:"first_value"`
	Operator    types.String `tfsdk:"operator"`
	SecondValue types.String `tfsdk:"second_value"`
}

// EdgeConditionAttrTypes returns the attribute types for EdgeConditionModel.
func EdgeConditionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":         types.StringType,
		"status":       types.StringType,
		"exit_code":    types.StringType,
		"first_value":  types.StringType,
		"operator":     types.StringType,
		"second_value": types.StringType,
	}
}

// WorkflowEdgeAPIModel represents the API request structure for creating an edge.
type WorkflowEdgeAPIModel struct {
	SourceId     *EdgeVertexRef         `json:"sourceId,omitempty"`
	TargetId     *EdgeVertexRef         `json:"targetId,omitempty"`
	StraightEdge bool                   `json:"straightEdge,omitempty"`
	Condition    *EdgeConditionAPIModel `json:"condition,omitempty"`
}

// EdgeVertexRef represents a vertex reference for edge endpoints.
type EdgeVertexRef struct {
	Value string `json:"value,omitempty"` // vertexId
}

// EdgeConditionAPIModel represents an edge branch condition in the API.
// The three shapes are distinguished by which fields are populated:
//   - Status:   {"value": "Success"}                 (Type is empty)
//   - ExitCode: {"type": "Exit Code", "value": "0"}
//   - Variable: {"type": "Variable", "firstValue": "...", "operator": "...", "secondValue": "..."}
type EdgeConditionAPIModel struct {
	Type        string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
	FirstValue  string `json:"firstValue,omitempty"`
	Operator    string `json:"operator,omitempty"`
	SecondValue string `json:"secondValue,omitempty"`
}

// WorkflowEdgeResponseModel represents the API response structure.
type WorkflowEdgeResponseModel struct {
	SysId        string                 `json:"sysId,omitempty"`
	SourceId     *EdgeVertexRefResp     `json:"sourceId,omitempty"`
	TargetId     *EdgeVertexRefResp     `json:"targetId,omitempty"`
	StraightEdge bool                   `json:"straightEdge,omitempty"`
	Condition    *EdgeConditionAPIModel `json:"condition,omitempty"`
}

type EdgeVertexRefResp struct {
	TaskName  string `json:"taskName,omitempty"`
	TaskAlias string `json:"taskAlias,omitempty"`
	Value     string `json:"value,omitempty"`
}

func (r *WorkflowEdgeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_edge"
}

func (r *WorkflowEdgeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a dependency (edge) between tasks within a StoneBranch Workflow. Use this resource to define execution order between workflow vertices.",

		Attributes: map[string]schema.Attribute{
			"workflow_name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow containing the dependency.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_id": schema.StringAttribute{
				MarkdownDescription: "Vertex ID of the source task (the predecessor). This task must complete before the target task runs.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": schema.StringAttribute{
				MarkdownDescription: "Vertex ID of the target task (the successor). This task waits for the source task to complete.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"straight_edge": schema.BoolAttribute{
				MarkdownDescription: "Whether to draw the edge as a straight line in the workflow diagram.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"condition": schema.SingleNestedAttribute{
				MarkdownDescription: "Branch condition controlling when the target task runs. Exactly one shape applies, selected by `type`: 'Status' (default, requires `status`), 'Exit Code' (requires `exit_code`), or 'Variable' (requires `first_value`, `operator`, `second_value`). If omitted entirely, UAC defaults to a Status condition of 'Success'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Condition shape. One of: 'Status' (default), 'Exit Code', 'Variable'.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"status": schema.StringAttribute{
						MarkdownDescription: "Required when `type` is 'Status'. One of: 'Success', 'Failure', 'Success/Failure'.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"exit_code": schema.StringAttribute{
						MarkdownDescription: "Required when `type` is 'Exit Code'. The exit code value to match.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"first_value": schema.StringAttribute{
						MarkdownDescription: "Required when `type` is 'Variable'. Left-hand side of the comparison (typically a `${variable}` reference).",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"operator": schema.StringAttribute{
						MarkdownDescription: "Required when `type` is 'Variable'. Comparison operator (e.g. '=', '!=').",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"second_value": schema.StringAttribute{
						MarkdownDescription: "Required when `type` is 'Variable'. Right-hand side of the comparison.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

// ValidateConfig enforces the mutually-exclusive field groups implied by
// condition.type, since the UAC API expects exactly one shape and would
// otherwise fail (or silently misbehave) with a generic error.
func (r *WorkflowEdgeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data WorkflowEdgeResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Condition.IsNull() || data.Condition.IsUnknown() {
		return
	}

	var cond EdgeConditionModel
	resp.Diagnostics.Append(data.Condition.As(ctx, &cond, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	conditionType := "Status"
	if isSet(cond.Type) {
		conditionType = cond.Type.ValueString()
		if !contains(validEdgeConditionTypes, conditionType) {
			resp.Diagnostics.AddAttributeError(
				path.Root("condition").AtName("type"),
				"Invalid Attribute Value",
				fmt.Sprintf("condition.type must be one of %v, got: %q.", validEdgeConditionTypes, conditionType),
			)
			return
		}
	}

	hasStatus := isSet(cond.Status)
	hasExitCode := isSet(cond.ExitCode)
	hasFirstValue := isSet(cond.FirstValue)
	hasOperator := isSet(cond.Operator)
	hasSecondValue := isSet(cond.SecondValue)

	conflict := func(field string, msg string) {
		resp.Diagnostics.AddAttributeError(path.Root("condition").AtName(field), "Conflicting Fields", msg)
	}
	missing := func(field string, msg string) {
		resp.Diagnostics.AddAttributeError(path.Root("condition").AtName(field), "Missing Required Field", msg)
	}

	switch conditionType {
	case "Status":
		if !hasStatus {
			missing("status", `"condition.status" is required when condition.type is "Status".`)
		} else if !contains(validEdgeConditionStatuses, cond.Status.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("condition").AtName("status"),
				"Invalid Attribute Value",
				fmt.Sprintf("condition.status must be one of %v, got: %q.", validEdgeConditionStatuses, cond.Status.ValueString()),
			)
		}
		if hasExitCode || hasFirstValue || hasOperator || hasSecondValue {
			conflict("type", `"condition.exit_code", "condition.first_value", "condition.operator", and "condition.second_value" must not be set when condition.type is "Status".`)
		}
	case "Exit Code":
		if !hasExitCode {
			missing("exit_code", `"condition.exit_code" is required when condition.type is "Exit Code".`)
		}
		if hasStatus || hasFirstValue || hasOperator || hasSecondValue {
			conflict("type", `"condition.status", "condition.first_value", "condition.operator", and "condition.second_value" must not be set when condition.type is "Exit Code".`)
		}
	case "Variable":
		if !hasFirstValue || !hasOperator || !hasSecondValue {
			missing("type", `"condition.first_value", "condition.operator", and "condition.second_value" are all required when condition.type is "Variable".`)
		}
		if hasStatus || hasExitCode {
			conflict("type", `"condition.status" and "condition.exit_code" must not be set when condition.type is "Variable".`)
		}
	}
}

// conditionToAPI converts a Terraform condition object to the API model.
// Returns nil (with no diagnostics) if the object is null or unknown, i.e.
// not set by the user - in that case UAC applies its own default (Success).
func conditionToAPI(ctx context.Context, obj types.Object) (*EdgeConditionAPIModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	var cond EdgeConditionModel
	diags.Append(obj.As(ctx, &cond, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	conditionType := "Status"
	if isSet(cond.Type) {
		conditionType = cond.Type.ValueString()
	}

	api := &EdgeConditionAPIModel{}
	switch conditionType {
	case "Exit Code":
		api.Type = "Exit Code"
		api.Value = cond.ExitCode.ValueString()
	case "Variable":
		api.Type = "Variable"
		api.FirstValue = cond.FirstValue.ValueString()
		api.Operator = cond.Operator.ValueString()
		api.SecondValue = cond.SecondValue.ValueString()
	default: // "Status"
		api.Value = cond.Status.ValueString()
	}
	return api, diags
}

// conditionFromAPI converts an API condition model to a Terraform object.
func conditionFromAPI(cond *EdgeConditionAPIModel) types.Object {
	if cond == nil {
		return types.ObjectNull(EdgeConditionAttrTypes())
	}

	conditionType := cond.Type
	if conditionType == "" {
		conditionType = "Status"
	}

	values := map[string]attr.Value{
		"type":         types.StringValue(conditionType),
		"status":       types.StringNull(),
		"exit_code":    types.StringNull(),
		"first_value":  types.StringNull(),
		"operator":     types.StringNull(),
		"second_value": types.StringNull(),
	}

	switch conditionType {
	case "Exit Code":
		values["exit_code"] = StringValueOrNull(cond.Value)
	case "Variable":
		values["first_value"] = StringValueOrNull(cond.FirstValue)
		values["operator"] = StringValueOrNull(cond.Operator)
		values["second_value"] = StringValueOrNull(cond.SecondValue)
	default:
		values["status"] = StringValueOrNull(cond.Value)
	}

	obj, _ := types.ObjectValue(EdgeConditionAttrTypes(), values)
	return obj
}

// findEdge looks up a single edge within a workflow by source/target vertex ID.
// Returns (nil, nil) if the workflow exists but no matching edge is found.
func (r *WorkflowEdgeResource) findEdge(ctx context.Context, workflowName, sourceId, targetId string) (*WorkflowEdgeResponseModel, error) {
	query := url.Values{}
	query.Set("workflowname", workflowName)

	respBody, err := r.client.Get(ctx, "/resources/workflow/edges", query)
	if err != nil {
		return nil, err
	}

	var edges []WorkflowEdgeResponseModel
	if err := json.Unmarshal(respBody, &edges); err != nil {
		return nil, fmt.Errorf("failed to parse edges response: %w", err)
	}

	for i := range edges {
		edge := &edges[i]
		if edge.SourceId != nil && edge.TargetId != nil &&
			edge.SourceId.Value == sourceId && edge.TargetId.Value == targetId {
			return edge, nil
		}
	}
	return nil, nil
}

func (r *WorkflowEdgeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowEdgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowEdgeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating workflow edge", map[string]any{
		"workflow":  data.WorkflowName.ValueString(),
		"source_id": data.SourceId.ValueString(),
		"target_id": data.TargetId.ValueString(),
	})

	// Build API model
	condAPI, condDiags := conditionToAPI(ctx, data.Condition)
	resp.Diagnostics.Append(condDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := &WorkflowEdgeAPIModel{
		SourceId: &EdgeVertexRef{
			Value: data.SourceId.ValueString(),
		},
		TargetId: &EdgeVertexRef{
			Value: data.TargetId.ValueString(),
		},
		StraightEdge: data.StraightEdge.ValueBool(),
		Condition:    condAPI,
	}

	// Add the edge to the workflow
	query := url.Values{}
	query.Set("workflowname", data.WorkflowName.ValueString())

	_, err := r.client.Post(ctx, "/resources/workflow/edges?"+query.Encode(), apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Workflow Edge",
			fmt.Sprintf("Could not create edge from %s to %s in workflow %s: %s",
				data.SourceId.ValueString(), data.TargetId.ValueString(), data.WorkflowName.ValueString(), err),
		)
		return
	}

	// Read back the created edge to populate computed fields (e.g. the
	// resolved `condition`, which UAC defaults to Success when unset).
	edge, err := r.findEdge(ctx, data.WorkflowName.ValueString(), data.SourceId.ValueString(), data.TargetId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Workflow Edge",
			fmt.Sprintf("Could not read edge from %s to %s after creation: %s",
				data.SourceId.ValueString(), data.TargetId.ValueString(), err),
		)
		return
	}
	if edge == nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Workflow Edge",
			fmt.Sprintf("Edge from %s to %s not found immediately after creation.",
				data.SourceId.ValueString(), data.TargetId.ValueString()),
		)
		return
	}
	data.StraightEdge = types.BoolValue(edge.StraightEdge)
	data.Condition = conditionFromAPI(edge.Condition)

	tflog.Debug(ctx, "Created workflow edge", map[string]any{
		"source_id": data.SourceId.ValueString(),
		"target_id": data.TargetId.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowEdgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowEdgeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Look up the specific edge matching source and target
	edge, err := r.findEdge(ctx, data.WorkflowName.ValueString(), data.SourceId.ValueString(), data.TargetId.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Workflow not found, removing edge from state", map[string]any{
				"workflow": data.WorkflowName.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Workflow Edges",
			fmt.Sprintf("Could not read edges for workflow %s: %s", data.WorkflowName.ValueString(), err),
		)
		return
	}

	if edge == nil {
		tflog.Debug(ctx, "Workflow edge not found, removing from state", map[string]any{
			"source_id": data.SourceId.ValueString(),
			"target_id": data.TargetId.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	data.StraightEdge = types.BoolValue(edge.StraightEdge)
	data.Condition = conditionFromAPI(edge.Condition)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowEdgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowEdgeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating workflow edge", map[string]any{
		"source_id": data.SourceId.ValueString(),
		"target_id": data.TargetId.ValueString(),
	})

	// Build API model for update
	condAPI, condDiags := conditionToAPI(ctx, data.Condition)
	resp.Diagnostics.Append(condDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := &WorkflowEdgeAPIModel{
		SourceId: &EdgeVertexRef{
			Value: data.SourceId.ValueString(),
		},
		TargetId: &EdgeVertexRef{
			Value: data.TargetId.ValueString(),
		},
		StraightEdge: data.StraightEdge.ValueBool(),
		Condition:    condAPI,
	}

	query := url.Values{}
	query.Set("workflowname", data.WorkflowName.ValueString())

	_, err := r.client.Put(ctx, "/resources/workflow/edges?"+query.Encode(), apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Workflow Edge",
			fmt.Sprintf("Could not update edge from %s to %s: %s",
				data.SourceId.ValueString(), data.TargetId.ValueString(), err),
		)
		return
	}

	// Read back the updated edge to populate computed fields.
	edge, err := r.findEdge(ctx, data.WorkflowName.ValueString(), data.SourceId.ValueString(), data.TargetId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Workflow Edge",
			fmt.Sprintf("Could not read edge from %s to %s after update: %s",
				data.SourceId.ValueString(), data.TargetId.ValueString(), err),
		)
		return
	}
	if edge == nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Workflow Edge",
			fmt.Sprintf("Edge from %s to %s not found immediately after update.",
				data.SourceId.ValueString(), data.TargetId.ValueString()),
		)
		return
	}
	data.StraightEdge = types.BoolValue(edge.StraightEdge)
	data.Condition = conditionFromAPI(edge.Condition)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowEdgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowEdgeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Removing workflow edge", map[string]any{
		"workflow":  data.WorkflowName.ValueString(),
		"source_id": data.SourceId.ValueString(),
		"target_id": data.TargetId.ValueString(),
	})

	query := url.Values{}
	query.Set("workflowname", data.WorkflowName.ValueString())
	query.Set("sourceid", data.SourceId.ValueString())
	query.Set("targetid", data.TargetId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/workflow/edges", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Removing Workflow Edge",
			fmt.Sprintf("Could not remove edge from %s to %s in workflow %s: %s",
				data.SourceId.ValueString(), data.TargetId.ValueString(), data.WorkflowName.ValueString(), err),
		)
		return
	}
}
