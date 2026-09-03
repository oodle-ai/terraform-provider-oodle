package oodlehttp

import (
	"context"
	"errors"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// mutingRulesResourcePath is the collection endpoint for muting rules.
const mutingRulesResourcePath = "muting-rules"

// MutingRuleClient manages muting rules.
//
// The API has no PUT for muting rules. POST is an upsert instead: a
// body carrying an id updates that rule, and a body without one
// creates a rule. Everything else is the standard model client.
type MutingRuleClient struct {
	*ModelClient[*clientmodels.MutingRule]
}

func NewMutingRuleClient(client *OodleApiClient) *MutingRuleClient {
	return &MutingRuleClient{
		ModelClient: NewModelClient[*clientmodels.MutingRule](
			client,
			mutingRulesResourcePath,
			func() *clientmodels.MutingRule {
				return &clientmodels.MutingRule{}
			},
		),
	}
}

// Update upserts the rule by POSTing it with its id.
func (c *MutingRuleClient) Update(
	ctx context.Context,
	rule *clientmodels.MutingRule,
) (*clientmodels.MutingRule, error) {
	return c.ModelClient.Create(ctx, rule)
}

// Get reads the rule by id, falling back to the collection.
//
// This covers instances predating the API fix for recurring rules.
// There, read-by-id only resolved one-off rules, which are stored as
// Alertmanager silences; a recurring rule — the kind that carries
// scheduleIds, and so the kind a muting schedule is used through —
// answered 404 however live it was, though it was listed by the
// collection endpoint and deleted by id perfectly well. Without the
// fallback Terraform reads every such rule as deleted, drops it from
// state, and creates a second copy on the next apply. Against a fixed
// API the direct read succeeds and the collection is never fetched.
func (c *MutingRuleClient) Get(
	ctx context.Context,
	id string,
) (*clientmodels.MutingRule, error) {
	rule, err := c.ModelClient.Get(ctx, id)
	if err == nil {
		return rule, nil
	}

	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	rules, listErr := c.ModelClient.List(ctx)
	if listErr != nil {
		return nil, listErr
	}

	for _, listed := range rules {
		if listed.GetID() == id {
			return listed, nil
		}
	}

	// Absent from the collection too, so it really is gone.
	return nil, ErrNotFound
}
