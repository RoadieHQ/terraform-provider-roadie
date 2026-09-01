package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IntegrationResource{}
var _ resource.ResourceWithImportState = &IntegrationResource{}

type IntegrationResource struct {
	client *client.RoadieClient
}

type IntegrationResourceModel struct {
	ID                types.String         `tfsdk:"id"`
	Name              types.String         `tfsdk:"name"`
	Slug              types.String         `tfsdk:"slug"`
	Type              types.String         `tfsdk:"type"`
	Host              types.String         `tfsdk:"host"`
	AuthType          types.String         `tfsdk:"auth_type"`
	AuthConfig        jsontypes.Normalized `tfsdk:"auth_config"`
	BackendType       types.String         `tfsdk:"backend_type"`
	Enabled           types.Bool           `tfsdk:"enabled"`
	RequestsPerHour   types.Float64        `tfsdk:"requests_per_hour"`
	RequestsPerSecond types.Float64        `tfsdk:"requests_per_second"`
	BurstCapacity     types.Float64        `tfsdk:"burst_capacity"`
	Config            jsontypes.Normalized `tfsdk:"config"`
	LogoSlug          types.String         `tfsdk:"logo_slug"`
	Adopted           types.Bool           `tfsdk:"adopted"`
	CreatedBy         types.String         `tfsdk:"created_by"`
	CreatedAt         types.String         `tfsdk:"created_at"`
	UpdatedAt         types.String         `tfsdk:"updated_at"`
}

type flexFloat64 float64

func (f *flexFloat64) UnmarshalJSON(data []byte) error {
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = flexFloat64(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("flexFloat64: cannot unmarshal %s", string(data))
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("flexFloat64: cannot parse %q: %w", s, err)
	}
	*f = flexFloat64(n)
	return nil
}

type integrationAPI struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Slug              string          `json:"slug"`
	Type              string          `json:"type,omitempty"`
	Host              string          `json:"host,omitempty"`
	AuthType          string          `json:"authType,omitempty"`
	AuthConfig        json.RawMessage `json:"authConfig,omitempty"`
	BackendType       string          `json:"backendType,omitempty"`
	Enabled           *bool           `json:"enabled,omitempty"`
	RequestsPerHour   *flexFloat64    `json:"requestsPerHour,omitempty"`
	RequestsPerSecond *flexFloat64    `json:"requestsPerSecond,omitempty"`
	BurstCapacity     *flexFloat64    `json:"burstCapacity,omitempty"`
	Config            json.RawMessage `json:"config,omitempty"`
	LogoSlug          string          `json:"logoSlug,omitempty"`
	CreatedBy         string          `json:"createdBy"`
	CreatedAt         string          `json:"createdAt"`
	UpdatedAt         string          `json:"updatedAt"`
}

type integrationInput struct {
	Name              string          `json:"name"`
	Slug              string          `json:"slug"`
	Type              string          `json:"type,omitempty"`
	Host              string          `json:"host,omitempty"`
	AuthType          string          `json:"authType,omitempty"`
	AuthConfig        json.RawMessage `json:"authConfig,omitempty"`
	BackendType       string          `json:"backendType,omitempty"`
	Enabled           bool            `json:"enabled"`
	RequestsPerHour   *float64        `json:"requestsPerHour,omitempty"`
	RequestsPerSecond *float64        `json:"requestsPerSecond,omitempty"`
	BurstCapacity     *float64        `json:"burstCapacity,omitempty"`
	Config            json.RawMessage `json:"config,omitempty"`
	LogoSlug          string          `json:"logoSlug,omitempty"`
}

func NewIntegrationResource() resource.Resource {
	return &IntegrationResource{}
}

