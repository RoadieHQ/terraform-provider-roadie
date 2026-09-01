package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ContextGroupResource{}
	_ resource.ResourceWithImportState = &ContextGroupResource{}
)

type ContextGroupResource struct {
	client *client.RoadieClient
}

type ContextGroupResourceModel struct {
	ID                       types.String              `tfsdk:"id"`
	Name                     types.String              `tfsdk:"name"`
	Slug                     types.String              `tfsdk:"slug"`
	Description              types.String              `tfsdk:"description"`
	Datasources              []DatasourceFilterModel   `tfsdk:"datasources"`
	MergeRelationshipTypes   types.List                `tfsdk:"merge_relationship_types"`
	Annotations              []AnnotationModel         `tfsdk:"annotations"`
	IncludeExternalRelations types.Bool                `tfsdk:"include_external_relations"`
	SeedVersion              types.Int64               `tfsdk:"seed_version"`
	CreatedAt                types.String              `tfsdk:"created_at"`
	UpdatedAt                types.String              `tfsdk:"updated_at"`
}

type DatasourceFilterModel struct {
	DatasourceID types.String         `tfsdk:"datasource_id"`
	SeedName     types.String         `tfsdk:"seed_name"`
	Filter       types.String         `tfsdk:"filter"`
	Projection   jsontypes.Normalized `tfsdk:"projection"`
	Annotation   *AnnotationModel     `tfsdk:"annotation"`
}

type AnnotationModel struct {
	Title types.String `tfsdk:"title"`
	Text  types.String `tfsdk:"text"`
}

type contextGroupAPI struct {
	ID                       string                `json:"id"`
	Name                     string                `json:"name"`
	Slug                     string                `json:"slug"`
	Description              *string               `json:"description"`
	Datasources              []datasourceFilterAPI `json:"datasources"`
	MergeRelationshipTypes   []string              `json:"mergeRelationshipTypes"`
	Annotations              []annotationAPI       `json:"annotations"`
	IncludeExternalRelations bool                  `json:"includeExternalRelations"`
	SeedVersion              *int64                `json:"seedVersion"`
	CreatedAt                string                `json:"createdAt"`
	UpdatedAt                string                `json:"updatedAt"`
}

type datasourceFilterAPI struct {
	DatasourceID string          `json:"datasourceId,omitempty"`
	SeedName     string          `json:"seedName,omitempty"`
	Filter       string          `json:"filter,omitempty"`
	Projection   json.RawMessage `json:"projection,omitempty"`
	Annotation   *annotationAPI  `json:"annotation,omitempty"`
	Status       json.RawMessage `json:"status,omitempty"`
}

type annotationAPI struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type contextGroupInput struct {
	Name                     string                `json:"name"`
	Slug                     string                `json:"slug,omitempty"`
	Description              string                `json:"description,omitempty"`
	Datasources              []datasourceFilterAPI `json:"datasources"`
	MergeRelationshipTypes   []string              `json:"mergeRelationshipTypes,omitempty"`
	Annotations              []annotationAPI       `json:"annotations,omitempty"`
	IncludeExternalRelations bool                  `json:"includeExternalRelations"`
	SeedVersion              *int64                `json:"seedVersion,omitempty"`
}

func NewContextGroupResource() resource.Resource {
	return &ContextGroupResource{}
}

func (r *ContextGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context_group"
}

