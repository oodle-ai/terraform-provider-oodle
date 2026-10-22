package notificationPolicy

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// warn and warning are one list: either name sends the warning
// notifiers, and a read keeps the name the config used.
func TestWarningIsTheWarnList(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	list := types.ListValueMust(types.StringType, []attr.Value{types.StringValue(id.String())})
	null := types.ListNull(types.StringType)
	for _, useWarning := range []bool{true, false} {
		n := &notificationPolicyResourceModel{
			Name: types.StringValue("p"),
			Notifiers: notifiersBySeverity{
				Warn: null, Warning: null, Critical: null,
			},
		}
		if useWarning {
			n.Notifiers.Warning = list
		} else {
			n.Notifiers.Warn = list
		}
		client := &clientmodels.NotificationPolicy{}
		assert.Nil(t, n.ToClientModel(ctx, client))
		assert.Equal(t, 1, len(client.Notifiers.Warn))
		assert.Equal(t, id, client.Notifiers.Warn[0].UUID)

		client.ID = clientmodels.ID{UUID: uuid.New()}
		diags := &diag.Diagnostics{}
		n.FromClientModel(ctx, client, diags)
		assert.False(t, diags.HasError())
		if useWarning {
			assert.Equal(t, 1, len(n.Notifiers.Warning.Elements()))
			assert.True(t, n.Notifiers.Warn.IsNull())
		} else {
			assert.Equal(t, 1, len(n.Notifiers.Warn.Elements()))
			assert.True(t, n.Notifiers.Warning.IsNull())
		}
	}
}
