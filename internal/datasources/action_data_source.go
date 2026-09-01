package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ActionDataSource{}

type ActionDataSource struct {
	client *client.RoadieClient
}

type ActionDataSourceModel struct {
	ID             types.String         `tfsdk:"id"`
	Slug           types.String         `tfsdk:"slug"`
	Name           types.String         `tfsdk:"name"`
	Description    types.String         `tfsdk:"description"`
	Enabled        types.Bool           `tfsdk:"enabled"`
	Mode           types.String         `tfsdk:"mode"`
	EffectiveMode  types.String         `tfsdk:"effective_mode"`
	CurrentVersion types.Int64          `tfsdk:"current_version"`
	InputSchema    jsontypes.Normalized `tfsdk:"input_schema"`
	CreatedAt      types.String         `tfsdk:"created_at"`
	UpdatedAt      types.String         `tfsdk:"updated_at"`
}

type actionDataSourceAPI struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Slug           string          `json:"slug"`
	Description    string          `json:"description"`
	Enabled        bool            `json:"enabled"`
	Mode           *string         `json:"mode"`
	EffectiveMode  string          `json:"effectiveMode"`
	CurrentVersion int64           `json:"currentVersion"`
	InputSchema    json.RawMessage `json:"inputSchema"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
}

func NewActionDataSource() datasource.DataSource {
	return &ActionDataSource{}
}

func (d *ActionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (d *ActionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Roadie action.",
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
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
			"mode": schema.StringAttribute{
				Computed: true,
			},
			"effective_mode": schema.StringAttribute{
				Computed: true,
			},
			"current_version": schema.Int64Attribute{
				Computed: true,
			},
			"input_schema": schema.StringAttribute{
				Description: "The computed JSON Schema for action inputs.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
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

func (d *ActionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ActionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := config.ID.ValueString()
	if identifier == "" {
		identifier = config.Slug.ValueString()
	}

	result, err := client.GetBare[actionDataSourceAPI](d.client, ctx, "/api/actions/"+identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error reading action", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Slug = types.StringValue(result.Slug)
	config.Name = types.StringValue(result.Name)
	config.Description = types.StringValue(result.Description)
	config.Enabled = types.BoolValue(result.Enabled)
	config.EffectiveMode = types.StringValue(result.EffectiveMode)
	config.CurrentVersion = types.Int64Value(result.CurrentVersion)
	config.CreatedAt = types.StringValue(result.CreatedAt)
	config.UpdatedAt = types.StringValue(result.UpdatedAt)

	if result.Mode != nil {
		config.Mode = types.StringValue(*result.Mode)
	} else {
		config.Mode = types.StringNull()
	}

	if result.InputSchema != nil && string(result.InputSchema) != "null" {
		config.InputSchema = jsontypes.NewNormalizedValue(string(result.InputSchema))
	} else {
		config.InputSchema = jsontypes.NewNormalizedNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
