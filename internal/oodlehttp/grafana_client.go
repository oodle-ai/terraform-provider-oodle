package oodlehttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

const grafanaBasePath = "%v/v1/api/instance/%v/grafana"

// GrafanaFolderClient handles Grafana folder operations.
type GrafanaFolderClient struct {
	*OodleApiClient
}

// NewGrafanaFolderClient creates a new GrafanaFolderClient.
func NewGrafanaFolderClient(client *OodleApiClient) *GrafanaFolderClient {
	return &GrafanaFolderClient{OodleApiClient: client}
}

// Create creates a new folder.
func (c *GrafanaFolderClient) Create(
	ctx context.Context,
	folder *clientmodels.GrafanaFolder,
) (*clientmodels.GrafanaFolder, error) {
	reqBody, err := jsoniter.Marshal(folder)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(grafanaBasePath+"/folders", c.DeploymentUrl, c.Instance),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header = c.Headers
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to create folder: %v, body: %v",
			resp.Status,
			string(bodyBytes),
		)
	}

	var result clientmodels.GrafanaFolder
	if err = jsoniter.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get gets a folder by UID.
func (c *GrafanaFolderClient) Get(
	ctx context.Context,
	uid string,
) (*clientmodels.GrafanaFolder, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			grafanaBasePath+"/folders/%s",
			c.DeploymentUrl,
			c.Instance,
			uid,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to get folder %s: %v, body: %v",
			uid,
			resp.Status,
			string(bodyBytes),
		)
	}

	var result clientmodels.GrafanaFolder
	if err = jsoniter.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates a folder.
func (c *GrafanaFolderClient) Update(
	ctx context.Context,
	folder *clientmodels.GrafanaFolder,
) (*clientmodels.GrafanaFolder, error) {
	reqBody, err := jsoniter.Marshal(map[string]interface{}{
		"title":     folder.Title,
		"version":   folder.Version,
		"overwrite": true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		fmt.Sprintf(
			grafanaBasePath+"/folders/%s",
			c.DeploymentUrl,
			c.Instance,
			folder.UID,
		),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header = c.Headers
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to update folder %s: %v, body: %v",
			folder.UID,
			resp.Status,
			string(bodyBytes),
		)
	}

	var result clientmodels.GrafanaFolder
	if err = jsoniter.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a folder by UID.
func (c *GrafanaFolderClient) Delete(ctx context.Context, uid string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf(
			grafanaBasePath+"/folders/%s",
			c.DeploymentUrl,
			c.Instance,
			uid,
		),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"failed to delete folder %s: %v, body: %v",
			uid,
			resp.Status,
			string(bodyBytes),
		)
	}

	return nil
}

// GrafanaDashboardClient handles Grafana dashboard operations.
type GrafanaDashboardClient struct {
	*OodleApiClient
}

// NewGrafanaDashboardClient creates a new GrafanaDashboardClient.
func NewGrafanaDashboardClient(client *OodleApiClient) *GrafanaDashboardClient {
	return &GrafanaDashboardClient{OodleApiClient: client}
}

// Create creates a new dashboard.
func (c *GrafanaDashboardClient) Create(
	ctx context.Context,
	dashboard *clientmodels.GrafanaDashboard,
) (*clientmodels.GrafanaDashboardResponse, error) {
	reqBody, err := jsoniter.Marshal(dashboard)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(grafanaBasePath+"/dashboards", c.DeploymentUrl, c.Instance),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header = c.Headers
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to create dashboard: %v, body: %v",
			resp.Status,
			string(bodyBytes),
		)
	}

	var result clientmodels.GrafanaDashboardResponse
	if err = jsoniter.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get gets a dashboard by UID.
func (c *GrafanaDashboardClient) Get(
	ctx context.Context,
	uid string,
) (*clientmodels.GrafanaDashboardGetResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			grafanaBasePath+"/dashboards/%s",
			c.DeploymentUrl,
			c.Instance,
			uid,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to get dashboard %s: %v, body: %v",
			uid,
			resp.Status,
			string(bodyBytes),
		)
	}

	var result clientmodels.GrafanaDashboardGetResponse
	if err = jsoniter.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates a dashboard.
func (c *GrafanaDashboardClient) Update(
	ctx context.Context,
	dashboard *clientmodels.GrafanaDashboard,
) (*clientmodels.GrafanaDashboardResponse, error) {
	dashboard.Overwrite = true
	return c.Create(ctx, dashboard)
}

// Delete deletes a dashboard by UID.
func (c *GrafanaDashboardClient) Delete(ctx context.Context, uid string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf(
			grafanaBasePath+"/dashboards/%s",
			c.DeploymentUrl,
			c.Instance,
			uid,
		),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"failed to delete dashboard %s: %v, body: %v",
			uid,
			resp.Status,
			string(bodyBytes),
		)
	}

	return nil
}
