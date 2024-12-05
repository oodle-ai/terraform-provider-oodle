package models

import (
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/common/model"
)

// ConditionOp is the operator for a condition.
type ConditionOp int

const (
	ConditionOpEqual ConditionOp = iota
	ConditionOpNotEqual
	ConditionOpGreaterThan
	ConditionOpGreaterThanOrEqual
	ConditionOpLessThan
	ConditionOpLessThanOrEqual
)

// String returns the string representation of the condition operator.
func (o ConditionOp) String() string {
	switch o {
	case ConditionOpEqual:
		return "=="
	case ConditionOpNotEqual:
		return "!="
	case ConditionOpGreaterThan:
		return ">"
	case ConditionOpGreaterThanOrEqual:
		return ">="
	case ConditionOpLessThan:
		return "<"
	case ConditionOpLessThanOrEqual:
		return "<="
	default:
		panic(fmt.Sprintf("unknown condition operator: %d", o))
	}
}

func ConditionOpFromString(op string) (ConditionOp, error) {
	switch op {
	case "==":
		return ConditionOpEqual, nil
	case "!=":
		return ConditionOpNotEqual, nil
	case ">":
		return ConditionOpGreaterThan, nil
	case ">=":
		return ConditionOpGreaterThanOrEqual, nil
	case "<":
		return ConditionOpLessThan, nil
	case "<=":
		return ConditionOpLessThanOrEqual, nil
	default:
		return ConditionOpEqual, fmt.Errorf("unknown condition operator: %s", op)
	}
}

// Condition is a model for a condition to be evaluated in monitors.
type Condition struct {
	Op    ConditionOp `json:"op" yaml:"op"`
	Value float64     `json:"value" yaml:"value"`
	// For is the duration for which the condition should be true
	// before the alert is triggered.
	For time.Duration `json:"for,omitempty" yaml:"for,omitempty"`
	// KeepFiringFor is the duration for which the alert should keep firing
	// after the condition is no longer true.
	KeepFiringFor time.Duration `json:"keep_firing_for,omitempty" yaml:"keep_firing_for,omitempty"`
}

// ToCompString returns the string representation of the condition.
func (c Condition) ToCompString() string {
	return fmt.Sprintf(" %s %g", c.Op, c.Value)
}

// ConditionBySeverity represents a condition for each severity level.
type ConditionBySeverity struct {
	Warn     *Condition `json:"warn,omitempty" yaml:"warn,omitempty"`
	Critical *Condition `json:"critical,omitempty" yaml:"critical,omitempty"`
}

// MarshalJSON customizes the JSON marshaling for Condition.
func (c Condition) MarshalJSON() ([]byte, error) {
	type Alias Condition
	return jsoniter.Marshal(&struct {
		*Alias
		For           model.Duration `json:"for,omitempty"`
		KeepFiringFor model.Duration `json:"keep_firing_for,omitempty"`
	}{
		Alias:         (*Alias)(&c),
		For:           model.Duration(c.For),
		KeepFiringFor: model.Duration(c.KeepFiringFor),
	})
}

// UnmarshalJSON customizes the JSON unmarshaling for Condition.
func (c *Condition) UnmarshalJSON(data []byte) error {
	type Alias Condition
	aux := &struct {
		*Alias
		For           model.Duration `json:"for,omitempty"`
		KeepFiringFor model.Duration `json:"keep_firing_for,omitempty"`
	}{
		Alias: (*Alias)(c),
	}
	if err := jsoniter.Unmarshal(data, aux); err != nil {
		return err
	}

	c.For = time.Duration(aux.For)
	c.KeepFiringFor = time.Duration(aux.KeepFiringFor)
	return nil
}
