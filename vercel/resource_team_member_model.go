package vercel

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vercel/terraform-provider-vercel/client"
)

// TeamVCS reflects the state terraform stores internally for a team member VCS information.
type TeamVCS struct {
	Login types.String `tfsdk:"login"`
}

type TeamJoinedFrom struct {
	Origin       types.String `tfsdk:"origin"`       // "link" | "mail" | "import" | "teams" | "github" | "gitlab" | "bitbucket" | "saml" | "dsync" | @"feedback" | "organization-teams"
	CommitID     types.String `tfsdk:"commitId"`     // omitempty
	RepoID       types.String `tfsdk:"repoId"`       // omitempty
	RepoPath     types.String `tfsdk:"repoPath"`     // omitempty
	GitUserID    types.String `tfsdk:"gitUserId"`    // omitempty, string | number
	GitUserLogin types.String `tfsdk:"gitUserLogin"` // omitempty
	SSOUserID    types.String `tfsdk:"ssoUserId"`    // omitempty
	IDPUserID    types.String `tfsdk:"idpUserId"`    // omitempty
	DsyncUserID  types.String `tfsdk:"dsyncUserId"`  // omitempty
}

type TeamEmailInviteCode struct {
	ID          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"` // omitempty
	Role        types.String `tfsdk:"role"`  // omitempty, "OWNER" | "MEMBER" | "DEVELOPER" | "VIEWER" | "BILLING" | "CONTRIBUTOR"
	IsDSyncUser types.Bool   `tfsdk:"isDSyncUser"`
	Projects    types.Map    `tfsdk:"projects"` // [key: string]: "ADMIN" | "PROJECT_DEVELOPER" | "PROJECT_VIEWER"
}

type TeamProject struct {
	Name types.String `tfsdk:"name"` // omitempty
	ID   types.String `tfsdk:"id"`   // omitempty
	Role types.String `tfsdk:"role"` // omitempty, "ADMIN" | "PROJECT_DEVELOPER" | "PROJECT_VIEWER"
}

// TeamMember reflects the state terraform stores internally for a team member.
type TeamMember struct {
	TeamID     types.String    `tfsdk:"teamID"`
	Avatar     types.String    `tfsdk:"avatar"` // omitempty
	Confirmed  types.Bool      `tfsdk:"confirmed"`
	Email      types.String    `tfsdk:"email"`
	Github     *TeamVCS        `tfsdk:"github"`    // omitempty
	Gitlab     *TeamVCS        `tfsdk:"gitlab"`    // omitempty
	Bitbucket  *TeamVCS        `tfsdk:"bitbucket"` // omitempty
	Role       types.String    `tfsdk:"role"`
	UID        types.String    `tfsdk:"uid"`
	Username   types.String    `tfsdk:"username"`
	Name       types.String    `tfsdk:"name"`       // omitempty
	JoinedFrom *TeamJoinedFrom `tfsdk:"joinedFrom"` // omitempty
	Projects   []*TeamProject  `tfsdk:"projects"`   // omitempty
}

// TeamMembers reflects the state terraform stores internally for a team members.
type TeamMembers struct {
	TeamID types.String `tfsdk:"teamID"`

	Members         []*TeamMember          `tfsdk:"members"`
	EmailInviteCode []*TeamEmailInviteCode `tfsdk:"emailInviteCodes"`
	// TODO(zchee): support pagination
	// pagination: {
	//   hasNext: boolean
	//   count: number
	//   next: number | null
	//   prev: number | null
	// }
}

func (e *TeamMembers) toUpdateTeamMembersRequest(uid string) client.UpdateTeamMemberRequest {
	var member *TeamMember
	for _, m := range e.Members {
		if m.UID.ValueString() == uid {
			member = m
			break
		}
	}
	req := client.UpdateTeamMemberRequest{
		Confirmed: member.Confirmed.ValueBool(),
		Role:      member.Role.ValueString(),
		UID:       member.UID.ValueString(),
		TeamID:    member.TeamID.ValueString(),
	}

	return req
}

func (e *TeamMembers) convertResponseToTeamMembers(response client.TeamMemberResponse) TeamMembers {
	// e.Members = append(e.Members, e)

	return *e
}
