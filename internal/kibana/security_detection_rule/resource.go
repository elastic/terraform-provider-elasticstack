package security_detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

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
