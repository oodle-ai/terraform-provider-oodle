package genaiprompt

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func strList(t *testing.T, values ...string) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), types.StringType, values)
	if diags.HasError() {
		t.Fatalf("list: %v", diags)
	}
	return list
}

func promptModel(t *testing.T, tags ...string) *genaiPromptResourceModel {
	t.Helper()
	return &genaiPromptResourceModel{
		Name:          types.StringValue("support-reply"),
		Type:          types.StringValue(promptTypeText),
		Prompt:        types.StringValue("Hello {{name}}"),
		Labels:        strList(t, "production", "canary"),
		Tags:          strList(t, tags...),
		CommitMessage: types.StringValue("first"),
	}
}

// The server stores tags and labels sorted. Read back in that order
// against a configuration written in another, every plan showed a
// change, and for a prompt any change published a new version.
func TestFromClientModelKeepsConfiguredTagOrder(t *testing.T) {
	m := promptModel(t, "shopassist", "demo")
	var diags diag.Diagnostics

	m.FromClientModel(context.Background(), &clientmodels.GenAIPrompt{
		ID:            "v1",
		Name:          "support-reply",
		Type:          promptTypeText,
		Prompt:        json.RawMessage(`"Hello {{name}}"`),
		Labels:        []string{"canary", "latest", "production"},
		Tags:          []string{"demo", "shopassist"},
		CommitMessage: "first",
		Version:       1,
	}, &diags)

	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if !m.Tags.Equal(strList(t, "shopassist", "demo")) {
		t.Errorf("tags = %v, want the configured order", m.Tags)
	}
	if !m.Labels.Equal(strList(t, "production", "canary")) {
		t.Errorf("labels = %v, want the configured order", m.Labels)
	}
}

// A configured tag the server no longer holds was removed outside
// Terraform, and the plan has to show it.
func TestFromClientModelReportsAMissingTag(t *testing.T) {
	m := promptModel(t, "shopassist", "demo")
	var diags diag.Diagnostics

	m.FromClientModel(context.Background(), &clientmodels.GenAIPrompt{
		Name:    "support-reply",
		Prompt:  json.RawMessage(`"Hello {{name}}"`),
		Tags:    []string{"demo", "other"},
		Version: 1,
	}, &diags)

	if !m.Tags.Equal(strList(t, "demo", "other")) {
		t.Errorf("tags = %v, want the server's tags when one is missing", m.Tags)
	}
}

// Reordering tags is not new content, so Update moves nothing and
// publishes nothing. A different tag is new content.
func TestSameContentIgnoresTagOrder(t *testing.T) {
	ctx := context.Background()
	state := promptModel(t, "a", "b")

	if !promptModel(t, "b", "a").sameContentAs(ctx, state) {
		t.Error("reordered tags must not read as new content")
	}
	if promptModel(t, "a", "c").sameContentAs(ctx, state) {
		t.Error("a changed tag must read as new content")
	}

	body := promptModel(t, "a", "b")
	body.Prompt = types.StringValue("Hi {{name}}")
	if body.sameContentAs(ctx, state) {
		t.Error("a changed body must read as new content")
	}
}

// Publishing v2 with more tags adds them to v1 on the server. The v1
// resource did not change, and reading it as changed published a new
// version on every apply.
func TestFromClientModelIgnoresTagsMergedFromOtherVersions(t *testing.T) {
	m := promptModel(t, "a")
	var diags diag.Diagnostics

	m.FromClientModel(context.Background(), &clientmodels.GenAIPrompt{
		Name:    "support-reply",
		Prompt:  json.RawMessage(`"Hello {{name}}"`),
		Tags:    []string{"b", "a"},
		Version: 1,
	}, &diags)

	if !m.Tags.Equal(strList(t, "a")) {
		t.Errorf("tags = %v, want the configured tags kept", m.Tags)
	}
}
