package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestMonitorModel(t *testing.T) {
	ctx := context.Background()
	dur1 := time.Duration(1)
	dur2 := time.Duration(2)
	dur3 := time.Duration(3)
	clientModel := &clientmodels.Monitor{
		ID: clientmodels.ID{
			UUID: uuid.New(),
		},
		Name:        "test",
		PromQLQuery: "test2",
		Interval:    1,
		Conditions: clientmodels.ConditionBySeverity{
			Warn: &clientmodels.Condition{
				Op:            3,
				Value:         1,
				For:           5,
				KeepFiringFor: 2,
			},
			Critical: &clientmodels.Condition{
				Op:            2,
				Value:         2,
				For:           6,
				KeepFiringFor: 3,
			},
		},
		Labels: map[string]string{
			"test1": "test2",
			"test3": "test4",
		},
		Annotations: map[string]string{
			"test5": "test6",
			"test7": "test8",
		},
		Grouping: clientmodels.Grouping{
			ByLabels: []string{"test9", "test10"},
		},
		NotificationPolicyID: &clientmodels.ID{
			UUID: uuid.New(),
		},
		LabelMatcherNotificationPolicies: []clientmodels.LabelMatcherNotificationPolicy{
			{
				Matchers: []clientmodels.LabelMatcher{
					{
						Name:  "test1",
						Value: "test2",
					},
				},
				NotificationPolicyID: clientmodels.ID{
					UUID: uuid.New(),
				},
			},
		},
		GroupWait:      &dur1,
		GroupInterval:  &dur2,
		RepeatInterval: &dur3,
	}

	resourceModel := &monitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.Monitor{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}
