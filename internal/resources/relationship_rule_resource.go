package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &RelationshipRuleResource{}
	_ resource.ResourceWithImportState = &RelationshipRuleResource{}
)

type RelationshipRuleResource struct {
	client *client.RoadieClient
}

type RelationshipRuleResourceModel struct {
	ID                         types.String         `tfsdk:"id"`
	Name                       types.String         `tfsdk:"name"`
	Description                types.String         `tfsdk:"description"`
	SourceDatasourceID         types.String         `tfsdk:"source_datasource_id"`
	TargetDatasourceID         types.String         `tfsdk:"target_datasource_id"`
	SourceFieldExpression      types.String         `tfsdk:"source_field_expression"`
	TargetFieldExpression      types.String         `tfsdk:"target_field_expression"`
	SourceFilterExpression     types.String         `tfsdk:"source_filter_expression"`
	TargetFilterExpression     types.String         `tfsdk:"target_filter_expression"`
	RelationshipType           types.String         `tfsdk:"relationship_type"`
	ReciprocalRelationshipType types.String         `tfsdk:"reciprocal_relationship_type"`
	Strategy                   types.String         `tfsdk:"strategy"`
	MatchStrategy              types.String         `tfsdk:"match_strategy"`
	IntegrationConfig          jsontypes.Normalized `tfsdk:"integration_config"`
	Origin                     types.String         `tfsdk:"origin"`
	State                      types.String         `tfsdk:"state"`
	CreatedAt                  types.String         `tfsdk:"created_at"`
	UpdatedAt                  types.String         `tfsdk:"updated_at"`
}

type relationshipRuleAPI struct {
	ID                         string          `json:"id"`
	Name                       string          `json:"name"`
	Description                *string         `json:"description"`
	SourceDatasourceID         string          `json:"sourceDatasourceId"`
	TargetDatasourceID         string          `json:"targetDatasourceId"`
	SourceFieldExpression      string          `json:"sourceFieldExpression"`
	TargetFieldExpression      string          `json:"targetFieldExpression"`
	SourceFilterExpression     *string         `json:"sourceFilterExpression"`
	TargetFilterExpression     *string         `json:"targetFilterExpression"`
	RelationshipType           string          `json:"relationshipType"`
	ReciprocalRelationshipType *string         `json:"reciprocalRelationshipType"`
	Strategy                   string          `json:"strategy"`
	MatchStrategy              string          `json:"matchStrategy"`
	IntegrationConfig          json.RawMessage `json:"integrationConfig"`
	Origin                     string          `json:"origin"`
	State                      string          `json:"state"`
	CreatedAt                  string          `json:"createdAt"`
	UpdatedAt                  string          `json:"updatedAt"`
}

type relationshipRuleInput struct {
	Name                       string          `json:"name"`
	Description                string          `json:"description,omitempty"`
	SourceDatasourceID         string          `json:"sourceDatasourceId"`
	TargetDatasourceID         string          `json:"targetDatasourceId"`
	SourceFieldExpression      string          `json:"sourceFieldExpression"`
	TargetFieldExpression      string          `json:"targetFieldExpression"`
	SourceFilterExpression     string          `json:"sourceFilterExpression,omitempty"`
	TargetFilterExpression     string          `json:"targetFilterExpression,omitempty"`
	RelationshipType           string          `json:"relationshipType"`
	ReciprocalRelationshipType string          `json:"reciprocalRelationshipType,omitempty"`
	Strategy                   string          `json:"strategy"`
	MatchStrategy              string          `json:"matchStrategy"`
	IntegrationConfig          json.RawMessage `json:"integrationConfig,omitempty"`
}

type stateInput struct {
	State string `json:"state"`
}

func NewRelationshipRuleResource() resource.Resource {
	return &RelationshipRuleResource{}
}

func (r *RelationshipRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_relationship_rule"
}

