package oodlehttp

import (
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// mutingSchedulesResourcePath is the collection endpoint for muting
// schedules.
const mutingSchedulesResourcePath = "schedules"

// NewMutingScheduleClient returns a client for muting schedules.
//
// Unlike muting rules, schedules are plain REST: POST creates, PUT
// updates, so the base model client needs no override.
func NewMutingScheduleClient(
	client *OodleApiClient,
) *ModelClient[*clientmodels.MutingSchedule] {
	return NewModelClient[*clientmodels.MutingSchedule](
		client,
		mutingSchedulesResourcePath,
		func() *clientmodels.MutingSchedule {
			return &clientmodels.MutingSchedule{}
		},
	)
}
