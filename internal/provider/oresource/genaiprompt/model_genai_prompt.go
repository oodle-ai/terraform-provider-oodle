package genaiprompt

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiPromptResourceModel struct {
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Type          types.String         `tfsdk:"type"`
	Prompt        types.String         `tfsdk:"prompt"`
	Labels        types.List           `tfsdk:"labels"`
	Tags          types.List           `tfsdk:"tags"`
	Config        jsontypes.Normalized `tfsdk:"config"`
	CommitMessage types.String         `tfsdk:"commit_message"`
	Version       types.Int64          `tfsdk:"version"`
}

func (m *genaiPromptResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.GenAIPrompt,
	diagnosticsOut *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Version = types.Int64Value(model.Version)

	promptType := model.Type
	if promptType == "" {
		promptType = promptTypeText
	}
	m.Type = types.StringValue(promptType)

	m.Prompt = rawToPrompt(model.Prompt, promptType, m.Prompt)
	m.Config = resourceutils.RawToNormalized(model.Config)

	m.Labels = resourceutils.SliceToStringList(
		ctx, withoutLatestLabel(model.Labels), m.Labels, diagnosticsOut,
	)
	m.Tags = resourceutils.SliceToStringList(
		ctx, model.Tags, m.Tags, diagnosticsOut,
	)

	if model.CommitMessage == "" {
		m.CommitMessage = types.StringNull()
	} else {
		m.CommitMessage = types.StringValue(model.CommitMessage)
	}
}

func (m *genaiPromptResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.GenAIPrompt,
) error {
	model.Name = m.Name.ValueString()
	model.CommitMessage = m.CommitMessage.ValueString()

	promptType := m.Type.ValueString()
	if promptType == "" {
		promptType = promptTypeText
	}
	model.Type = promptType

	prompt, err := promptToRaw(m.Prompt.ValueString(), promptType)
	if err != nil {
		return err
	}
	model.Prompt = prompt

	config, err := resourceutils.NormalizedToRaw(m.Config, "config")
	if err != nil {
		return err
	}
	model.Config = config

	labels, err := resourceutils.StringListToSlice(ctx, m.Labels)
	if err != nil {
		return err
	}
	model.Labels = labels

	tags, err := resourceutils.StringListToSlice(ctx, m.Tags)
	if err != nil {
		return err
	}
	model.Tags = tags

	return nil
}

// withoutLatestLabel drops the server-managed "latest" label.
//
// The API attaches it to whichever version was published most
// recently and refuses it as input. Left in, it would appear as a
// label the configuration never asked for and fail the apply.
func withoutLatestLabel(labels []string) []string {
	filtered := make([]string, 0, len(labels))
	for _, label := range labels {
		if label == latestLabel {
			continue
		}

		filtered = append(filtered, label)
	}

	return filtered
}

// sameContentAs reports whether the two models describe the same
// prompt version, ignoring labels. Labels move between existing
// versions; everything else here is what a version is made of, so a
// change to any of it has to be published as a new version.
func (m *genaiPromptResourceModel) sameContentAs(
	other *genaiPromptResourceModel,
) bool {
	return m.Name.Equal(other.Name) &&
		m.Type.Equal(other.Type) &&
		m.Prompt.Equal(other.Prompt) &&
		m.Config.Equal(other.Config) &&
		m.Tags.Equal(other.Tags) &&
		m.CommitMessage.Equal(other.CommitMessage)
}