func (r *ContextGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Roadie context group.",
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
			},
			"datasources": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"datasource_id": schema.StringAttribute{
							Optional: true,
						},
						"seed_name": schema.StringAttribute{
							Optional: true,
						},
						"filter": schema.StringAttribute{
							Optional: true,
						},
						"projection": schema.StringAttribute{
							Description: "Projection configuration as JSON (include/exclude).",
							Optional:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
						"annotation": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"title": schema.StringAttribute{
									Required: true,
								},
								"text": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},
			"merge_relationship_types": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"annotations": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"title": schema.StringAttribute{
							Required: true,
						},
						"text": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"include_external_relations": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"seed_version": schema.Int64Attribute{
				Optional: true,
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

func (r *ContextGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContextGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContextGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildContextGroupInput(ctx, &plan)

	result, err := client.CreateBare[contextGroupAPI](r.client, ctx, "/api/catalog-datastore/context-groups/rules", input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating context group", err.Error())
		return
	}

	mapContextGroupToState(ctx, result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContextGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContextGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetBare[contextGroupAPI](r.client, ctx, "/api/catalog-datastore/context-groups/rules/"+state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading context group", err.Error())
		return
	}

	mapContextGroupToState(ctx, result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContextGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ContextGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildContextGroupInput(ctx, &plan)

	result, err := client.PatchBare[contextGroupAPI](r.client, ctx, "/api/catalog-datastore/context-groups/rules/"+plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating context group", err.Error())
		return
	}

	mapContextGroupToState(ctx, result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContextGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContextGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/api/catalog-datastore/context-groups/rules/"+state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting context group", err.Error())
	}
}

func (r *ContextGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildContextGroupInput(ctx context.Context, plan *ContextGroupResourceModel) *contextGroupInput {
	input := &contextGroupInput{
		Name:                     plan.Name.ValueString(),
		Description:              plan.Description.ValueString(),
		IncludeExternalRelations: plan.IncludeExternalRelations.ValueBool(),
	}

	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		input.Slug = plan.Slug.ValueString()
	}

	if !plan.SeedVersion.IsNull() && !plan.SeedVersion.IsUnknown() {
		v := plan.SeedVersion.ValueInt64()
		input.SeedVersion = &v
	}

	input.Datasources = make([]datasourceFilterAPI, len(plan.Datasources))
	for i, ds := range plan.Datasources {
		input.Datasources[i] = datasourceFilterAPI{
			DatasourceID: ds.DatasourceID.ValueString(),
			SeedName:     ds.SeedName.ValueString(),
			Filter:       ds.Filter.ValueString(),
		}
		if !ds.Projection.IsNull() && !ds.Projection.IsUnknown() {
			input.Datasources[i].Projection = json.RawMessage(ds.Projection.ValueString())
		}
		if ds.Annotation != nil {
			input.Datasources[i].Annotation = &annotationAPI{
				Title: ds.Annotation.Title.ValueString(),
				Text:  ds.Annotation.Text.ValueString(),
			}
		}
	}

	if plan.Annotations != nil {
		input.Annotations = make([]annotationAPI, len(plan.Annotations))
		for i, a := range plan.Annotations {
			input.Annotations[i] = annotationAPI{
				Title: a.Title.ValueString(),
				Text:  a.Text.ValueString(),
			}
		}
	}

	if !plan.MergeRelationshipTypes.IsNull() && !plan.MergeRelationshipTypes.IsUnknown() {
		var mergeTypes []string
		plan.MergeRelationshipTypes.ElementsAs(ctx, &mergeTypes, false)
		input.MergeRelationshipTypes = mergeTypes
	}

	return input
}

func mapContextGroupToState(ctx context.Context, api *contextGroupAPI, state *ContextGroupResourceModel) {
	state.ID = types.StringValue(api.ID)
	state.Name = types.StringValue(api.Name)
	state.Slug = types.StringValue(api.Slug)
	state.IncludeExternalRelations = types.BoolValue(api.IncludeExternalRelations)
	state.CreatedAt = types.StringValue(api.CreatedAt)
	state.UpdatedAt = types.StringValue(api.UpdatedAt)

	if api.Description != nil {
		state.Description = types.StringValue(*api.Description)
	} else {
		state.Description = types.StringNull()
	}

	if api.SeedVersion != nil {
		state.SeedVersion = types.Int64Value(*api.SeedVersion)
	} else {
		state.SeedVersion = types.Int64Null()
	}

	state.Datasources = make([]DatasourceFilterModel, len(api.Datasources))
	for i, ds := range api.Datasources {
		state.Datasources[i] = DatasourceFilterModel{
			DatasourceID: optionalString(ds.DatasourceID),
			SeedName:     optionalString(ds.SeedName),
			Filter:       optionalString(ds.Filter),
		}
		if ds.Projection != nil && string(ds.Projection) != "null" {
			state.Datasources[i].Projection = jsontypes.NewNormalizedValue(string(ds.Projection))
		} else {
			state.Datasources[i].Projection = jsontypes.NewNormalizedNull()
		}
		if ds.Annotation != nil {
			state.Datasources[i].Annotation = &AnnotationModel{
				Title: types.StringValue(ds.Annotation.Title),
				Text:  types.StringValue(ds.Annotation.Text),
			}
		}
	}

	if len(api.Annotations) > 0 {
		state.Annotations = make([]AnnotationModel, len(api.Annotations))
		for i, a := range api.Annotations {
			state.Annotations[i] = AnnotationModel{
				Title: types.StringValue(a.Title),
				Text:  types.StringValue(a.Text),
			}
		}
	} else {
		state.Annotations = nil
	}

	if len(api.MergeRelationshipTypes) > 0 {
		listVal, _ := types.ListValueFrom(ctx, types.StringType, api.MergeRelationshipTypes)
		state.MergeRelationshipTypes = listVal
	} else {
		state.MergeRelationshipTypes = types.ListNull(types.StringType)
	}
}

func optionalString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
