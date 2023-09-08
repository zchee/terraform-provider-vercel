package client

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TeamMemberResponse struct {
	ID string `json:"id"`
}

type JoinedFrom struct {
	SSOUserID string `json:"ssoUserId"`
}

type Project struct {
	ProjectID string `json:"projectId"`
	Role      string `json:"role,omitempty"`
}

type UpdateTeamMemberRequest struct {
	Confirmed bool   `json:"confirmed"`
	Role      string `json:"role"`
	// JoinedFrom *JoinedFrom `json:"joinedFrom,omitempty"`
	// Projects   []*Project  `json:"projects,omitempty"`
	TeamID string `json:"-"`
	UID    string `json:"-"`
}

func (c *Client) UpdateTeamMember(ctx context.Context, request UpdateTeamMemberRequest) (e TeamMemberResponse, err error) {
	url := fmt.Sprintf("%s/v1/teams/%s/members/%s", c.baseURL, c.teamID(request.TeamID), request.UID)
	payload := string(mustMarshal(&request))

	tflog.Trace(ctx, "updating team member", map[string]interface{}{
		"url":     url,
		"payload": payload,
	})
	var response TeamMemberResponse
	err = c.doRequest(clientRequest{
		ctx:    ctx,
		method: "PATCH",
		url:    url,
		body:   payload,
	}, &response)
	if err != nil {
		return e, err
	}
	return response, err
}
