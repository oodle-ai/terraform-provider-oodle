package validatorutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type notifiersValidator struct{}

var _ validator.Object = (*notifiersValidator)(nil)

func NewNotifiersValidator() validator.Object {
	return &notifiersValidator{}
}

func (n notifiersValidator) Description(ctx context.Context) string {
	return "Validates that 'any' notifiers are mutually exclusive with warn, critical, or no_data notifiers"
}

func (n notifiersValidator) MarkdownDescription(ctx context.Context) string {
	return n.Description(ctx)
}

func (n notifiersValidator) ValidateObject(
	ctx context.Context,
	request validator.ObjectRequest,
	response *validator.ObjectResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	attrs := request.ConfigValue.Attributes()
	if attrs == nil {
		return
	}

	// Check if 'any' field is set and not empty
	anyAttr, hasAny := attrs["any"]
	var anyHasElements bool
	if hasAny && !anyAttr.IsNull() && !anyAttr.IsUnknown() {
		if anyList, ok := anyAttr.(types.List); ok && len(anyList.Elements()) > 0 {
			anyHasElements = true
		}
	}

	if anyHasElements {
		// Check if any other fields are set
		otherFields := []string{"warn", "critical", "no_data"}
		for _, field := range otherFields {
			if attr, exists := attrs[field]; exists && !attr.IsNull() && !attr.IsUnknown() {
				if list, ok := attr.(types.List); ok && len(list.Elements()) > 0 {
					response.Diagnostics.AddAttributeError(
						request.Path,
						"Invalid notifiers configuration",
						fmt.Sprintf("'any' notifiers must be empty if %s notifiers are set", field),
					)
					return
				}
			}
		}
	}
}
