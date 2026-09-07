package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ---------------------------------------------------------------------------
// Shared: vertex reference (used by systemOperations.vertices)
// ---------------------------------------------------------------------------

// VertexRefModel describes a workflow vertex reference in Terraform.
type VertexRefModel struct {
	TaskName   types.String `tfsdk:"task_name"`
	VertexName types.String `tfsdk:"vertex_name"`
	VertexId   types.String `tfsdk:"vertex_id"`
}

// VertexRefAPIModel represents a workflow vertex reference in the API.
type VertexRefAPIModel struct {
	TaskName   string `json:"taskName,omitempty"`
	VertexName string `json:"vertexName,omitempty"`
	VertexId   string `json:"vertexId,omitempty"`
}

func vertexRefAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"task_name":   types.StringType,
		"vertex_name": types.StringType,
		"vertex_id":   types.StringType,
	}
}

func vertexRefsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Specific workflow vertices to target when launching a workflow task (used when `vertex_selection` is enabled).",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"task_name": schema.StringAttribute{
					MarkdownDescription: "Name of the task at this vertex.",
					Optional:            true,
				},
				"vertex_name": schema.StringAttribute{
					MarkdownDescription: "Name of the vertex.",
					Optional:            true,
				},
				"vertex_id": schema.StringAttribute{
					MarkdownDescription: "ID of the vertex.",
					Optional:            true,
				},
			},
		},
	}
}

func vertexRefsToAPI(ctx context.Context, list types.List) []VertexRefAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var models []VertexRefModel
	list.ElementsAs(ctx, &models, false)

	result := make([]VertexRefAPIModel, len(models))
	for i, m := range models {
		result[i] = VertexRefAPIModel{
			TaskName:   m.TaskName.ValueString(),
			VertexName: m.VertexName.ValueString(),
			VertexId:   m.VertexId.ValueString(),
		}
	}
	return result
}

func vertexRefsFromAPI(api []VertexRefAPIModel) types.List {
	if len(api) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: vertexRefAttrTypes()})
	}
	values := make([]attr.Value, len(api))
	for i, v := range api {
		values[i], _ = types.ObjectValue(vertexRefAttrTypes(), map[string]attr.Value{
			"task_name":   StringValueOrNull(v.TaskName),
			"vertex_name": StringValueOrNull(v.VertexName),
			"vertex_id":   StringValueOrNull(v.VertexId),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: vertexRefAttrTypes()}, values)
	return result
}

// ---------------------------------------------------------------------------
// Shared: report reference (used by emailNotifications.report)
// ---------------------------------------------------------------------------

// ReportRefModel describes a report reference in Terraform.
type ReportRefModel struct {
	Title      types.String `tfsdk:"title"`
	UserName   types.String `tfsdk:"user_name"`
	GroupName  types.String `tfsdk:"group_name"`
	GroupNames types.List   `tfsdk:"group_names"`
}

// ReportRefAPIModel represents a report reference in the API.
type ReportRefAPIModel struct {
	Title      string   `json:"title,omitempty"`
	UserName   string   `json:"userName,omitempty"`
	GroupName  string   `json:"groupName,omitempty"`
	GroupNames []string `json:"groupNames,omitempty"`
}

func reportRefAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"title":       types.StringType,
		"user_name":   types.StringType,
		"group_name":  types.StringType,
		"group_names": types.ListType{ElemType: types.StringType},
	}
}

func reportRefSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Report to attach to the email notification.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the report.",
				Optional:            true,
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "User name the report is scoped to.",
				Optional:            true,
			},
			"group_name": schema.StringAttribute{
				MarkdownDescription: "Group name the report is scoped to.",
				Optional:            true,
			},
			"group_names": schema.ListAttribute{
				MarkdownDescription: "Group names the report is scoped to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func reportRefToAPI(ctx context.Context, obj types.Object) *ReportRefAPIModel {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var m ReportRefModel
	obj.As(ctx, &m, basetypes.ObjectAsOptions{})

	var groupNames []string
	if !m.GroupNames.IsNull() && !m.GroupNames.IsUnknown() {
		m.GroupNames.ElementsAs(ctx, &groupNames, false)
	}

	return &ReportRefAPIModel{
		Title:      m.Title.ValueString(),
		UserName:   m.UserName.ValueString(),
		GroupName:  m.GroupName.ValueString(),
		GroupNames: groupNames,
	}
}

func reportRefFromAPI(ctx context.Context, api *ReportRefAPIModel) types.Object {
	if api == nil {
		return types.ObjectNull(reportRefAttrTypes())
	}
	groupNames := types.ListNull(types.StringType)
	if len(api.GroupNames) > 0 {
		groupNames, _ = types.ListValueFrom(ctx, types.StringType, api.GroupNames)
	}
	obj, _ := types.ObjectValue(reportRefAttrTypes(), map[string]attr.Value{
		"title":       StringValueOrNull(api.Title),
		"user_name":   StringValueOrNull(api.UserName),
		"group_name":  StringValueOrNull(api.GroupName),
		"group_names": groupNames,
	})
	return obj
}

// ---------------------------------------------------------------------------
// Shared: common fields present on every action type (abort, email, snmp,
// system-operation, set-variable). Mirrors the common properties of
// AbortActionWsData / EmailNotificationWsData / SnmpNotificationWsData /
// SystemOperationNotificationWsData / VariableActionWsData in the UAC API.
// ---------------------------------------------------------------------------

func commonActionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":                   types.StringType,
		"notify_on_late_start":     types.BoolType,
		"notify_on_late_finish":    types.BoolType,
		"notify_on_early_finish":   types.BoolType,
		"notify_on_projected_late": types.BoolType,
		"exit_codes":               types.StringType,
		"description":              types.StringType,
		"inheritance":              types.StringType,
		"sys_id":                   types.StringType,
	}
}

func commonActionSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"status": schema.StringAttribute{
			MarkdownDescription: "Status(es) that trigger this action. Comma/newline-separated list of statuses (e.g. 'Success', 'Failed', 'Running/Problems').",
			Optional:            true,
			Computed:            true,
		},
		"notify_on_late_start": schema.BoolAttribute{
			MarkdownDescription: "Whether to trigger this action when the task starts late.",
			Optional:            true,
			Computed:            true,
		},
		"notify_on_late_finish": schema.BoolAttribute{
			MarkdownDescription: "Whether to trigger this action when the task finishes late.",
			Optional:            true,
			Computed:            true,
		},
		"notify_on_early_finish": schema.BoolAttribute{
			MarkdownDescription: "Whether to trigger this action when the task finishes early.",
			Optional:            true,
			Computed:            true,
		},
		"notify_on_projected_late": schema.BoolAttribute{
			MarkdownDescription: "Whether to trigger this action when the task is projected to be late.",
			Optional:            true,
			Computed:            true,
		},
		"exit_codes": schema.StringAttribute{
			MarkdownDescription: "Comma-separated list of exit codes that trigger this action.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of this action.",
			Optional:            true,
		},
		"inheritance": schema.StringAttribute{
			MarkdownDescription: "Inheritance setting for this action (e.g. 'Children').",
			Optional:            true,
		},
		"sys_id": schema.StringAttribute{
			MarkdownDescription: "System ID of this action (assigned by StoneBranch).",
			Computed:            true,
		},
	}
}

// commonActionModel holds the fields shared by every action type.
// commonActionAPIModel holds the API fields shared by every action type.
// It is embedded (safe for encoding/json, which flattens anonymous struct
// fields) by each concrete *AttrTypes/API model below. It is intentionally
// not mirrored by an equivalent embedded Terraform model struct, since the
// framework's reflection-based ElementsAs/As conversion does not flatten
// embedded struct fields tagged with `tfsdk`.
type commonActionAPIModel struct {
	Status                string `json:"status,omitempty"`
	NotifyOnLateStart     bool   `json:"notifyOnLateStart,omitempty"`
	NotifyOnLateFinish    bool   `json:"notifyOnLateFinish,omitempty"`
	NotifyOnEarlyFinish   bool   `json:"notifyOnEarlyFinish,omitempty"`
	NotifyOnProjectedLate bool   `json:"notifyOnProjectedLate,omitempty"`
	ExitCodes             string `json:"exitCodes,omitempty"`
	Description           string `json:"description,omitempty"`
	Inheritance           string `json:"inheritance,omitempty"`
	SysId                 string `json:"sysId,omitempty"`
}

