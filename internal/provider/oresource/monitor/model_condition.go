package monitor

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/validatorutils"
)

type conditionsModel struct {
	Warning  *conditionModel       `tfsdk:"warning"`
	Critical *conditionModel       `tfsdk:"critical"`
	NoData   *noDataConditionModel `tfsdk:"no_data"`
}

type conditionModel struct {
	// Operation - The operation to perform for the condition. Possible values are: ">", "<", ">=", "<=", "==", "!=".
	Operation     types.String                 `tfsdk:"operation"`
	Value         types.Float64                `tfsdk:"value"`
	For           validatorutils.DurationValue `tfsdk:"for"`
	KeepFiringFor validatorutils.DurationValue `tfsdk:"keep_firing_for"`
	// Deprecated: Use conditions.no_data instead
	AlertOnNoData types.Bool `tfsdk:"alert_on_no_data"`
}

type noDataConditionModel struct {
	// NoData conditions don't need Operation, Value, or AlertOnNoData - they default to Equal, 1, and true respectively
	For           validatorutils.DurationValue `tfsdk:"for"`
	KeepFiringFor validatorutils.DurationValue `tfsdk:"keep_firing_for"`
}

// durationValueFromModel converts a time.Duration to a DurationValue.
// For zero durations, this returns null since zero is semantically equivalent to
// "not set" for optional duration fields like keep_firing_for and interval.
// A zero duration configured as "0s" by the user will be round-tripped through the
// API as zero and converted back to null, which matches the user's intent (zero
// duration is functionally equivalent to omitting the field).
func durationValueFromModel(d time.Duration) validatorutils.DurationValue {
	if d > 0 {
		return validatorutils.NewDurationValue(validatorutils.ShortDur(d))
	}
	return validatorutils.NewDurationNull()
}

// durationValueFromDurationPtr converts a *time.Duration to a DurationValue.
// nil pointers map to null; non-nil values (including zero) are preserved.
func durationValueFromDurationPtr(d *time.Duration) validatorutils.DurationValue {
	if d != nil {
		return validatorutils.NewDurationValue(validatorutils.ShortDur(*d))
	}
	return validatorutils.NewDurationNull()
}

// parseDurationValue parses a DurationValue into a time.Duration.
// Returns zero duration if the value is null or unknown.
func parseDurationValue(v validatorutils.DurationValue, fieldName string) (time.Duration, error) {
	if v.IsNull() || v.IsUnknown() {
		return 0, nil
	}
	d, err := time.ParseDuration(v.ValueString())
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %v", fieldName, err)
	}
	return d, nil
}

func newConditionFromModel(model *clientmodels.Condition) *conditionModel {
	c := conditionModel{}
	c.Operation = types.StringValue(model.Op.String())
	c.Value = types.Float64Value(model.Value)
	c.AlertOnNoData = types.BoolValue(model.AlertOnNoData)

	c.For = validatorutils.NewDurationValue(validatorutils.ShortDur(model.For))
	c.KeepFiringFor = durationValueFromModel(model.KeepFiringFor)
	return &c
}

func newNoDataConditionFromModel(model *clientmodels.Condition) *noDataConditionModel {
	c := noDataConditionModel{}

	c.For = validatorutils.NewDurationValue(validatorutils.ShortDur(model.For))
	c.KeepFiringFor = durationValueFromModel(model.KeepFiringFor)
	return &c
}

func (c *conditionModel) toModel() (*clientmodels.Condition, error) {
	op, err := clientmodels.ConditionOpFromString(c.Operation.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ConditionOp: %v", err)
	}

	forVal, err := parseDurationValue(c.For, "warning forVal")
	if err != nil {
		return nil, err
	}

	keepFiringForVal, err := parseDurationValue(c.KeepFiringFor, "warning keepFiringFor")
	if err != nil {
		return nil, err
	}

	var alertOnNoData bool
	if !c.AlertOnNoData.IsNull() && !c.AlertOnNoData.IsUnknown() {
		alertOnNoData = c.AlertOnNoData.ValueBool()
	}

	return &clientmodels.Condition{
		Op:            op,
		Value:         c.Value.ValueFloat64(),
		For:           forVal,
		KeepFiringFor: keepFiringForVal,
		AlertOnNoData: alertOnNoData,
	}, nil
}

func (c *noDataConditionModel) toModel() (*clientmodels.Condition, error) {
	forVal, err := parseDurationValue(c.For, "no_data forVal")
	if err != nil {
		return nil, err
	}

	keepFiringForVal, err := parseDurationValue(c.KeepFiringFor, "no_data keepFiringFor")
	if err != nil {
		return nil, err
	}

	// Default NoData condition to Equal, 1
	return &clientmodels.Condition{
		Op:            clientmodels.ConditionOpEqual,
		Value:         1,
		For:           forVal,
		KeepFiringFor: keepFiringForVal,
	}, nil
}
