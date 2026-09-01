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

var _ datasource.DataSource = &CapabilityDataSource{}

type CapabilityDataSource struct {
	client *client.RoadieClient
}

type CapabilityDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Slug           types.String `tfsdk:"slug"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Instructions   types.String `tfsdk:"instructions"`
	CurrentVersion types.Int64  `tfsdk:"current_version"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

type capabilityDataSourceAPI struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
	Instructions   string `json:"instructions"`
	CurrentVersion int64  `json:"currentVersion"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

func NewCapabilityDataSource() datasource.DataSource {
	return &CapabilityDataSource{}
}

func (d *CapabilityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_capability"
}

func (d *CapabilityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Roadie capability.",
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
			"description": schema.StringAttribute{
				Computed: true,
			},
			"instructions": schema.StringAttribute{
				Computed: true,
			},
			"current_version": schema.Int64Attribute{
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

func (d *CapabilityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CapabilityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CapabilityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := config.ID.ValueString()
	if identifier == "" {
		identifier = config.Slug.ValueString()
	}

	result, err := client.GetBare[capabilityDataSourceAPI](d.client, ctx, "/api/capabilities/"+identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error reading capability", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Slug = types.StringValue(result.Slug)
	config.Name = types.StringValue(result.Name)
	config.Description = types.StringValue(result.Description)
	config.Instructions = types.StringValue(result.Instructions)
	config.CurrentVersion = types.Int64Value(result.CurrentVersion)
	config.CreatedAt = types.StringValue(result.CreatedAt)
	config.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
