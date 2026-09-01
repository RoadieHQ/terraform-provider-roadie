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

var _ datasource.DataSource = &ContextGroupDataSource{}

type ContextGroupDataSource struct {
	client *client.RoadieClient
}

type ContextGroupDataSourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Slug                     types.String `tfsdk:"slug"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	IncludeExternalRelations types.Bool   `tfsdk:"include_external_relations"`
	SeedVersion              types.Int64  `tfsdk:"seed_version"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
}

type contextGroupDataSourceAPI struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	Slug                     string  `json:"slug"`
	Description              *string `json:"description"`
	IncludeExternalRelations bool    `json:"includeExternalRelations"`
	SeedVersion              *int64  `json:"seedVersion"`
	CreatedAt                string  `json:"createdAt"`
	UpdatedAt                string  `json:"updatedAt"`
}

func NewContextGroupDataSource() datasource.DataSource {
	return &ContextGroupDataSource{}
}

func (d *ContextGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context_group"
}

func (d *ContextGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Roadie context group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The UUID. Provide either id or slug.",
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
			"include_external_relations": schema.BoolAttribute{
				Computed: true,
			},
			"seed_version": schema.Int64Attribute{
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

func (d *ContextGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ContextGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ContextGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := config.ID.ValueString()
	if identifier == "" {
		identifier = config.Slug.ValueString()
	}

	result, err := client.GetBare[contextGroupDataSourceAPI](d.client, ctx, "/api/catalog-datastore/context-groups/rules/"+identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error reading context group", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Slug = types.StringValue(result.Slug)
	config.Name = types.StringValue(result.Name)
	config.IncludeExternalRelations = types.BoolValue(result.IncludeExternalRelations)
	config.CreatedAt = types.StringValue(result.CreatedAt)
	config.UpdatedAt = types.StringValue(result.UpdatedAt)

	if result.Description != nil {
		config.Description = types.StringValue(*result.Description)
	} else {
		config.Description = types.StringNull()
	}
	if result.SeedVersion != nil {
		config.SeedVersion = types.Int64Value(*result.SeedVersion)
	} else {
		config.SeedVersion = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
