package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/prometheus/alertmanager/pkg/labels"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
	"terraform-provider-oodle/internal/validatorutils"
)

type monitorResourceModel struct {
	ID                               types.String     `tfsdk:"id"`
	Name                             types.String     `tfsdk:"name"`
	Interval                         types.String     `tfsdk:"interval"`
	PromQLQuery                      types.String     `tfsdk:"promql_query"`
	Conditions                       *conditionsModel `tfsdk:"conditions"`
	Labels                           types.Map        `tfsdk:"labels"`
	Annotations                      types.Map        `tfsdk:"annotations"`
	Grouping                         *grouping        `tfsdk:"grouping"`
	NotificationPolicyID             types.String     `tfsdk:"notification_policy_id"`
	LabelMatcherNotificationPolicies types.List       `tfsdk:"label_matcher_notification_policies"`
	GroupWait                        types.String     `tfsdk:"group_wait"`
	GroupInterval                    types.String     `tfsdk:"group_interval"`
	RepeatInterval                   types.String     `tfsdk:"repeat_interval"`
}

type labelMatcherNotificationPolicyModel struct {
	Matchers             types.List   `tfsdk:"matchers"`
	NotificationPolicyID types.String `tfsdk:"notification_policy_id"`
}

type labelMatcherModel struct {
	Type  types.String `tfsdk:"type"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
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

	if len(model.LabelMatcherNotificationPolicies) > 0 {
		policies := make([]attr.Value, 0, len(model.LabelMatcherNotificationPolicies))
		for _, policy := range model.LabelMatcherNotificationPolicies {
			matchers := make([]attr.Value, 0, len(policy.Matchers))
			for _, matcher := range policy.Matchers {
				matcherObj, diags := types.ObjectValue(
					map[string]attr.Type{
						"type":  types.StringType,
						"name":  types.StringType,
						"value": types.StringType,
					},
					map[string]attr.Value{
						"type":  types.StringValue(matcher.Type.String()),
						"name":  types.StringValue(matcher.Name),
						"value": types.StringValue(matcher.Value),
					},
				)
				if diags.HasError() {
					diagnosticsOut.Append(diags...)
					continue
				}
				matchers = append(matchers, matcherObj)
			}

			matchersList, diags := types.ListValue(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"type":  types.StringType,
						"name":  types.StringType,
						"value": types.StringType,
					},
				},
				matchers,
			)
			if diags.HasError() {
				diagnosticsOut.Append(diags...)
				continue
			}

			policyObj, diags := types.ObjectValue(
				map[string]attr.Type{
					"matchers":               matchersList.Type(context.Background()),
					"notification_policy_id": types.StringType,
				},
				map[string]attr.Value{
					"matchers":               matchersList,
					"notification_policy_id": types.StringValue(policy.NotificationPolicyID.UUID.String()),
				},
			)
			if diags.HasError() {
				diagnosticsOut.Append(diags...)
				continue
			}
			policies = append(policies, policyObj)
		}

		listType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"matchers": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":  types.StringType,
							"name":  types.StringType,
							"value": types.StringType,
						},
					},
				},
				"notification_policy_id": types.StringType,
			},
		}
		m.LabelMatcherNotificationPolicies = types.ListValueMust(listType, policies)
	} else {
		m.LabelMatcherNotificationPolicies = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"matchers": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":  types.StringType,
							"name":  types.StringType,
							"value": types.StringType,
						},
					},
				},
				"notification_policy_id": types.StringType,
			},
		})
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

	if !m.LabelMatcherNotificationPolicies.IsNull() && !m.LabelMatcherNotificationPolicies.IsUnknown() {
		policies := make([]clientmodels.LabelMatcherNotificationPolicy, 0, len(m.LabelMatcherNotificationPolicies.Elements()))
		for _, policyElem := range m.LabelMatcherNotificationPolicies.Elements() {
			policyObj, ok := policyElem.(types.Object)
			if !ok {
				return fmt.Errorf("failed to parse label matcher notification policy: %v, type is %T", policyElem, policyElem)
			}

			var policy labelMatcherNotificationPolicyModel
			diags := policyObj.As(context.Background(), &policy, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return fmt.Errorf("failed to parse label matcher notification policy: %v", diags)
			}

			matchers := make([]clientmodels.LabelMatcher, 0, len(policy.Matchers.Elements()))
			for _, matcherElem := range policy.Matchers.Elements() {
				matcherObj, ok := matcherElem.(types.Object)
				if !ok {
					return fmt.Errorf("failed to parse label matcher: %v, type is %T", matcherElem, matcherElem)
				}

				var matcher labelMatcherModel
				diags = matcherObj.As(context.Background(), &matcher, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					return fmt.Errorf("failed to parse label matcher fields: %v", diags)
				}

				var matchType labels.MatchType
				switch matcher.Type.ValueString() {
				case "=":
					matchType = labels.MatchEqual
				case "!=":
					matchType = labels.MatchNotEqual
				case "=~":
					matchType = labels.MatchRegexp
				case "!~":
					matchType = labels.MatchNotRegexp
				default:
					return fmt.Errorf("invalid match type: %s", matcher.Type.ValueString())
				}

				matchers = append(matchers, clientmodels.LabelMatcher{
					Type:  matchType,
					Name:  matcher.Name.ValueString(),
					Value: matcher.Value.ValueString(),
				})
			}

			uid, err := uuid.Parse(policy.NotificationPolicyID.ValueString())
			if err != nil {
				return fmt.Errorf("failed to parse notification policy UUID: %v", err)
			}

			policies = append(policies, clientmodels.LabelMatcherNotificationPolicy{
				Matchers:             matchers,
				NotificationPolicyID: clientmodels.ID{UUID: uid},
			})
		}
		model.LabelMatcherNotificationPolicies = policies
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
