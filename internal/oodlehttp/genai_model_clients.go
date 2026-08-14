package oodlehttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// Clients for the individual GenAI objects.
//
// The endpoints are not uniform — some update with PUT and some with
// PATCH, some are keyed by name rather than id, and two have no
// read-by-id at all — so each object gets its own small client rather
// than sharing the generic ModelClient.

// GenAIDatasetClient manages datasets, which are keyed by name.
type GenAIDatasetClient struct {
	*GenAIClient
}

func NewGenAIDatasetClient(client *OodleApiClient) *GenAIDatasetClient {
	return &GenAIDatasetClient{GenAIClient: NewGenAIClient(client)}
}

func (c *GenAIDatasetClient) Create(
	ctx context.Context,
	dataset *clientmodels.GenAIDataset,
) (*clientmodels.GenAIDataset, error) {
	created := &clientmodels.GenAIDataset{}
	if err := c.Do(
		ctx, http.MethodPost, "datasets", dataset, created,
	); err != nil {
		return nil, err
	}

	return created, nil
}

// Get reads a dataset by name.
func (c *GenAIDatasetClient) Get(
	ctx context.Context,
	name string,
) (*clientmodels.GenAIDataset, error) {
	dataset := &clientmodels.GenAIDataset{}
	if err := c.Do(
		ctx,
		http.MethodGet,
		"datasets/"+url.PathEscape(name),
		nil,
		dataset,
	); err != nil {
		return nil, err
	}

	return dataset, nil
}

// Update is never reached: the API has no dataset update endpoint,
// so every attribute of the dataset resource forces replacement.
func (c *GenAIDatasetClient) Update(
	_ context.Context,
	_ *clientmodels.GenAIDataset,
) (*clientmodels.GenAIDataset, error) {
	return nil, errors.New(
		"datasets cannot be updated in place; change forces replacement",
	)
}

// Delete removes a dataset by name, along with its items.
func (c *GenAIDatasetClient) Delete(ctx context.Context, name string) error {
	return c.Do(
		ctx,
		http.MethodDelete,
		"datasets/"+url.PathEscape(name),
		nil,
		nil,
	)
}

// GenAIDatasetItemClient manages the rows of a dataset.
type GenAIDatasetItemClient struct {
	*GenAIClient
}

func NewGenAIDatasetItemClient(
	client *OodleApiClient,
) *GenAIDatasetItemClient {
	return &GenAIDatasetItemClient{GenAIClient: NewGenAIClient(client)}
}

func (c *GenAIDatasetItemClient) Create(
	ctx context.Context,
	item *clientmodels.GenAIDatasetItem,
) (*clientmodels.GenAIDatasetItem, error) {
	created := &clientmodels.GenAIDatasetItem{}
	if err := c.Do(
		ctx, http.MethodPost, "dataset-items", item, created,
	); err != nil {
		return nil, err
	}

	return created, nil
}

func (c *GenAIDatasetItemClient) Get(
	ctx context.Context,
	id string,
) (*clientmodels.GenAIDatasetItem, error) {
	item := &clientmodels.GenAIDatasetItem{}
	if err := c.Do(
		ctx,
		http.MethodGet,
		"dataset-items/"+url.PathEscape(id),
		nil,
		item,
	); err != nil {
		return nil, err
	}

	return item, nil
}

func (c *GenAIDatasetItemClient) Update(
	ctx context.Context,
	item *clientmodels.GenAIDatasetItem,
) (*clientmodels.GenAIDatasetItem, error) {
	updated := &clientmodels.GenAIDatasetItem{}
	if err := c.Do(
		ctx,
		http.MethodPut,
		"dataset-items/"+url.PathEscape(item.ID),
		item,
		updated,
	); err != nil {
		return nil, err
	}

	return updated, nil
}

func (c *GenAIDatasetItemClient) Delete(ctx context.Context, id string) error {
	return c.Do(
		ctx,
		http.MethodDelete,
		"dataset-items/"+url.PathEscape(id),
		nil,
		nil,
	)
}

// GenAIEvalTemplateClient manages eval templates. Updates are a
// PATCH that merges: a field left empty keeps its stored value.
type GenAIEvalTemplateClient struct {
	*GenAIClient
}

