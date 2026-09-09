package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DataSourceResource{}
var _ resource.ResourceWithImportState = &DataSourceResource{}

type DataSourceResource struct {
	client *client.RoadieClient
}

type DataSourceResourceModel struct {
	ID           types.String         `tfsdk:"id"`
	Name         types.String         `tfsdk:"name"`
	Slug         types.String         `tfsdk:"slug"`
	Description  types.String         `tfsdk:"description"`
	WorkflowType types.String         `tfsdk:"workflow_type"`
	Nodes        jsontypes.Normalized `tfsdk:"nodes"`
	Edges        jsontypes.Normalized `tfsdk:"edges"`
	Enabled      types.Bool           `tfsdk:"enabled"`
	Adopted      types.Bool           `tfsdk:"adopted"`
	Version      types.Int64          `tfsdk:"version"`
	CreatedBy    types.String         `tfsdk:"created_by"`
	CreatedAt    types.String         `tfsdk:"created_at"`
	UpdatedAt    types.String         `tfsdk:"updated_at"`
}

type dataSourceAPI struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	Description  string          `json:"description"`
	WorkflowType string          `json:"workflowType"`
	Nodes        json.RawMessage `json:"nodes"`
	Edges        json.RawMessage `json:"edges"`
	Enabled      bool            `json:"enabled"`
	Version      int64           `json:"version"`
	CreatedBy    string          `json:"createdBy"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
}

type dataSourceInput struct {
	Name         string          `json:"name"`
	Slug         string          `json:"slug,omitempty"`
	Description  string          `json:"description"`
	WorkflowType string          `json:"workflowType"`
	Nodes        json.RawMessage `json:"nodes,omitempty"`
	Edges        json.RawMessage `json:"edges,omitempty"`
	Enabled      bool            `json:"enabled"`
}

func NewDataSourceResource() resource.Resource {
	return &DataSourceResource{}
}

func (r *DataSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_source"
}

func (r *DataSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Roadie data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"slug": schema.StringAttribute{
				Description: "URL-friendly identifier. Auto-derived from name if omitted.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"workflow_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("data-ingestion"),
				},
			},
			"nodes": schema.StringAttribute{
				Description: "Data source nodes as JSON. UI-only fields (width, height, selected, dragging) are stripped automatically.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"edges": schema.StringAttribute{
				Description: "Data source edges as JSON. UI-only fields (animated, style) are stripped automatically.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"adopted": schema.BoolAttribute{
				Description: "Whether this resource was adopted from a pre-existing data source rather than created by Terraform.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *DataSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.RoadieClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.RoadieClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *DataSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DataSourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := buildDataSourceInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Input", err.Error())
		return
	}

	result, createErr := client.CreateWrapped[dataSourceAPI](r.client, ctx, "/api/catalog-workflow/workflows", input)
	if createErr != nil {
		if strings.Contains(createErr.Error(), "already exists") {
			r.adoptExistingDataSource(ctx, &plan, input, resp)
			return
		}
		resp.Diagnostics.AddError("Error creating data source", createErr.Error())
		return
	}

	plan.Adopted = types.BoolValue(false)
	mapDataSourceToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DataSourceResource) adoptExistingDataSource(ctx context.Context, plan *DataSourceResourceModel, input *dataSourceInput, resp *resource.CreateResponse) {
	body, err := r.client.Get(ctx, "/api/catalog-workflow/workflows?limit=500")
	if err != nil {
		resp.Diagnostics.AddError("Error listing data sources", err.Error())
		return
	}
	var listResp struct {
		Data []struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		resp.Diagnostics.AddError("Error decoding data sources list", err.Error())
		return
	}

	slug := plan.Slug.ValueString()
	var existingID string
	for _, ds := range listResp.Data {
		if ds.Slug == slug {
			existingID = ds.ID
			break
		}
	}
	if existingID == "" {
		resp.Diagnostics.AddError("Error adopting data source", fmt.Sprintf("Data source with slug %q reported as existing but not found", slug))
		return
	}

	result, err := client.UpdateWrapped[dataSourceAPI](r.client, ctx, "/api/catalog-workflow/workflows/"+existingID, input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating adopted data source", err.Error())
		return
	}

	plan.Adopted = types.BoolValue(true)
	mapDataSourceToState(result, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DataSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DataSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetWrapped[dataSourceAPI](r.client, ctx, "/api/catalog-workflow/workflows/"+state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading data source", err.Error())
		return
	}

	mapDataSourceToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DataSourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := buildDataSourceInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Input", err.Error())
		return
	}

	result, err := client.UpdateWrapped[dataSourceAPI](r.client, ctx, "/api/catalog-workflow/workflows/"+plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating data source", err.Error())
		return
	}

	mapDataSourceToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DataSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DataSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Adopted.ValueBool() {
		return
	}

	err := r.client.Delete(ctx, "/api/catalog-workflow/workflows/"+state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting data source", err.Error())
	}
}

func (r *DataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildDataSourceInput(plan *DataSourceResourceModel) (*dataSourceInput, error) {
	input := &dataSourceInput{
		Name:         plan.Name.ValueString(),
		Slug:         plan.Slug.ValueString(),
		Description:  plan.Description.ValueString(),
		WorkflowType: plan.WorkflowType.ValueString(),
		Enabled:      plan.Enabled.ValueBool(),
	}

	if !plan.Nodes.IsNull() && !plan.Nodes.IsUnknown() {
		raw := json.RawMessage(plan.Nodes.ValueString())
		input.Nodes = raw
	}
	if !plan.Edges.IsNull() && !plan.Edges.IsUnknown() {
		raw := json.RawMessage(plan.Edges.ValueString())
		input.Edges = raw
	}

	return input, nil
}

func mapDataSourceToState(api *dataSourceAPI, state *DataSourceResourceModel) {
	state.ID = types.StringValue(api.ID)
	state.Name = types.StringValue(api.Name)
	state.Slug = types.StringValue(api.Slug)
	state.Description = types.StringValue(api.Description)
	state.WorkflowType = types.StringValue(api.WorkflowType)
	state.Enabled = types.BoolValue(api.Enabled)
	state.Version = types.Int64Value(api.Version)
	state.CreatedBy = types.StringValue(api.CreatedBy)
	state.CreatedAt = types.StringValue(api.CreatedAt)
	state.UpdatedAt = types.StringValue(api.UpdatedAt)

	if api.Nodes != nil && string(api.Nodes) != "[]" && string(api.Nodes) != "null" {
		state.Nodes = jsontypes.NewNormalizedValue(string(stripNodeUIFields(api.Nodes)))
	} else if state.Nodes.IsUnknown() {
		state.Nodes = jsontypes.NewNormalizedNull()
	}
	if api.Edges != nil && string(api.Edges) != "[]" && string(api.Edges) != "null" {
		state.Edges = jsontypes.NewNormalizedValue(string(stripEdgeUIFields(api.Edges)))
	} else if state.Edges.IsUnknown() {
		state.Edges = jsontypes.NewNormalizedNull()
	}
}

func stripNodeUIFields(raw json.RawMessage) json.RawMessage {
	var nodes []map[string]interface{}
	if err := json.Unmarshal(raw, &nodes); err != nil {
		return raw
	}
	for _, node := range nodes {
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