// ---------------------------------------------------------------------------
// AbortAction
// ---------------------------------------------------------------------------

// AbortActionModel describes an abort action in Terraform.
type AbortActionModel struct {
	Status                types.String `tfsdk:"status"`
	NotifyOnLateStart     types.Bool   `tfsdk:"notify_on_late_start"`
	NotifyOnLateFinish    types.Bool   `tfsdk:"notify_on_late_finish"`
	NotifyOnEarlyFinish   types.Bool   `tfsdk:"notify_on_early_finish"`
	NotifyOnProjectedLate types.Bool   `tfsdk:"notify_on_projected_late"`
	ExitCodes             types.String `tfsdk:"exit_codes"`
	Description           types.String `tfsdk:"description"`
	Inheritance           types.String `tfsdk:"inheritance"`
	SysId                 types.String `tfsdk:"sys_id"`
	CancelProcess         types.Bool   `tfsdk:"cancel_process"`
	OverrideExitCode      types.String `tfsdk:"override_exit_code"`
	HaltOnFinish          types.Bool   `tfsdk:"halt_on_finish"`
}

// AbortActionAPIModel represents an abort action in the API.
type AbortActionAPIModel struct {
	commonActionAPIModel
	CancelProcess    bool   `json:"cancelProcess,omitempty"`
	OverrideExitCode string `json:"overrideExitCode,omitempty"`
	HaltOnFinish     bool   `json:"haltOnFinish,omitempty"`
}

func abortActionAttrTypes() map[string]attr.Type {
	t := commonActionAttrTypes()
	t["cancel_process"] = types.BoolType
	t["override_exit_code"] = types.StringType
	t["halt_on_finish"] = types.BoolType
	return t
}

func abortActionsSchema() schema.ListNestedAttribute {
	attrs := commonActionSchemaAttrs()
	attrs["cancel_process"] = schema.BoolAttribute{
		MarkdownDescription: "Whether to cancel the task's process.",
		Optional:            true,
		Computed:            true,
	}
	attrs["override_exit_code"] = schema.StringAttribute{
		MarkdownDescription: "Exit code to report instead of the task's actual exit code.",
		Optional:            true,
	}
	attrs["halt_on_finish"] = schema.BoolAttribute{
		MarkdownDescription: "Whether to halt the workflow when this task finishes.",
		Optional:            true,
		Computed:            true,
	}
	return schema.ListNestedAttribute{
		MarkdownDescription: "Abort actions to take when the task matches the configured status/exit code.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
}

func abortActionsToAPI(ctx context.Context, list types.List) []AbortActionAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var models []AbortActionModel
	list.ElementsAs(ctx, &models, false)

	result := make([]AbortActionAPIModel, len(models))
	for i, m := range models {
		result[i] = AbortActionAPIModel{
			commonActionAPIModel: commonActionAPIModel{
				Status:                m.Status.ValueString(),
				NotifyOnLateStart:     m.NotifyOnLateStart.ValueBool(),
				NotifyOnLateFinish:    m.NotifyOnLateFinish.ValueBool(),
				NotifyOnEarlyFinish:   m.NotifyOnEarlyFinish.ValueBool(),
				NotifyOnProjectedLate: m.NotifyOnProjectedLate.ValueBool(),
				ExitCodes:             m.ExitCodes.ValueString(),
				Description:           m.Description.ValueString(),
				Inheritance:           m.Inheritance.ValueString(),
				SysId:                 m.SysId.ValueString(),
			},
			CancelProcess:    m.CancelProcess.ValueBool(),
			OverrideExitCode: m.OverrideExitCode.ValueString(),
			HaltOnFinish:     m.HaltOnFinish.ValueBool(),
		}
	}
	return result
}

func abortActionsFromAPIOrdered(ctx context.Context, api []AbortActionAPIModel, priorOrder types.List) types.List {
	if len(api) == 0 || priorOrder.IsNull() || priorOrder.IsUnknown() {
		return abortActionsFromAPI(api)
	}
	return abortActionsFromAPI(reorderToMatchPrior(api, abortActionsToAPI(ctx, priorOrder)))
}

func abortActionsFromAPI(api []AbortActionAPIModel) types.List {
	if len(api) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: abortActionAttrTypes()})
	}
	values := make([]attr.Value, len(api))
	for i, a := range api {
		values[i], _ = types.ObjectValue(abortActionAttrTypes(), map[string]attr.Value{
			"status":                   StringValueOrNull(a.Status),
			"notify_on_late_start":     types.BoolValue(a.NotifyOnLateStart),
			"notify_on_late_finish":    types.BoolValue(a.NotifyOnLateFinish),
			"notify_on_early_finish":   types.BoolValue(a.NotifyOnEarlyFinish),
			"notify_on_projected_late": types.BoolValue(a.NotifyOnProjectedLate),
			"exit_codes":               StringValueOrNull(a.ExitCodes),
			"description":              StringValueOrNull(a.Description),
			"inheritance":              StringValueOrNull(a.Inheritance),
			"sys_id":                   StringValueOrNull(a.SysId),
			"cancel_process":           types.BoolValue(a.CancelProcess),
			"override_exit_code":       StringValueOrNull(a.OverrideExitCode),
			"halt_on_finish":           types.BoolValue(a.HaltOnFinish),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: abortActionAttrTypes()}, values)
	return result
}

// ---------------------------------------------------------------------------
// EmailNotification
// ---------------------------------------------------------------------------

// EmailNotificationActionModel describes an email notification action in Terraform.
type EmailNotificationActionModel struct {
	Status                types.String `tfsdk:"status"`
	NotifyOnLateStart     types.Bool   `tfsdk:"notify_on_late_start"`
	NotifyOnLateFinish    types.Bool   `tfsdk:"notify_on_late_finish"`
	NotifyOnEarlyFinish   types.Bool   `tfsdk:"notify_on_early_finish"`
	NotifyOnProjectedLate types.Bool   `tfsdk:"notify_on_projected_late"`
	ExitCodes             types.String `tfsdk:"exit_codes"`
	Description           types.String `tfsdk:"description"`
	Inheritance           types.String `tfsdk:"inheritance"`
	SysId                 types.String `tfsdk:"sys_id"`

	Subject              types.String `tfsdk:"subject"`
	Body                 types.String `tfsdk:"body"`
	To                   types.String `tfsdk:"to"`
	Cc                   types.String `tfsdk:"cc"`
	Bcc                  types.String `tfsdk:"bcc"`
	ReplyTo              types.String `tfsdk:"reply_to"`
	EmailTemplate        types.String `tfsdk:"email_template"`
	EmailTemplateVar     types.String `tfsdk:"email_template_var"`
	EmailConnection      types.String `tfsdk:"email_connection"`
	AttachStdError       types.Bool   `tfsdk:"attach_std_error"`
	AttachStdOut         types.Bool   `tfsdk:"attach_std_out"`
	AttachFile           types.Bool   `tfsdk:"attach_file"`
	FileName             types.String `tfsdk:"file_name"`
	FileNumLines         types.Int64  `tfsdk:"file_num_lines"`
	FileStartLine        types.Int64  `tfsdk:"file_start_line"`
	FileScanText         types.String `tfsdk:"file_scan_text"`
	StderrNumLines       types.Int64  `tfsdk:"stderr_num_lines"`
	StderrStartLine      types.Int64  `tfsdk:"stderr_start_line"`
	StderrScanText       types.String `tfsdk:"stderr_scan_text"`
	StdoutNumLines       types.Int64  `tfsdk:"stdout_num_lines"`
	StdoutStartLine      types.Int64  `tfsdk:"stdout_start_line"`
	StdoutScanText       types.String `tfsdk:"stdout_scan_text"`
	AttachJobLog         types.Bool   `tfsdk:"attach_job_log"`
	JoblogStartLine      types.Int64  `tfsdk:"joblog_start_line"`
	JoblogNumLines       types.Int64  `tfsdk:"joblog_num_lines"`
	JoblogScanText       types.String `tfsdk:"joblog_scan_text"`
	ReportId             types.String `tfsdk:"report_id"`
	ReportVar            types.String `tfsdk:"report_var"`
	UseReportVar         types.String `tfsdk:"use_report_var"`
	ListReportFormat     types.String `tfsdk:"list_report_format"`
	AttachLocalFile      types.Bool   `tfsdk:"attach_local_file"`
	LocalAttachment      types.String `tfsdk:"local_attachment"`
	LocalAttachmentsPath types.String `tfsdk:"local_attachments_path"`
	Report               types.Object `tfsdk:"report"`
}

