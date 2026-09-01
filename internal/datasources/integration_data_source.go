package datasources

import (
	"context"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IntegrationDataSource{}

type IntegrationDataSource struct {
	client *client.RoadieClient
}

type IntegrationDataSourceModel struct {
	ID              types.String  `tfsdk:"id"`
	Slug            types.String  `tfsdk:"slug"`
	Name            types.String  `tfsdk:"name"`
	Type            types.String  `tfsdk:"type"`
	Host            types.String  `tfsdk:"host"`
	AuthType        types.String  `tfsdk:"auth_type"`
	BackendType     types.String  `tfsdk:"backend_type"`
	Enabled         types.Bool    `tfsdk:"enabled"`
	RequestsPerHour types.Float64 `tfsdk:"requests_per_hour"`
	RequestsPerSec  types.Float64 `tfsdk:"requests_per_second"`
	BurstCapacity   types.Float64 `tfsdk:"burst_capacity"`
	CreatedAt       types.String  `tfsdk:"created_at"`
	UpdatedAt       types.String  `tfsdk:"updated_at"`
}

type integrationDataSourceAPI struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Type            string  `json:"type"`
	Host            string  `json:"host"`
	AuthType        string  `json:"authType"`
	BackendType     string  `json:"backendType"`
	Enabled         bool    `json:"enabled"`
	RequestsPerHour float64 `json:"requestsPerHour"`
	RequestsPerSec  float64 `json:"requestsPerSecond"`
	BurstCapacity   float64 `json:"burstCapacity"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func NewIntegrationDataSource() datasource.DataSource {
	return &IntegrationDataSource{}
}

func (d *IntegrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (d *IntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Roadie integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier. Provide either id or slug.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("slug")),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The slug identifier. Provide either id or slug.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"host": schema.StringAttribute{
				Computed: true,
			},
			"auth_type": schema.StringAttribute{
				Computed: true,
			},
			"backend_type": schema.StringAttribute{
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
			"requests_per_hour": schema.Float64Attribute{
				Computed: true,
			},
			"requests_per_second": schema.Float64Attribute{
				Computed: true,
			},
			"burst_capacity": schema.Float64Attribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *IntegrationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.RoadieClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.RoadieClient, got: %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *IntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IntegrationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := config.ID.ValueString()
	if identifier == "" {
		identifier = config.Slug.ValueString()
	}

	result, err := client.GetWrapped[integrationDataSourceAPI](d.client, ctx, "/api/integrations/"+identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Slug = types.StringValue(result.Slug)
	config.Name = types.StringValue(result.Name)
	config.Type = types.StringValue(result.Type)
	config.Host = types.StringValue(result.Host)
	config.AuthType = types.StringValue(result.AuthType)
	config.BackendType = types.StringValue(result.BackendType)
	config.Enabled = types.BoolValue(result.Enabled)
	config.RequestsPerHour = types.Float64Value(result.RequestsPerHour)
	config.RequestsPerSec = types.Float64Value(result.RequestsPerSec)
	config.BurstCapacity = types.Float64Value(result.BurstCapacity)
	config.CreatedAt = types.StringValue(result.CreatedAt)
	config.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
