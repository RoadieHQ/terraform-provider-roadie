package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DataSourceDataSource{}

type DataSourceDataSource struct {
	client *client.RoadieClient
}

type DataSourceDataSourceModel struct {
	ID           types.String         `tfsdk:"id"`
	Name         types.String         `tfsdk:"name"`
	Description  types.String         `tfsdk:"description"`
	WorkflowType types.String         `tfsdk:"workflow_type"`
	Nodes        jsontypes.Normalized `tfsdk:"nodes"`
	Edges        jsontypes.Normalized `tfsdk:"edges"`
	Enabled      types.Bool           `tfsdk:"enabled"`
	Version      types.Int64          `tfsdk:"version"`
	CreatedAt    types.String         `tfsdk:"created_at"`
	UpdatedAt    types.String         `tfsdk:"updated_at"`
}

type dataSourceReadAPI struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	WorkflowType string          `json:"workflowType"`
	Nodes        json.RawMessage `json:"nodes"`
	Edges        json.RawMessage `json:"edges"`
	Enabled      bool            `json:"enabled"`
	Version      int64           `json:"version"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
}

func NewDataSourceDataSource() datasource.DataSource {
	return &DataSourceDataSource{}
}

func (d *DataSourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_source"
}

func (d *DataSourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Roadie data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the data source.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"workflow_type": schema.StringAttribute{
				Computed: true,
			},
			"nodes": schema.StringAttribute{
				Description: "The data source nodes as JSON. UI-only fields are stripped.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"edges": schema.StringAttribute{
				Description: "The data source edges as JSON. UI-only fields are stripped.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
			"version": schema.Int64Attribute{
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

func (d *DataSourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetWrapped[dataSourceReadAPI](d.client, ctx, "/api/catalog-workflow/workflows/"+config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading data source", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Name = types.StringValue(result.Name)
	config.Description = types.StringValue(result.Description)
	config.WorkflowType = types.StringValue(result.WorkflowType)
	config.Enabled = types.BoolValue(result.Enabled)
	config.Version = types.Int64Value(result.Version)
	config.CreatedAt = types.StringValue(result.CreatedAt)
	config.UpdatedAt = types.StringValue(result.UpdatedAt)

	if result.Nodes != nil {
		config.Nodes = jsontypes.NewNormalizedValue(string(stripNodeUIFields(result.Nodes)))
	} else {
		config.Nodes = jsontypes.NewNormalizedNull()
	}
	if result.Edges != nil {
		config.Edges = jsontypes.NewNormalizedValue(string(stripEdgeUIFields(result.Edges)))
	} else {
		config.Edges = jsontypes.NewNormalizedNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func stripNodeUIFields(raw json.RawMessage) json.RawMessage {
	var nodes []map[string]interface{}
	if err := json.Unmarshal(raw, &nodes); err != nil {
		return raw
	}
	for _, node := range nodes {
		delete(node, "position")
		delete(node, "width")
		delete(node, "height")
		delete(node, "selected")
		delete(node, "dragging")
		delete(node, "positionAbsolute")
	}
	out, err := json.Marshal(nodes)
	if err != nil {
		return raw
	}
	return out
}

func stripEdgeUIFields(raw json.RawMessage) json.RawMessage {
	var edges []map[string]interface{}
	if err := json.Unmarshal(raw, &edges); err != nil {
		return raw
	}
	for _, edge := range edges {
		delete(edge, "animated")
		delete(edge, "style")
	}
	out, err := json.Marshal(edges)
	if err != nil {
		return raw
	}
	return out
}
