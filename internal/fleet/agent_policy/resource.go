package agent_policy

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &agentPolicyResource{}
	_ resource.ResourceWithConfigure   = &agentPolicyResource{}
	_ resource.ResourceWithImportState = &agentPolicyResource{}
)

var (
	MinVersionGlobalDataTags      = version.Must(version.NewVersion("8.15.0"))
	MinSupportsAgentlessVersion   = version.Must(version.NewVersion("8.15.0"))
	MinVersionInactivityTimeout   = version.Must(version.NewVersion("8.7.0"))
	MinVersionUnenrollmentTimeout = version.Must(version.NewVersion("8.15.0"))
	MinVersionSpaceIds            = version.Must(version.NewVersion("9.1.0"))
	MinVersionRequiredVersions    = version.Must(version.NewVersion("9.1.0"))
	MinVersionAgentFeatures       = version.Must(version.NewVersion("8.7.0"))
	MinVersionAdvancedMonitoring  = version.Must(version.NewVersion("8.16.0"))
	MinVersionAdvancedSettings    = version.Must(version.NewVersion("8.17.0"))
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &agentPolicyResource{}
}

type agentPolicyResource struct {
	client *clients.ApiClient
}

func (r *agentPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *agentPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_agent_policy")
}

func (r *agentPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("policy_id"), req, resp)
}

func (r *agentPolicyResource) buildFeatures(ctx context.Context) (features, diag.Diagnostics) {
	supportsGDT, diags := r.client.EnforceMinVersion(ctx, MinVersionGlobalDataTags)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsSupportsAgentless, diags := r.client.EnforceMinVersion(ctx, MinSupportsAgentlessVersion)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsInactivityTimeout, diags := r.client.EnforceMinVersion(ctx, MinVersionInactivityTimeout)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsUnenrollmentTimeout, diags := r.client.EnforceMinVersion(ctx, MinVersionUnenrollmentTimeout)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsSpaceIds, diags := r.client.EnforceMinVersion(ctx, MinVersionSpaceIds)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsRequiredVersions, diags := r.client.EnforceMinVersion(ctx, MinVersionRequiredVersions)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsAgentFeatures, diags := r.client.EnforceMinVersion(ctx, MinVersionAgentFeatures)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsAdvancedMonitoring, diags := r.client.EnforceMinVersion(ctx, MinVersionAdvancedMonitoring)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	return features{
		SupportsGlobalDataTags:      supportsGDT,
		SupportsSupportsAgentless:   supportsSupportsAgentless,
		SupportsInactivityTimeout:   supportsInactivityTimeout,
		SupportsUnenrollmentTimeout: supportsUnenrollmentTimeout,
		SupportsSpaceIds:            supportsSpaceIds,
		SupportsRequiredVersions:    supportsRequiredVersions,
		SupportsAgentFeatures:       supportsAgentFeatures,
		SupportsAdvancedMonitoring:  supportsAdvancedMonitoring,
	}, nil
}