// EmailNotificationAPIModel represents an email notification action in the API.
type EmailNotificationAPIModel struct {
	commonActionAPIModel

	Subject              string             `json:"subject,omitempty"`
	Body                 string             `json:"body,omitempty"`
	To                   string             `json:"to,omitempty"`
	Cc                   string             `json:"cc,omitempty"`
	Bcc                  string             `json:"bcc,omitempty"`
	ReplyTo              string             `json:"replyTo,omitempty"`
	EmailTemplate        string             `json:"emailTemplate,omitempty"`
	EmailTemplateVar     string             `json:"emailTemplateVar,omitempty"`
	EmailConnection      string             `json:"emailConnection,omitempty"`
	AttachStdError       bool               `json:"attachStdError,omitempty"`
	AttachStdOut         bool               `json:"attachStdOut,omitempty"`
	AttachFile           bool               `json:"attachFile,omitempty"`
	FileName             string             `json:"fileName,omitempty"`
	FileNumLines         int64              `json:"fileNumLines,omitempty"`
	FileStartLine        int64              `json:"fileStartLine,omitempty"`
	FileScanText         string             `json:"fileScanText,omitempty"`
	StderrNumLines       int64              `json:"stderrNumLines,omitempty"`
	StderrStartLine      int64              `json:"stderrStartLine,omitempty"`
	StderrScanText       string             `json:"stderrScanText,omitempty"`
	StdoutNumLines       int64              `json:"stdoutNumLines,omitempty"`
	StdoutStartLine      int64              `json:"stdoutStartLine,omitempty"`
	StdoutScanText       string             `json:"stdoutScanText,omitempty"`
	AttachJobLog         bool               `json:"attachJobLog,omitempty"`
	JoblogStartLine      int64              `json:"joblogStartLine,omitempty"`
	JoblogNumLines       int64              `json:"joblogNumLines,omitempty"`
	JoblogScanText       string             `json:"joblogScanText,omitempty"`
	ReportId             string             `json:"reportId,omitempty"`
	ReportVar            string             `json:"reportVar,omitempty"`
	UseReportVar         string             `json:"useReportVar,omitempty"`
	ListReportFormat     string             `json:"listReportFormat,omitempty"`
	AttachLocalFile      bool               `json:"attachLocalFile,omitempty"`
	LocalAttachment      string             `json:"localAttachment,omitempty"`
	LocalAttachmentsPath string             `json:"localAttachmentsPath,omitempty"`
	Report               *ReportRefAPIModel `json:"report,omitempty"`
}

func emailNotificationAttrTypes() map[string]attr.Type {
	t := commonActionAttrTypes()
	t["subject"] = types.StringType
	t["body"] = types.StringType
	t["to"] = types.StringType
	t["cc"] = types.StringType
	t["bcc"] = types.StringType
	t["reply_to"] = types.StringType
	t["email_template"] = types.StringType
	t["email_template_var"] = types.StringType
	t["email_connection"] = types.StringType
	t["attach_std_error"] = types.BoolType
	t["attach_std_out"] = types.BoolType
	t["attach_file"] = types.BoolType
	t["file_name"] = types.StringType
	t["file_num_lines"] = types.Int64Type
	t["file_start_line"] = types.Int64Type
	t["file_scan_text"] = types.StringType
	t["stderr_num_lines"] = types.Int64Type
	t["stderr_start_line"] = types.Int64Type
	t["stderr_scan_text"] = types.StringType
	t["stdout_num_lines"] = types.Int64Type
	t["stdout_start_line"] = types.Int64Type
	t["stdout_scan_text"] = types.StringType
	t["attach_job_log"] = types.BoolType
	t["joblog_start_line"] = types.Int64Type
	t["joblog_num_lines"] = types.Int64Type
	t["joblog_scan_text"] = types.StringType
	t["report_id"] = types.StringType
	t["report_var"] = types.StringType
	t["use_report_var"] = types.StringType
	t["list_report_format"] = types.StringType
	t["attach_local_file"] = types.BoolType
	t["local_attachment"] = types.StringType
	t["local_attachments_path"] = types.StringType
	t["report"] = types.ObjectType{AttrTypes: reportRefAttrTypes()}
	return t
}

func emailNotificationsSchema() schema.ListNestedAttribute {
	attrs := commonActionSchemaAttrs()
	stringOpt := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Optional: true}
	}
	boolOptComputed := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{MarkdownDescription: desc, Optional: true, Computed: true}
	}
	intOptComputed := func(desc string) schema.Int64Attribute {
		return schema.Int64Attribute{MarkdownDescription: desc, Optional: true, Computed: true}
	}

	attrs["subject"] = stringOpt("Subject of the notification email.")
	attrs["body"] = stringOpt("Body of the notification email.")
	attrs["to"] = stringOpt("Comma-separated list of 'To' recipients.")
	attrs["cc"] = stringOpt("Comma-separated list of 'Cc' recipients.")
	attrs["bcc"] = stringOpt("Comma-separated list of 'Bcc' recipients.")
	attrs["reply_to"] = stringOpt("Reply-To address.")
	attrs["email_template"] = stringOpt("Name of the email template to use.")
	attrs["email_template_var"] = stringOpt("Variable containing the email template name.")
	attrs["email_connection"] = stringOpt("Name of the email connection to send through.")
	attrs["attach_std_error"] = boolOptComputed("Whether to attach the task's stderr output.")
	attrs["attach_std_out"] = boolOptComputed("Whether to attach the task's stdout output.")
	attrs["attach_file"] = boolOptComputed("Whether to attach a file.")
	attrs["file_name"] = stringOpt("Name of the file to attach.")
	attrs["file_num_lines"] = intOptComputed("Number of lines of the file to include.")
	attrs["file_start_line"] = intOptComputed("Starting line of the file to include.")
	attrs["file_scan_text"] = stringOpt("Text to scan for within the file attachment.")
	attrs["stderr_num_lines"] = intOptComputed("Number of lines of stderr to include.")
	attrs["stderr_start_line"] = intOptComputed("Starting line of stderr to include.")
	attrs["stderr_scan_text"] = stringOpt("Text to scan for within stderr.")
	attrs["stdout_num_lines"] = intOptComputed("Number of lines of stdout to include.")
	attrs["stdout_start_line"] = intOptComputed("Starting line of stdout to include.")
	attrs["stdout_scan_text"] = stringOpt("Text to scan for within stdout.")
	attrs["attach_job_log"] = boolOptComputed("Whether to attach the job log.")
	attrs["joblog_start_line"] = intOptComputed("Starting line of the job log to include.")
	attrs["joblog_num_lines"] = intOptComputed("Number of lines of the job log to include.")
	attrs["joblog_scan_text"] = stringOpt("Text to scan for within the job log.")
	attrs["report_id"] = stringOpt("ID of the report to attach.")
	attrs["report_var"] = stringOpt("Variable containing the report ID.")
	attrs["use_report_var"] = stringOpt("Whether/how to use the report variable.")
	attrs["list_report_format"] = schema.StringAttribute{
		MarkdownDescription: "Format of the attached report listing. Values: 'PDF', 'CSV', 'HTML'.",
		Optional:            true,
		Computed:            true,
	}
	attrs["attach_local_file"] = boolOptComputed("Whether to attach a local file.")
	attrs["local_attachment"] = stringOpt("Name of the local file to attach.")
	attrs["local_attachments_path"] = stringOpt("Path to local file attachments.")
	attrs["report"] = reportRefSchema()

	return schema.ListNestedAttribute{
		MarkdownDescription: "Email notifications to send when the task matches the configured status/exit code.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
}

