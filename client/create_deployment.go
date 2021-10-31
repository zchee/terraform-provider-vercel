package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type ProjectSettings struct {
	BuildCommand    string `json:"buildCommand,omitempty"`
	DevCommand      string `json:"devCommand,omitempty"`
	Framework       string `json:"framework,omitempty"`
	InstallCommand  string `json:"installCommand,omitempty"`
	OutputDirectory string `json:"outputDirectory,omitempty"`
	RootDirectory   string `json:"rootDirectory,omitempty"`
}

type DeploymentFile struct {
	File string `json:"file,omitempty"`
	Sha  string `json:"sha,omitempty"`
	Size int    `json:"size,omitempty"`
}

type Build struct {
	Environment map[string]string `json:"env,omitempty"`
}

type CreateDeploymentRequest struct {
	Aliases         []string               `json:"alias,omitempty"`
	Build           Build                  `json:"build"`
	Environment     map[string]string      `json:"env,omitempty"`
	Files           []DeploymentFile       `json:"files,omitempty"`
	Functions       map[string]interface{} `json:"functions,omitempty"`
	Meta            map[string]string      `json:"meta,omitempty"`
	ProjectID       string                 `json:"project,omitempty"`
	ProjectName     string                 `json:"name,omitempty"`
	ProjectSettings ProjectSettings        `json:"projectSettings,omitempty"`
	Public          bool                   `json:"public,omitempty"`
	Regions         []string               `json:"regions,omitempty"`
	Routes          []interface{}          `json:"routes,omitempty"`
	Target          string                 `json:"target,omitempty"`
}

type CreateDeploymentResponse struct {
	ID        string            `json:"id"`
	URL       string            `json:"url"`
	Meta      map[string]string `json:"meta"`
	CreatedIn string            `json:"createdIn"`
}

func buildDetectedRequest(cr CreateDeploymentRequest, apiErr APIError) (CreateDeploymentRequest, error) {
	var frameworkDetection struct {
		Error struct {
			Framework struct {
				Slug string `json:"slug"`
			} `json:"framework"`
		} `json:"error"`
	}

	err := json.Unmarshal(apiErr.RawMessage, &frameworkDetection)
	cr.ProjectSettings.Framework = frameworkDetection.Error.Framework.Slug
	return cr, err
}

func (c *Client) CreateDeployment(ctx context.Context, createRequest CreateDeploymentRequest) (r CreateDeploymentResponse, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/v12/now/deployments?skipAutoDetectionConfirmation=1",
		strings.NewReader(string(mustMarshal(createRequest))),
	)
	if err != nil {
		return r, err
	}

	err = c.doRequest(req, &r)
	return r, err
}
