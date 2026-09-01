package datasources

import (
	"context"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SecretDataSource{}

type SecretDataSource struct {
	client *client.RoadieClient
}

type SecretDataSourceModel struct {
	Ref         types.String `tfsdk:"ref"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	HelpURL     types.String `tfsdk:"help_url"`
	Status      types.String `tfsdk:"status"`
}

type secretDataSourceAPI struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	HelpURL      string `json:"helpUrl"`
	IsCustom     bool   `json:"isCustom"`
	CreatedAt    string `json:"createdAt"`
	LastModified string `json:"lastModified"`
}

func NewSecretDataSource() datasource.DataSource {
	return &SecretDataSource{}
}

func (d *SecretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *SecretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Roadie secret's metadata and status.",
		Attributes: map[string]schema.Attribute{
			"ref": schema.StringAttribute{
				Description: "The secret reference key (e.g. GITHUB_TOKEN).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Human-friendly display name.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the secret.",
				Computed:    true,
			},
			"help_url": schema.StringAttribute{
				Description: "URL linking to documentation about obtaining this secret.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status: Available or Not Set.",
				Computed:    true,
			},
		},
	}
}

func (d *SecretDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SecretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetBare[secretDataSourceAPI](d.client, ctx, "/api/secrets-settings/secret-value/"+config.Ref.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret", err.Error())
		return
	}

	config.Name = types.StringValue(result.Name)
	config.Status = types.StringValue(result.Status)
	if result.Description != "" {
		config.Description = types.StringValue(result.Description)
	} else {
		config.Description = types.StringNull()
	}
	if result.HelpURL != "" {
		config.HelpURL = types.StringValue(result.HelpURL)
	} else {
		config.HelpURL = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
