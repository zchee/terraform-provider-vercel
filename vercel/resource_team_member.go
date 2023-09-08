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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vercel/terraform-provider-vercel/client"
)

type teamMemberResource struct {
	client *client.Client
}

var (
	_ resource.Resource              = &teamMemberResource{}
	_ resource.ResourceWithConfigure = &teamMemberResource{}
)

func newteamMemberResource() resource.Resource {
	return &teamMemberResource{}
}

func (r *teamMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_member"
}

func (r *teamMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema returns the schema information for a team member resource.
func (r *teamMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Provides a Team member resource.

TODO(zchee): more descriptions.
        `,
		Attributes: map[string]schema.Attribute{
			"uid": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringLengthBetween(1, 48),
					stringRegex(
						regexp.MustCompile(`^[a-z0-9\-]{1,48}$`),
						"The slug of a Team can only contain up to 48 alphanumeric lowercase characters and hyphens",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
				Description:   "The desired slug for the Team",
			},
		},
	}
}

func (r *teamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TeamMembers
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetTeam(ctx, plan.TeamID.ValueString())
	if client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Error creating team member",
			"Could not find team, please make sure team_id match the team you wish to deploy to.",
		)
		return
	}

	attr := req.Config.Schema.GetAttributes()
	uid, ok := attr["uid"]
	if !ok {
		resp.Diagnostics.AddError(
			"Error creating team member",
			"Could not find uid, please make sure match uid wish to deploy to.",
		)
		return
	}
	_ = uid
	out, err := r.client.UpdateTeamMember(ctx, plan.toUpdateTeamMembersRequest("")) // TODO(zchee): use UID
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding domain to project",
			fmt.Sprintf(
				"Could not create %s to %s team, unexpected error: %s",
				uid,
				plan.TeamID.ValueString(),
				err,
			),
		)
		return
	}

	result := plan.convertResponseToTeamMembers(out)
	tflog.Trace(ctx, "added uid to team", map[string]interface{}{
		"uid":     uid,
		"team_id": result.TeamID.ValueString(),
	})

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read will read a team member from the vercel API and provide terraform with information about it.
// It is called by the provider whenever values should be read to update state.
func (r *teamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var members TeamMembers
	diags := req.State.Get(ctx, &members)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.GetTeamMembers(ctx, members.TeamID.ValueString())
	if client.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading team",
			fmt.Sprintf("Could not read team %s, unexpected error: %s",
				members.TeamID.ValueString(),
				err,
			),
		)
		return
	}

	members.Members = make([]*TeamMember, len(out.Members))
	for i, member := range out.Members {
		members.Members[i] = &TeamMember{
			Confirmed: types.BoolValue(member.Confirmed),
			Email:     types.StringValue(member.Email),
			Role:      types.StringValue(member.Role),
			UID:       types.StringValue(member.UID),
			Username:  types.StringValue(member.Username),
			JoinedFrom: &TeamJoinedFrom{
				SSOUserID: types.StringValue(member.JoinedFrom),
			},
		}
		if member.Github != nil {
			members.Members[i].Github = &TeamVCS{
				Login: types.StringValue(member.Github.Login),
			}
		}
		if member.Gitlab != nil {
			members.Members[i].Gitlab = &TeamVCS{
				Login: types.StringValue(member.Gitlab.Login),
			}
		}
		if member.Bitbucket != nil {
			members.Members[i].Gitlab = &TeamVCS{
				Login: types.StringValue(member.Bitbucket.Login),
			}
		}
	}
	tflog.Trace(ctx, "read team member", map[string]interface{}{
		"team_id": members.TeamID.ValueString(),
	})

	diags = resp.State.Set(ctx, members)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update will update a members of target team via the vercel API.
func (r *teamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TeamMembers
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateTeamMember(ctx, plan.toUpdateTeamMembersRequest("")) // TODO(zchee): use UID
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating shared environment variable",
			"Could not update shared environment variable, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "updated team member", map[string]interface{}{
		"team_id": response.ID,
	})

	diags = resp.State.Set(ctx, response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete a team from within terraform.
// Environment variables do not need to be explicitly deleted, as Vercel will automatically prune them.
//
// TODO(zchee): implements
func (r *teamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// ImportState takes an identifier and reads all the team information from the Vercel API.
// The results are then stored in terraform state.
//
// TODO(zchee): implements
func (r *teamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
