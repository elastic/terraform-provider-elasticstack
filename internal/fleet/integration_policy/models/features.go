package models

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	MinVersionPolicyIds = version.Must(version.NewVersion("8.15.0"))
	MinVersionOutputId  = version.Must(version.NewVersion("8.16.0"))
)

type Features struct {
	SupportsPolicyIds bool
	SupportsOutputId  bool
}

func NewFeatures(ctx context.Context, client *clients.ApiClient) (Features, diag.Diagnostics) {
	supportsPolicyIds, diags := client.EnforceMinVersion(ctx, MinVersionPolicyIds)
	if diags.HasError() {
		return Features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsOutputId, outputIdDiags := client.EnforceMinVersion(ctx, MinVersionOutputId)
	if outputIdDiags.HasError() {
		return Features{}, diagutil.FrameworkDiagsFromSDK(outputIdDiags)
	}

	return Features{
		SupportsPolicyIds: supportsPolicyIds,
		SupportsOutputId:  supportsOutputId,
	}, nil
}
