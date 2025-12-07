package security_detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &securityDetectionRuleResource{}
var _ resource.ResourceWithConfigure = &securityDetectionRuleResource{}
var _ resource.ResourceWithImportState = &securityDetectionRuleResource{}

func NewSecurityDetectionRuleResource() resource.Resource {
	return &securityDetectionRuleResource{}
}

type securityDetectionRuleResource struct {
	client *clients.ApiClient
}

func (r *securityDetectionRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kibana_security_detection_rule"
}

func (r *securityDetectionRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *securityDetectionRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
