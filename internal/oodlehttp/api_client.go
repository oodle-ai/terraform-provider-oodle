package oodlehttp

import (
	"net/http"
)

const (
	maxConnections    = 30
	OodleApiKeyHeader = "X-API-KEY"
)

type OodleApiClient struct {
	HttpClient    *http.Client
	DeploymentUrl string
	Instance      string
	ApiKey        string
	Headers       map[string][]string
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

func NewInstanceClient(
	deploymentUrl string,
	instance string,
	apiKey string,
) (*OodleApiClient, error) {
	return &OodleApiClient{
		HttpClient:    newHttpClient(),
		DeploymentUrl: deploymentUrl,
		Instance:      instance,
		ApiKey:        apiKey,
		Headers: map[string][]string{
			OodleApiKeyHeader: {apiKey},
		},
	}, nil
}
