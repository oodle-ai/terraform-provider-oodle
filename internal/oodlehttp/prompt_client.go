package oodlehttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// The "langfuse" segment is on the wire because the Langfuse
// SDKs point at it; the data is Oodle's.
const promptBasePath = "%v/v1/api/instance/%v/langfuse/api/public/v2/prompts"

// PromptClient handles GenAI prompt operations.
//
// Prompts do not fit the generic ModelClient. A version is never
// mutated: an update is a POST that creates the next one, and
// which version is live is decided by a label rather than by the
// row a PUT would overwrite.
type PromptClient struct {
	*OodleApiClient
}

// NewPromptClient creates a new PromptClient.
func NewPromptClient(client *OodleApiClient) *PromptClient {
	return &PromptClient{OodleApiClient: client}
}

func (c *PromptClient) collectionURL() string {
	return fmt.Sprintf(promptBasePath, c.DeploymentUrl, c.Instance)
}

// promptURL builds the URL for one prompt.
//
// A name routinely contains "/" (`groot/sub-agent/logs`).
// Unencoded it splits into extra path segments and 404s against
// a route that does not exist.
func (c *PromptClient) promptURL(name string) string {
	return c.collectionURL() + "/" + url.PathEscape(name)
}

// Get returns the prompt carrying the given label, or nil when
// there is none.
//
// `resolve=false` is not optional. The default expands a
// prompt's `@@@oodlePrompt@@@` references into the child's text,
// and an expansion never equals the reference it was stored as
// -- so a resolving read reports drift on every plan and every
// apply writes a new version that changes nothing.
func (c *PromptClient) Get(
	ctx context.Context,
	name string,
	label string,
) (*clientmodels.Prompt, error) {
	target := fmt.Sprintf(
		"%s?resolve=false&label=%s",
		c.promptURL(name),
		url.QueryEscape(label),
	)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, target, nil,
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

	// A prompt that is not there is not an error: it is how
	// Terraform learns the resource has to be created, or that
	// something removed it outside the state.
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to read prompt %v: %v, body: %v",
			name,
			resp.Status,
			string(bodyBytes),
		)
	}

	var result clientmodels.Prompt
	if err = jsoniter.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create publishes a new version and gives it the labels it
// carries. It is used for an update too: a version is immutable,
// so changing a prompt means adding the next one.
func (c *PromptClient) Create(
	ctx context.Context,
	prompt *clientmodels.Prompt,
) (*clientmodels.Prompt, error) {
	reqBody, err := jsoniter.Marshal(prompt)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.collectionURL(),
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

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(
			"failed to create prompt %v: %v, body: %v",
			prompt.Name,
			resp.Status,
			string(bodyBytes),
		)
	}

	var result clientmodels.Prompt
	if err = jsoniter.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes every version of a prompt.
func (c *PromptClient) Delete(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, c.promptURL(name), nil,
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

	// Already gone is the outcome a destroy wanted.
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(
			"failed to delete prompt %v: %v, body: %v",
			name,
			resp.Status,
			string(bodyBytes),
		)
	}
	return nil
}