func emailNotificationsToAPI(ctx context.Context, list types.List) []EmailNotificationAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var models []EmailNotificationActionModel
	list.ElementsAs(ctx, &models, false)

	result := make([]EmailNotificationAPIModel, len(models))
	for i, m := range models {
		result[i] = EmailNotificationAPIModel{
			commonActionAPIModel: commonActionAPIModel{
				Status:                m.Status.ValueString(),
				NotifyOnLateStart:     m.NotifyOnLateStart.ValueBool(),
				NotifyOnLateFinish:    m.NotifyOnLateFinish.ValueBool(),
				NotifyOnEarlyFinish:   m.NotifyOnEarlyFinish.ValueBool(),
				NotifyOnProjectedLate: m.NotifyOnProjectedLate.ValueBool(),
				ExitCodes:             m.ExitCodes.ValueString(),
				Description:           m.Description.ValueString(),
				Inheritance:           m.Inheritance.ValueString(),
				SysId:                 m.SysId.ValueString(),
			},
			Subject:              m.Subject.ValueString(),
			Body:                 m.Body.ValueString(),
			To:                   m.To.ValueString(),
			Cc:                   m.Cc.ValueString(),
			Bcc:                  m.Bcc.ValueString(),
			ReplyTo:              m.ReplyTo.ValueString(),
			EmailTemplate:        m.EmailTemplate.ValueString(),
			EmailTemplateVar:     m.EmailTemplateVar.ValueString(),
			EmailConnection:      m.EmailConnection.ValueString(),
			AttachStdError:       m.AttachStdError.ValueBool(),
			AttachStdOut:         m.AttachStdOut.ValueBool(),
			AttachFile:           m.AttachFile.ValueBool(),
			FileName:             m.FileName.ValueString(),
			FileNumLines:         m.FileNumLines.ValueInt64(),
			FileStartLine:        m.FileStartLine.ValueInt64(),
			FileScanText:         m.FileScanText.ValueString(),
			StderrNumLines:       m.StderrNumLines.ValueInt64(),
			StderrStartLine:      m.StderrStartLine.ValueInt64(),
			StderrScanText:       m.StderrScanText.ValueString(),
			StdoutNumLines:       m.StdoutNumLines.ValueInt64(),
			StdoutStartLine:      m.StdoutStartLine.ValueInt64(),
			StdoutScanText:       m.StdoutScanText.ValueString(),
			AttachJobLog:         m.AttachJobLog.ValueBool(),
			JoblogStartLine:      m.JoblogStartLine.ValueInt64(),
			JoblogNumLines:       m.JoblogNumLines.ValueInt64(),
			JoblogScanText:       m.JoblogScanText.ValueString(),
			ReportId:             m.ReportId.ValueString(),
			ReportVar:            m.ReportVar.ValueString(),
			UseReportVar:         m.UseReportVar.ValueString(),
			ListReportFormat:     m.ListReportFormat.ValueString(),
			AttachLocalFile:      m.AttachLocalFile.ValueBool(),
			LocalAttachment:      m.LocalAttachment.ValueString(),
			LocalAttachmentsPath: m.LocalAttachmentsPath.ValueString(),
			Report:               reportRefToAPI(ctx, m.Report),
		}
	}
	return result
}

func emailNotificationsFromAPIOrdered(ctx context.Context, api []EmailNotificationAPIModel, priorOrder types.List) types.List {
	if len(api) == 0 || priorOrder.IsNull() || priorOrder.IsUnknown() {
		return emailNotificationsFromAPI(ctx, api)
	}
	return emailNotificationsFromAPI(ctx, reorderToMatchPrior(api, emailNotificationsToAPI(ctx, priorOrder)))
}

func emailNotificationsFromAPI(ctx context.Context, api []EmailNotificationAPIModel) types.List {
	if len(api) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: emailNotificationAttrTypes()})
	}
	values := make([]attr.Value, len(api))
	for i, e := range api {
		values[i], _ = types.ObjectValue(emailNotificationAttrTypes(), map[string]attr.Value{
			"status":                   StringValueOrNull(e.Status),
			"notify_on_late_start":     types.BoolValue(e.NotifyOnLateStart),
			"notify_on_late_finish":    types.BoolValue(e.NotifyOnLateFinish),
			"notify_on_early_finish":   types.BoolValue(e.NotifyOnEarlyFinish),
			"notify_on_projected_late": types.BoolValue(e.NotifyOnProjectedLate),
			"exit_codes":               StringValueOrNull(e.ExitCodes),
			"description":              StringValueOrNull(e.Description),
			"inheritance":              StringValueOrNull(e.Inheritance),
			"sys_id":                   StringValueOrNull(e.SysId),
			"subject":                  StringValueOrNull(e.Subject),
			"body":                     StringValueOrNull(e.Body),
			"to":                       StringValueOrNull(e.To),
			"cc":                       StringValueOrNull(e.Cc),
			"bcc":                      StringValueOrNull(e.Bcc),
			"reply_to":                 StringValueOrNull(e.ReplyTo),
			"email_template":           StringValueOrNull(e.EmailTemplate),
			"email_template_var":       StringValueOrNull(e.EmailTemplateVar),
			"email_connection":         StringValueOrNull(e.EmailConnection),
			"attach_std_error":         types.BoolValue(e.AttachStdError),
			"attach_std_out":           types.BoolValue(e.AttachStdOut),
			"attach_file":              types.BoolValue(e.AttachFile),
			"file_name":                StringValueOrNull(e.FileName),
			"file_num_lines":           types.Int64Value(e.FileNumLines),
			"file_start_line":          types.Int64Value(e.FileStartLine),
			"file_scan_text":           StringValueOrNull(e.FileScanText),
			"stderr_num_lines":         types.Int64Value(e.StderrNumLines),
			"stderr_start_line":        types.Int64Value(e.StderrStartLine),
			"stderr_scan_text":         StringValueOrNull(e.StderrScanText),
			"stdout_num_lines":         types.Int64Value(e.StdoutNumLines),
			"stdout_start_line":        types.Int64Value(e.StdoutStartLine),
			"stdout_scan_text":         StringValueOrNull(e.StdoutScanText),
			"attach_job_log":           types.BoolValue(e.AttachJobLog),
			"joblog_start_line":        types.Int64Value(e.JoblogStartLine),
			"joblog_num_lines":         types.Int64Value(e.JoblogNumLines),
			"joblog_scan_text":         StringValueOrNull(e.JoblogScanText),
			"report_id":                StringValueOrNull(e.ReportId),
			"report_var":               StringValueOrNull(e.ReportVar),
			"use_report_var":           StringValueOrNull(e.UseReportVar),
			"list_report_format":       StringValueOrNull(e.ListReportFormat),
			"attach_local_file":        types.BoolValue(e.AttachLocalFile),
			"local_attachment":         StringValueOrNull(e.LocalAttachment),
			"local_attachments_path":   StringValueOrNull(e.LocalAttachmentsPath),
			"report":                   reportRefFromAPI(ctx, e.Report),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: emailNotificationAttrTypes()}, values)
	return result
}

// ---------------------------------------------------------------------------
// SetVariableAction
// ---------------------------------------------------------------------------

// SetVariableActionModel describes a set-variable action in Terraform.
type SetVariableActionModel struct {
	Status                types.String `tfsdk:"status"`
	NotifyOnLateStart     types.Bool   `tfsdk:"notify_on_late_start"`
	NotifyOnLateFinish    types.Bool   `tfsdk:"notify_on_late_finish"`
	NotifyOnEarlyFinish   types.Bool   `tfsdk:"notify_on_early_finish"`
	NotifyOnProjectedLate types.Bool   `tfsdk:"notify_on_projected_late"`
	ExitCodes             types.String `tfsdk:"exit_codes"`
	Description           types.String `tfsdk:"description"`
	Inheritance           types.String `tfsdk:"inheritance"`
	SysId                 types.String `tfsdk:"sys_id"`

	VariableScope       types.String `tfsdk:"variable_scope"`
	VariableName        types.String `tfsdk:"variable_name"`
	VariableDescription types.String `tfsdk:"variable_description"`
	VariableValue       types.String `tfsdk:"variable_value"`
	NotificationOption  types.String `tfsdk:"notification_option"`
}

