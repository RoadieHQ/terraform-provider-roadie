package provider

import (
	"context"
	"os"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/RoadieHQ/terraform-provider-roadie/internal/datasources"
	"github.com/RoadieHQ/terraform-provider-roadie/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &RoadieProvider{}

type RoadieProvider struct {
	version string
}

type RoadieProviderModel struct {
	Host     types.String `tfsdk:"host"`
	ApiToken types.String `tfsdk:"api_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RoadieProvider{version: version}
	}
}

func (p *RoadieProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "roadie"
	resp.Version = p.version
}

func (p *RoadieProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Roadie developer portal resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The Roadie backend URL (e.g. https://api.roadie.so). Can also be set via ROADIE_HOST environment variable.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "Service token for authentication (format: rst_...). Can also be set via ROADIE_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *RoadieProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config RoadieProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("ROADIE_HOST")
	if !config.Host.IsNull() && !config.Host.IsUnknown() {
		host = config.Host.ValueString()
	}

	apiToken := os.Getenv("ROADIE_API_TOKEN")
	if !config.ApiToken.IsNull() && !config.ApiToken.IsUnknown() {
		apiToken = config.ApiToken.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Missing Roadie Host",
			"The provider cannot create the Roadie API client as there is a missing or empty value for the Roadie host. "+
				"Set the host value in the configuration or use the ROADIE_HOST environment variable.",
		)
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing Roadie API Token",
			"The provider cannot create the Roadie API client as there is a missing or empty value for the API token. "+
				"Set the api_token value in the configuration or use the ROADIE_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(host, apiToken, p.version)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *RoadieProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewCapabilityResource,
		resources.NewActionResource,
		resources.NewIntegrationResource,
		resources.NewDataSourceResource,
		resources.NewRelationshipRuleResource,
		resources.NewContextGroupResource,
		resources.NewSecretResource,
	}
}

func (p *RoadieProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewCapabilityDataSource,
		datasources.NewActionDataSource,
		datasources.NewIntegrationDataSource,
		datasources.NewDataSourceDataSource,
		datasources.NewRelationshipRuleDataSource,
		datasources.NewContextGroupDataSource,
		datasources.NewSecretDataSource,
	}
}
