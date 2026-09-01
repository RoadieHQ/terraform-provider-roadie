package datasources

import (
	"context"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RelationshipRuleDataSource{}

type RelationshipRuleDataSource struct {
	client *client.RoadieClient
}

type RelationshipRuleDataSourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	SourceDatasourceID         types.String `tfsdk:"source_datasource_id"`
	TargetDatasourceID         types.String `tfsdk:"target_datasource_id"`
	SourceFieldExpression      types.String `tfsdk:"source_field_expression"`
	TargetFieldExpression      types.String `tfsdk:"target_field_expression"`
	RelationshipType           types.String `tfsdk:"relationship_type"`
	ReciprocalRelationshipType types.String `tfsdk:"reciprocal_relationship_type"`
	Strategy                   types.String `tfsdk:"strategy"`
	MatchStrategy              types.String `tfsdk:"match_strategy"`
	Origin                     types.String `tfsdk:"origin"`
	State                      types.String `tfsdk:"state"`
	CreatedAt                  types.String `tfsdk:"created_at"`
	UpdatedAt                  types.String `tfsdk:"updated_at"`
}

type relationshipRuleDataSourceAPI struct {
	ID                         string  `json:"id"`
	Name                       string  `json:"name"`
	Description                *string `json:"description"`
	SourceDatasourceID         string  `json:"sourceDatasourceId"`
	TargetDatasourceID         string  `json:"targetDatasourceId"`
	SourceFieldExpression      string  `json:"sourceFieldExpression"`
	TargetFieldExpression      string  `json:"targetFieldExpression"`
	RelationshipType           string  `json:"relationshipType"`
	ReciprocalRelationshipType *string `json:"reciprocalRelationshipType"`
	Strategy                   string  `json:"strategy"`
	MatchStrategy              string  `json:"matchStrategy"`
	Origin                     string  `json:"origin"`
	State                      string  `json:"state"`
	CreatedAt                  string  `json:"createdAt"`
	UpdatedAt                  string  `json:"updatedAt"`
}

func NewRelationshipRuleDataSource() datasource.DataSource {
	return &RelationshipRuleDataSource{}
}

func (d *RelationshipRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_relationship_rule"
}

func (d *RelationshipRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Roadie relationship rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The UUID of the relationship rule.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"source_datasource_id": schema.StringAttribute{
				Computed: true,
			},
			"target_datasource_id": schema.StringAttribute{
				Computed: true,
			},
			"source_field_expression": schema.StringAttribute{
				Computed: true,
			},
			"target_field_expression": schema.StringAttribute{
				Computed: true,
			},
			"relationship_type": schema.StringAttribute{
				Computed: true,
			},
			"reciprocal_relationship_type": schema.StringAttribute{
				Computed: true,
			},
			"strategy": schema.StringAttribute{
				Computed: true,
			},
			"match_strategy": schema.StringAttribute{
				Computed: true,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
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

func (d *RelationshipRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RelationshipRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RelationshipRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetBare[relationshipRuleDataSourceAPI](d.client, ctx, "/api/catalog-datastore/relationship-rules/"+config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading relationship rule", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Name = types.StringValue(result.Name)
	config.SourceDatasourceID = types.StringValue(result.SourceDatasourceID)
	config.TargetDatasourceID = types.StringValue(result.TargetDatasourceID)
	config.SourceFieldExpression = types.StringValue(result.SourceFieldExpression)
	config.TargetFieldExpression = types.StringValue(result.TargetFieldExpression)
	config.RelationshipType = types.StringValue(result.RelationshipType)
	config.Strategy = types.StringValue(result.Strategy)
	config.MatchStrategy = types.StringValue(result.MatchStrategy)
	config.Origin = types.StringValue(result.Origin)
	config.State = types.StringValue(result.State)
	config.CreatedAt = types.StringValue(result.CreatedAt)
	config.UpdatedAt = types.StringValue(result.UpdatedAt)

	if result.Description != nil {
		config.Description = types.StringValue(*result.Description)
	} else {
		config.Description = types.StringNull()
	}
	if result.ReciprocalRelationshipType != nil {
		config.ReciprocalRelationshipType = types.StringValue(*result.ReciprocalRelationshipType)
	} else {
		config.ReciprocalRelationshipType = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
