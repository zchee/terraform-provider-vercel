package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Repository defines the information about a projects git connection.
type Repository struct {
	Type             string
	Repo             string
	ProductionBranch *string
}

// getRepoNameFromURL is a helper method to extract the repo name from a GitLab URL.
// This is necessary as GitLab doesn't return the repository slug in the API response,
// Because this information isn't present, the only way to obtain it is to parse the URL.
func getRepoNameFromURL(url string) string {
	url = strings.TrimSuffix(url, ".git")
	urlParts := strings.Split(url, "/")

	return urlParts[len(urlParts)-1]
}

// Repository is a helper method to convert the ProjectResponse Repository information into a more
// digestible format.
func (r *ProjectResponse) Repository() *Repository {
	if r.Link == nil {
		return nil
	}
	switch r.Link.Type {
	case "github":
		return &Repository{
			Type:             "github",
			Repo:             fmt.Sprintf("%s/%s", r.Link.Org, r.Link.Repo),
			ProductionBranch: r.Link.ProductionBranch,
		}
	case "gitlab":
		return &Repository{
			Type:             "gitlab",
			Repo:             fmt.Sprintf("%s/%s", r.Link.ProjectNamespace, getRepoNameFromURL(r.Link.ProjectURL)),
			ProductionBranch: r.Link.ProductionBranch,
		}
	case "bitbucket":
		return &Repository{
			Type:             "bitbucket",
			Repo:             fmt.Sprintf("%s/%s", r.Link.Owner, r.Link.Slug),
			ProductionBranch: r.Link.ProductionBranch,
		}
	}
	return nil
}

type Protection struct {
	DeploymentType string `json:"deploymentType"`
}

type ProtectionBypass struct {
	Scope string `json:"scope"`
}

// ProjectResponse defines the information Vercel returns about a project.
type ProjectResponse struct {
	BuildCommand                *string               `json:"buildCommand"`
	CommandForIgnoringBuildStep *string               `json:"commandForIgnoringBuildStep"`
	DevCommand                  *string               `json:"devCommand"`
	EnvironmentVariables        []EnvironmentVariable `json:"env"`
	Framework                   *string               `json:"framework"`
	ID                          string                `json:"id"`
	TeamID                      string                `json:"-"`
	InstallCommand              *string               `json:"installCommand"`
	Link                        *struct {
		Type string `json:"type"`
		// github
		Org  string `json:"org"`
		Repo string `json:"repo"`
		// bitbucket
		Owner string `json:"owner"`
		Slug  string `json:"slug"`
		// gitlab
		ProjectNamespace string `json:"projectNamespace"`
		ProjectURL       string `json:"projectUrl"`
		ProjectID        int64  `json:"projectId,string"`
		// production branch
		ProductionBranch *string `json:"productionBranch"`
	} `json:"link"`
	Name                     string                      `json:"name"`
	OutputDirectory          *string                     `json:"outputDirectory"`
	PublicSource             *bool                       `json:"publicSource"`
	RootDirectory            *string                     `json:"rootDirectory"`
	ServerlessFunctionRegion *string                     `json:"serverlessFunctionRegion"`
	SSOProtection            *Protection                 `json:"ssoProtection"`
	PasswordProtection       *Protection                 `json:"passwordProtection"`
	ProtectionBypass         map[string]ProtectionBypass `json:"protectionBypass"`
}

// GetProject retrieves information about an existing project from Vercel.
func (c *Client) GetProject(ctx context.Context, projectID, teamID string, shouldFetchEnvironmentVariables bool) (r ProjectResponse, err error) {
	url := fmt.Sprintf("%s/v10/projects/%s", c.baseURL, projectID)
	if c.teamID(teamID) != "" {
		url = fmt.Sprintf("%s?teamId=%s", url, c.teamID(teamID))
	}
	tflog.Trace(ctx, "getting project", map[string]interface{}{
		"url":                    url,
		"shouldFetchEnvironment": shouldFetchEnvironmentVariables,
	})
	err = c.doRequest(clientRequest{
		ctx:    ctx,
		method: "GET",
		url:    url,
		body:   "",
	}, &r)
	if err != nil {
		return r, fmt.Errorf("unable to get project: %w", err)
	}

	if shouldFetchEnvironmentVariables {
		r.EnvironmentVariables, err = c.getEnvironmentVariables(ctx, projectID, teamID)
		if err != nil {
			return r, fmt.Errorf("error getting environment variables for project: %w", err)
		}
	} else {
		// The get project endpoint returns environment variables, but returns them fully
		// encrypted. This isn't useful, so we just remove them.
		r.EnvironmentVariables = nil
	}
	r.TeamID = c.teamID(teamID)
	return r, err
}
