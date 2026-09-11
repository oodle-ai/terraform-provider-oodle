// Package prompt manages GenAI prompts.
//
// A prompt is not an ordinary CRUD resource. Versions are
// immutable and assigned by the server, and which one is live is
// decided by a label, so an update here publishes the next
// version and points the label at it rather than overwriting a
// row.
//
// That has a consequence worth stating plainly before anybody
// adopts this: Terraform reverts drift. A prompt edited in the
// Oodle UI is republished from the configuration on the next
// apply. That is the point when one definition has to serve
// several deployments, and it is the wrong tool if the prompt is
// meant to be tuned by hand -- use a non-production label for
// that, and let this own `production`.
package prompt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &promptResource{}
	_ resource.ResourceWithConfigure   = &promptResource{}
	_ resource.ResourceWithImportState = &promptResource{}
)

// productionLabel is the label an application resolves when it
// asks for a prompt by name and says nothing else.
const productionLabel = "production"

type promptResource struct {
	client *oodlehttp.PromptClient
}

type promptResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Prompt        types.String `tfsdk:"prompt"`
	Label         types.String `tfsdk:"label"`
	CommitMessage types.String `tfsdk:"commit_message"`
	Version       types.Int64  `tfsdk:"version"`
}

func NewPromptResource() resource.Resource {
	return &promptResource{}
}

func (r *promptResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (r *promptResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a GenAI prompt. Changing the text " +
			"publishes a new version and moves the label to it; " +
			"a prompt edited outside Terraform is republished " +
			"from this configuration on the next apply. Declare " +
			"one resource per prompt name: destroying it removes " +
			"every version, whatever label they carry. Experiment " +
			"on another label in the UI, and let this own " +
			"production.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The prompt name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "The prompt name, which is the key " +
					"applications fetch it by. A '/' groups " +
					"prompts into folders in the UI. There is no " +
					"rename on the API, so changing this " +
					"replaces the prompt.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				Required:    true,
				Description: "The prompt text.",
			},
			"label": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The label this version carries, and " +
					"the one read back to detect drift. Defaults " +
					"to 'production', which is what an " +
					"application resolves when it names no label.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"commit_message": schema.StringAttribute{
				Optional: true,
				Description: "Recorded on each published version. " +
					"Version numbers are assigned per " +
					"deployment, so this is the only thing that " +
					"identifies the same text across them.",
			},
			"version": schema.Int64Attribute{
				Computed: true,
				Description: "The version this label points at. " +
					"Assigned by the server, and local to one " +
					"deployment.",
			},
		},
	}
}

func (r *promptResource) Configure(
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
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *oodlehttp.OodleApiClient, got: %T.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = oodlehttp.NewPromptClient(client)
}

func (r *promptResource) label(model *promptResourceModel) string {
	if model.Label.IsNull() || model.Label.IsUnknown() ||
		model.Label.ValueString() == "" {
		return productionLabel
	}
	return model.Label.ValueString()
}

// publish creates the next version and points the label at it.
// It backs both Create and Update, because a version is never
// edited in place.
func (r *promptResource) publish(
	ctx context.Context,
	model *promptResourceModel,
) (*clientmodels.Prompt, error) {
	label := r.label(model)
	return r.client.Create(ctx, &clientmodels.Prompt{
		Name:          model.Name.ValueString(),
		Type:          "text",
		Prompt:        model.Prompt.ValueString(),
		Labels:        []string{label},
		CommitMessage: model.CommitMessage.ValueString(),
	})
}

func (r *promptResource) applyResult(
	model *promptResourceModel,
	published *clientmodels.Prompt,
	label string,
) {
	model.ID = types.StringValue(model.Name.ValueString())
	model.Label = types.StringValue(label)
	model.Version = types.Int64Value(int64(published.Version))
}

func (r *promptResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan promptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	published, err := r.publish(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating prompt", err.Error())
		return
	}

	r.applyResult(&plan, published, r.label(&plan))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *promptResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state promptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	label := r.label(&state)
	live, err := r.client.Get(ctx, state.Name.ValueString(), label)
	if err != nil {
		resp.Diagnostics.AddError("Error reading prompt", err.Error())
		return
	}

	// Nothing carries the label any more, so the resource is
	// gone as far as the configuration is concerned and the next
	// apply recreates it.
	if live == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// The text is read back so an edit made outside Terraform
	// shows up as a diff rather than being invisible until
	// somebody wonders why two deployments disagree.
	state.ID = types.StringValue(state.Name.ValueString())
	state.Prompt = types.StringValue(live.Prompt)
	state.Label = types.StringValue(label)
	state.Version = types.Int64Value(int64(live.Version))
	if live.CommitMessage != "" {
		state.CommitMessage = types.StringValue(live.CommitMessage)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *promptResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan promptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	published, err := r.publish(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating prompt", err.Error())
		return
	}

	r.applyResult(&plan, published, r.label(&plan))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *promptResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state promptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// This removes every version, not just the one the label
	// points at. There is no per-version destroy that would
	// leave a prompt half-managed.
	if err := r.client.Delete(
		ctx, state.Name.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error deleting prompt", err.Error())
	}
}

func (r *promptResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// The name is the identifier, so importing an existing
	// prompt is `terraform import oodle_prompt.x team/reply`.
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...,
	)
}
