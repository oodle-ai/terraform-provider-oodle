package validatorutils

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/prometheus/prometheus/promql/parser"
)

type promqlValidator struct {
}

var _ validator.String = (*promqlValidator)(nil)

func NewPromQLValidator() validator.String {
	return &promqlValidator{}
}

func (p promqlValidator) Description(ctx context.Context) string {
	return "Validates that the string is a valid promql query"
}

func (p promqlValidator) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}

func (p promqlValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() {
		return
	}

	_, err := parser.NewParser(
		request.ConfigValue.ValueString(),
		parser.WithFunctions(GetFunctions()),
	).ParseExpr()

	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid promql query",
			fmt.Sprintf(
				"The value %v is not a valid promql query: %v",
				request.ConfigValue.ValueString(),
				err,
			),
		)
	}
}

var functions map[string]*parser.Function
var once sync.Once

var XFunctions = map[string]*parser.Function{
	"xdelta": {
		Name:       "xdelta",
		ArgTypes:   []parser.ValueType{parser.ValueTypeMatrix},
		ReturnType: parser.ValueTypeVector,
	},
	"xincrease": {
		Name:       "xincrease",
		ArgTypes:   []parser.ValueType{parser.ValueTypeMatrix},
		ReturnType: parser.ValueTypeVector,
	},
	"xrate": {
		Name:       "xrate",
		ArgTypes:   []parser.ValueType{parser.ValueTypeMatrix},
		ReturnType: parser.ValueTypeVector,
	},
	// The below correspond to prometheus rate/increase/delta.
	"prate": {
		Name:       "prate",
		ArgTypes:   []parser.ValueType{parser.ValueTypeMatrix},
		ReturnType: parser.ValueTypeVector,
	},
	"pincrease": {
		Name:       "pincrease",
		ArgTypes:   []parser.ValueType{parser.ValueTypeMatrix},
		ReturnType: parser.ValueTypeVector,
	},
	"pdelta": {
		Name:       "pdelta",
		ArgTypes:   []parser.ValueType{parser.ValueTypeMatrix},
		ReturnType: parser.ValueTypeVector,
	},
}

func GetFunctions() map[string]*parser.Function {
	once.Do(func() {
		functions = make(map[string]*parser.Function, len(parser.Functions))
		for k, v := range parser.Functions {
			functions[k] = v
		}

		functions["xdelta"] = XFunctions["xdelta"]
		functions["xincrease"] = XFunctions["xincrease"]
		functions["xrate"] = XFunctions["xrate"]
		functions["prate"] = XFunctions["prate"]
		functions["pincrease"] = XFunctions["pincrease"]
		functions["pdelta"] = XFunctions["pdelta"]
	})

	return functions
}
