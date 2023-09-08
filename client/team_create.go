package client

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// TeamCreateRequest defines the information needed to create a team within vercel.
type TeamCreateRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Enabled struct {
	Value *bool `json:"enabled"`
}

// TeamResponse is the information returned by the vercel api when a team is created.
type TeamResponse struct {
	ID         string  `json:"id"`
	Name       *string `json:"name"`
	Avatar     *string `json:"avatar"`
	Membership *struct {
		Role       *string `json:"role"`
		Confirmed  *bool   `json:"confirmed"`
		JoinedFrom *struct {
			Origin *string
		} `json:"joinedFrom"`
		UID *string `json:"uid"`
	} `json:"membership"`
	EnablePreviewFeedback         *string `json:"enablePreviewFeedback"`
	IsMigratingToSensitiveEnvVars *bool   `json:"isMigratingToSensitiveEnvVars"`
	InviteCode                    *string `json:"inviteCode"`
	Description                   *string `json:"description"`
	StagingPrefix                 *string `json:"stagingPrefix"`
	ResourceConfig                *struct {
		ConcurrentBuilds *int64 `json:"concurrentBuilds"`
	}
	RemoteCaching       *Enabled `json:"remoteCaching"`
	EnabledInvoiceItems *struct {
		ConcurrentBuilds *Enabled `json:"concurrentBuilds"`
		Monitoring       *Enabled `json:"monitoring"`
	} `json:"enabledInvoiceItems"`
	Spaces *Enabled `json:"spaces"`
}

// CreateTeam creates a team within vercel.
func (c *Client) CreateTeam(ctx context.Context, request TeamCreateRequest) (r TeamResponse, err error) {
	url := fmt.Sprintf("%s/v1/teams", c.baseURL)

	payload := string(mustMarshal(request))
	tflog.Trace(ctx, "creating team", map[string]interface{}{
		"url":     url,
		"payload": payload,
	})
	err = c.doRequest(clientRequest{
		ctx:    ctx,
		method: "POST",
		url:    url,
		body:   payload,
	}, &r)
	return r, err
}
