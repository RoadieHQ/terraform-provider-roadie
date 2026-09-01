package resources

import (
	"context"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &SecretResource{}
	_ resource.ResourceWithConfigure = &SecretResource{}
)

type SecretResource struct {
	client *client.RoadieClient
}

type SecretResourceModel struct {
	Ref         types.String `tfsdk:"ref"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	HelpURL     types.String `tfsdk:"help_url"`
	Value       types.String `tfsdk:"value"`
	Status      types.String `tfsdk:"status"`
}

type secretMetadataInput struct {
	Name            string `json:"name"`
	InternalKeyName string `json:"internalKeyName,omitempty"`
	Description     string `json:"description,omitempty"`
	HelpURL         string `json:"helpUrl,omitempty"`
}

type secretValueInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type secretStatusResponse struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	HelpURL      string `json:"helpUrl"`
	IsCustom     bool   `json:"isCustom"`
	CreatedAt    string `json:"createdAt"`
	LastModified string `json:"lastModified"`
}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

func (r *SecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Roadie secret. Creates the secret metadata and sets its value.",
		Attributes: map[string]schema.Attribute{
			"ref": schema.StringAttribute{
				Description: "The secret reference key used in ${...} placeholders (e.g. GITHUB_TOKEN).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-friendly display name for the secret.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of what this secret is used for.",
				Optional:    true,
			},
			"help_url": schema.StringAttribute{
				Description: "URL linking to documentation about obtaining this secret.",
				Optional:    true,
			},
			"value": schema.StringAttribute{
				Description: "The secret value. Write-only — never returned by the API.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					writeOnlyPlanModifier{},
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the secret: Available or Not Set.",
				Computed:    true,
			},
		},
	}
}

func (r *SecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadataInput := secretMetadataInput{
		Name:            plan.Name.ValueString(),
		InternalKeyName: plan.Ref.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		metadataInput.Description = plan.Description.ValueString()
	}
	if !plan.HelpURL.IsNull() && !plan.HelpURL.IsUnknown() {
		metadataInput.HelpURL = plan.HelpURL.ValueString()
	}

	_, err := r.client.Post(ctx, "/api/secrets-settings/secret-metadata", metadataInput)
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret metadata", err.Error())
		return
	}

	valueInput := secretValueInput{
		Name:  plan.Ref.ValueString(),
		Value: plan.Value.ValueString(),
	}
	_, err = r.client.Post(ctx, "/api/secrets-settings/keys", valueInput)
	if err != nil {
		resp.Diagnostics.AddError("Error setting secret value", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	savedValue := state.Value

	diags := r.readIntoModel(ctx, &state)
	if diags.HasError() {
		for _, d := range diags {
			if d.Summary() == "Secret not found" {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.Append(diags...)
		return
	}

	state.Value = savedValue
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priorState SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadataInput := map[string]string{}
	if plan.Name.ValueString() != priorState.Name.ValueString() {
		metadataInput["name"] = plan.Name.ValueString()
	}
	if plan.Description.ValueString() != priorState.Description.ValueString() {
		metadataInput["description"] = plan.Description.ValueString()
	}
	if plan.HelpURL.ValueString() != priorState.HelpURL.ValueString() {
		metadataInput["helpUrl"] = plan.HelpURL.ValueString()
	}

	if len(metadataInput) > 0 {
		_, err := r.client.Patch(ctx, "/api/secrets-settings/secret-metadata/"+plan.Ref.ValueString(), metadataInput)
		if err != nil {
			resp.Diagnostics.AddError("Error updating secret metadata", err.Error())
			return
		}
	}

	if !plan.Value.Equal(priorState.Value) {
		valueInput := secretValueInput{
			Name:  plan.Ref.ValueString(),
			Value: plan.Value.ValueString(),
		}
		_, err := r.client.Post(ctx, "/api/secrets-settings/keys", valueInput)
		if err != nil {
			resp.Diagnostics.AddError("Error updating secret value", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := state.Ref.ValueString()

	err := r.client.Delete(ctx, "/api/secrets-settings/secret-value/"+ref)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting secret value", err.Error())
		return
	}

	err = r.client.Delete(ctx, "/api/secrets-settings/secret-metadata/"+ref)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting secret metadata", err.Error())
	}
}

func (r *SecretResource) readIntoModel(ctx context.Context, model *SecretResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	result, err := client.GetBare[secretStatusResponse](r.client, ctx, "/api/secrets-settings/secret-value/"+model.Ref.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			diags.AddError("Secret not found", fmt.Sprintf("Secret with ref %q was not found", model.Ref.ValueString()))
			return diags
		}
		diags.AddError("Error reading secret", err.Error())
		return diags
	}

	model.Status = types.StringValue(result.Status)
	if result.Name != "" {
		model.Name = types.StringValue(result.Name)
	}
	if result.Description != "" {
		model.Description = types.StringValue(result.Description)
	} else if model.Description.IsNull() || model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}
	if result.HelpURL != "" {
		model.HelpURL = types.StringValue(result.HelpURL)
	} else if model.HelpURL.IsNull() || model.HelpURL.IsUnknown() {
		model.HelpURL = types.StringNull()
	}

	return diags
}
