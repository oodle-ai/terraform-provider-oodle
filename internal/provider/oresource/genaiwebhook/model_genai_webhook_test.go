package genaiwebhook

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestWebhookRoundTrip(t *testing.T) {
	ctx := context.Background()

	// What the API answers: header names, never values.
	answered := &clientmodels.GenAIWebhook{
		ID:              "webhook-uuid",
		Name:            "support-agent",
		Description:     "LangGraph support triage agent",
		URL:             "https://agent.example.com/run",
		HeaderNames:     []string{"Authorization"},
		TimeoutSeconds:  120,
		RequestTemplate: `{"query": {{input.question}}}`,
		OutputPath:      "answer",
	}

	headers, diags := types.MapValueFrom(
		ctx, types.StringType,
		map[string]string{"Authorization": "Bearer secret"},
	)
	assert.False(t, diags.HasError())
	resourceModel := &genaiWebhookResourceModel{Headers: headers}

	out := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, answered, out)
	assert.False(t, out.HasError())

	assert.Equal(t, "webhook-uuid", resourceModel.ID.ValueString())
	assert.Equal(t, "support-agent", resourceModel.Name.ValueString())
	assert.Equal(t, int64(120), resourceModel.TimeoutSeconds.ValueInt64())
	assert.Equal(t, "answer", resourceModel.OutputPath.ValueString())
	assert.Equal(
		t, `{"query": {{input.question}}}`,
		resourceModel.RequestTemplate.ValueString(),
	)
	// The configured headers survive a read that cannot return them.
	assert.Equal(t, 1, len(resourceModel.Headers.Elements()))

	written := &clientmodels.GenAIWebhook{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, written))
	assert.Equal(t, "Bearer secret", written.Headers["Authorization"])
	assert.Equal(t, 120, written.TimeoutSeconds)
	assert.Equal(t, "answer", written.OutputPath)
}

// A configuration with no headers sends an empty object, not
// nothing: the update endpoint keeps stored headers when the
// request carries none, and removing them from the configuration
// means "none".
func TestWebhookWithoutHeadersSendsNone(t *testing.T) {
	resourceModel := &genaiWebhookResourceModel{
		Name:    types.StringValue("bare"),
		URL:     types.StringValue("https://agent.example.com/run"),
		Headers: types.MapNull(types.StringType),
	}
	written := &clientmodels.GenAIWebhook{}
	assert.Nil(t, resourceModel.ToClientModel(context.Background(), written))
	assert.NotNil(t, written.Headers)
	assert.Equal(t, 0, len(written.Headers))
}

// An empty description stays null rather than flipping to "" and
// showing a diff on every plan.
func TestWebhookEmptyDescriptionStaysNull(t *testing.T) {
	resourceModel := &genaiWebhookResourceModel{}
	resourceModel.FromClientModel(
		context.Background(),
		&clientmodels.GenAIWebhook{ID: "x", Name: "n", URL: "https://a.example.com"},
		&diag.Diagnostics{},
	)
	assert.True(t, resourceModel.Description.IsNull())
	// The defaults are real values the server stored.
	assert.Equal(t, "", resourceModel.RequestTemplate.ValueString())
	assert.False(t, resourceModel.RequestTemplate.IsNull())
}
