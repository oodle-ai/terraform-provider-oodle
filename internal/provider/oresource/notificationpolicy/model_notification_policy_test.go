package notificationPolicy

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestNotificationPolicyModel(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.NotificationPolicy{
		ID: clientmodels.ID{
			UUID: uuid.New(),
		},
		Name: "test",
		Notifiers: clientmodels.NotifiersByCondition{
			Warn:     []clientmodels.ID{{UUID: uuid.New()}, {UUID: uuid.New()}},
			Critical: []clientmodels.ID{{UUID: uuid.New()}, {UUID: uuid.New()}, {UUID: uuid.New()}},
		},
		Global:        true,
		MuteGlobal:    true,
		MuteNonGlobal: true,
	}

	resourceModel := &notificationPolicyResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.NotificationPolicy{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}
