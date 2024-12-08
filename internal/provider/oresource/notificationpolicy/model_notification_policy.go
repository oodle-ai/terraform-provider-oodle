// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notificationPolicy

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
	"terraform-provider-oodle/internal/validatorutils"
)

type notificationPolicyResourceModel struct {
	ID            types.String        `tfsdk:"id"`
	Name          types.String        `tfsdk:"name"`
	Notifiers     notifiersBySeverity `tfsdk:"notifiers"`
	Global        types.Bool          `tfsdk:"global"`
	MuteGlobal    types.Bool          `tfsdk:"mute_global"`
	MuteNonGlobal types.Bool          `tfsdk:"mute_non_global"`
}

type notifiersBySeverity struct {
	Warn     types.List `tfsdk:"warn"`
	Critical types.List `tfsdk:"critical"`
}

var _ resourceutils.ResourceModel[*clientmodels.NotificationPolicy] = (*notificationPolicyResourceModel)(nil)

func (n *notificationPolicyResourceModel) FromClientModel(model *clientmodels.NotificationPolicy, diagnosticsOut *diag.Diagnostics) {
	n.ID = types.StringValue(model.ID.UUID.String())
	n.Name = types.StringValue(model.Name)
	n.Global = types.BoolValue(model.Global)
	n.MuteGlobal = types.BoolValue(model.MuteGlobal)
	n.MuteNonGlobal = types.BoolValue(model.MuteNonGlobal)

	if len(model.Notifiers.Critical) > 0 {
		n.Notifiers.Critical = validatorutils.IDsToAttrList(model.Notifiers.Critical, diagnosticsOut)
	} else {
		n.Notifiers.Critical = types.ListNull(types.StringType)
	}

	if len(model.Notifiers.Warn) > 0 {
		n.Notifiers.Warn = validatorutils.IDsToAttrList(model.Notifiers.Warn, diagnosticsOut)
	} else {
		n.Notifiers.Warn = types.ListNull(types.StringType)
	}
}

func (n *notificationPolicyResourceModel) ToClientModel(model *clientmodels.NotificationPolicy) error {
	var err error
	if !n.ID.IsNull() && !n.ID.IsUnknown() {
		model.ID.UUID, err = uuid.Parse(n.ID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse ID UUID %v: %v", n.ID.ValueString(), err)
		}
	}

	model.Name = n.Name.ValueString()
	model.Global = n.Global.ValueBool()
	model.MuteGlobal = n.MuteGlobal.ValueBool()
	model.MuteNonGlobal = n.MuteNonGlobal.ValueBool()

	criticalNotifIDs := n.Notifiers.Critical.Elements()
	if len(criticalNotifIDs) > 0 {
		model.Notifiers.Critical, err = validatorutils.AttrListToIDs(n.Notifiers.Critical)
		if err != nil {
			return err
		}
	}

	warnNotifIDs := n.Notifiers.Warn.Elements()
	if len(warnNotifIDs) > 0 {
		model.Notifiers.Warn, err = validatorutils.AttrListToIDs(n.Notifiers.Warn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *notificationPolicyResourceModel) SetID(id types.String) {
	n.ID = id
}

func (n *notificationPolicyResourceModel) GetID() types.String {
	return n.ID
}
