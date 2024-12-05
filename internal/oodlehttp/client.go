package oodlehttp

import (
	"fmt"
	"github.com/pkg/errors"
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
		return nil, errors.Wrapf(err, "ASDASDASD 1: %v", req.URL.String())
	}

	defer resp.Body.Close()
	var monitor models.Monitor
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "ASDASDASD 2")
	}

	if err = monitor.UnmarshalJSON(bodyBytes); err != nil {
		return nil, errors.Wrapf(err, "ASDASDASD %v", string(bodyBytes))
	}

	return &monitor, nil
}
