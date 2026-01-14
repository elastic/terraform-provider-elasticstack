package proxy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type proxyModel struct {
	ID                     types.String `tfsdk:"id"`
	ProxyID                types.String `tfsdk:"proxy_id"`
	Name                   types.String `tfsdk:"name"`
	URL                    types.String `tfsdk:"url"`
	Certificate            types.String `tfsdk:"certificate"`
	CertificateAuthorities types.String `tfsdk:"certificate_authorities"`
	CertificateKey         types.String `tfsdk:"certificate_key"`
	IsPreconfigured        types.Bool   `tfsdk:"is_preconfigured"`
	SpaceIds               types.Set    `tfsdk:"space_ids"` //> string
}

func (model *proxyModel) populateFromAPI(ctx context.Context, proxy *kbapi.GetFleetProxiesItemidResponse) (diags diag.Diagnostics) {
	if proxy == nil || proxy.JSON200 == nil {
		return
	}

	item := proxy.JSON200.Item

	// Set computed fields
	model.ID = types.StringValue(item.Id)
	model.ProxyID = types.StringValue(item.Id)

	// Set required fields
	model.Name = types.StringValue(item.Name)
	model.URL = types.StringValue(item.Url)

	// Set optional fields
	model.Certificate = types.StringPointerValue(item.Certificate)
	model.CertificateAuthorities = types.StringPointerValue(item.CertificateAuthorities)
	model.CertificateKey = types.StringPointerValue(item.CertificateKey)
	model.IsPreconfigured = types.BoolPointerValue(item.IsPreconfigured)

	// Note: space_ids is not returned in the API response, but it's managed through
	// the space-aware API endpoints. We'll preserve it from the plan/state.

	return
}

func (model proxyModel) toAPICreateModel(ctx context.Context) (kbapi.PostFleetProxiesJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostFleetProxiesJSONBody{
		Name: model.Name.ValueString(),
		Url:  model.URL.ValueString(),
	}

	// Optional: proxy_id
	if utils.IsKnown(model.ProxyID) {
		body.Id = model.ProxyID.ValueStringPointer()
	}

	// Optional: certificates
	if utils.IsKnown(model.Certificate) {
		body.Certificate = model.Certificate.ValueStringPointer()
	}
	if utils.IsKnown(model.CertificateAuthorities) {
		body.CertificateAuthorities = model.CertificateAuthorities.ValueStringPointer()
	}
	if utils.IsKnown(model.CertificateKey) {
		body.CertificateKey = model.CertificateKey.ValueStringPointer()
	}

	// Optional: is_preconfigured
	if utils.IsKnown(model.IsPreconfigured) {
		body.IsPreconfigured = model.IsPreconfigured.ValueBoolPointer()
	}

	return kbapi.PostFleetProxiesJSONRequestBody(body), diags
}

func (model proxyModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutFleetProxiesItemidJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PutFleetProxiesItemidJSONBody{}

	// All fields are optional in update
	if utils.IsKnown(model.Name) {
		body.Name = model.Name.ValueStringPointer()
	}
	if utils.IsKnown(model.URL) {
		body.Url = model.URL.ValueStringPointer()
	}

	// Optional: certificates
	if utils.IsKnown(model.Certificate) {
		body.Certificate = model.Certificate.ValueStringPointer()
	}
	if utils.IsKnown(model.CertificateAuthorities) {
		body.CertificateAuthorities = model.CertificateAuthorities.ValueStringPointer()
	}
	if utils.IsKnown(model.CertificateKey) {
		body.CertificateKey = model.CertificateKey.ValueStringPointer()
	}

	// Note: is_preconfigured is not supported by the Fleet Proxy Update API.
	// It can only be set during creation. Attempting to update it would require
	// ForceNew behavior (destroy and recreate), which is not appropriate for
	// this field since it's meant to mark externally-managed proxies.

	return kbapi.PutFleetProxiesItemidJSONRequestBody(body), diags
}
