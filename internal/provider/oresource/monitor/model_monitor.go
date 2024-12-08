// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
	"terraform-provider-oodle/internal/validatorutils"
)

type monitorResourceModel struct {
	ID                   types.String     `tfsdk:"id"`
	Name                 types.String     `tfsdk:"name"`
	Interval             types.String     `tfsdk:"interval"`
	PromQLQuery          types.String     `tfsdk:"promql_query"`
	Conditions           *conditionsModel `tfsdk:"conditions"`
	Labels               types.Map        `tfsdk:"labels"`
	Annotations          types.Map        `tfsdk:"annotations"`
	Grouping             *grouping        `tfsdk:"grouping"`
	NotificationPolicyID types.String     `tfsdk:"notification_policy_id"`
	GroupWait            types.String     `tfsdk:"group_wait"`
	GroupInterval        types.String     `tfsdk:"group_interval"`
	RepeatInterval       types.String     `tfsdk:"repeat_interval"`
}

var _ resourceutils.ResourceModel[*clientmodels.Monitor] = (*monitorResourceModel)(nil)

func (m *monitorResourceModel) GetID() types.String {
	return m.ID
}

func (m *monitorResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *monitorResourceModel) FromClientModel(
	model *clientmodels.Monitor,
	diagnosticsOut *diag.Diagnostics,
) {
	// Reset the model to clear any existing data.
	*m = monitorResourceModel{}

	m.ID = types.StringValue(model.ID.UUID.String())
	m.Name = types.StringValue(model.Name)
	m.PromQLQuery = types.StringValue(model.PromQLQuery)
	if model.Interval > 0 {
		m.Interval = types.StringValue(validatorutils.ShortDur(model.Interval))
	}
	if model.Conditions.Warn != nil {
		if m.Conditions == nil {
			m.Conditions = &conditionsModel{}
		}

		m.Conditions.Warning = newConditionFromModel(model.Conditions.Warn)
	}

	if model.Conditions.Critical != nil {
		if m.Conditions == nil {
			m.Conditions = &conditionsModel{}
		}

		m.Conditions.Critical = newConditionFromModel(model.Conditions.Critical)
	}

	if len(model.Labels) > 0 {
		m.Labels = validatorutils.ToAttrMap(model.Labels, diagnosticsOut)
	} else {
		m.Labels = types.MapNull(basetypes.StringType{})
	}

	if len(model.Annotations) > 0 {
		m.Annotations = validatorutils.ToAttrMap(model.Annotations, diagnosticsOut)
	} else {
		m.Annotations = types.MapNull(basetypes.StringType{})
	}

	if len(model.Grouping.ByLabels) > 0 || model.Grouping.Disabled || model.Grouping.ByMonitor {
		m.Grouping = &grouping{}
		m.Grouping.ByMonitor = types.BoolValue(model.Grouping.ByMonitor)
		m.Grouping.ByLabels = validatorutils.ToAttrList(model.Grouping.ByLabels, diagnosticsOut)
		m.Grouping.Disabled = types.BoolValue(model.Grouping.Disabled)
		m.Grouping.ByMonitor = types.BoolValue(model.Grouping.ByMonitor)
	}

	if model.NotificationPolicyID != nil {
		m.NotificationPolicyID = types.StringValue(model.NotificationPolicyID.UUID.String())
	}

	if model.GroupWait != nil {
		m.GroupWait = types.StringValue(validatorutils.ShortDur(*model.GroupWait))
	}

	if model.GroupInterval != nil {
		m.GroupInterval = types.StringValue(validatorutils.ShortDur(*model.GroupInterval))
	}

	if model.RepeatInterval != nil {
		m.RepeatInterval = types.StringValue(validatorutils.ShortDur(*model.RepeatInterval))
	}
}

func (m *monitorResourceModel) ToClientModel(
	model *clientmodels.Monitor,
) error {
	var err error
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID.UUID, err = uuid.Parse(m.ID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse ID UUID %v: %v", m.ID.ValueString(), err)
		}
	}

	model.Name = m.Name.ValueString()
	model.PromQLQuery = m.PromQLQuery.ValueString()
	if !m.Interval.IsNull() {
		model.Interval, err = time.ParseDuration(m.Interval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse interval duration: %v", err)
		}
	}

	if m.Conditions != nil {
		if m.Conditions.Warning != nil {
			model.Conditions.Warn, err = m.Conditions.Warning.toModel()
			if err != nil {
				return fmt.Errorf("failed to parse warning condition: %v", err)
			}
		}

		if m.Conditions.Critical != nil {
			model.Conditions.Critical, err = m.Conditions.Critical.toModel()
			if err != nil {
				return fmt.Errorf("failed to parse critical condition: %v", err)
			}
		}
	}

	if len(m.Labels.Elements()) > 0 {
		model.Labels = make(map[string]string)
		for k, v := range m.Labels.Elements() {
			strVal, ok := v.(validatorutils.StringValue)
			if !ok {
				return fmt.Errorf("failed to parse label value as string: %v, type is %T", v, v)
			}

			model.Labels[k] = strVal.ValueString()
		}
	}

	if len(m.Annotations.Elements()) > 0 {
		model.Annotations = make(map[string]string)
		for k, v := range m.Annotations.Elements() {
			strVal, ok := v.(validatorutils.StringValue)
			if !ok {
				return fmt.Errorf("failed to parse label value as string: %v, type is %T", v, v)
			}

			model.Annotations[k] = strVal.ValueString()
		}
	}

	if m.Grouping != nil {
		model.Grouping.ByMonitor = m.Grouping.ByMonitor.ValueBool()
		if len(m.Grouping.ByLabels.Elements()) > 0 {
			model.Grouping.ByLabels = make([]string, 0, len(m.Grouping.ByLabels.Elements()))
			for _, v := range m.Grouping.ByLabels.Elements() {
				strVal, ok := v.(validatorutils.StringValue)
				if !ok {
					return fmt.Errorf("failed to parse grouping labels as string: %v, type is %T", v, v)
				}

				model.Grouping.ByLabels = append(model.Grouping.ByLabels, strVal.ValueString())
			}

			model.Grouping.Disabled = m.Grouping.Disabled.ValueBool()
			model.Grouping.ByMonitor = m.Grouping.ByMonitor.ValueBool()
		}
	}

	if len(m.NotificationPolicyID.ValueString()) > 0 {
		uid, err := uuid.Parse(m.NotificationPolicyID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse notification policy UUID: %v", err)
		}

		model.NotificationPolicyID = &clientmodels.ID{UUID: uid}
	}

	if len(m.GroupWait.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.GroupWait.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse group wait duration: %v", err)
		}

		model.GroupWait = &dur
	}

	if len(m.GroupInterval.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.GroupInterval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse group interval duration: %v", err)
		}

		model.GroupInterval = &dur
	}

	if len(m.RepeatInterval.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.RepeatInterval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse repeat interval duration: %v", err)
		}

		model.RepeatInterval = &dur
	}

	return nil
}
