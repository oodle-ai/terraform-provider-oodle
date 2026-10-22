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

// NameKeeper is a model where two attribute names can hold one API
// value (for example warn and warning). The API answer does not say
// which name the config used, so before a create or an update reads
// the answer, the model takes the names from the plan. Without it,
// the state uses the other name and Terraform reports an error.
type NameKeeper[R any] interface {
	TakeNamesFrom(plan R)
}