func NewGenAIEvalTemplateClient(
	client *OodleApiClient,
) *GenAIEvalTemplateClient {
	return &GenAIEvalTemplateClient{GenAIClient: NewGenAIClient(client)}
}

func (c *GenAIEvalTemplateClient) Create(
	ctx context.Context,
	template *clientmodels.GenAIEvalTemplate,
) (*clientmodels.GenAIEvalTemplate, error) {
	created := &clientmodels.GenAIEvalTemplate{}
	if err := c.Do(
		ctx, http.MethodPost, "eval-templates", template, created,
	); err != nil {
		return nil, err
	}

	return created, nil
}

func (c *GenAIEvalTemplateClient) Get(
	ctx context.Context,
	id string,
) (*clientmodels.GenAIEvalTemplate, error) {
	template := &clientmodels.GenAIEvalTemplate{}
	if err := c.Do(
		ctx,
		http.MethodGet,
		"eval-templates/"+url.PathEscape(id),
		nil,
		template,
	); err != nil {
		return nil, err
	}

	return template, nil
}

func (c *GenAIEvalTemplateClient) Update(
	ctx context.Context,
	template *clientmodels.GenAIEvalTemplate,
) (*clientmodels.GenAIEvalTemplate, error) {
	updated := &clientmodels.GenAIEvalTemplate{}
	if err := c.Do(
		ctx,
		http.MethodPatch,
		"eval-templates/"+url.PathEscape(template.ID),
		template,
		updated,
	); err != nil {
		return nil, err
	}

	return updated, nil
}

func (c *GenAIEvalTemplateClient) Delete(ctx context.Context, id string) error {
	return c.Do(
		ctx,
		http.MethodDelete,
		"eval-templates/"+url.PathEscape(id),
		nil,
		nil,
	)
}

// GenAIEvaluationRuleClient manages evaluation rules. The API has no
// read-by-id for rules, so Get filters the list.
type GenAIEvaluationRuleClient struct {
	*GenAIClient
}

func NewGenAIEvaluationRuleClient(
	client *OodleApiClient,
) *GenAIEvaluationRuleClient {
	return &GenAIEvaluationRuleClient{GenAIClient: NewGenAIClient(client)}
}

func (c *GenAIEvaluationRuleClient) Create(
	ctx context.Context,
	rule *clientmodels.GenAIEvaluationRule,
) (*clientmodels.GenAIEvaluationRule, error) {
	created := &clientmodels.GenAIEvaluationRule{}
	if err := c.Do(
		ctx, http.MethodPost, "evaluation-rules", rule, created,
	); err != nil {
		return nil, err
	}

	return created, nil
}

func (c *GenAIEvaluationRuleClient) Get(
	ctx context.Context,
	id string,
) (*clientmodels.GenAIEvaluationRule, error) {
	rules, err := genaiList[*clientmodels.GenAIEvaluationRule](
		ctx, c.GenAIClient, "evaluation-rules", nil,
	)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if rule.ID == id {
			return rule, nil
		}
	}

	return nil, ErrNotFound
}

func (c *GenAIEvaluationRuleClient) Update(
	ctx context.Context,
	rule *clientmodels.GenAIEvaluationRule,
) (*clientmodels.GenAIEvaluationRule, error) {
	updated := &clientmodels.GenAIEvaluationRule{}
	if err := c.Do(
		ctx,
		http.MethodPatch,
		"evaluation-rules/"+url.PathEscape(rule.ID),
		rule,
		updated,
	); err != nil {
		return nil, err
	}

	return updated, nil
}

func (c *GenAIEvaluationRuleClient) Delete(
	ctx context.Context,
	id string,
) error {
	return c.Do(
		ctx,
		http.MethodDelete,
		"evaluation-rules/"+url.PathEscape(id),
		nil,
		nil,
	)
}

// GenAILLMConnectionClient manages LLM provider connections. Like
// evaluation rules these have no read-by-id, so Get filters the list.
type GenAILLMConnectionClient struct {
	*GenAIClient
}