func (r *RelationshipRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Roadie relationship rule.",
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
			"description": schema.StringAttribute{
				Optional: true,
			},
			"source_datasource_id": schema.StringAttribute{
				Required: true,
			},
			"target_datasource_id": schema.StringAttribute{
				Required: true,
			},
			"source_field_expression": schema.StringAttribute{
				Required: true,
			},
			"target_field_expression": schema.StringAttribute{
				Required: true,
			},
			"source_filter_expression": schema.StringAttribute{
				Optional: true,
			},
			"target_filter_expression": schema.StringAttribute{
				Optional: true,
			},
			"relationship_type": schema.StringAttribute{
				Required: true,
			},
			"reciprocal_relationship_type": schema.StringAttribute{
				Optional: true,
			},
			"strategy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("field-matching"),
				Validators: []validator.String{
					stringvalidator.OneOf("field-matching", "integration-backed"),
				},
			},
			"match_strategy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("exact"),
				Validators: []validator.String{
					stringvalidator.OneOf("exact", "contains", "array_contains", "regex", "person_name_alias"),
				},
			},
			"integration_config": schema.StringAttribute{
				Description: "Integration configuration as JSON. Only used when strategy is integration-backed.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"origin": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("active"),
				Validators: []validator.String{
					stringvalidator.OneOf("suggested", "active", "inactive"),
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

func (r *RelationshipRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RelationshipRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RelationshipRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildRelationshipRuleInput(&plan)

	result, err := client.CreateBare[relationshipRuleAPI](r.client, ctx, "/api/catalog-datastore/relationship-rules", input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating relationship rule", err.Error())
		return
	}

	// Always set the desired state after creation
	desiredState := plan.State.ValueString()
	_, err = r.client.Put(ctx, "/api/catalog-datastore/relationship-rules/"+result.ID+"/state", stateInput{State: desiredState})
	if err != nil {
		resp.Diagnostics.AddError("Error setting relationship rule state", err.Error())
		return
	}

	final, err := client.GetBare[relationshipRuleAPI](r.client, ctx, "/api/catalog-datastore/relationship-rules/"+result.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading relationship rule after create", err.Error())
		return
	}

	mapRelationshipRuleToState(final, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RelationshipRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RelationshipRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetBare[relationshipRuleAPI](r.client, ctx, "/api/catalog-datastore/relationship-rules/"+state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading relationship rule", err.Error())
		return
	}

	mapRelationshipRuleToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RelationshipRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RelationshipRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildRelationshipRuleInput(&plan)

	result, err := client.UpdateBare[relationshipRuleAPI](r.client, ctx, "/api/catalog-datastore/relationship-rules/"+plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating relationship rule", err.Error())
		return
	}

	desiredState := plan.State.ValueString()
	if result.State != desiredState {
		_, err = r.client.Put(ctx, "/api/catalog-datastore/relationship-rules/"+result.ID+"/state", stateInput{State: desiredState})
		if err != nil {
			resp.Diagnostics.AddError("Error setting relationship rule state", err.Error())
			return
		}
	}

	final, err := client.GetBare[relationshipRuleAPI](r.client, ctx, "/api/catalog-datastore/relationship-rules/"+result.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading relationship rule after update", err.Error())
		return
	}

	mapRelationshipRuleToState(final, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RelationshipRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RelationshipRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/api/catalog-datastore/relationship-rules/"+state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting relationship rule", err.Error())
	}
}

func (r *RelationshipRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildRelationshipRuleInput(plan *RelationshipRuleResourceModel) *relationshipRuleInput {
	input := &relationshipRuleInput{
		Name:                       plan.Name.ValueString(),
		Description:                plan.Description.ValueString(),
		SourceDatasourceID:         plan.SourceDatasourceID.ValueString(),
		TargetDatasourceID:         plan.TargetDatasourceID.ValueString(),
		SourceFieldExpression:      plan.SourceFieldExpression.ValueString(),
		TargetFieldExpression:      plan.TargetFieldExpression.ValueString(),
		SourceFilterExpression:     plan.SourceFilterExpression.ValueString(),
		TargetFilterExpression:     plan.TargetFilterExpression.ValueString(),
		RelationshipType:           plan.RelationshipType.ValueString(),
		ReciprocalRelationshipType: plan.ReciprocalRelationshipType.ValueString(),
		Strategy:                   plan.Strategy.ValueString(),
		MatchStrategy:              plan.MatchStrategy.ValueString(),
	}
	if !plan.IntegrationConfig.IsNull() && !plan.IntegrationConfig.IsUnknown() {
		input.IntegrationConfig = json.RawMessage(plan.IntegrationConfig.ValueString())
	}
	return input
}

func mapRelationshipRuleToState(api *relationshipRuleAPI, state *RelationshipRuleResourceModel) {
	state.ID = types.StringValue(api.ID)
	state.Name = types.StringValue(api.Name)
	state.SourceDatasourceID = types.StringValue(api.SourceDatasourceID)
	state.TargetDatasourceID = types.StringValue(api.TargetDatasourceID)
	state.SourceFieldExpression = types.StringValue(api.SourceFieldExpression)
	state.TargetFieldExpression = types.StringValue(api.TargetFieldExpression)
	state.RelationshipType = types.StringValue(api.RelationshipType)
	state.Strategy = types.StringValue(api.Strategy)
	state.MatchStrategy = types.StringValue(api.MatchStrategy)
	state.Origin = types.StringValue(api.Origin)
	state.State = types.StringValue(api.State)
	state.CreatedAt = types.StringValue(api.CreatedAt)
	state.UpdatedAt = types.StringValue(api.UpdatedAt)

	if api.Description != nil {
		state.Description = types.StringValue(*api.Description)
	} else {
		state.Description = types.StringNull()
	}
	if api.SourceFilterExpression != nil {
		state.SourceFilterExpression = types.StringValue(*api.SourceFilterExpression)
	} else {
		state.SourceFilterExpression = types.StringNull()
	}
	if api.TargetFilterExpression != nil {
		state.TargetFilterExpression = types.StringValue(*api.TargetFilterExpression)
	} else {
		state.TargetFilterExpression = types.StringNull()
	}
	if api.ReciprocalRelationshipType != nil {
		state.ReciprocalRelationshipType = types.StringValue(*api.ReciprocalRelationshipType)
	} else {
		state.ReciprocalRelationshipType = types.StringNull()
	}
	if api.IntegrationConfig != nil && string(api.IntegrationConfig) != "null" {
		state.IntegrationConfig = jsontypes.NewNormalizedValue(string(api.IntegrationConfig))
	} else {
		state.IntegrationConfig = jsontypes.NewNormalizedNull()
	}
}
