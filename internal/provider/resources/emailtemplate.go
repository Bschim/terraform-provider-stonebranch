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
	_ resource.Resource                   = &EmailTemplateResource{}
	_ resource.ResourceWithImportState    = &EmailTemplateResource{}
	_ resource.ResourceWithValidateConfig = &EmailTemplateResource{}
)

func NewEmailTemplateResource() resource.Resource {
	return &EmailTemplateResource{}
}

// EmailTemplateResource defines the resource implementation.
type EmailTemplateResource struct {
	client *client.Client
}

// EmailTemplateResourceModel describes the resource data model.
type EmailTemplateResourceModel struct {
	// Identity
	SysId   types.String `tfsdk:"sys_id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`

	// Content
	Description types.String `tfsdk:"description"`
	Connection  types.String `tfsdk:"email_connection"`
	ReplyTo     types.String `tfsdk:"reply_to"`
	To          types.String `tfsdk:"to"`
	Cc          types.String `tfsdk:"cc"`
	Bcc         types.String `tfsdk:"bcc"`
	Subject     types.String `tfsdk:"subject"`
	Body        types.String `tfsdk:"body"`

	// Business services
	OpswiseGroups types.List `tfsdk:"opswise_groups"`
}

// EmailTemplateAPIModel represents the API request/response structure.
type EmailTemplateAPIModel struct {
	SysId         string   `json:"sysId,omitempty"`
	TemplateName  string   `json:"templateName"`
	Version       int64    `json:"version,omitempty"`
	Description   string   `json:"description,omitempty"`
	Connection    string   `json:"connection,omitempty"`
	ReplyTo       string   `json:"replyTo,omitempty"`
	To            string   `json:"to,omitempty"`
	Cc            string   `json:"cc,omitempty"`
	Bcc           string   `json:"bcc,omitempty"`
	Subject       string   `json:"subject,omitempty"`
	Body          string   `json:"body,omitempty"`
	OpswiseGroups []string `json:"opswiseGroups,omitempty"`
}

func (r *EmailTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template"
}

func (r *EmailTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a StoneBranch Email Template. Email templates define reusable " +
			"subject/body content and recipients for email notifications sent via a `stonebranch_email_connection`. " +
			"At least one of `to`, `cc`, or `bcc` must be set.",

		Attributes: map[string]schema.Attribute{
			// Identity
			"sys_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the email template (assigned by StoneBranch).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name of the email template.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number of the email template (for optimistic locking).",
				Computed:            true,
			},

			// Content
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the email template.",
				Optional:            true,
			},
			"email_connection": schema.StringAttribute{
				MarkdownDescription: "Name of the `stonebranch_email_connection` used to send emails from this template.",
				Required:            true,
			},
			"reply_to": schema.StringAttribute{
				MarkdownDescription: "Reply-To address for emails sent from this template.",
				Optional:            true,
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of To recipients. At least one of `to`, `cc`, or `bcc` must be set.",
				Optional:            true,
			},
			"cc": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of Cc recipients. At least one of `to`, `cc`, or `bcc` must be set.",
				Optional:            true,
			},
			"bcc": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of Bcc recipients. At least one of `to`, `cc`, or `bcc` must be set.",
				Optional:            true,
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "Subject line of the email.",
				Optional:            true,
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "Body content of the email.",
				Optional:            true,
			},

			// Business services
			"opswise_groups": schema.ListAttribute{
				MarkdownDescription: "List of business service names this email template belongs to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *EmailTemplateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data EmailTemplateResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isSet(data.To) || isSet(data.Cc) || isSet(data.Bcc) {
		return
	}
	// If any of the three are unknown (e.g. computed from another resource),
	// defer this check to the server rather than failing the plan early.
	if data.To.IsUnknown() || data.Cc.IsUnknown() || data.Bcc.IsUnknown() {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("to"),
		"Missing Required Field",
		`At least one of "to", "cc", or "bcc" must be set.`,
	)
}

