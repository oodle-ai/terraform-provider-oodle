package awsintegration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func filterList(t *testing.T, filters ...clientmodels.CloudWatchMetricNameFilter) types.List {
	t.Helper()
	diags := &diag.Diagnostics{}
	list := metricNameFiltersFromClient(&filters, diags)
	assert.False(t, diags.HasError())
	return list
}

// An omitted attribute must not delete the filters that were set in the
// Oodle UI: the server keeps them only when the field is absent.
func TestMetricNameFiltersToClientKeepsStoredFilters(t *testing.T) {
	ctx := context.Background()
	for name, list := range map[string]types.List{
		"null":    types.ListNull(metricNameFilterObjectType),
		"unknown": types.ListUnknown(metricNameFilterObjectType),
	} {
		t.Run(name, func(t *testing.T) {
			filters, err := metricNameFiltersToClient(ctx, list)
			assert.Nil(t, err)
			assert.Nil(t, filters)
		})
	}
}

func TestMetricNameFiltersToClientEmptyListRemovesFilters(t *testing.T) {
	filters, err := metricNameFiltersToClient(context.Background(), filterList(t))
	assert.Nil(t, err)
	assert.NotNil(t, filters)
	assert.Equal(t, 0, len(*filters))
}

func TestMetricNameFiltersFromClientWithoutFilters(t *testing.T) {
	diags := &diag.Diagnostics{}
	list := metricNameFiltersFromClient(nil, diags)
	assert.False(t, diags.HasError())
	assert.False(t, list.IsNull())
	assert.False(t, list.IsUnknown())
	assert.Equal(t, 0, len(list.Elements()))
}

func TestValidateMetricNameFilters(t *testing.T) {
	stringList := func(values ...string) types.List {
		elements := make([]attr.Value, len(values))
		for i, v := range values {
			elements[i] = types.StringValue(v)
		}
		return types.ListValueMust(types.StringType, elements)
	}
	filter := func(namespace string, include, exclude types.List) attr.Value {
		return types.ObjectValueMust(metricNameFilterAttrTypes, map[string]attr.Value{
			"namespace":    types.StringValue(namespace),
			"include_only": include,
			"exclude_only": exclude,
		})
	}
	null := types.ListNull(types.StringType)

	cases := []struct {
		name    string
		filters []attr.Value
		valid   bool
	}{
		{
			name: "include and exclude on different namespaces",
			filters: []attr.Value{
				filter("AWS/EC2", stringList("CPU*"), null),
				filter("AWS/RDS", null, stringList("Aurora*")),
			},
			valid: true,
		},
		{
			name:    "unknown patterns are checked at apply time",
			filters: []attr.Value{filter("AWS/EC2", types.ListUnknown(types.StringType), null)},
			valid:   true,
		},
		{
			name:    "neither include nor exclude",
			filters: []attr.Value{filter("AWS/EC2", null, null)},
		},
		{
			name:    "empty include list",
			filters: []attr.Value{filter("AWS/EC2", stringList(), null)},
		},
		{
			name:    "both include and exclude",
			filters: []attr.Value{filter("AWS/EC2", stringList("A"), stringList("B"))},
		},
		{
			name: "two filters for one namespace",
			filters: []attr.Value{
				filter("AWS/EC2", stringList("A"), null),
				filter("AWS/EC2", null, stringList("B")),
			},
		},
		{
			name:    "Datadog metric name",
			filters: []attr.Value{filter("AWS/EC2", stringList("aws.ec2.network_in"), null)},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := &diag.Diagnostics{}
			list := types.ListValueMust(metricNameFilterObjectType, c.filters)
			validateMetricNameFilters(context.Background(), list, diags)
			assert.Equal(t, c.valid, !diags.HasError())
		})
	}
}
