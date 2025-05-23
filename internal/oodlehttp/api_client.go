package oodlehttp

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

const (
	maxConnections    = 30
	OodleApiKeyHeader = "X-API-KEY"
)

// OodleApiClient is a http client with credentials to access Oodle APIs.
type OodleApiClient struct {
	HttpClient    *http.Client
	DeploymentUrl string
	Instance      string
	ApiKey        string
	Headers       map[string][]string
}

func newHttpClient() *http.Client {
	tr, _ := http.DefaultTransport.(*http.Transport)
	t := tr.Clone()
	t.MaxIdleConns = maxConnections
	t.MaxConnsPerHost = maxConnections
	t.MaxIdleConnsPerHost = maxConnections
	return &http.Client{
		Transport: logging.NewLoggingHTTPTransport(t),
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
