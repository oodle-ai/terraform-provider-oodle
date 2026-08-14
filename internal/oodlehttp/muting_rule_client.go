package oodlehttp

import (
	"context"

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
