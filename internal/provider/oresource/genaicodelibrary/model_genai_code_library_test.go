package genaicodelibrary

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/validatorutils"
)

func ptr(s string) *string { return &s }

func TestCodeLibraryRoundTrip(t *testing.T) {
	ctx := context.Background()
	m := &genaiCodeLibraryResourceModel{
		Name:        types.StringValue("helpers"),
		SourceCode:  types.StringValue("def f():\n    return 1\n"),
		Description: types.StringNull(),
	}

	written := &clientmodels.GenAICodeLibrary{}
	assert.Nil(t, m.ToClientModel(ctx, written))
	assert.Equal(t, "helpers", written.Name)
	assert.Equal(t, "def f():\n    return 1\n", *written.SourceCode)
	// An unset description is sent as "", which clears a stored one.
	assert.Equal(t, "", *written.Description)

	out := &diag.Diagnostics{}
	m.FromClientModel(ctx, &clientmodels.GenAICodeLibrary{
		ID:          "lib-1",
		Name:        "helpers",
		Description: ptr(""),
		SourceCode:  ptr("def f():\n    return 1\n"),
		Version:     3,
	}, out)
	assert.False(t, out.HasError())
	assert.Equal(t, "lib-1", m.ID.ValueString())
	assert.Equal(t, int64(3), m.Version.ValueInt64())
	// The API's "" does not turn an unset description into a diff.
	assert.True(t, m.Description.IsNull())
}

func TestCodeLibraryDescription(t *testing.T) {
	m := &genaiCodeLibraryResourceModel{Description: types.StringValue("")}
	m.FromClientModel(context.Background(), &clientmodels.GenAICodeLibrary{
		ID: "lib-1", Description: ptr(""),
	}, &diag.Diagnostics{})
	assert.False(t, m.Description.IsNull())
	assert.Equal(t, "", m.Description.ValueString())

	m = &genaiCodeLibraryResourceModel{Description: types.StringNull()}
	m.FromClientModel(context.Background(), &clientmodels.GenAICodeLibrary{
		ID: "lib-1", Description: ptr("text helpers"),
	}, &diag.Diagnostics{})
	assert.Equal(t, "text helpers", m.Description.ValueString())
}

func TestLibraryNameValidator(t *testing.T) {
	v := libraryNameValidator{}
	for _, name := range []string{"helpers", "text_utils2", "a"} {
		assert.True(t, validatorutils.IsValidForValidator(types.StringValue(name), v))
	}
	for _, name := range []string{
		"Helpers", "2x", "_x", "text-utils", "shared", "oodle_eval",
		"import", "a2345678901234567890123456789012345678901234567890123456789012345",
	} {
		assert.False(t, validatorutils.IsValidForValidator(types.StringValue(name), v))
	}
	assert.True(t, validatorutils.IsValidForValidator(types.StringUnknown(), v))
}
