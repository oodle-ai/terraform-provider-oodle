package oodlehttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

const apiBasePath = "%v/v1/api/instance/%v/"

// ModelClient is a client is used to access and update models from Oodle APIs.
type ModelClient[T clientmodels.ClientModel] struct {
	*OodleApiClient
	resourcePath string
	nilVal       T
	modelCreator func() T
}

// NewModelClient creates a new model client.
func NewModelClient[T clientmodels.ClientModel](
	client *OodleApiClient,
	resourcePath string,
	modelCreator func() T,
) *ModelClient[T] {
	return &ModelClient[T]{
		OodleApiClient: client,
		resourcePath:   resourcePath,
		modelCreator:   modelCreator,
	}
}

func (c *ModelClient[T]) Get(id string) (T, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(apiBasePath+c.resourcePath+"/%s", c.DeploymentUrl, c.Instance, id),
		nil,
	)
	if err != nil {
		return c.nilVal, err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return c.nilVal, err
	}

	defer resp.Body.Close()
	model := c.modelCreator()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.nilVal, err
	}
	if resp.StatusCode != http.StatusOK {
		return c.nilVal, fmt.Errorf("failed to get model %T: %v, body: %v", c.nilVal, resp.Status, string(bodyBytes))
	}

	if err = jsoniter.Unmarshal(bodyBytes, model); err != nil {
		return c.nilVal, err
	}

	return model, nil
}

func (c *ModelClient[T]) Delete(id string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf(apiBasePath+c.resourcePath+"/%s", c.DeploymentUrl, c.Instance, id),
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
		return fmt.Errorf("failed to delete model %T: %v, body: %v", c.nilVal, resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *ModelClient[T]) Create(model T) (T, error) {
	reqBody, err := jsoniter.Marshal(model)
	if err != nil {
		return c.nilVal, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(apiBasePath+c.resourcePath, c.DeploymentUrl, c.Instance),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return c.nilVal, err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return c.nilVal, err
	}

	defer resp.Body.Close()
	resModel := c.modelCreator()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.nilVal, err
	}

	if resp.StatusCode != http.StatusOK {
		return c.nilVal, fmt.Errorf("failed to update model %T: %v, body: %v", c.nilVal, resp.Status, string(bodyBytes))
	}

	if err = jsoniter.Unmarshal(bodyBytes, resModel); err != nil {
		return c.nilVal, err
	}

	return resModel, nil
}

func (c *ModelClient[T]) Update(model T) (T, error) {
	reqBody, err := jsoniter.Marshal(model)
	if err != nil {
		return c.nilVal, err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf(apiBasePath+c.resourcePath+"/%s", c.DeploymentUrl, c.Instance, model.GetID()),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return c.nilVal, err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return c.nilVal, err
	}

	defer resp.Body.Close()
	resModel := c.modelCreator()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.nilVal, err
	}

	if resp.StatusCode != http.StatusOK {
		return c.nilVal, fmt.Errorf("failed to update model %T: %v, body: %v", c.nilVal, resp.Status, string(bodyBytes))
	}

	if err = jsoniter.Unmarshal(bodyBytes, resModel); err != nil {
		return c.nilVal, err
	}

	return resModel, nil
}
