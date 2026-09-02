package genaiprompt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
	"terraform-provider-oodle/internal/validatorutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiPromptResource{}
	_ resource.ResourceWithConfigure   = &genaiPromptResource{}
	_ resource.ResourceWithImportState = &genaiPromptResource{}
)

const (
	promptTypeText = "text"
	promptTypeChat = "chat"

	// latestLabel is applied by the server to the newest version and
	// is not accepted as input.
	latestLabel = "latest"
)

var validPromptTypes = map[string]struct{}{
	promptTypeText: {},
	promptTypeChat: {},
}

// genaiPromptResource manages one version of a named prompt.
//
// Prompts are append-only: the API has no endpoint that rewrites a
// version in place, so an edit to the body publishes a new version
// and the resource tracks that version from then on. Only the
// versions Terraform created are destroyed, which leaves versions
// published from the UI or the SDK alone.
type genaiPromptResource struct {
	client *oodlehttp.GenAIPromptClient
}

func NewGenAIPromptResource() resource.Resource {
	return &genaiPromptResource{}
}

func (r *genaiPromptResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_prompt"
}

func (r *genaiPromptResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*oodlehttp.OodleApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *oodlehttp.OodleApiClient, got: %T. "+
					"Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = oodlehttp.NewGenAIPromptClient(client)
}

func (r *genaiPromptResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "A versioned GenAI prompt.\n\n" +
			"Prompts are append-only: editing the body publishes a new " +
			"version rather than rewriting the current one, and " +
			"applications resolve a prompt by label. Moving the " +
			"\"production\" label is how a new version is rolled out " +
			"without a deploy.\n\n" +
			"The resource tracks one version — the one it last " +
			"published — and destroying it removes only that version. " +
			"Earlier versions it superseded are left in place, along " +
			"with any published from the UI or an SDK, so the rollback " +
			"history survives. Delete the prompt outright to remove " +
			"every version.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the prompt version.",
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "Name of the prompt. Renaming forces " +
					"replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "'text' for a single prompt string, or 'chat' " +
					"for a list of chat messages. Defaults to 'text' and " +
					"must match the other versions of the same name, so " +
					"changing it forces replacement.",
				Validators: []validator.String{
					validatorutils.NewChoiceValidator(validPromptTypes),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				Required: true,
				Description: "The prompt body. For type 'text' this is the " +
					"prompt itself, with {{var}} placeholders. For type " +
					"'chat' it is a JSON array of messages, for example " +
					"jsonencode([{role = \"system\", content = \"...\"}]).",
			},
			"labels": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Labels pointing at this version, for example " +
					"[\"production\"]. A label moves off whichever version " +
					"held it. The server-managed \"latest\" label is not " +
					"accepted here.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Tags applied to the prompt.",
			},
			"config": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON object of model configuration stored " +
					"alongside the prompt.",
			},
			"commit_message": schema.StringAttribute{
				Optional:    true,
				Description: "Message describing this version's change.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned version number.",
			},
		},
	}
}

func (r *genaiPromptResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan genaiPromptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.publish(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating prompt", err.Error())
		return
	}

	plan.FromClientModel(ctx, created, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *genaiPromptResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state genaiPromptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Name.IsNull() || state.Version.IsNull() {
		resp.Diagnostics.AddError(
			"Incomplete state",
			"name and version are required to read a prompt.",
		)
		return
	}

	prompt, err := r.client.GetVersion(
		ctx, state.Name.ValueString(), state.Version.ValueInt64(),
	)
	if err != nil {
		if errors.Is(err, oodlehttp.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading prompt",
			fmt.Sprintf(
				"Could not read %q version %d: %v",
				state.Name.ValueString(),
				state.Version.ValueInt64(),
				err,
			),
		)
		return
	}

	state.FromClientModel(ctx, prompt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *genaiPromptResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state genaiPromptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Moving a label is not an edit to the prompt, so it must not
	// publish a version. Rolling "production" onto an existing
	// version is the common case and would otherwise leave a trail
	// of identical versions behind.
	if plan.sameContentAs(&state) {
		labels, err := resourceutils.StringListToSlice(ctx, plan.Labels)
		if err != nil {
			resp.Diagnostics.AddError("Invalid labels", err.Error())
			return
		}

		if err := r.client.SetVersionLabels(
			ctx,
			state.Name.ValueString(),
			state.Version.ValueInt64(),
			labels,
		); err != nil {
			resp.Diagnostics.AddError(
				"Error updating prompt labels", err.Error(),
			)
			return
		}

		plan.ID = state.ID
		plan.Version = state.Version
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

		return
	}

	published, err := r.publish(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating prompt", err.Error())
		return
	}

	plan.FromClientModel(ctx, published, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *genaiPromptResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state genaiPromptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only this version is removed. Deleting the prompt outright
	// would take versions Terraform never created with it.
	err := r.client.DeleteVersion(
		ctx, state.Name.ValueString(), state.Version.ValueInt64(),
	)
	if err != nil && !errors.Is(err, oodlehttp.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting prompt version",
			fmt.Sprintf(
				"Could not delete %q version %d: %v",
				state.Name.ValueString(),
				state.Version.ValueInt64(),
				err,
			),
		)
		return
	}
}

// ImportState imports a prompt version addressed as "name:version",
// because a prompt name alone does not identify one.
func (r *genaiPromptResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	name, versionText, found := strings.Cut(req.ID, ":")
	if !found || name == "" || versionText == "" {
		resp.Diagnostics.AddError(
			"Unexpected import identifier",
			fmt.Sprintf(
				"Expected \"name:version\", for example "+
					"\"support-reply:3\", got: %q.",
				req.ID,
			),
		)
		return
	}

	version, err := strconv.ParseInt(versionText, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected import identifier",
			fmt.Sprintf("Version %q is not a number.", versionText),
		)
		return
	}

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("name"), name)...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("version"), version)...,
	)
}

// publish creates a new version of the prompt from the plan.
func (r *genaiPromptResource) publish(
	ctx context.Context,
	plan *genaiPromptResourceModel,
) (*clientmodels.GenAIPrompt, error) {
	model := &clientmodels.GenAIPrompt{}
	if err := plan.ToClientModel(ctx, model); err != nil {
		return nil, err
	}

	return r.client.Create(ctx, model)
}

// promptToRaw encodes the prompt body for the API, which carries a
// JSON string for text prompts and a JSON array for chat prompts.
func promptToRaw(
	promptText string,
	promptType string,
) (json.RawMessage, error) {
	if promptType == promptTypeChat {
		if !json.Valid([]byte(promptText)) {
			return nil, errors.New(
				"prompt must be a JSON array of chat messages when " +
					"type is \"chat\"",
			)
		}

		return json.RawMessage(promptText), nil
	}

	encoded, err := json.Marshal(promptText)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

// rawToPrompt decodes the prompt body back into the configured form.
func rawToPrompt(
	raw json.RawMessage,
	promptType string,
	prior types.String,
) types.String {
	if promptType != promptTypeChat {
		var text string
		if err := json.Unmarshal(raw, &text); err == nil {
			return types.StringValue(text)
		}
	}

	return resourceutils.RawToJSONString(raw, prior)
}
