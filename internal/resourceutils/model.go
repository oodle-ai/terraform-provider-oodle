package resourceutils

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

type ResourceModel[M clientmodels.ClientModel] interface {
	FromModel(
		model M,
		diagnosticsOut *diag.Diagnostics,
	)

	ToModel(model M) error

	SetID(id types.String)

	GetID() types.String
}
