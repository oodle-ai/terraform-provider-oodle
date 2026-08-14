package oodlehttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"
)

// genaiBasePath is the prefix every GenAI (llmops) endpoint sits
// under. The path is Langfuse-SDK compatible, which is why it does
// not follow the flat "/v1/api/instance/{instance}/{resource}"
// shape the other Oodle APIs use.
const genaiBasePath = "%v/v1/api/instance/%v/langfuse/api/public/%v"

// GenAIClient issues requests against the GenAI endpoints. It is a
// thin transport: the per-object clients below layer the resource
// paths and update verbs on top.
type GenAIClient struct {
	*OodleApiClient
}

// NewGenAIClient creates a new GenAIClient.
func NewGenAIClient(client *OodleApiClient) *GenAIClient {
	return &GenAIClient{OodleApiClient: client}
}

// Do issues a request against subPath, relative to the GenAI base
// path, and decodes a JSON response into out when out is non-nil.
//
// Callers are responsible for escaping any path segment that comes
// from user input; use url.PathEscape. Prompt and dataset names are
// user-chosen and may legitimately contain "/".
func (c *GenAIClient) Do(
	ctx context.Context,
	method string,
	subPath string,
	body any,
	out any,
) error {
	var reqBody io.Reader
	if body != nil {
		encoded, err := jsoniter.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf(genaiBasePath, c.DeploymentUrl, c.Instance, subPath),
		reqBody,
	)
	if err != nil {
		return err
	}

	req.Header = http.Header(c.Headers).Clone()
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf(
			"%s %s failed: %v, body: %v",
			method,
			subPath,
			resp.Status,
			string(bodyBytes),
		)
	}

	if out == nil {
		return nil
	}

	return jsoniter.Unmarshal(bodyBytes, out)
}

// genaiListEnvelope is the shape every GenAI list endpoint returns.
// Unlike the other Oodle list APIs, the rows are wrapped rather than
// returned as a bare array.
type genaiListEnvelope[T any] struct {
	Data []T `json:"data"`
}

// genaiList reads one page of a list endpoint.
func genaiList[T any](
	ctx context.Context,
	c *GenAIClient,
	subPath string,
	params url.Values,
) ([]T, error) {
	if len(params) > 0 {
		subPath += "?" + params.Encode()
	}

	var envelope genaiListEnvelope[T]
	if err := c.Do(ctx, http.MethodGet, subPath, nil, &envelope); err != nil {
		return nil, err
	}

	return envelope.Data, nil
}
