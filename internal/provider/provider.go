package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/provider/data_sources"
	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/provider/resources"
)

// Ensure StonebranchProvider satisfies various provider interfaces.
var _ provider.Provider = &StonebranchProvider{}

// StonebranchProvider defines the provider implementation.
type StonebranchProvider struct {
	version string
}

// StonebranchProviderModel describes the provider data model.
type StonebranchProviderModel struct {
	APIToken types.String `tfsdk:"api_token"`
	BaseURL  types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StonebranchProvider{
			version: version,
		}
	}
}

func (p *StonebranchProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "stonebranch"
	resp.Version = p.version
}

func (p *StonebranchProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with StoneBranch Universal Controller API.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "Bearer token for StoneBranch API authentication. Can also be set via STONEBRANCH_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Base URL for the StoneBranch API. Can also be set via STONEBRANCH_BASE_URL environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *StonebranchProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring StoneBranch client")

	var config StonebranchProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values
	apiToken := os.Getenv("STONEBRANCH_API_TOKEN")
	baseURL := os.Getenv("STONEBRANCH_BASE_URL")

	// Override with config values if provided
	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}

	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	// Validate required configuration
	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing StoneBranch API Token",
			"The provider cannot create the StoneBranch API client as there is a missing or empty value for the StoneBranch API token. "+
				"Set the api_token value in the configuration or use the STONEBRANCH_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing StoneBranch Base URL",
			"The provider cannot create the StoneBranch API client as there is a missing or empty value for the base URL. "+
				"Set the base_url value in the configuration or use the STONEBRANCH_BASE_URL environment variable. "+
				"Example: https://your-instance.stonebranch.cloud",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating StoneBranch client", map[string]any{
		"base_url": baseURL,
	})

	// Create the API client
	apiClient := client.NewClient(baseURL, apiToken)

	// Make the client available to resources and data sources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient

	tflog.Info(ctx, "Configured StoneBranch client", map[string]any{
		"base_url": baseURL,
	})
}

func (p *StonebranchProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewTaskUnixResource,
		resources.NewTaskWindowsResource,
		resources.NewTaskFileTransferResource,
		resources.NewTaskSQLResource,
		resources.NewTaskEmailResource,
		resources.NewTaskWorkflowResource,
		resources.NewScriptResource,
		resources.NewTriggerTimeResource,
		resources.NewTriggerCronResource,
		resources.NewCredentialResource,
		resources.NewVariableResource,
		resources.NewDatabaseConnectionResource,
		resources.NewEmailConnectionResource,
		resources.NewWorkflowVertexResource,
		resources.NewWorkflowEdgeResource,
		resources.NewBusinessServiceResource,
		resources.NewEmailTemplateResource,
		resources.NewTriggerFileMonitorResource,
		resources.NewTaskFileMonitorResource,
		resources.NewCalendarResource,
		resources.NewCustomDayResource,
		resources.NewAgentClusterResource,
		resources.NewTriggerTaskMonitorResource,
		resources.NewTaskMonitorResource,
		resources.NewTaskStoredProcedureResource,
		resources.NewTaskWebServiceResource,
		resources.NewTaskTimerResource,
		resources.NewTaskUniversalAwsS3Resource,
		resources.NewTaskUniversalResource,
		resources.NewTaskRecurringResource,
		resources.NewUniversalTemplateResource,
		resources.NewVirtualResourceResource,
	}
}

func (p *StonebranchProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		data_sources.NewAgentsDataSource,
		data_sources.NewAgentClustersDataSource,
		data_sources.NewTasksDataSource,
		data_sources.NewTaskInstancesDataSource,
		data_sources.NewTaskDataSource,
		data_sources.NewTriggerDataSource,
	}
}