// SetVariableActionAPIModel represents a set-variable action in the API.
type SetVariableActionAPIModel struct {
	commonActionAPIModel

	VariableScope       string `json:"variableScope,omitempty"`
	VariableName        string `json:"variableName,omitempty"`
	VariableDescription string `json:"variableDescription,omitempty"`
	VariableValue       string `json:"variableValue,omitempty"`
	NotificationOption  string `json:"notificationOption,omitempty"`
}

func setVariableActionAttrTypes() map[string]attr.Type {
	t := commonActionAttrTypes()
	t["variable_scope"] = types.StringType
	t["variable_name"] = types.StringType
	t["variable_description"] = types.StringType
	t["variable_value"] = types.StringType
	t["notification_option"] = types.StringType
	return t
}

func setVariableActionsSchema() schema.ListNestedAttribute {
	attrs := commonActionSchemaAttrs()
	attrs["variable_scope"] = schema.StringAttribute{
		MarkdownDescription: "Scope of the variable. Values: 'Task', 'Workflow', 'Global'.",
		Optional:            true,
		Computed:            true,
	}
	attrs["variable_name"] = schema.StringAttribute{
		MarkdownDescription: "Name of the variable to set.",
		Required:            true,
	}
	attrs["variable_description"] = schema.StringAttribute{
		MarkdownDescription: "Description of the variable.",
		Optional:            true,
	}
	attrs["variable_value"] = schema.StringAttribute{
		MarkdownDescription: "Value to set the variable to.",
		Optional:            true,
	}
	attrs["notification_option"] = schema.StringAttribute{
		MarkdownDescription: "Notification option for this action.",
		Optional:            true,
		Computed:            true,
	}
	return schema.ListNestedAttribute{
		MarkdownDescription: "Actions to set a variable when the task matches the configured status/exit code.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
}

func setVariableActionsToAPI(ctx context.Context, list types.List) []SetVariableActionAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var models []SetVariableActionModel
	list.ElementsAs(ctx, &models, false)

	result := make([]SetVariableActionAPIModel, len(models))
	for i, m := range models {
		result[i] = SetVariableActionAPIModel{
			commonActionAPIModel: commonActionAPIModel{
				Status:                m.Status.ValueString(),
				NotifyOnLateStart:     m.NotifyOnLateStart.ValueBool(),
				NotifyOnLateFinish:    m.NotifyOnLateFinish.ValueBool(),
				NotifyOnEarlyFinish:   m.NotifyOnEarlyFinish.ValueBool(),
				NotifyOnProjectedLate: m.NotifyOnProjectedLate.ValueBool(),
				ExitCodes:             m.ExitCodes.ValueString(),
				Description:           m.Description.ValueString(),
				Inheritance:           m.Inheritance.ValueString(),
				SysId:                 m.SysId.ValueString(),
			},
			VariableScope:       m.VariableScope.ValueString(),
			VariableName:        m.VariableName.ValueString(),
			VariableDescription: m.VariableDescription.ValueString(),
			VariableValue:       m.VariableValue.ValueString(),
			NotificationOption:  m.NotificationOption.ValueString(),
		}
	}
	return result
}

func setVariableActionsFromAPIOrdered(ctx context.Context, api []SetVariableActionAPIModel, priorOrder types.List) types.List {
	if len(api) == 0 || priorOrder.IsNull() || priorOrder.IsUnknown() {
		return setVariableActionsFromAPI(api)
	}
	return setVariableActionsFromAPI(reorderToMatchPrior(api, setVariableActionsToAPI(ctx, priorOrder)))
}

func setVariableActionsFromAPI(api []SetVariableActionAPIModel) types.List {
	if len(api) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: setVariableActionAttrTypes()})
	}
	values := make([]attr.Value, len(api))
	for i, v := range api {
		values[i], _ = types.ObjectValue(setVariableActionAttrTypes(), map[string]attr.Value{
			"status":                   StringValueOrNull(v.Status),
			"notify_on_late_start":     types.BoolValue(v.NotifyOnLateStart),
			"notify_on_late_finish":    types.BoolValue(v.NotifyOnLateFinish),
			"notify_on_early_finish":   types.BoolValue(v.NotifyOnEarlyFinish),
			"notify_on_projected_late": types.BoolValue(v.NotifyOnProjectedLate),
			"exit_codes":               StringValueOrNull(v.ExitCodes),
			"description":              StringValueOrNull(v.Description),
			"inheritance":              StringValueOrNull(v.Inheritance),
			"sys_id":                   StringValueOrNull(v.SysId),
			"variable_scope":           StringValueOrNull(v.VariableScope),
			"variable_name":            StringValueOrNull(v.VariableName),
			"variable_description":     StringValueOrNull(v.VariableDescription),
			"variable_value":           StringValueOrNull(v.VariableValue),
			"notification_option":      StringValueOrNull(v.NotificationOption),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: setVariableActionAttrTypes()}, values)
	return result
}

// ---------------------------------------------------------------------------
// SnmpNotification
// ---------------------------------------------------------------------------

// SnmpNotificationModel describes an SNMP notification action in Terraform.
type SnmpNotificationModel struct {
	Status                types.String `tfsdk:"status"`
	NotifyOnLateStart     types.Bool   `tfsdk:"notify_on_late_start"`
	NotifyOnLateFinish    types.Bool   `tfsdk:"notify_on_late_finish"`
	NotifyOnEarlyFinish   types.Bool   `tfsdk:"notify_on_early_finish"`
	NotifyOnProjectedLate types.Bool   `tfsdk:"notify_on_projected_late"`
	ExitCodes             types.String `tfsdk:"exit_codes"`
	Description           types.String `tfsdk:"description"`
	Inheritance           types.String `tfsdk:"inheritance"`
	SysId                 types.String `tfsdk:"sys_id"`

	SnmpManager types.String `tfsdk:"snmp_manager"`
	Severity    types.String `tfsdk:"severity"`
}

// SnmpNotificationAPIModel represents an SNMP notification action in the API.
type SnmpNotificationAPIModel struct {
	commonActionAPIModel

	SnmpManager string `json:"snmpManager,omitempty"`
	Severity    string `json:"severity,omitempty"`
}

func snmpNotificationAttrTypes() map[string]attr.Type {
	t := commonActionAttrTypes()
	t["snmp_manager"] = types.StringType
	t["severity"] = types.StringType
	return t
}

func snmpNotificationsSchema() schema.ListNestedAttribute {
	attrs := commonActionSchemaAttrs()
	attrs["snmp_manager"] = schema.StringAttribute{
		MarkdownDescription: "Name of the SNMP manager connection to send the trap to.",
		Optional:            true,
	}
	attrs["severity"] = schema.StringAttribute{
		MarkdownDescription: "Severity of the SNMP notification.",
		Optional:            true,
	}
	return schema.ListNestedAttribute{
		MarkdownDescription: "SNMP notifications to send when the task matches the configured status/exit code.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
}

func snmpNotificationsToAPI(ctx context.Context, list types.List) []SnmpNotificationAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var models []SnmpNotificationModel
	list.ElementsAs(ctx, &models, false)

	result := make([]SnmpNotificationAPIModel, len(models))
	for i, m := range models {
		result[i] = SnmpNotificationAPIModel{
			commonActionAPIModel: commonActionAPIModel{
				Status:                m.Status.ValueString(),
				NotifyOnLateStart:     m.NotifyOnLateStart.ValueBool(),
				NotifyOnLateFinish:    m.NotifyOnLateFinish.ValueBool(),
				NotifyOnEarlyFinish:   m.NotifyOnEarlyFinish.ValueBool(),
				NotifyOnProjectedLate: m.NotifyOnProjectedLate.ValueBool(),
				ExitCodes:             m.ExitCodes.ValueString(),
				Description:           m.Description.ValueString(),
				Inheritance:           m.Inheritance.ValueString(),
				SysId:                 m.SysId.ValueString(),
			},
			SnmpManager: m.SnmpManager.ValueString(),
			Severity:    m.Severity.ValueString(),
		}
	}
	return result
}

