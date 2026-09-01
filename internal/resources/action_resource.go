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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ActionResource{}
	_ resource.ResourceWithImportState = &ActionResource{}
)

type ActionResource struct {
	client *client.RoadieClient
}

type ActionResourceModel struct {
	ID             types.String         `tfsdk:"id"`
	Name           types.String         `tfsdk:"name"`
	Slug           types.String         `tfsdk:"slug"`
	Description    types.String         `tfsdk:"description"`
	Parameters     []ActionParamModel   `tfsdk:"parameters"`
	Steps          []ActionStepModel    `tfsdk:"steps"`
	Enabled        types.Bool           `tfsdk:"enabled"`
	Mode           types.String         `tfsdk:"mode"`
	Metadata       jsontypes.Normalized `tfsdk:"metadata"`
	EffectiveMode  types.String         `tfsdk:"effective_mode"`
	CurrentVersion types.Int64          `tfsdk:"current_version"`
	InputSchema    jsontypes.Normalized `tfsdk:"input_schema"`
	CreatedAt      types.String         `tfsdk:"created_at"`
	UpdatedAt      types.String         `tfsdk:"updated_at"`
}

type ActionParamModel struct {
	Name        types.String         `tfsdk:"name"`
	Type        types.String         `tfsdk:"type"`
	Description types.String         `tfsdk:"description"`
	Required    types.Bool           `tfsdk:"required"`
	Default     jsontypes.Normalized `tfsdk:"default"`
}

type ActionStepModel struct {
	ID            types.String         `tfsdk:"id"`
	IntegrationID types.String         `tfsdk:"integration_id"`
	Request       jsontypes.Normalized `tfsdk:"request"`
}

