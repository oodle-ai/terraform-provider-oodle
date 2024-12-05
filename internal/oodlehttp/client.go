package oodlehttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"terraform-provider-oodle/internal/oodlehttp/models"
)

const (
	maxConnections    = 30
	OodleApiKeyHeader = "X-API-KEY"
)

type Client struct {
	httpClient    *http.Client
	deploymentUrl string
	instance      string
	apiKey        string
	headers       map[string][]string
}

func newHttpClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = maxConnections
	t.MaxConnsPerHost = maxConnections
	t.MaxIdleConnsPerHost = maxConnections
	return &http.Client{
		Transport: t,
	}
}

func NewClient(
	deploymentUrl string,
	instance string,
	apiKey string,
) (*Client, error) {
	return &Client{
		httpClient:    newHttpClient(),
		deploymentUrl: deploymentUrl,
		instance:      instance,
		apiKey:        apiKey,
		headers: map[string][]string{
			OodleApiKeyHeader: {apiKey},
		},
	}, nil
}

func (c *Client) GetMonitor(monitorId string) (*models.Monitor, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(`%v/v1/api/instance/%v/monitors/%v`, c.deploymentUrl, c.instance, monitorId),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header = c.headers
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var monitor models.Monitor
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = monitor.UnmarshalJSON(bodyBytes); err != nil {
		return nil, err
	}

	return &monitor, nil
}

func (c *Client) CreateMonitor(monitor *models.Monitor) (*models.Monitor, error) {
	reqBody, err := monitor.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(`%v/v1/api/instance/%v/monitors`, c.deploymentUrl, c.instance),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header = c.headers
	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()
	var resMonitor models.Monitor
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = resMonitor.UnmarshalJSON(bodyBytes); err != nil {
		return nil, err
	}

	return &resMonitor, nil
}