func snmpNotificationsFromAPIOrdered(ctx context.Context, api []SnmpNotificationAPIModel, priorOrder types.List) types.List {
	if len(api) == 0 || priorOrder.IsNull() || priorOrder.IsUnknown() {
		return snmpNotificationsFromAPI(api)
	}
	return snmpNotificationsFromAPI(reorderToMatchPrior(api, snmpNotificationsToAPI(ctx, priorOrder)))
}

func snmpNotificationsFromAPI(api []SnmpNotificationAPIModel) types.List {
	if len(api) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: snmpNotificationAttrTypes()})
	}
	values := make([]attr.Value, len(api))
	for i, s := range api {
		values[i], _ = types.ObjectValue(snmpNotificationAttrTypes(), map[string]attr.Value{
			"status":                   StringValueOrNull(s.Status),
			"notify_on_late_start":     types.BoolValue(s.NotifyOnLateStart),
			"notify_on_late_finish":    types.BoolValue(s.NotifyOnLateFinish),
			"notify_on_early_finish":   types.BoolValue(s.NotifyOnEarlyFinish),
			"notify_on_projected_late": types.BoolValue(s.NotifyOnProjectedLate),
			"exit_codes":               StringValueOrNull(s.ExitCodes),
			"description":              StringValueOrNull(s.Description),
			"inheritance":              StringValueOrNull(s.Inheritance),
			"sys_id":                   StringValueOrNull(s.SysId),
			"snmp_manager":             StringValueOrNull(s.SnmpManager),
			"severity":                 StringValueOrNull(s.Severity),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: snmpNotificationAttrTypes()}, values)
	return result
}

// ---------------------------------------------------------------------------
// SystemOperation ("Launch Task" and other system operations)
// ---------------------------------------------------------------------------

// SystemOperationModel describes a system operation action in Terraform.
type SystemOperationModel struct {
	Status                types.String `tfsdk:"status"`
	NotifyOnLateStart     types.Bool   `tfsdk:"notify_on_late_start"`
	NotifyOnLateFinish    types.Bool   `tfsdk:"notify_on_late_finish"`
	NotifyOnEarlyFinish   types.Bool   `tfsdk:"notify_on_early_finish"`
	NotifyOnProjectedLate types.Bool   `tfsdk:"notify_on_projected_late"`
	ExitCodes             types.String `tfsdk:"exit_codes"`
	Description           types.String `tfsdk:"description"`
	Inheritance           types.String `tfsdk:"inheritance"`
	SysId                 types.String `tfsdk:"sys_id"`

	Operation                 types.String `tfsdk:"operation"`
	Task                      types.String `tfsdk:"task"`
	TaskVar                   types.String `tfsdk:"task_var"`
	TaskLimitType             types.String `tfsdk:"task_limit_type"`
	Limit                     types.String `tfsdk:"limit"`
	VirtualResource           types.String `tfsdk:"virtual_resource"`
	VirtualResourceVar        types.String `tfsdk:"virtual_resource_var"`
	ExecCommand               types.String `tfsdk:"exec_command"`
	ExecCriteria              types.String `tfsdk:"exec_criteria"`
	ExecLookupOption          types.String `tfsdk:"exec_lookup_option"`
	ExecName                  types.String `tfsdk:"exec_name"`
	ExecId                    types.String `tfsdk:"exec_id"`
	ExecWorkflowName          types.String `tfsdk:"exec_workflow_name"`
	ExecWorkflowNameCond      types.String `tfsdk:"exec_workflow_name_cond"`
	Agent                     types.String `tfsdk:"agent"`
	AgentVar                  types.String `tfsdk:"agent_var"`
	AgentCluster              types.String `tfsdk:"agent_cluster"`
	AgentClusterVar           types.String `tfsdk:"agent_cluster_var"`
	Trigger                   types.String `tfsdk:"trigger"`
	TriggerVar                types.String `tfsdk:"trigger_var"`
	OverrideTriggerTime       types.String `tfsdk:"override_trigger_time"`
	OverrideTriggerDateTime   types.Bool   `tfsdk:"override_trigger_date_time"`
	OverrideTriggerDateOffset types.String `tfsdk:"override_trigger_date_offset"`
	VertexSelection           types.Bool   `tfsdk:"vertex_selection"`
	Vertices                  types.List   `tfsdk:"vertices"`
	NotificationOption        types.String `tfsdk:"notification_option"`
	VariablesUnresolved       types.Bool   `tfsdk:"variables_unresolved"`
	Variables                 types.List   `tfsdk:"variables"`
}

// SystemOperationAPIModel represents a system operation action in the API.
type SystemOperationAPIModel struct {
	commonActionAPIModel

	Operation                 string                 `json:"operation,omitempty"`
	Task                      string                 `json:"task,omitempty"`
	TaskVar                   string                 `json:"taskVar,omitempty"`
	TaskLimitType             string                 `json:"taskLimitType,omitempty"`
	Limit                     string                 `json:"limit,omitempty"`
	VirtualResource           string                 `json:"virtualResource,omitempty"`
	VirtualResourceVar        string                 `json:"virtualResourceVar,omitempty"`
	ExecCommand               string                 `json:"execCommand,omitempty"`
	ExecCriteria              string                 `json:"execCriteria,omitempty"`
	ExecLookupOption          string                 `json:"execLookupOption,omitempty"`
	ExecName                  string                 `json:"execName,omitempty"`
	ExecId                    string                 `json:"execId,omitempty"`
	ExecWorkflowName          string                 `json:"execWorkflowName,omitempty"`
	ExecWorkflowNameCond      string                 `json:"execWorkflowNameCond,omitempty"`
	Agent                     string                 `json:"agent,omitempty"`
	AgentVar                  string                 `json:"agentVar,omitempty"`
	AgentCluster              string                 `json:"agentCluster,omitempty"`
	AgentClusterVar           string                 `json:"agentClusterVar,omitempty"`
	Trigger                   string                 `json:"trigger,omitempty"`
	TriggerVar                string                 `json:"triggerVar,omitempty"`
	OverrideTriggerTime       string                 `json:"overrideTriggerTime,omitempty"`
	OverrideTriggerDateTime   bool                   `json:"overrideTriggerDateTime,omitempty"`
	OverrideTriggerDateOffset string                 `json:"overrideTriggerDateOffset,omitempty"`
	VertexSelection           bool                   `json:"vertexSelection,omitempty"`
	Vertices                  []VertexRefAPIModel    `json:"vertices,omitempty"`
	NotificationOption        string                 `json:"notificationOption,omitempty"`
	VariablesUnresolved       bool                   `json:"variablesUnresolved,omitempty"`
	Variables                 []TaskVariableAPIModel `json:"variables,omitempty"`
}

func systemOperationAttrTypes() map[string]attr.Type {
	t := commonActionAttrTypes()
	t["operation"] = types.StringType
	t["task"] = types.StringType
	t["task_var"] = types.StringType
	t["task_limit_type"] = types.StringType
	t["limit"] = types.StringType
	t["virtual_resource"] = types.StringType
	t["virtual_resource_var"] = types.StringType
	t["exec_command"] = types.StringType
	t["exec_criteria"] = types.StringType
	t["exec_lookup_option"] = types.StringType
	t["exec_name"] = types.StringType
	t["exec_id"] = types.StringType
	t["exec_workflow_name"] = types.StringType
	t["exec_workflow_name_cond"] = types.StringType
	t["agent"] = types.StringType
	t["agent_var"] = types.StringType
	t["agent_cluster"] = types.StringType
	t["agent_cluster_var"] = types.StringType
	t["trigger"] = types.StringType
	t["trigger_var"] = types.StringType
	t["override_trigger_time"] = types.StringType
	t["override_trigger_date_time"] = types.BoolType
	t["override_trigger_date_offset"] = types.StringType
	t["vertex_selection"] = types.BoolType
	t["vertices"] = types.ListType{ElemType: types.ObjectType{AttrTypes: vertexRefAttrTypes()}}
	t["notification_option"] = types.StringType
	t["variables_unresolved"] = types.BoolType
	t["variables"] = types.ListType{ElemType: types.ObjectType{AttrTypes: TaskVariableAttrTypes()}}
	return t
}

