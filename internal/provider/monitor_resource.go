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
	if model.For > 0 {
		c.For = types.StringValue(validatorutils.ShortDur(model.For))
	}

	if model.KeepFiringFor > 0 {
		c.KeepFiringFor = types.StringValue(validatorutils.ShortDur(model.KeepFiringFor))
	}
	return &c
}

func (c *conditionModel) toModel() (*models.Condition, error) {
	op, err := models.ConditionOpFromString(c.Operation.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ConditionOp: %v", err)
	}

	var forVal time.Duration
	if !c.For.IsNull() && !c.For.IsUnknown() && len(c.For.ValueString()) > 0 {
		forVal, err = time.ParseDuration(c.For.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to parse warning forVal: %v", err)
		}
	}

	var keepFiringForVal time.Duration
	if !c.KeepFiringFor.IsNull() && !c.KeepFiringFor.IsUnknown() && len(c.KeepFiringFor.ValueString()) > 0 {
		keepFiringForVal, err = time.ParseDuration(c.KeepFiringFor.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to parse warning keepFiringFor: %v", err)
		}
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
	diagnosticsOut *diag.Diagnostics,
) {
	// Reset the model to clear any existing data.
	*m = monitorResourceModel{}

	m.ID = types.StringValue(model.ID.UUID.String())
	m.Name = types.StringValue(model.Name)
	m.PromQLQuery = types.StringValue(model.PromQLQuery)
	if model.Interval > 0 {
		m.Interval = types.StringValue(validatorutils.ShortDur(model.Interval))
	}
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
		m.Labels = validatorutils.ToAttrMap(model.Labels, diagnosticsOut)
	} else {
		m.Labels = types.MapNull(basetypes.StringType{})
	}

	if len(model.Annotations) > 0 {
		m.Annotations = validatorutils.ToAttrMap(model.Annotations, diagnosticsOut)
	} else {
		m.Annotations = types.MapNull(basetypes.StringType{})
	}

	if len(model.Grouping.ByLabels) > 0 || model.Grouping.Disabled || model.Grouping.ByMonitor {
		m.Grouping = &grouping{}
		m.Grouping.ByMonitor = types.BoolValue(model.Grouping.ByMonitor)
		m.Grouping.ByLabels = validatorutils.ToAttrList(model.Grouping.ByLabels, diagnosticsOut)
		m.Grouping.Disabled = types.BoolValue(model.Grouping.Disabled)
		m.Grouping.ByMonitor = types.BoolValue(model.Grouping.ByMonitor)
	}

	if model.NotificationPolicyID != nil {
		m.NotificationPolicyID = types.StringValue(model.NotificationPolicyID.UUID.String())
	}

	if model.GroupWait != nil {
		m.GroupWait = types.StringValue(validatorutils.ShortDur(*model.GroupWait))
	}

	if model.GroupInterval != nil {
		m.GroupInterval = types.StringValue(validatorutils.ShortDur(*model.GroupInterval))
	}

	if model.RepeatInterval != nil {
		m.RepeatInterval = types.StringValue(validatorutils.ShortDur(*model.RepeatInterval))
	}
}

func (m *monitorResourceModel) toModel(
	model *models.Monitor,
) error {
	var err error
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID.UUID, err = uuid.Parse(m.ID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse ID UUID %v: %v", m.ID.ValueString(), err)
		}
	}

	model.Name = m.Name.ValueString()
	model.PromQLQuery = m.PromQLQuery.ValueString()
	if !m.Interval.IsNull() {
		model.Interval, err = time.ParseDuration(m.Interval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse interval duration: %v", err)
		}
	}

	if m.Conditions != nil {
		if m.Conditions.Warning != nil {
			model.Conditions.Warn, err = m.Conditions.Warning.toModel()
			if err != nil {
				return fmt.Errorf("failed to parse warning condition: %v", err)
			}
		}

		if m.Conditions.Critical != nil {
			model.Conditions.Critical, err = m.Conditions.Critical.toModel()
			if err != nil {
				return fmt.Errorf("failed to parse critical condition: %v", err)
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

	if m.Grouping != nil {
		model.Grouping.ByMonitor = m.Grouping.ByMonitor.ValueBool()
		if len(m.Grouping.ByLabels.Elements()) > 0 {
			model.Grouping.ByLabels = make([]string, 0, len(m.Grouping.ByLabels.Elements()))
			for _, v := range m.Grouping.ByLabels.Elements() {
				model.Grouping.ByLabels = append(model.Grouping.ByLabels, v.String())
			}

			model.Grouping.Disabled = m.Grouping.Disabled.ValueBool()
			model.Grouping.ByMonitor = m.Grouping.ByMonitor.ValueBool()
		}
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
			return fmt.Errorf("failed to parse group wait duration: %v", err)
		}

		model.GroupWait = &dur
	}

	if len(m.GroupInterval.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.GroupInterval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse group interval duration: %v", err)
		}

		model.GroupInterval = &dur
	}

	if len(m.RepeatInterval.ValueString()) > 0 {
		dur, err := time.ParseDuration(m.RepeatInterval.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse repeat interval duration: %v", err)
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
						Required:    true,
						Description: "If true, only one notification will be sent for this monitor irrespective of how many series match.",
					},
					"by_labels": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "List of labels to group by. One notification is sent for each unique grouping when the monitor fires.",
					},
					"disabled": schema.BoolAttribute{
						Required:    true,
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
	var newPlan monitorResourceModel
	newPlan.fromModel(createdMonitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newPlan)
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

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddError("ID is not set", fmt.Sprintf("ID is required to read monitor: %+v", state))
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

	state.fromModel(monitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *monitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan monitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Assign ID to plan from state.
	var state monitorResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		resp.Diagnostics.AddError("ID is not set", fmt.Sprintf("ID is required to update monitor: %+v", plan))
		return
	}

	monitor := &models.Monitor{}
	err := plan.toModel(monitor)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert plan to model", err.Error())
		return
	}

	updatedMonitor, err := r.client.UpdateMonitor(monitor)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Monitor",
			"Could not update monitor, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with newly created monitor.
	var newState monitorResourceModel
	newState.fromModel(updatedMonitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state monitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddError("ID is not set", fmt.Sprintf("ID is required to delete monitor: %+v", state))
		return
	}

	// Delete existing order
	err := r.client.DeleteMonitor(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Monitor",
			fmt.Sprintf("Could not delete monitor ID %s: %v", state.ID.ValueString(), err),
		)
		return
	}
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
