package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/models"
	"terraform-provider-oodle/internal/validatorutils"
	"time"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &monitorResource{}
	_ resource.ResourceWithConfigure   = &monitorResource{}
	_ resource.ResourceWithImportState = &monitorResource{}
)

// orderResourceModel maps the resource schema data.
type orderResourceModel struct {
	ID          types.String     `tfsdk:"id"`
	Items       []orderItemModel `tfsdk:"items"`
	LastUpdated types.String     `tfsdk:"last_updated"`
}

// orderItemModel maps order item data.
type orderItemModel struct {
	Coffee   orderItemCoffeeModel `tfsdk:"coffee"`
	Quantity types.Int64          `tfsdk:"quantity"`
}

// orderItemCoffeeModel maps coffee order item data.
type orderItemCoffeeModel struct {
	ID          types.Int64   `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Teaser      types.String  `tfsdk:"teaser"`
	Description types.String  `tfsdk:"description"`
	Price       types.Float64 `tfsdk:"price"`
	Image       types.String  `tfsdk:"image"`
}

type conditionModel struct {
	// Operation - The operation to perform for the condition. Possible values are: ">", "<", ">=", "<=", "==", "!=".
	Operation     types.String  `tfsdk:"operation"`
	Value         types.Float64 `tfsdk:"value"`
	For           types.String  `tfsdk:"for"`
	KeepFiringFor types.String  `tfsdk:"keep_firing_for"`
}

func newConditionFromModel(model *models.Condition) *conditionModel {
	c := conditionModel{}
	c.Operation = types.StringValue(model.Op.String())
	c.Value = types.Float64Value(model.Value)
	c.For = types.StringValue(model.For.String())
	c.KeepFiringFor = types.StringValue(model.KeepFiringFor.String())
	return &c
}

func (c *conditionModel) toModel() (*models.Condition, error) {
	op, err := models.ConditionOpFromString(c.Operation.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ConditionOp: %w", err)
	}

	forVal, err := time.ParseDuration(c.For.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse warning forVal: %w", err)
	}

	keepFiringForVal, err := time.ParseDuration(c.KeepFiringFor.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse warning keepFiringFor: %w", err)
	}

	return &models.Condition{
		Op:            op,
		Value:         c.Value.ValueFloat64(),
		For:           forVal,
		KeepFiringFor: keepFiringForVal,
	}, nil
}

type conditionsModel struct {
	Warning  *conditionModel `tfsdk:"warning"`
	Critical *conditionModel `tfsdk:"critical"`
}

type grouping struct {
	ByMonitor types.Bool `tfsdk:"by_monitor"`
	ByLabels  types.List `tfsdk:"by_labels"`
	Disabled  types.Bool `tfsdk:"disabled"`
}

type monitorResourceModel struct {
	ID                   types.String     `tfsdk:"id"`
	Name                 types.String     `tfsdk:"name"`
	Interval             types.String     `tfsdk:"interval"`
	PromQLQuery          types.String     `tfsdk:"promql_query"`
	Conditions           *conditionsModel `tfsdk:"conditions"`
	Labels               types.Map        `tfsdk:"labels"`
	Annotations          types.Map        `tfsdk:"annotations"`
	Grouping             *grouping        `tfsdk:"grouping"`
	NotificationPolicyID types.String     `tfsdk:"notification_policy_id"`
	GroupWait            types.String     `tfsdk:"group_wait"`
	GroupInterval        types.String     `tfsdk:"group_interval"`
	RepeatInterval       types.String     `tfsdk:"repeat_interval"`
}

func (m *monitorResourceModel) fromModel(
	model *models.Monitor,
	diagnosticsOut diag.Diagnostics,
) {
	// Reset the model to clear any existing data.
	*m = monitorResourceModel{
		Grouping: &grouping{},
	}

	m.ID = types.StringValue(model.ID.UUID.String())
	m.Name = types.StringValue(model.Name)
	m.PromQLQuery = types.StringValue(model.PromQLQuery)
	if model.Conditions.Warn != nil {
		if m.Conditions == nil {
			m.Conditions = &conditionsModel{}
		}

		m.Conditions.Warning = newConditionFromModel(model.Conditions.Warn)
	}

	if model.Conditions.Critical != nil {
		if m.Conditions == nil {
			m.Conditions = &conditionsModel{}
		}

		m.Conditions.Critical = newConditionFromModel(model.Conditions.Critical)
	}

	if len(model.Labels) > 0 {
		m.Labels = validatorutils.ToAttrMap(model.Labels, &diagnosticsOut)
	} else {
		m.Labels, _ = types.MapValue(basetypes.StringType{}, nil)
	}

	if len(model.Annotations) > 0 {
		m.Annotations = validatorutils.ToAttrMap(model.Annotations, &diagnosticsOut)
	} else {
		m.Annotations, _ = types.MapValue(basetypes.StringType{}, nil)
	}

	m.Grouping.ByMonitor = types.BoolValue(model.Grouping.ByMonitor)
	if len(model.Grouping.ByLabels) > 0 {
		m.Grouping.ByLabels = validatorutils.ToAttrList(model.Grouping.ByLabels, &diagnosticsOut)
		m.Grouping.Disabled = types.BoolValue(model.Grouping.Disabled)
		m.Grouping.ByMonitor = types.BoolValue(model.Grouping.ByMonitor)
	} else {
		m.Grouping.ByLabels, _ = types.ListValue(basetypes.StringType{}, nil)
	}

	if model.NotificationPolicyID != nil {
		m.NotificationPolicyID = types.StringValue(model.NotificationPolicyID.UUID.String())
	}

	if model.GroupWait != nil {
		m.GroupWait = types.StringValue(model.GroupWait.String())
	}

	if model.GroupInterval != nil {
		m.GroupInterval = types.StringValue(model.GroupInterval.String())
	}

	if model.RepeatInterval != nil {
		m.RepeatInterval = types.StringValue(model.RepeatInterval.String())
	}
}

func (m *monitorResourceModel) toModel(
	model *models.Monitor,
) error {
	var err error
	model.ID.UUID, err = uuid.Parse(m.ID.String())
	if err != nil {
		return fmt.Errorf("failed to parse UUID: %w", err)
	}

	model.Name = m.Name.ValueString()
	model.PromQLQuery = m.PromQLQuery.ValueString()
	if m.Conditions != nil {
		if m.Conditions.Warning != nil {
			model.Conditions.Warn, err = m.Conditions.Warning.toModel()
			if err != nil {
				return fmt.Errorf("failed to parse warning condition: %w", err)
			}
		}

		if m.Conditions.Critical != nil {
			model.Conditions.Critical, err = m.Conditions.Critical.toModel()
			if err != nil {
				return fmt.Errorf("failed to parse critical condition: %w", err)
			}
		}
	}

	if len(m.Labels.Elements()) > 0 {
		model.Labels = make(map[string]string)
		for k, v := range m.Labels.Elements() {
			model.Labels[k] = v.String()
		}
	}

	if len(m.Annotations.Elements()) > 0 {
		model.Annotations = make(map[string]string)
		for k, v := range m.Annotations.Elements() {
			model.Annotations[k] = v.String()
		}
	}

	model.Grouping.ByMonitor = m.Grouping.ByMonitor.ValueBool()
	if len(m.Grouping.ByLabels.Elements()) > 0 {
		model.Grouping.ByLabels = make([]string, 0, len(m.Grouping.ByLabels.Elements()))
		for _, v := range m.Grouping.ByLabels.Elements() {
			model.Grouping.ByLabels = append(model.Grouping.ByLabels, v.String())
		}

		model.Grouping.Disabled = m.Grouping.Disabled.ValueBool()
		model.Grouping.ByMonitor = m.Grouping.ByMonitor.ValueBool()
	}

	if len(m.NotificationPolicyID.ValueString()) > 0 {
		uid, err := uuid.Parse(m.NotificationPolicyID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse notification policy UUID: %v", err)
		}

		model.NotificationPolicyID = &models.ID{UUID: uid}
	}

	if len(m.GroupWait.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.GroupWait.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse group wait duration: %w", err)
		}

		model.GroupWait = &dur
	}

	if len(m.GroupInterval.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.GroupInterval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse group interval duration: %w", err)
		}

		model.GroupInterval = &dur
	}

	if len(m.RepeatInterval.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.RepeatInterval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse repeat interval duration: %w", err)
		}

		model.RepeatInterval = &dur
	}

	return nil
}

// NewMonitorResource is a helper function to simplify the provider implementation.
func NewMonitorResource() resource.Resource {
	return &monitorResource{}
}

// monitorResource is the resource implementation.
type monitorResource struct {
	client *oodlehttp.Client
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
				Required: true,
				Validators: []validator.String{
					validatorutils.NewPromQLValidator(),
				},
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
									validatorutils.NewComparatorValidator(),
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
					"critical": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"operation": schema.StringAttribute{
								Required:    true,
								Description: "The operation to perform for the condition. Possible values are: '>', '<', '>=', '<=', '==', '!='.",
								Validators: []validator.String{
									validatorutils.NewComparatorValidator(),
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
			},
			"notification_policy_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the notification policy to use for the monitor.",
			},
			"group_wait": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
				Description: "Interval at which to send alerts for the same group of alerts after the first alert.",
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

// Create a new resource.
func (r *monitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan monitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor := &models.Monitor{}
	err := plan.toModel(monitor)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert plan to model", err.Error())
		return
	}

	// Create new monitor
	createdMonitor, err := r.client.CreateMonitor(monitor)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Monitor",
			"Could not create monitor, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with newly created monitor.
	plan.fromModel(createdMonitor, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *monitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state monitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from HashiCups
	monitor, err := r.client.GetMonitor(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Monitors",
			"Could not read Oodle Monitor ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.fromModel(monitor, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("ASDASDASD1", "ASDASDASD1")
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("ASDASDASD2", fmt.Sprintf("%+v", state.Labels))
		return
	}
}

func (r *monitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//// Retrieve values from plan
	//var plan orderResourceModel
	//diags := req.Plan.Get(ctx, &plan)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// Generate API request body from plan
	//var hashicupsItems []hashicups.OrderItem
	//for _, item := range plan.Items {
	//	hashicupsItems = append(hashicupsItems, hashicups.OrderItem{
	//		Coffee: hashicups.Coffee{
	//			ID: int(item.Coffee.ID.ValueInt64()),
	//		},
	//		Quantity: int(item.Quantity.ValueInt64()),
	//	})
	//}
	//
	//// Update existing order
	//_, err := r.client.UpdateOrder(plan.ID.ValueString(), hashicupsItems)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"Error Updating HashiCups Order",
	//		"Could not update order, unexpected error: "+err.Error(),
	//	)
	//	return
	//}
	//
	//// Fetch updated items from GetOrder as UpdateOrder items are not
	//// populated.
	//order, err := r.client.GetOrder(plan.ID.ValueString())
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"Error Reading HashiCups Order",
	//		"Could not read HashiCups order ID "+plan.ID.ValueString()+": "+err.Error(),
	//	)
	//	return
	//}
	//
	//// Update resource state with updated items and timestamp
	//plan.Items = []orderItemModel{}
	//for _, item := range order.Items {
	//	plan.Items = append(plan.Items, orderItemModel{
	//		Coffee: orderItemCoffeeModel{
	//			ID:          types.Int64Value(int64(item.Coffee.ID)),
	//			Name:        types.StringValue(item.Coffee.Name),
	//			Teaser:      types.StringValue(item.Coffee.Teaser),
	//			Description: types.StringValue(item.Coffee.Description),
	//			Price:       types.Float64Value(item.Coffee.Price),
	//			Image:       types.StringValue(item.Coffee.Image),
	//		},
	//		Quantity: types.Int64Value(int64(item.Quantity)),
	//	})
	//}
	//plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	//
	//diags = resp.State.Set(ctx, plan)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
}

func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state orderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//// Delete existing order
	//err := r.client.DeleteOrder(state.ID.ValueString())
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"Error Deleting HashiCups Order",
	//		"Could not delete order, unexpected error: "+err.Error(),
	//	)
	//	return
	//}
}

// Configure adds the provider configured client to the resource.
func (r *monitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*oodlehttp.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *oodlehttp.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
