package genaicodelibrary

import (
	"context"
	"fmt"
	"regexp"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiCodeLibraryResource{}
	_ resource.ResourceWithConfigure   = &genaiCodeLibraryResource{}
	_ resource.ResourceWithImportState = &genaiCodeLibraryResource{}
)

// genaiCodeLibraryResource is the resource implementation.
type genaiCodeLibraryResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAICodeLibrary,
		*genaiCodeLibraryResourceModel,
	]
}

func NewGenAICodeLibraryResource() resource.Resource {
	return &genaiCodeLibraryResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAICodeLibrary,
			*genaiCodeLibraryResourceModel,
		](
			func() *genaiCodeLibraryResourceModel {
				return &genaiCodeLibraryResourceModel{}
			},
			func() *clientmodels.GenAICodeLibrary {
				return &clientmodels.GenAICodeLibrary{}
			},
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAICodeLibrary] {
				return oodlehttp.NewGenAICodeLibraryClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiCodeLibraryResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_code_library"
}

// Schema defines the schema for the resource.
func (r *genaiCodeLibraryResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "A shared Python module that 'code' eval templates " +
			"import as `shared.<name>`, so that helper code is written " +
			"once. Each change of source_code adds a version; a template " +
			"can pin the version it runs in its library_pins. The API " +
			"refuses to delete a library while a template or another " +
			"library imports it. Requires an enterprise plan and the " +
			"code evaluator feature.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the code library.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "Module name the code imports, as " +
					"`shared.<name>`, for example " +
					"`from shared.<name> import f`. A lower-case letter " +
					"followed by at most 63 lower-case letters, digits " +
					"and underscores; not a Python keyword, " +
					"'oodle_eval' or 'shared'. Cannot be changed after " +
					"creation, because the code that imports the " +
					"library names it.",
				Validators: []validator.String{libraryNameValidator{}},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "What the library provides.",
			},
			"source_code": schema.StringAttribute{
				Required:    true,
				Description: "Python source of the module, at most 256 KB.",
			},
			"version": schema.Int64Attribute{
				Computed: true,
				Description: "Latest version of the library. Each change " +
					"of source_code adds one. Reference it from a " +
					"template's library_pins to pin the version the " +
					"template runs.",
			},
		},
	}
}

// libraryNameRE is the module name the API accepts.
var libraryNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// reservedLibraryNames are refused by the API as module names: the
// Python keywords that match libraryNameRE, the Oodle library, and the
// package the libraries are imported from.
var reservedLibraryNames = []string{
	"and", "as", "assert", "async", "await", "break", "class",
	"continue", "def", "del", "elif", "else", "except", "finally",
	"for", "from", "global", "if", "import", "in", "is", "lambda",
	"nonlocal", "not", "or", "pass", "raise", "return", "try",
	"while", "with", "yield", "oodle_eval", "shared",
}

// libraryNameValidator refuses a name the API would refuse, so the plan
// fails rather than the apply.
type libraryNameValidator struct{}

func (libraryNameValidator) Description(_ context.Context) string {
	return "name must start with a lower-case letter and hold only " +
		"lower-case letters, digits and underscores, at most 64 " +
		"characters, and must not be a Python keyword, 'oodle_eval' " +
		"or 'shared'"
}

func (v libraryNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v libraryNameValidator) ValidateString(
	ctx context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	name := req.ConfigValue.ValueString()
	if !libraryNameRE.MatchString(name) ||
		slices.Contains(reservedLibraryNames, name) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid code library name",
			fmt.Sprintf("%q: %s", name, v.Description(ctx)),
		)
	}
}
