package client

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TeamMember struct {
	Confirmed bool   `json:"confirmed"`
	Email     string `json:"email"`
	Github    *struct {
		Login string `tfsdk:"login"`
	} `json:"github"`
	Gitlab *struct {
		Login string `tfsdk:"login"`
	} `json:"gitlab"`
	Bitbucket *struct {
		Login string `tfsdk:"login"`
	} `json:"bitbucket"`
	Role       string `json:"role"`
	UID        string `json:"uid"`
	Username   string `json:"username"`
	JoinedFrom string `json:"joinedFrom"`
}

type TeamMembers struct {
	Members []TeamMember `json:"members"`
}

// GetTeamMembers returns information about a members of existing team within vercel.
func (c *Client) GetTeamMembers(ctx context.Context, teamID string) (r TeamMembers, err error) {
	url := fmt.Sprintf("%s/v2/teams/%s/members", c.baseURL, c.teamID(teamID))
	tflog.Trace(ctx, "getting team members", map[string]interface{}{
		"url": url,
	})
	err = c.doRequest(clientRequest{
		ctx:    ctx,
		method: "GET",
		url:    url,
		body:   "",
	}, &r)
	return r, err
}
