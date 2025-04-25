package resourceutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

type ResourceModel[M clientmodels.ClientModel] interface {
	// FromClientModel converts a client model received
	// from oodle APIs to a resource model.
	FromClientModel(
		ctx context.Context,
		model M,
		diagnosticsOut *diag.Diagnostics,
	)

	// ToClientModel converts a resource model to a client model to
	// use in oodle APIs.
	ToClientModel(
		ctx context.Context,
		model M,
	) error

	// SetID sets the ID of the resource model.
	SetID(id types.String)

	// GetID returns the ID of the resource model.
	GetID() types.String
}