func NewGenAILLMConnectionClient(
	client *OodleApiClient,
) *GenAILLMConnectionClient {
	return &GenAILLMConnectionClient{GenAIClient: NewGenAIClient(client)}
}

func (c *GenAILLMConnectionClient) Create(
	ctx context.Context,
	connection *clientmodels.GenAILLMConnection,
) (*clientmodels.GenAILLMConnection, error) {
	created := &clientmodels.GenAILLMConnection{}
	if err := c.Do(
		ctx, http.MethodPost, "llm-connections", connection, created,
	); err != nil {
		return nil, err
	}

	return created, nil
}

func (c *GenAILLMConnectionClient) Get(
	ctx context.Context,
	id string,
) (*clientmodels.GenAILLMConnection, error) {
	connections, err := genaiList[*clientmodels.GenAILLMConnection](
		ctx, c.GenAIClient, "llm-connections", nil,
	)
	if err != nil {
		return nil, err
	}

	for _, connection := range connections {
		if connection.ID == id {
			return connection, nil
		}
	}

	return nil, ErrNotFound
}

func (c *GenAILLMConnectionClient) Update(
	ctx context.Context,
	connection *clientmodels.GenAILLMConnection,
) (*clientmodels.GenAILLMConnection, error) {
	updated := &clientmodels.GenAILLMConnection{}
	if err := c.Do(
		ctx,
		http.MethodPut,
		"llm-connections/"+url.PathEscape(connection.ID),
		connection,
		updated,
	); err != nil {
		return nil, err
	}

	return updated, nil
}

func (c *GenAILLMConnectionClient) Delete(
	ctx context.Context,
	id string,
) error {
	return c.Do(
		ctx,
		http.MethodDelete,
		"llm-connections/"+url.PathEscape(id),
		nil,
		nil,
	)
}

// GenAIPromptClient manages prompts, which are append-only: creating
// a version and moving a label are the only writes.
type GenAIPromptClient struct {
	*GenAIClient
}

func NewGenAIPromptClient(client *OodleApiClient) *GenAIPromptClient {
	return &GenAIPromptClient{GenAIClient: NewGenAIClient(client)}
}

// Create adds a new version of the named prompt.
func (c *GenAIPromptClient) Create(
	ctx context.Context,
	prompt *clientmodels.GenAIPrompt,
) (*clientmodels.GenAIPrompt, error) {
	created := &clientmodels.GenAIPrompt{}
	if err := c.Do(
		ctx, http.MethodPost, "v2/prompts", prompt, created,
	); err != nil {
		return nil, err
	}

	return created, nil
}

// GetVersion reads one exact version of a prompt.
func (c *GenAIPromptClient) GetVersion(
	ctx context.Context,
	name string,
	version int64,
) (*clientmodels.GenAIPrompt, error) {
	params := url.Values{}
	params.Set("version", strconv.FormatInt(version, 10))
	// Dependency tags in the prompt body are stored literally;
	// resolving them would make Terraform see a permanent diff
	// against the configured text.
	params.Set("resolve", "false")

	prompt := &clientmodels.GenAIPrompt{}
	if err := c.Do(
		ctx,
		http.MethodGet,
		"v2/prompts/"+url.PathEscape(name)+"?"+params.Encode(),
		nil,
		prompt,
	); err != nil {
		return nil, err
	}

	return prompt, nil
}

// SetVersionLabels makes the given version's labels exactly labels.
func (c *GenAIPromptClient) SetVersionLabels(
	ctx context.Context,
	name string,
	version int64,
	labels []string,
) error {
	if labels == nil {
		labels = []string{}
	}

	return c.Do(
		ctx,
		http.MethodPatch,
		"v2/prompts/"+url.PathEscape(name),
		map[string]any{"version": version, "labels": labels},
		nil,
	)
}

// DeleteVersion removes a single version of a prompt.
func (c *GenAIPromptClient) DeleteVersion(
	ctx context.Context,
	name string,
	version int64,
) error {
	return c.Do(
		ctx,
		http.MethodDelete,
		"v2/prompts/"+url.PathEscape(name)+"/versions/"+strconv.FormatInt(version, 10),
		nil,
		nil,
	)
}
