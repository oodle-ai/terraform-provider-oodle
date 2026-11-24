package awsintegration

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// datadogMetricNamePrefix starts every Datadog AWS metric name (for example
// aws.ec2.network_in). No CloudWatch metric name starts with it.
const datadogMetricNamePrefix = "aws."

type metricNameFilterModel struct {
	Namespace   types.String `tfsdk:"namespace"`
	IncludeOnly types.List   `tfsdk:"include_only"`
	ExcludeOnly types.List   `tfsdk:"exclude_only"`
}

var metricNameFilterAttrTypes = map[string]attr.Type{
	"namespace":    types.StringType,
	"include_only": types.ListType{ElemType: types.StringType},
	"exclude_only": types.ListType{ElemType: types.StringType},
}

var metricNameFilterObjectType = types.ObjectType{AttrTypes: metricNameFilterAttrTypes}

// metricNameFiltersFromClient always returns a known list, empty when the
// server returns no filters, so that `terraform import` and a config with
// `metric_name_filters = []` show no diff.
func metricNameFiltersFromClient(
	filters *[]clientmodels.CloudWatchMetricNameFilter,
	diags *diag.Diagnostics,
) types.List {
	var values []attr.Value
	if filters != nil {
		for _, filter := range *filters {
			obj, d := types.ObjectValue(metricNameFilterAttrTypes, map[string]attr.Value{
				"namespace":    types.StringValue(filter.Namespace),
				"include_only": stringListOrNull(filter.IncludeOnly, diags),
				"exclude_only": stringListOrNull(filter.ExcludeOnly, diags),
			})
			diags.Append(d...)
			values = append(values, obj)
		}
	}
	list, d := types.ListValue(metricNameFilterObjectType, values)
	diags.Append(d...)
	return list
}

// stringListOrNull returns null for an empty list, because a filter sets only
// one of include_only and exclude_only and the other one is null in config.
func stringListOrNull(values []string, diags *diag.Diagnostics) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	list, d := types.ListValueFrom(context.Background(), types.StringType, values)
	diags.Append(d...)
	return list
}

// metricNameFiltersToClient returns nil when the value is null or unknown, so
// that the server keeps the stored filters. A known empty list removes them.
func metricNameFiltersToClient(
	ctx context.Context,
	list types.List,
) (*[]clientmodels.CloudWatchMetricNameFilter, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var models []metricNameFilterModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("failed to read metric_name_filters: %v", d.Errors())
	}

	filters := make([]clientmodels.CloudWatchMetricNameFilter, 0, len(models))
	for i, model := range models {
		filter := clientmodels.CloudWatchMetricNameFilter{Namespace: model.Namespace.ValueString()}
		var err error
		if filter.IncludeOnly, err = knownStrings(ctx, model.IncludeOnly); err != nil {
			return nil, fmt.Errorf("failed to read metric_name_filters[%d].include_only: %w", i, err)
		}
		if filter.ExcludeOnly, err = knownStrings(ctx, model.ExcludeOnly); err != nil {
			return nil, fmt.Errorf("failed to read metric_name_filters[%d].exclude_only: %w", i, err)
		}
		filters = append(filters, filter)
	}
	return &filters, nil
}

func knownStrings(ctx context.Context, list types.List) ([]string, error) {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil, nil
	}
	var values []string
	if d := list.ElementsAs(ctx, &values, false); d.HasError() {
		return nil, fmt.Errorf("%v", d.Errors())
	}
	return values, nil
}

// validateMetricNameFilters repeats the server's checks at plan time. Values
// that are not known yet are skipped, because `for_each` and references are
// not resolved at validate time.
func validateMetricNameFilters(ctx context.Context, list types.List, diags *diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return
	}

	var models []metricNameFilterModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		diags.Append(d...)
		return
	}

	seen := make(map[string]struct{}, len(models))
	for i, model := range models {
		at := path.Root("metric_name_filters").AtListIndex(i)

		if model.Namespace.IsNull() {
			diags.AddAttributeError(
				at.AtName("namespace"),
				"Missing namespace in metric name filter",
				"Set namespace, for example \"AWS/EC2\".",
			)
		}
		if !model.Namespace.IsUnknown() && !model.Namespace.IsNull() {
			namespace := model.Namespace.ValueString()
			if _, ok := seen[namespace]; ok {
				diags.AddAttributeError(
					at.AtName("namespace"),
					"Duplicate metric name filter",
					fmt.Sprintf(
						"Namespace %q has more than one filter. Put all its patterns in one filter.",
						namespace,
					),
				)
			}
			seen[namespace] = struct{}{}
		}

		if model.IncludeOnly.IsUnknown() || model.ExcludeOnly.IsUnknown() {
			continue
		}
		hasInclude := len(model.IncludeOnly.Elements()) > 0
		hasExclude := len(model.ExcludeOnly.Elements()) > 0
		if hasInclude == hasExclude {
			diags.AddAttributeError(
				at,
				"Invalid metric name filter",
				"Set exactly one of include_only and exclude_only, with at least one pattern.",
			)
		}

		for _, name := range []string{"include_only", "exclude_only"} {
			patterns := model.IncludeOnly
			if name == "exclude_only" {
				patterns = model.ExcludeOnly
			}
			for j, element := range patterns.Elements() {
				pattern, ok := element.(types.String)
				if !ok || pattern.IsUnknown() || pattern.IsNull() {
					continue
				}
				// A Datadog name never matches a CloudWatch name. An
				// include_only list copied from datadog_integration_aws_account
				// would stop all data for the namespace.
				if strings.HasPrefix(strings.ToLower(pattern.ValueString()), datadogMetricNamePrefix) {
					diags.AddAttributeError(
						at.AtName(name).AtListIndex(j),
						"Datadog metric name in metric name filter",
						fmt.Sprintf(
							"Pattern %q is a Datadog metric name. Use the CloudWatch metric name, "+
								"for example NetworkIn in place of aws.ec2.network_in.",
							pattern.ValueString(),
						),
					)
				}
			}
		}
	}
}
