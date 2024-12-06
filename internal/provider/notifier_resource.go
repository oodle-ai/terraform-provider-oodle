package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type notifierResource struct {
	baseResource
}

// NewNotifierResource is a helper function to simplify the provider implementation.
func NewNotifierResource() resource.Resource {
	return &notifierResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &notifierResource{}
	_ resource.ResourceWithConfigure   = &notifierResource{}
	_ resource.ResourceWithImportState = &notifierResource{}
)

type notifierConfigCommonModel struct {
	SendResolved types.Bool `tfsdk:"send_resolved"`
}

type hostPort struct {
	Host types.String `tfsdk:"host"`
	Port types.Number `tfsdk:"port"`
}

type tlsConfigModel struct {
	CA                 types.String `tfsdk:"ca"`
	Cert               types.String `tfsdk:"cert"`
	Key                types.String `tfsdk:"key"`
	Insecure           types.Bool   `tfsdk:"insecure"`
	CAFile             types.String `tfsdk:"ca_file"`
	CertFile           types.String `tfsdk:"cert_file"`
	KeyFile            types.String `tfsdk:"key_file"`
	CARef              types.String `tfsdk:"ca_ref"`
	CertRef            types.String `tfsdk:"cert_ref"`
	KeyRef             types.String `tfsdk:"key_ref"`
	ServerName         types.String `tfsdk:"server_name"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	MinVersion         types.Int32  `tfsdk:"min_version"`
	MaxVersion         types.Int32  `tfsdk:"max_version"`
}

type emailConfigModel struct {
	notifierConfigCommonModel
	To           types.String   `tfsdk:"to"`
	From         types.String   `tfsdk:"from"`
	Hello        types.String   `tfsdk:"hello"`
	Smarthost    hostPort       `tfsdk:"smart_host"`
	AuthUsername types.String   `tfsdk:"auth_username"`
	AuthPassword types.String   `tfsdk:"auth_password"`
	AuthSecret   types.String   `tfsdk:"auth_secret"`
	AuthIdentify types.String   `tfsdk:"auth_identity"`
	Headers      types.Map      `tfsdk:"headers"`
	HTML         types.String   `tfsdk:"html"`
	Text         types.String   `tfsdk:"text"`
	RequireTLS   types.Bool     `tfsdk:"require_tls"`
	TLSConfig    tlsConfigModel `tfsdk:"tls_config"`
}

type pagerdutyConfigModel struct {
	notifierConfigCommonModel
}

type notifierModel struct {
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	Type        types.String      `tfsdk:"type"`
	EmailConfig *emailConfigModel `tfsdk:"email_config"`
}

func (n notifierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notifier"
}

func (n notifierResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (n notifierResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (n notifierResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (n notifierResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (n notifierResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