func (r *EmailTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EmailTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EmailTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating email template", map[string]any{"name": data.Name.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Create the email template
	_, err := r.client.Post(ctx, "/resources/emailtemplate", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Email Template",
			fmt.Sprintf("Could not create email template %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back the created email template to get sysId and other computed fields
	err = r.readEmailTemplate(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Email Template",
			fmt.Sprintf("Could not read email template %s after creation: %s", data.Name.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Created email template", map[string]any{"sys_id": data.SysId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EmailTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EmailTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readEmailTemplate(ctx, &data)
	if err != nil {
		// Check if email template was deleted outside of Terraform
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Email template not found, removing from state", map[string]any{"name": data.Name.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Email Template",
			fmt.Sprintf("Could not read email template %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EmailTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EmailTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for sysId
	var state EmailTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sysId from state
	data.SysId = state.SysId

	tflog.Debug(ctx, "Updating email template", map[string]any{"sys_id": data.SysId.ValueString()})

	// Build API model
	apiModel := r.toAPIModel(ctx, &data)

	// Update the email template
	_, err := r.client.Put(ctx, "/resources/emailtemplate", apiModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Email Template",
			fmt.Sprintf("Could not update email template %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get updated version
	err = r.readEmailTemplate(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Email Template",
			fmt.Sprintf("Could not read email template %s after update: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EmailTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EmailTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting email template", map[string]any{"sys_id": data.SysId.ValueString()})

	query := url.Values{}
	query.Set("templateid", data.SysId.ValueString())

	_, err := r.client.Delete(ctx, "/resources/emailtemplate", query)
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Email Template",
			fmt.Sprintf("Could not delete email template %s: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *EmailTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readEmailTemplate fetches the email template from the API and updates the model.
func (r *EmailTemplateResource) readEmailTemplate(ctx context.Context, data *EmailTemplateResourceModel) error {
	query := url.Values{}
	query.Set("templatename", data.Name.ValueString())

	respBody, err := r.client.Get(ctx, "/resources/emailtemplate", query)
	if err != nil {
		return err
	}

	var apiModel EmailTemplateAPIModel
	if err := json.Unmarshal(respBody, &apiModel); err != nil {
		return fmt.Errorf("failed to parse email template response: %w", err)
	}

	r.fromAPIModel(ctx, &apiModel, data)
	return nil
}

// toAPIModel converts the Terraform model to an API model.
func (r *EmailTemplateResource) toAPIModel(ctx context.Context, data *EmailTemplateResourceModel) *EmailTemplateAPIModel {
	model := &EmailTemplateAPIModel{
		SysId:        data.SysId.ValueString(),
		TemplateName: data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		Connection:   data.Connection.ValueString(),
		ReplyTo:      data.ReplyTo.ValueString(),
		To:           data.To.ValueString(),
		Cc:           data.Cc.ValueString(),
		Bcc:          data.Bcc.ValueString(),
		Subject:      data.Subject.ValueString(),
		Body:         data.Body.ValueString(),
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
func (r *EmailTemplateResource) fromAPIModel(ctx context.Context, apiModel *EmailTemplateAPIModel, data *EmailTemplateResourceModel) {
	// Identity fields - always set
	data.SysId = types.StringValue(apiModel.SysId)
	data.Name = types.StringValue(apiModel.TemplateName)
	data.Version = types.Int64Value(apiModel.Version)

	// Content
	data.Description = StringValueOrNull(apiModel.Description)
	data.Connection = StringValueOrNull(apiModel.Connection)
	data.ReplyTo = StringValueOrNull(apiModel.ReplyTo)
	data.To = StringValueOrNull(apiModel.To)
	data.Cc = StringValueOrNull(apiModel.Cc)
	data.Bcc = StringValueOrNull(apiModel.Bcc)
	data.Subject = StringValueOrNull(apiModel.Subject)
	data.Body = StringValueOrNull(apiModel.Body)

	// Handle opswise_groups
	if len(apiModel.OpswiseGroups) > 0 {
		groups, _ := types.ListValueFrom(ctx, types.StringType, apiModel.OpswiseGroups)
		data.OpswiseGroups = groups
	}
}
