package mutingrule

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	amlabels "github.com/prometheus/alertmanager/pkg/labels"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

// zeroTimestamp is what the API returns for an unset time. The
// muting rule model serializes time.Time unconditionally, so an
// absent starts_at comes back as the zero value rather than being
// omitted, and must be read back as null.
const zeroTimestamp = "0001-01-01T00:00:00Z"

var matcherAttrTypes = map[string]attr.Type{
	"type":  types.StringType,
	"name":  types.StringType,
	"value": types.StringType,
}

type labelMatcherModel struct {
	Type  types.String `tfsdk:"type"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type mutingRuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Comment     types.String `tfsdk:"comment"`
	Matchers    types.List   `tfsdk:"matchers"`
	StartsAt    types.String `tfsdk:"starts_at"`
	EndsAt      types.String `tfsdk:"ends_at"`
	ScheduleIDs types.List   `tfsdk:"schedule_ids"`
	CreatedBy   types.String `tfsdk:"created_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (m *mutingRuleResourceModel) GetID() types.String {
	return m.ID
}

func (m *mutingRuleResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *mutingRuleResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.MutingRule,
	diagnosticsOut *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = optionalString(model.Name)
	m.Comment = optionalString(model.Comment)
	m.StartsAt = optionalTimestamp(model.StartsAt)
	m.EndsAt = optionalTimestamp(model.EndsAt)
	m.CreatedBy = optionalString(model.CreatedBy)
	m.CreatedAt = optionalTimestamp(model.CreatedAt)
	m.UpdatedAt = optionalTimestamp(model.UpdatedAt)

	m.ScheduleIDs = resourceutils.SliceToStringList(
		ctx, model.ScheduleIDs, m.ScheduleIDs, diagnosticsOut,
	)
	m.Matchers = matchersFromClient(model.Matchers, diagnosticsOut)
}

func (m *mutingRuleResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.MutingRule,
) error {
	model.ID = m.ID.ValueString()
	model.Name = m.Name.ValueString()
	model.Comment = m.Comment.ValueString()
	model.StartsAt = m.StartsAt.ValueString()
	model.EndsAt = m.EndsAt.ValueString()

	scheduleIDs, err := resourceutils.StringListToSlice(ctx, m.ScheduleIDs)
	if err != nil {
		return err
	}
	model.ScheduleIDs = scheduleIDs

	matchers, err := matchersToClient(ctx, m.Matchers)
	if err != nil {
		return err
	}
	model.Matchers = matchers

	return nil
}

// matchersToClient converts the configured matchers into the client
// model, translating the operator into its alertmanager match type.
func matchersToClient(
	ctx context.Context,
	list types.List,
) ([]clientmodels.LabelMatcher, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	matchers := make([]clientmodels.LabelMatcher, 0, len(list.Elements()))
	for _, element := range list.Elements() {
		object, ok := element.(types.Object)
		if !ok {
			return nil, fmt.Errorf(
				"failed to parse label matcher: %v, type is %T",
				element,
				element,
			)
		}

		var matcher labelMatcherModel
		diags := object.As(ctx, &matcher, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf(
				"failed to parse label matcher fields: %v", diags,
			)
		}

		matchType, err := parseMatchType(matcher.Type.ValueString())
		if err != nil {
			return nil, err
		}

		matchers = append(matchers, clientmodels.LabelMatcher{
			Type:  matchType,
			Name:  matcher.Name.ValueString(),
			Value: matcher.Value.ValueString(),
		})
	}

	return matchers, nil
}

// matchersFromClient converts the API's matchers back into state.
func matchersFromClient(
	matchers []clientmodels.LabelMatcher,
	diagnosticsOut *diag.Diagnostics,
) types.List {
	elements := make([]attr.Value, 0, len(matchers))
	for _, matcher := range matchers {
		object, diags := types.ObjectValue(
			matcherAttrTypes,
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

		elements = append(elements, object)
	}

	list, diags := types.ListValue(
		types.ObjectType{AttrTypes: matcherAttrTypes},
		elements,
	)
	if diags.HasError() {
		diagnosticsOut.Append(diags...)
		return types.ListNull(types.ObjectType{AttrTypes: matcherAttrTypes})
	}

	return list
}

func parseMatchType(value string) (amlabels.MatchType, error) {
	switch value {
	case "=":
		return amlabels.MatchEqual, nil
	case "!=":
		return amlabels.MatchNotEqual, nil
	case "=~":
		return amlabels.MatchRegexp, nil
	case "!~":
		return amlabels.MatchNotRegexp, nil
	default:
		return 0, fmt.Errorf("invalid match type: %s", value)
	}
}

// optionalString keeps an unset optional attribute null rather than
// turning it into "" when the API omits the field.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

// optionalTimestamp additionally treats the zero time as unset.
func optionalTimestamp(value string) types.String {
	if value == zeroTimestamp {
		return types.StringNull()
	}

	return optionalString(value)
}
