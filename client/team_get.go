package client

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// GetTeam returns information about an existing team within vercel.
func (c *Client) GetTeam(ctx context.Context, teamID string) (r TeamResponse, err error) {
	url := fmt.Sprintf("%s/v2/teams/%s", c.baseURL, c.teamID(teamID))
	tflog.Trace(ctx, "getting team", map[string]interface{}{
		"url":     url,
		"team_id": teamID,
	})
	err = c.doRequest(clientRequest{
		ctx:    ctx,
		method: "GET",
		url:    url,
		body:   "",
	}, &r)
	return r, err
}
