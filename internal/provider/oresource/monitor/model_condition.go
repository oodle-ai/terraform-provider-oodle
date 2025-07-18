package monitor

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/validatorutils"
)

type conditionsModel struct {
	Warning  *conditionModel `tfsdk:"warning"`
	Critical *conditionModel `tfsdk:"critical"`
}

type conditionModel struct {
	// Operation - The operation to perform for the condition. Possible values are: ">", "<", ">=", "<=", "==", "!=".
	Operation     types.String  `tfsdk:"operation"`
	Value         types.Float64 `tfsdk:"value"`
	For           types.String  `tfsdk:"for"`
	KeepFiringFor types.String  `tfsdk:"keep_firing_for"`
	AlertOnNoData types.Bool    `tfsdk:"alert_on_no_data"`
}

func newConditionFromModel(model *clientmodels.Condition) *conditionModel {
	c := conditionModel{}
	c.Operation = types.StringValue(model.Op.String())
	c.Value = types.Float64Value(model.Value)
	c.For = types.StringValue(validatorutils.ShortDur(model.For))
	c.AlertOnNoData = types.BoolValue(model.AlertOnNoData)

	if model.KeepFiringFor > 0 {
		c.KeepFiringFor = types.StringValue(validatorutils.ShortDur(model.KeepFiringFor))
	}
	return &c
}

func (c *conditionModel) toModel() (*clientmodels.Condition, error) {
	op, err := clientmodels.ConditionOpFromString(c.Operation.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ConditionOp: %v", err)
	}

	var forVal time.Duration
	if !c.For.IsNull() && !c.For.IsUnknown() && len(c.For.ValueString()) > 0 {
		forVal, err = time.ParseDuration(c.For.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to parse warning forVal: %v", err)
		}
	}

	var keepFiringForVal time.Duration
	if !c.KeepFiringFor.IsNull() && !c.KeepFiringFor.IsUnknown() && len(c.KeepFiringFor.ValueString()) > 0 {
		keepFiringForVal, err = time.ParseDuration(c.KeepFiringFor.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to parse warning keepFiringFor: %v", err)
		}
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