type actionAPI struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Slug           string            `json:"slug"`
	Description    string            `json:"description"`
	Parameters     []actionParamAPI  `json:"parameters"`
	Steps          []actionStepAPI   `json:"steps"`
	Enabled        bool              `json:"enabled"`
	Mode           *string           `json:"mode"`
	Metadata       json.RawMessage   `json:"metadata"`
	EffectiveMode  string            `json:"effectiveMode"`
	CurrentVersion int64             `json:"currentVersion"`
	InputSchema    json.RawMessage   `json:"inputSchema"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

type actionParamAPI struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Required    bool            `json:"required,omitempty"`
	Default     json.RawMessage `json:"default,omitempty"`
}

type actionStepAPI struct {
	ID            string          `json:"id"`
	IntegrationID string          `json:"integrationId"`
	Request       json.RawMessage `json:"request"`
}

type actionCreateInput struct {
	Name        string           `json:"name"`
	Slug        string           `json:"slug,omitempty"`
	Description string           `json:"description"`
	Parameters  []actionParamAPI `json:"parameters"`
	Steps       []actionStepAPI  `json:"steps"`
	Enabled     bool             `json:"enabled"`
}

type actionUpdateInput struct {
	Name        string           `json:"name"`
	Slug        string           `json:"slug,omitempty"`
	Description string           `json:"description"`
	Parameters  []actionParamAPI `json:"parameters"`
	Steps       []actionStepAPI  `json:"steps"`
	Enabled     bool             `json:"enabled"`
}

func NewActionResource() resource.Resource {
	return &ActionResource{}
}

func (r *ActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (r *ActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Roadie action.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the action.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the action.",
				Required:    true,
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
				Description: "A description of the action.",
				Optional:    true,
			},
			"parameters": schema.ListNestedAttribute{
				Description: "Input parameters for the action.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The parameter name.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "The parameter type.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("string", "number", "integer", "boolean", "array<string>"),
							},
						},
						"description": schema.StringAttribute{
							Description: "A description of the parameter.",
							Optional:    true,
						},
						"required": schema.BoolAttribute{
							Description: "Whether the parameter is required.",
							Optional:    true,
						},
						"default": schema.StringAttribute{
							Description: "Default value as JSON.",
							Optional:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
					},
				},
			},
			"steps": schema.ListNestedAttribute{
				Description: "The steps that make up the action.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Step identifier (must match /^[A-Za-z_][A-Za-z0-9_]*$/).",
							Required:    true,
						},
						"integration_id": schema.StringAttribute{
							Description: "The integration to use for this step.",
							Required:    true,
						},
						"request": schema.StringAttribute{
							Description: "The request configuration as JSON (HTTP or AWS request object).",
							Required:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
					},
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the action is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"mode": schema.StringAttribute{
				Description: "The action mode (read or write). If omitted, derived from steps.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("read", "write"),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary metadata as JSON.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"effective_mode": schema.StringAttribute{
				Description: "The resolved mode (explicit or derived from steps).",
				Computed:    true,
			},
			"current_version": schema.Int64Attribute{
				Description: "The current version number, auto-incremented on each update.",
				Computed:    true,
			},
			"input_schema": schema.StringAttribute{
				Description: "The computed JSON Schema for action inputs.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"created_at": schema.StringAttribute{
				Description: "When the action was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "When the action was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := actionCreateInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Parameters:  mapActionParamsToAPI(plan.Parameters),
		Steps:       mapActionStepsToAPI(plan.Steps),
		Enabled:     plan.Enabled.ValueBool(),
	}
	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		input.Slug = plan.Slug.ValueString()
	}

	result, err := client.CreateBare[actionAPI](r.client, ctx, "/api/actions/", input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating action", err.Error())
		return
	}

	mapActionToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetBare[actionAPI](r.client, ctx, "/api/actions/"+state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading action", err.Error())
		return
	}

	mapActionToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := actionUpdateInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Parameters:  mapActionParamsToAPI(plan.Parameters),
		Steps:       mapActionStepsToAPI(plan.Steps),
		Enabled:     plan.Enabled.ValueBool(),
	}
	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		input.Slug = plan.Slug.ValueString()
	}

	result, err := client.UpdateBare[actionAPI](r.client, ctx, "/api/actions/"+plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating action", err.Error())
		return
	}

	mapActionToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/api/actions/"+state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting action", err.Error())
	}
}

func (r *ActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapActionParamsToAPI(params []ActionParamModel) []actionParamAPI {
	if params == nil {
		return []actionParamAPI{}
	}
	result := make([]actionParamAPI, len(params))
	for i, p := range params {
		result[i] = actionParamAPI{
			Name:        p.Name.ValueString(),
			Type:        p.Type.ValueString(),
			Description: p.Description.ValueString(),
			Required:    p.Required.ValueBool(),
		}
		if !p.Default.IsNull() && !p.Default.IsUnknown() {
			result[i].Default = json.RawMessage(p.Default.ValueString())
		}
	}
	return result
}

func mapActionStepsToAPI(steps []ActionStepModel) []actionStepAPI {
	result := make([]actionStepAPI, len(steps))
	for i, s := range steps {
		result[i] = actionStepAPI{
			ID:            s.ID.ValueString(),
			IntegrationID: s.IntegrationID.ValueString(),
			Request:       json.RawMessage(s.Request.ValueString()),
		}
	}
	return result
}

func mapActionToState(api *actionAPI, state *ActionResourceModel) {
	state.ID = types.StringValue(api.ID)
	state.Name = types.StringValue(api.Name)
	state.Slug = types.StringValue(api.Slug)
	state.Description = types.StringValue(api.Description)
	state.Enabled = types.BoolValue(api.Enabled)
	state.EffectiveMode = types.StringValue(api.EffectiveMode)
	state.CurrentVersion = types.Int64Value(api.CurrentVersion)
	state.CreatedAt = types.StringValue(api.CreatedAt)
	state.UpdatedAt = types.StringValue(api.UpdatedAt)

	if api.Mode != nil {
		state.Mode = types.StringValue(*api.Mode)
	} else {
		state.Mode = types.StringNull()
	}

	if api.Metadata != nil && string(api.Metadata) != "null" {
		state.Metadata = jsontypes.NewNormalizedValue(string(api.Metadata))
	} else {
		state.Metadata = jsontypes.NewNormalizedNull()
	}

	if api.InputSchema != nil && string(api.InputSchema) != "null" {
		state.InputSchema = jsontypes.NewNormalizedValue(string(api.InputSchema))
	} else {
		state.InputSchema = jsontypes.NewNormalizedNull()
	}

	state.Parameters = mapActionParamsFromAPI(api.Parameters)
	state.Steps = mapActionStepsFromAPI(api.Steps)
}

func mapActionParamsFromAPI(params []actionParamAPI) []ActionParamModel {
	if len(params) == 0 {
		return nil
	}
	result := make([]ActionParamModel, len(params))
	for i, p := range params {
		result[i] = ActionParamModel{
			Name:        types.StringValue(p.Name),
			Type:        types.StringValue(p.Type),
			Description: types.StringValue(p.Description),
			Required:    types.BoolValue(p.Required),
		}
		if p.Default != nil && string(p.Default) != "null" {
			result[i].Default = jsontypes.NewNormalizedValue(string(p.Default))
		} else {
			result[i].Default = jsontypes.NewNormalizedNull()
		}
	}
	return result
}

func mapActionStepsFromAPI(steps []actionStepAPI) []ActionStepModel {
	result := make([]ActionStepModel, len(steps))
	for i, s := range steps {
		result[i] = ActionStepModel{
			ID:            types.StringValue(s.ID),
			IntegrationID: types.StringValue(s.IntegrationID),
			Request:       jsontypes.NewNormalizedValue(string(s.Request)),
		}
	}
	return result
}
