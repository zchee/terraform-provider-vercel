package vercel

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vercel/terraform-provider-vercel/client"
)

type teamResource struct {
	client *client.Client
}

var _ resource.ResourceWithConfigure = &teamResource{}

func newTeamResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Schema returns the schema information for a deployment resource.
func (r *teamResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Provides a Team resource.

TODO(zchee): more descriptions.
        `,
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringLengthBetween(1, 48),
					stringRegex(
						regexp.MustCompile(`^[a-z][a-z0-9_-]{0,47}$`),
						"The slug of a Team can only contain up to 48 alphanumeric lowercase characters and hyphens",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
				Description:   "The desired slug for the Team",
			},
			"name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringLengthBetween(1, 256),
					stringRegex(
						regexp.MustCompile(`^[ a-zA-Z0-9_-]{1,256}$`),
						"The name of a Team can only contain up to 256 alphanumeric characters, hyphens and space",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   "The desired name for the Team. It will be generated from the provided slug if nothing is provided",
			},
		},
	}
}

// Create will create a team within Vercel by calling the Vercel API.
//
// TODO(zchee): implements
func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

// Read will read a team from the vercel API and provide terraform with information about it.
// It is called by the provider whenever values should be read to update state.
func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Team
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if client.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading team",
			fmt.Sprintf("Could not read team %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	result := convertResponseToTeam(out, state)
	tflog.Trace(ctx, "read team", map[string]interface{}{
		"team":    result,
		"team_id": result.ID.ValueString(),
	})

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update will update a team and it's associated environment variables via the vercel API.
//
// TODO(zchee): implements
func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete a team from within terraform.
func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Team
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTeam(ctx, state.ID.ValueString())
	if client.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting team",
			fmt.Sprintf(
				"Could not delete team %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Trace(ctx, "deleted team", map[string]interface{}{
		"team_id": state.ID.ValueString(),
	})
}

// ImportState takes an identifier and reads all the team information from the Vercel API.
// The results are then stored in terraform state.
func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	out, err := r.client.GetTeam(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading team",
			fmt.Sprintf("Could not get team %s, unexpected error: %s",
				req.ID,
				err,
			),
		)
		return
	}

	result := convertResponseToTeam(out, nullTeam)
	tflog.Trace(ctx, "imported team", map[string]interface{}{
		"team_id": result.ID.ValueString(),
	})

	diags := resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
