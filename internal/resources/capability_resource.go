package resources

import (
	"context"
	"fmt"

	"github.com/RoadieHQ/terraform-provider-roadie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &CapabilityResource{}
	_ resource.ResourceWithImportState = &CapabilityResource{}
)

type CapabilityResource struct {
	client *client.RoadieClient
}

type CapabilityResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	Description    types.String `tfsdk:"description"`
	Instructions   types.String `tfsdk:"instructions"`
	CurrentVersion types.Int64  `tfsdk:"current_version"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

type capabilityAPI struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
	Instructions   string `json:"instructions"`
	CurrentVersion int64  `json:"currentVersion"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type capabilityCreateInput struct {
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
}

type capabilityUpdateInput struct {
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
}

func NewCapabilityResource() resource.Resource {
	return &CapabilityResource{}
}

func (r *CapabilityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_capability"
}

func (r *CapabilityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Roadie capability.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the capability.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the capability.",
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
				Description: "A description of what the capability does.",
				Required:    true,
			},
			"instructions": schema.StringAttribute{
				Description: "The instructions for the capability.",
				Required:    true,
			},
			"current_version": schema.Int64Attribute{
				Description: "The current version number, auto-incremented on each update.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the capability was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "When the capability was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *CapabilityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CapabilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CapabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := capabilityCreateInput{
		Name:         plan.Name.ValueString(),
		Description:  plan.Description.ValueString(),
		Instructions: plan.Instructions.ValueString(),
	}
	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		input.Slug = plan.Slug.ValueString()
	}

	result, err := client.CreateBare[capabilityAPI](r.client, ctx, "/api/capabilities/", input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating capability", err.Error())
		return
	}

	mapCapabilityToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CapabilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := client.GetBare[capabilityAPI](r.client, ctx, "/api/capabilities/"+state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading capability", err.Error())
		return
	}

	mapCapabilityToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CapabilityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CapabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := capabilityUpdateInput{
		Name:         plan.Name.ValueString(),
		Description:  plan.Description.ValueString(),
		Instructions: plan.Instructions.ValueString(),
	}
	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		input.Slug = plan.Slug.ValueString()
	}

	result, err := client.UpdateBare[capabilityAPI](r.client, ctx, "/api/capabilities/"+plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating capability", err.Error())
		return
	}

	mapCapabilityToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CapabilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/api/capabilities/"+state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting capability", err.Error())
	}
}

func (r *CapabilityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapCapabilityToState(api *capabilityAPI, state *CapabilityResourceModel) {
	state.ID = types.StringValue(api.ID)
	state.Name = types.StringValue(api.Name)
	state.Slug = types.StringValue(api.Slug)
	state.Description = types.StringValue(api.Description)
	state.Instructions = types.StringValue(api.Instructions)
	state.CurrentVersion = types.Int64Value(api.CurrentVersion)
	state.CreatedAt = types.StringValue(api.CreatedAt)
	state.UpdatedAt = types.StringValue(api.UpdatedAt)
}