func systemOperationsSchema() schema.ListNestedAttribute {
	attrs := commonActionSchemaAttrs()
	stringOpt := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Optional: true}
	}
	stringOptComputed := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Optional: true, Computed: true}
	}
	boolOptComputed := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{MarkdownDescription: desc, Optional: true, Computed: true}
	}

	attrs["operation"] = stringOptComputed("System operation to perform. Values: 'Launch Task', 'Cancel', 'Force Finish', 'Force Finish/Success', 'Skip On', 'Skip Off', 'Hold', 'Release', etc. Defaults to 'Launch Task'.")
	attrs["task"] = stringOpt("Name of the task to operate on.")
	attrs["task_var"] = stringOpt("Variable containing the task name.")
	attrs["task_limit_type"] = stringOptComputed("Limit type for launching multiple task instances. Values: 'Unlimited', 'Limited'. Defaults to 'Unlimited'.")
	attrs["limit"] = stringOpt("Maximum number of task instances (when `task_limit_type` is 'Limited').")
	attrs["virtual_resource"] = stringOpt("Name of the virtual resource to operate on.")
	attrs["virtual_resource_var"] = stringOpt("Variable containing the virtual resource name.")
	attrs["exec_command"] = stringOptComputed("Command to execute against the matched task instance(s). Defaults to 'Cancel'.")
	attrs["exec_criteria"] = stringOptComputed("Criteria for selecting the task instance to operate on. Defaults to 'Oldest Active Instance'.")
	attrs["exec_lookup_option"] = stringOptComputed("How to look up the task instance. Values: 'Instance Name', 'Instance Id'. Defaults to 'Instance Name'.")
	attrs["exec_name"] = stringOpt("Name of the task instance to operate on.")
	attrs["exec_id"] = stringOpt("ID of the task instance to operate on.")
	attrs["exec_workflow_name"] = stringOpt("Workflow instance name to match against.")
	attrs["exec_workflow_name_cond"] = stringOptComputed("Condition used to match the workflow instance name. Defaults to 'Equals'.")
	attrs["agent"] = stringOpt("Name of the agent to launch the task on.")
	attrs["agent_var"] = stringOpt("Variable containing the agent name.")
	attrs["agent_cluster"] = stringOpt("Name of the agent cluster to launch the task on.")
	attrs["agent_cluster_var"] = stringOpt("Variable containing the agent cluster name.")
	attrs["trigger"] = stringOpt("Name of the trigger to use when launching the task.")
	attrs["trigger_var"] = stringOpt("Variable containing the trigger name.")
	attrs["override_trigger_time"] = stringOptComputed("Overridden trigger time (HH:MM). Defaults to '00:00'.")
	attrs["override_trigger_date_time"] = boolOptComputed("Whether to override the trigger date/time.")
	attrs["override_trigger_date_offset"] = stringOpt("Date offset to apply to the overridden trigger.")
	attrs["vertex_selection"] = boolOptComputed("Whether to target specific workflow vertices instead of the whole workflow.")
	attrs["vertices"] = vertexRefsSchema()
	attrs["notification_option"] = stringOptComputed("When this action is evaluated relative to the task's completion. Defaults to 'Operation Failure'.")
	attrs["variables_unresolved"] = boolOptComputed("Whether variable references in this action are left unresolved.")
	attrs["variables"] = TaskVariablesSchema()

	return schema.ListNestedAttribute{
		MarkdownDescription: "System operations (e.g. Launch Task) to perform when the task matches the configured status/exit code.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
}

func systemOperationsToAPI(ctx context.Context, list types.List) []SystemOperationAPIModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var models []SystemOperationModel
	list.ElementsAs(ctx, &models, false)

	result := make([]SystemOperationAPIModel, len(models))
	for i, m := range models {
		result[i] = SystemOperationAPIModel{
			commonActionAPIModel: commonActionAPIModel{
				Status:                m.Status.ValueString(),
				NotifyOnLateStart:     m.NotifyOnLateStart.ValueBool(),
				NotifyOnLateFinish:    m.NotifyOnLateFinish.ValueBool(),
				NotifyOnEarlyFinish:   m.NotifyOnEarlyFinish.ValueBool(),
				NotifyOnProjectedLate: m.NotifyOnProjectedLate.ValueBool(),
				ExitCodes:             m.ExitCodes.ValueString(),
				Description:           m.Description.ValueString(),
				Inheritance:           m.Inheritance.ValueString(),
				SysId:                 m.SysId.ValueString(),
			},
			Operation:                 m.Operation.ValueString(),
			Task:                      m.Task.ValueString(),
			TaskVar:                   m.TaskVar.ValueString(),
			TaskLimitType:             m.TaskLimitType.ValueString(),
			Limit:                     m.Limit.ValueString(),
			VirtualResource:           m.VirtualResource.ValueString(),
			VirtualResourceVar:        m.VirtualResourceVar.ValueString(),
			ExecCommand:               m.ExecCommand.ValueString(),
			ExecCriteria:              m.ExecCriteria.ValueString(),
			ExecLookupOption:          m.ExecLookupOption.ValueString(),
			ExecName:                  m.ExecName.ValueString(),
			ExecId:                    m.ExecId.ValueString(),
			ExecWorkflowName:          m.ExecWorkflowName.ValueString(),
			ExecWorkflowNameCond:      m.ExecWorkflowNameCond.ValueString(),
			Agent:                     m.Agent.ValueString(),
			AgentVar:                  m.AgentVar.ValueString(),
			AgentCluster:              m.AgentCluster.ValueString(),
			AgentClusterVar:           m.AgentClusterVar.ValueString(),
			Trigger:                   m.Trigger.ValueString(),
			TriggerVar:                m.TriggerVar.ValueString(),
			OverrideTriggerTime:       m.OverrideTriggerTime.ValueString(),
			OverrideTriggerDateTime:   m.OverrideTriggerDateTime.ValueBool(),
			OverrideTriggerDateOffset: m.OverrideTriggerDateOffset.ValueString(),
			VertexSelection:           m.VertexSelection.ValueBool(),
			Vertices:                  vertexRefsToAPI(ctx, m.Vertices),
			NotificationOption:        m.NotificationOption.ValueString(),
			VariablesUnresolved:       m.VariablesUnresolved.ValueBool(),
			Variables:                 TaskVariablesToAPI(ctx, m.Variables),
		}
	}
	return result
}

func systemOperationsFromAPIOrdered(ctx context.Context, api []SystemOperationAPIModel, priorOrder types.List) types.List {
	if len(api) == 0 || priorOrder.IsNull() || priorOrder.IsUnknown() {
		return systemOperationsFromAPI(ctx, api)
	}
	return systemOperationsFromAPI(ctx, reorderToMatchPrior(api, systemOperationsToAPI(ctx, priorOrder)))
}