func (r *IntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *IntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Roadie integration.",
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
				Required: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"scm",
						"ci-cd",
						"monitoring",
						"incident-management",
						"infrastructure",
						"security",
						"communication",
						"project-management",
						"analytics",
						"other",
					),
				},
			},
			"host": schema.StringAttribute{
				Optional: true,
			},
			"auth_type": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"header",
						"basic",
						"bearer-token",
						"oauth2-client-credentials",
						"oauth2-jwt-bearer",
						"github-app",
						"none",
					),
				},
			},
			"auth_config": schema.StringAttribute{
				Description: "Auth configuration JSON containing ${SECRET_REF} placeholders, not raw credentials.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"backend_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("http"),
				Validators: []validator.String{
					stringvalidator.OneOf("http", "aws"),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"requests_per_hour": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"requests_per_second": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"burst_capacity": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"config": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"logo_slug": schema.StringAttribute{
				Optional: true,
			},
			"adopted": schema.BoolAttribute{
				Description: "Whether this resource was adopted from a pre-existing integration rather than created by Terraform.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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

func (r *IntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildIntegrationInput(&plan)

	result, err := client.CreateWrapped[integrationAPI](r.client, ctx, "/api/integrations", input)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			r.adoptExisting(ctx, &plan, input, resp)
			return
		}
		resp.Diagnostics.AddError("Error creating integration", err.Error())
		return
	}

	plan.Adopted = types.BoolValue(false)
	full, readErr := client.GetWrapped[integrationAPI](r.client, ctx, "/api/integrations/"+result.ID)
	if readErr != nil {
		mapIntegrationToState(result, &plan)
	} else {
		mapIntegrationToState(full, &plan)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IntegrationResource) adoptExisting(ctx context.Context, plan *IntegrationResourceModel, input *integrationInput, resp *resource.CreateResponse) {
	body, err := r.client.Get(ctx, "/api/integrations?limit=500")
	if err != nil {
		resp.Diagnostics.AddError("Error listing integrations", err.Error())
		return
	}
	var listResp struct {
		Data []struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		resp.Diagnostics.AddError("Error decoding integrations list", err.Error())
		return
	}

	slug := plan.Slug.ValueString()
	var existingID string
	for _, integration := range listResp.Data {
		if integration.Slug == slug {
			existingID = integration.ID
			break
		}
	}
	if existingID == "" {
		resp.Diagnostics.AddError("Error adopting integration", fmt.Sprintf("Integration with slug %q reported as existing but not found", slug))
		return
	}

	_, err = client.UpdateWrapped[integrationAPI](r.client, ctx, "/api/integrations/"+existingID, input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating adopted integration", err.Error())
		return
	}

	full, readErr := client.GetWrapped[integrationAPI](r.client, ctx, "/api/integrations/"+existingID)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading adopted integration", readErr.Error())
		return
	}

	plan.Adopted = types.BoolValue(true)
	mapIntegrationToState(full, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *IntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetWrapped[integrationAPI](r.client, ctx, "/api/integrations/"+state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading integration", err.Error())
		return
	}

	mapIntegrationToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildIntegrationInput(&plan)

	_, err := client.UpdateWrapped[integrationAPI](r.client, ctx, "/api/integrations/"+plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating integration", err.Error())
		return
	}

	full, readErr := client.GetWrapped[integrationAPI](r.client, ctx, "/api/integrations/"+plan.ID.ValueString())
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading integration after update", readErr.Error())
		return
	}

	mapIntegrationToState(full, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Adopted.ValueBool() {
		return
	}

	err := r.client.Delete(ctx, "/api/integrations/"+state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting integration", err.Error())
	}
}

func (r *IntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildIntegrationInput(plan *IntegrationResourceModel) *integrationInput {
	input := &integrationInput{
		Name:        plan.Name.ValueString(),
		Slug:        plan.Slug.ValueString(),
		Type:        plan.Type.ValueString(),
		Host:        plan.Host.ValueString(),
		AuthType:    plan.AuthType.ValueString(),
		BackendType: plan.BackendType.ValueString(),
		Enabled:     plan.Enabled.ValueBool(),
		LogoSlug:    plan.LogoSlug.ValueString(),
	}

	if !plan.AuthConfig.IsNull() && !plan.AuthConfig.IsUnknown() {
		input.AuthConfig = json.RawMessage(plan.AuthConfig.ValueString())
	}
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		input.Config = json.RawMessage(plan.Config.ValueString())
	}
	if !plan.RequestsPerHour.IsNull() && !plan.RequestsPerHour.IsUnknown() {
		v := plan.RequestsPerHour.ValueFloat64()
		input.RequestsPerHour = &v
	}
	if !plan.RequestsPerSecond.IsNull() && !plan.RequestsPerSecond.IsUnknown() {
		v := plan.RequestsPerSecond.ValueFloat64()
		input.RequestsPerSecond = &v
	}
	if !plan.BurstCapacity.IsNull() && !plan.BurstCapacity.IsUnknown() {
		v := plan.BurstCapacity.ValueFloat64()
		input.BurstCapacity = &v
	}

	return input
}

func mapIntegrationToState(api *integrationAPI, state *IntegrationResourceModel) {
	state.ID = types.StringValue(api.ID)
	state.Name = types.StringValue(api.Name)
	state.Slug = types.StringValue(api.Slug)
	if api.Enabled != nil {
		state.Enabled = types.BoolValue(*api.Enabled)
	}
	state.CreatedBy = types.StringValue(api.CreatedBy)
	state.CreatedAt = types.StringValue(api.CreatedAt)
	state.UpdatedAt = types.StringValue(api.UpdatedAt)

	if api.Type != "" {
		state.Type = types.StringValue(api.Type)
	} else {
		state.Type = types.StringNull()
	}
	if api.Host != "" {
		state.Host = types.StringValue(api.Host)
	} else {
		state.Host = types.StringNull()
	}
	if api.AuthType != "" {
		state.AuthType = types.StringValue(api.AuthType)
	} else {
		state.AuthType = types.StringNull()
	}
	if api.BackendType != "" {
		state.BackendType = types.StringValue(api.BackendType)
	} else {
		state.BackendType = types.StringNull()
	}
	if api.LogoSlug != "" {
		state.LogoSlug = types.StringValue(api.LogoSlug)
	} else {
		state.LogoSlug = types.StringNull()
	}

	if api.RequestsPerHour != nil {
		state.RequestsPerHour = types.Float64Value(float64(*api.RequestsPerHour))
	} else {
		state.RequestsPerHour = types.Float64Null()
	}
	if api.RequestsPerSecond != nil {
		state.RequestsPerSecond = types.Float64Value(float64(*api.RequestsPerSecond))
	} else {
		state.RequestsPerSecond = types.Float64Null()
	}
	if api.BurstCapacity != nil {
		state.BurstCapacity = types.Float64Value(float64(*api.BurstCapacity))
	} else {
		state.BurstCapacity = types.Float64Null()
	}

	if api.AuthConfig != nil {
		state.AuthConfig = jsontypes.NewNormalizedValue(string(api.AuthConfig))
	} else {
		state.AuthConfig = jsontypes.NewNormalizedNull()
	}

	if api.Config != nil && string(api.Config) != "{}" && string(api.Config) != "null" {
		state.Config = jsontypes.NewNormalizedValue(string(api.Config))
	} else if state.Config.IsUnknown() {
		state.Config = jsontypes.NewNormalizedNull()
	}
}

// writeOnlyPlanModifier preserves the prior state value for write-only attributes
// that are never returned by the API.
type writeOnlyPlanModifier struct{}

func (m writeOnlyPlanModifier) Description(_ context.Context) string {
	return "Preserves the prior state value for write-only attributes."
}

func (m writeOnlyPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m writeOnlyPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsUnknown() {
		return
	}
	// If there's no config change, keep whatever was in state.
	if !req.StateValue.IsNull() && req.PlanValue.Equal(req.StateValue) {
		return
	}
}
