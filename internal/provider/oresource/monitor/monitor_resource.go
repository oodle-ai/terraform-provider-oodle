package monitor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &monitorResource{}
	_ resource.ResourceWithConfigure   = &monitorResource{}
	_ resource.ResourceWithImportState = &monitorResource{}
)

const monitorsResource = "monitors"

var validComparators = map[string]struct{}{
	"==": {},
	"!=": {},
	">":  {},
	"<":  {},
	">=": {},
	"<=": {},
}

var validMatchTypes = map[string]struct{}{
	"=":  {},
	"!=": {},
	"=~": {},
	"!~": {},
}

// monitorResource is the resource implementation.
type monitorResource struct {
	oresource.BaseResource[*clientmodels.Monitor, *monitorResourceModel]
}

func NewMonitorResource() resource.Resource {
	modelCreator := func() *clientmodels.Monitor {
		return &clientmodels.Monitor{}
	}
	return &monitorResource{
		BaseResource: oresource.NewBaseResource[*clientmodels.Monitor, *monitorResourceModel](
			func() *monitorResourceModel {
				return &monitorResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.Monitor] {
				return oodlehttp.NewModelClient[*clientmodels.Monitor](
					oodleHttpClient,
					monitorsResource,
					modelCreator,
				)
			},
		),
	}
}

// Metadata returns the resource type name.
func (r *monitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

// Schema defines the schema for the resource.
func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the monitor.",
				Validators: []validator.String{
					validatorutils.NewUUIDValidator(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the monitor.",
			},
			"interval": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
				Description: "Interval at which the monitor should be evaluated. Default is 1m.",
			},
			"promql_query": schema.StringAttribute{
				Required:    true,
				Description: "Prometheus query for the monitor.",
			},
			"conditions": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"warning": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"operation": schema.StringAttribute{
								Required:    true,
								Description: "The operation to perform for the condition. Possible values are: '>', '<', '>=', '<=', '==', '!='.",
								Validators: []validator.String{
									validatorutils.NewChoiceValidator(validComparators),
								},
							},
							"value": schema.Float64Attribute{
								Required:    true,
								Description: "Value to compare against.",
							},
							"for": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									validatorutils.NewDurationValidator(),
								},
								Description: "Duration for which the condition should be true before the alert is triggered.",
							},
							"keep_firing_for": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									validatorutils.NewDurationValidator(),
								},
								Description: "Duration for which the alert should keep firing after the condition is no longer true.",
							},
							"alert_on_no_data": schema.BoolAttribute{
								Optional:    true,
								Description: "If true, the monitor is considered firing when there is no data for the query.",
							},
						},
					},
					"critical": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"operation": schema.StringAttribute{
								Required:    true,
								Description: "The operation to perform for the condition. Possible values are: '>', '<', '>=', '<=', '==', '!='.",
								Validators: []validator.String{
									validatorutils.NewChoiceValidator(validComparators),
								},
							},
							"value": schema.Float64Attribute{
								Required:    true,
								Description: "Value to compare against.",
							},
							"for": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									validatorutils.NewDurationValidator(),
								},
								Description: "Duration for which the condition should be true before the alert is triggered.",
							},
							"keep_firing_for": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									validatorutils.NewDurationValidator(),
								},
								Description: "Duration for which the alert should keep firing after the condition is no longer true.",
							},
						},
					},
				},
				Description: "Warning and Critical thresholds for the monitor.",
			},
			"labels": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Additional labels to attach to the fired alerts.",
			},
			"annotations": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Additional metadata to attach to each monitor.",
			},
			"grouping": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"by_monitor": schema.BoolAttribute{
						Optional:    true,
						Description: "If true, only one notification will be sent for this monitor irrespective of how many series match.",
					},
					"by_labels": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "List of labels to group by. One notification is sent for each unique grouping when the monitor fires.",
					},
					"disabled": schema.BoolAttribute{
						Optional:    true,
						Description: "If true, grouping is disabled.",
					},
				},
				Validators: []validator.Object{
					validatorutils.NewGroupingValidator(),
				},
			},
			"notification_policy_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the notification policy to use for the monitor.",
			},
			"label_matcher_notification_policies": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of label matcher notification policies. These policies are evaluated in order, and the first matching policy is used. Within a label matcher, all matchers must match for policy to be effective. If no policy matches, the default notification_policy_id is used if set.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"matchers": schema.ListNestedAttribute{
							Required:    true,
							Description: "List of label matchers that determine when this policy applies.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:    true,
										Description: "The type of match to perform. Valid values are: '=' (equals), '!=' (not equals), '=~' (regex match), '!~' (regex not match).",
										Validators: []validator.String{
											validatorutils.NewChoiceValidator(validMatchTypes),
										},
									},
									"name": schema.StringAttribute{
										Required:    true,
										Description: "The name of the label to match against.",
									},
									"value": schema.StringAttribute{
										Required:    true,
										Description: "The value to match against. For regex matches, this must be a valid regular expression.",
									},
								},
							},
						},
						"notification_policy_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the notification policy to use when labels match.",
							Validators: []validator.String{
								validatorutils.NewUUIDValidator(),
							},
						},
					},
				},
			},
			"group_wait": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
				Description: "Time to wait before sending the first alert for a group of alerts.",
			},
			"group_interval": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
				Description: "Interval at which to send alerts for the same group of alerts after the first alert.",
			},
			"repeat_interval": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
				Description: "Interval at which to send alerts for the same alert after firing. RepeatInterval should be a multiple of GroupInterval.",
			},
		},
	}
}