func systemOperationsFromAPI(ctx context.Context, api []SystemOperationAPIModel) types.List {
	if len(api) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: systemOperationAttrTypes()})
	}
	values := make([]attr.Value, len(api))
	for i, s := range api {
		values[i], _ = types.ObjectValue(systemOperationAttrTypes(), map[string]attr.Value{
			"status":                       StringValueOrNull(s.Status),
			"notify_on_late_start":         types.BoolValue(s.NotifyOnLateStart),
			"notify_on_late_finish":        types.BoolValue(s.NotifyOnLateFinish),
			"notify_on_early_finish":       types.BoolValue(s.NotifyOnEarlyFinish),
			"notify_on_projected_late":     types.BoolValue(s.NotifyOnProjectedLate),
			"exit_codes":                   StringValueOrNull(s.ExitCodes),
			"description":                  StringValueOrNull(s.Description),
			"inheritance":                  StringValueOrNull(s.Inheritance),
			"sys_id":                       StringValueOrNull(s.SysId),
			"operation":                    StringValueOrNull(s.Operation),
			"task":                         StringValueOrNull(s.Task),
			"task_var":                     StringValueOrNull(s.TaskVar),
			"task_limit_type":              StringValueOrNull(s.TaskLimitType),
			"limit":                        StringValueOrNull(s.Limit),
			"virtual_resource":             StringValueOrNull(s.VirtualResource),
			"virtual_resource_var":         StringValueOrNull(s.VirtualResourceVar),
			"exec_command":                 StringValueOrNull(s.ExecCommand),
			"exec_criteria":                StringValueOrNull(s.ExecCriteria),
			"exec_lookup_option":           StringValueOrNull(s.ExecLookupOption),
			"exec_name":                    StringValueOrNull(s.ExecName),
			"exec_id":                      StringValueOrNull(s.ExecId),
			"exec_workflow_name":           StringValueOrNull(s.ExecWorkflowName),
			"exec_workflow_name_cond":      StringValueOrNull(s.ExecWorkflowNameCond),
			"agent":                        StringValueOrNull(s.Agent),
			"agent_var":                    StringValueOrNull(s.AgentVar),
			"agent_cluster":                StringValueOrNull(s.AgentCluster),
			"agent_cluster_var":            StringValueOrNull(s.AgentClusterVar),
			"trigger":                      StringValueOrNull(s.Trigger),
			"trigger_var":                  StringValueOrNull(s.TriggerVar),
			"override_trigger_time":        StringValueOrNull(s.OverrideTriggerTime),
			"override_trigger_date_time":   types.BoolValue(s.OverrideTriggerDateTime),
			"override_trigger_date_offset": StringValueOrNull(s.OverrideTriggerDateOffset),
			"vertex_selection":             types.BoolValue(s.VertexSelection),
			"vertices":                     vertexRefsFromAPI(s.Vertices),
			"notification_option":          StringValueOrNull(s.NotificationOption),
			"variables_unresolved":         types.BoolValue(s.VariablesUnresolved),
			"variables":                    TaskVariablesFromAPI(ctx, s.Variables),
		})
	}
	result, _ := types.ListValue(types.ObjectType{AttrTypes: systemOperationAttrTypes()}, values)
	return result
}

// ---------------------------------------------------------------------------
// Actions (top-level wrapper: the `actions` attribute on every task type)
// ---------------------------------------------------------------------------

// ActionsModel describes the `actions` attribute in Terraform.
type ActionsModel struct {
	SystemOperations   types.List `tfsdk:"system_operations"`
	EmailNotifications types.List `tfsdk:"email_notifications"`
	AbortActions       types.List `tfsdk:"abort_actions"`
	SetVariableActions types.List `tfsdk:"set_variable_actions"`
	SnmpNotifications  types.List `tfsdk:"snmp_notifications"`
}

// ActionsAPIModel represents the `actions` object in the API.
type ActionsAPIModel struct {
	SystemOperations   []SystemOperationAPIModel   `json:"systemOperations,omitempty"`
	EmailNotifications []EmailNotificationAPIModel `json:"emailNotifications,omitempty"`
	AbortActions       []AbortActionAPIModel       `json:"abortActions,omitempty"`
	SetVariableActions []SetVariableActionAPIModel `json:"setVariableActions,omitempty"`
	SnmpNotifications  []SnmpNotificationAPIModel  `json:"snmpNotifications,omitempty"`
}

// ActionsAttrTypes returns the attribute types for ActionsModel.
func ActionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"system_operations":    types.ListType{ElemType: types.ObjectType{AttrTypes: systemOperationAttrTypes()}},
		"email_notifications":  types.ListType{ElemType: types.ObjectType{AttrTypes: emailNotificationAttrTypes()}},
		"abort_actions":        types.ListType{ElemType: types.ObjectType{AttrTypes: abortActionAttrTypes()}},
		"set_variable_actions": types.ListType{ElemType: types.ObjectType{AttrTypes: setVariableActionAttrTypes()}},
		"snmp_notifications":   types.ListType{ElemType: types.ObjectType{AttrTypes: snmpNotificationAttrTypes()}},
	}
}

// TaskActionsSchema returns the schema for the `actions` attribute, shared by
// every task resource type.
func TaskActionsSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Actions to trigger based on the task's status or exit code: launching/controlling other tasks, sending email or SNMP notifications, or setting variables.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"system_operations":    systemOperationsSchema(),
			"email_notifications":  emailNotificationsSchema(),
			"abort_actions":        abortActionsSchema(),
			"set_variable_actions": setVariableActionsSchema(),
			"snmp_notifications":   snmpNotificationsSchema(),
		},
	}
}

// TaskActionsToAPI converts the Terraform `actions` object to an API model.
func TaskActionsToAPI(ctx context.Context, actions types.Object) *ActionsAPIModel {
	if actions.IsNull() || actions.IsUnknown() {
		return nil
	}

	var m ActionsModel
	actions.As(ctx, &m, basetypes.ObjectAsOptions{})

	return &ActionsAPIModel{
		SystemOperations:   systemOperationsToAPI(ctx, m.SystemOperations),
		EmailNotifications: emailNotificationsToAPI(ctx, m.EmailNotifications),
		AbortActions:       abortActionsToAPI(ctx, m.AbortActions),
		SetVariableActions: setVariableActionsToAPI(ctx, m.SetVariableActions),
		SnmpNotifications:  snmpNotificationsToAPI(ctx, m.SnmpNotifications),
	}
}

// TaskActionsFromAPI converts an API actions model to the Terraform `actions`
// object. prior is the actions value from the plan/prior state (i.e.
// data.Actions as it stood before this call) and is used only to reorder
// each sub-list to match what was submitted, since the UAC API does not
// preserve submission order for these list-typed attributes (see
// reorderToMatchPrior).
func TaskActionsFromAPI(ctx context.Context, api *ActionsAPIModel, prior types.Object) types.Object {
	// The API always returns an "actions" object, even when nothing is
	// configured (its sub-lists are simply empty in that case). Treat that
	// as equivalent to no actions at all so that an unconfigured `actions`
	// block round-trips to null instead of an object full of null lists.
	if api == nil || (len(api.SystemOperations) == 0 && len(api.EmailNotifications) == 0 &&
		len(api.AbortActions) == 0 && len(api.SetVariableActions) == 0 && len(api.SnmpNotifications) == 0) {
		return types.ObjectNull(ActionsAttrTypes())
	}

	priorModel := ActionsModel{
		SystemOperations:   types.ListNull(types.ObjectType{AttrTypes: systemOperationAttrTypes()}),
		EmailNotifications: types.ListNull(types.ObjectType{AttrTypes: emailNotificationAttrTypes()}),
		AbortActions:       types.ListNull(types.ObjectType{AttrTypes: abortActionAttrTypes()}),
		SetVariableActions: types.ListNull(types.ObjectType{AttrTypes: setVariableActionAttrTypes()}),
		SnmpNotifications:  types.ListNull(types.ObjectType{AttrTypes: snmpNotificationAttrTypes()}),
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		prior.As(ctx, &priorModel, basetypes.ObjectAsOptions{})
	}

	obj, _ := types.ObjectValue(ActionsAttrTypes(), map[string]attr.Value{
		"system_operations":    systemOperationsFromAPIOrdered(ctx, api.SystemOperations, priorModel.SystemOperations),
		"email_notifications":  emailNotificationsFromAPIOrdered(ctx, api.EmailNotifications, priorModel.EmailNotifications),
		"abort_actions":        abortActionsFromAPIOrdered(ctx, api.AbortActions, priorModel.AbortActions),
		"set_variable_actions": setVariableActionsFromAPIOrdered(ctx, api.SetVariableActions, priorModel.SetVariableActions),
		"snmp_notifications":   snmpNotificationsFromAPIOrdered(ctx, api.SnmpNotifications, priorModel.SnmpNotifications),
	})
	return obj
}
