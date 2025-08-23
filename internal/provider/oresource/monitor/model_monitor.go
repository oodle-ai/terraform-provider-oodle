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
	Notifications                    types.List       `tfsdk:"notifications"`
	GroupWait                        types.String     `tfsdk:"group_wait"`
	GroupInterval                    types.String     `tfsdk:"group_interval"`
	RepeatInterval                   types.String     `tfsdk:"repeat_interval"`
}

type labelMatcherNotificationPolicyModel struct {
	Matchers             types.List   `tfsdk:"matchers"`
	NotificationPolicyID types.String `tfsdk:"notification_policy_id"`
}

type labelMatcherNotificationsModel struct {
	Matchers             types.List            `tfsdk:"matchers"`
	NotificationPolicyID types.String          `tfsdk:"notification_policy_id"`
	Notifiers            *notifiersByCondition `tfsdk:"notifiers"`
}

type notifiersByCondition struct {
	Any      types.List `tfsdk:"any"`
	Warn     types.List `tfsdk:"warn"`
	Critical types.List `tfsdk:"critical"`
	NoData   types.List `tfsdk:"no_data"`
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
	ctx context.Context,
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

		// Only set the fields that are actually configured
		if model.Grouping.ByMonitor {
			m.Grouping.ByMonitor = types.BoolValue(true)
			m.Grouping.ByLabels = types.ListNull(types.StringType)
			m.Grouping.Disabled = types.BoolNull()
		} else if len(model.Grouping.ByLabels) > 0 {
			m.Grouping.ByMonitor = types.BoolNull()
			m.Grouping.ByLabels = validatorutils.ToAttrList(model.Grouping.ByLabels, diagnosticsOut)
			m.Grouping.Disabled = types.BoolNull()
		} else if model.Grouping.Disabled {
			m.Grouping.ByMonitor = types.BoolNull()
			m.Grouping.ByLabels = types.ListNull(types.StringType)
			m.Grouping.Disabled = types.BoolValue(true)
		}
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
					"matchers":               matchersList.Type(ctx),
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

	if len(model.Notifications) > 0 {
		notifications := make([]attr.Value, 0, len(model.Notifications))
		for _, notification := range model.Notifications {
			var matchersList attr.Value
			if len(notification.Matchers) > 0 {
				matchers := make([]attr.Value, 0, len(notification.Matchers))
				for _, matcher := range notification.Matchers {
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

				matchersListValue, diags := types.ListValue(
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
				matchersList = matchersListValue
			} else {
				// When no matchers, use null instead of empty list
				matchersList = types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"type":  types.StringType,
						"name":  types.StringType,
						"value": types.StringType,
					},
				})
			}

			// Convert NotifiersByCondition to terraform structure
			var notifiersValue attr.Value
			if len(notification.Notifiers.Any) > 0 || len(notification.Notifiers.Warn) > 0 ||
				len(notification.Notifiers.Critical) > 0 || len(notification.Notifiers.NoData) > 0 {
				anyList := types.ListNull(types.StringType)
				if len(notification.Notifiers.Any) > 0 {
					anyList = validatorutils.IDsToAttrList(notification.Notifiers.Any, diagnosticsOut)
				}

				warnList := types.ListNull(types.StringType)
				if len(notification.Notifiers.Warn) > 0 {
					warnList = validatorutils.IDsToAttrList(notification.Notifiers.Warn, diagnosticsOut)
				}

				criticalList := types.ListNull(types.StringType)
				if len(notification.Notifiers.Critical) > 0 {
					criticalList = validatorutils.IDsToAttrList(notification.Notifiers.Critical, diagnosticsOut)
				}

				noDataList := types.ListNull(types.StringType)
				if len(notification.Notifiers.NoData) > 0 {
					noDataList = validatorutils.IDsToAttrList(notification.Notifiers.NoData, diagnosticsOut)
				}

				notifiersObj, diags := types.ObjectValue(
					map[string]attr.Type{
						"any":      types.ListType{ElemType: types.StringType},
						"warn":     types.ListType{ElemType: types.StringType},
						"critical": types.ListType{ElemType: types.StringType},
						"no_data":  types.ListType{ElemType: types.StringType},
					},
					map[string]attr.Value{
						"any":      anyList,
						"warn":     warnList,
						"critical": criticalList,
						"no_data":  noDataList,
					},
				)
				if diags.HasError() {
					diagnosticsOut.Append(diags...)
					continue
				}
				notifiersValue = notifiersObj
			} else {
				notifiersValue = types.ObjectNull(map[string]attr.Type{
					"any":      types.ListType{ElemType: types.StringType},
					"warn":     types.ListType{ElemType: types.StringType},
					"critical": types.ListType{ElemType: types.StringType},
					"no_data":  types.ListType{ElemType: types.StringType},
				})
			}

			var notificationPolicyIDValue attr.Value
			if notification.NotificationPolicyID.UUID != uuid.Nil {
				notificationPolicyIDValue = types.StringValue(notification.NotificationPolicyID.UUID.String())
			} else {
				notificationPolicyIDValue = types.StringNull()
			}

			notificationObj, diags := types.ObjectValue(
				map[string]attr.Type{
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
					"notifiers": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"any":      types.ListType{ElemType: types.StringType},
							"warn":     types.ListType{ElemType: types.StringType},
							"critical": types.ListType{ElemType: types.StringType},
							"no_data":  types.ListType{ElemType: types.StringType},
						},
					},
				},
				map[string]attr.Value{
					"matchers":               matchersList,
					"notification_policy_id": notificationPolicyIDValue,
					"notifiers":              notifiersValue,
				},
			)
			if diags.HasError() {
				diagnosticsOut.Append(diags...)
				continue
			}
			notifications = append(notifications, notificationObj)
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
				"notifiers": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"any":      types.ListType{ElemType: types.StringType},
						"warn":     types.ListType{ElemType: types.StringType},
						"critical": types.ListType{ElemType: types.StringType},
						"no_data":  types.ListType{ElemType: types.StringType},
					},
				},
			},
		}
		m.Notifications = types.ListValueMust(listType, notifications)
	} else {
		m.Notifications = types.ListNull(types.ObjectType{
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
				"notifiers": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"any":      types.ListType{ElemType: types.StringType},
						"warn":     types.ListType{ElemType: types.StringType},
						"critical": types.ListType{ElemType: types.StringType},
						"no_data":  types.ListType{ElemType: types.StringType},
					},
				},
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
	ctx context.Context,
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
		// Only set values that are not null
		if !m.Grouping.ByMonitor.IsNull() {
			model.Grouping.ByMonitor = m.Grouping.ByMonitor.ValueBool()
		}
		if !m.Grouping.Disabled.IsNull() {
			model.Grouping.Disabled = m.Grouping.Disabled.ValueBool()
		}

		if !m.Grouping.ByLabels.IsNull() && len(m.Grouping.ByLabels.Elements()) > 0 {
			model.Grouping.ByLabels = make([]string, 0, len(m.Grouping.ByLabels.Elements()))
			for _, v := range m.Grouping.ByLabels.Elements() {
				strVal, ok := v.(validatorutils.StringValue)
				if !ok {
					return fmt.Errorf("failed to parse grouping labels as string: %v, type is %T", v, v)
				}

				model.Grouping.ByLabels = append(model.Grouping.ByLabels, strVal.ValueString())
			}
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
			diags := policyObj.As(ctx, &policy, basetypes.ObjectAsOptions{})
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
				diags = matcherObj.As(ctx, &matcher, basetypes.ObjectAsOptions{})
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

	if !m.Notifications.IsNull() && !m.Notifications.IsUnknown() {
		notifications := make([]clientmodels.LabelMatcherNotifications, 0, len(m.Notifications.Elements()))
		for _, notificationElem := range m.Notifications.Elements() {
			notificationObj, ok := notificationElem.(types.Object)
			if !ok {
				return fmt.Errorf("failed to parse notification: %v, type is %T", notificationElem, notificationElem)
			}

			var notification labelMatcherNotificationsModel
			diags := notificationObj.As(ctx, &notification, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return fmt.Errorf("failed to parse notification: %v", diags)
			}

			matchers := make([]clientmodels.LabelMatcher, 0)
			if !notification.Matchers.IsNull() && !notification.Matchers.IsUnknown() {
				matchers = make([]clientmodels.LabelMatcher, 0, len(notification.Matchers.Elements()))
				for _, matcherElem := range notification.Matchers.Elements() {
					matcherObj, ok := matcherElem.(types.Object)
					if !ok {
						return fmt.Errorf("failed to parse notification matcher: %v, type is %T", matcherElem, matcherElem)
					}

					var matcher labelMatcherModel
					diags = matcherObj.As(ctx, &matcher, basetypes.ObjectAsOptions{})
					if diags.HasError() {
						return fmt.Errorf("failed to parse notification matcher fields: %v", diags)
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
			}

			var notificationPolicyID clientmodels.ID
			if !notification.NotificationPolicyID.IsNull() && !notification.NotificationPolicyID.IsUnknown() {
				uid, err := uuid.Parse(notification.NotificationPolicyID.ValueString())
				if err != nil {
					return fmt.Errorf("failed to parse notification policy UUID: %v", err)
				}
				notificationPolicyID = clientmodels.ID{UUID: uid}
			}

			var notifiers clientmodels.NotifiersByCondition
			if notification.Notifiers != nil {
				if !notification.Notifiers.Any.IsNull() && len(notification.Notifiers.Any.Elements()) > 0 {
					anyIDs, err := validatorutils.AttrListToIDs(notification.Notifiers.Any)
					if err != nil {
						return fmt.Errorf("failed to parse any notifier IDs: %v", err)
					}
					notifiers.Any = anyIDs
				}

				if !notification.Notifiers.Warn.IsNull() && len(notification.Notifiers.Warn.Elements()) > 0 {
					warnIDs, err := validatorutils.AttrListToIDs(notification.Notifiers.Warn)
					if err != nil {
						return fmt.Errorf("failed to parse warn notifier IDs: %v", err)
					}
					notifiers.Warn = warnIDs
				}

				if !notification.Notifiers.Critical.IsNull() && len(notification.Notifiers.Critical.Elements()) > 0 {
					criticalIDs, err := validatorutils.AttrListToIDs(notification.Notifiers.Critical)
					if err != nil {
						return fmt.Errorf("failed to parse critical notifier IDs: %v", err)
					}
					notifiers.Critical = criticalIDs
				}

				if !notification.Notifiers.NoData.IsNull() && len(notification.Notifiers.NoData.Elements()) > 0 {
					noDataIDs, err := validatorutils.AttrListToIDs(notification.Notifiers.NoData)
					if err != nil {
						return fmt.Errorf("failed to parse no_data notifier IDs: %v", err)
					}
					notifiers.NoData = noDataIDs
				}
			}

			notifications = append(notifications, clientmodels.LabelMatcherNotifications{
				Matchers:             matchers,
				NotificationPolicyID: notificationPolicyID,
				Notifiers:            notifiers,
			})
		}
		model.Notifications = notifications
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
