package api_key

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithUpgradeState = &Resource{}
var (
	MinVersion                         = version.Must(version.NewVersion("8.0.0")) // Enabled in 8.0
	MinVersionWithUpdate               = version.Must(version.NewVersion("8.4.0"))
	MinVersionReturningRoleDescriptors = version.Must(version.NewVersion("8.5.0"))
	MinVersionWithRestriction          = version.Must(version.NewVersion("8.9.0")) // Enabled in 8.0
)

type Resource struct {
	client *clients.ApiClient
}

var configuredResources = []*Resource{}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
	configuredResources = append(configuredResources, r)
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_elasticsearch_security_api_key"
}

// Equivalent to privatestate.ProviderData
type privateData interface {
	// GetKey returns the private state data associated with the given key.
	//
	// If the key is reserved for framework usage, an error diagnostic
	// is returned. If the key is valid, but private state data is not found,
	// nil is returned.
	//
	// The naming of keys only matters in context of a single resource,
	// however care should be taken that any historical keys are not reused
	// without accounting for older resource instances that may still have
	// older data at the key.
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)

	// SetKey sets the private state data at the given key.
	//
	// If the key is reserved for framework usage, an error diagnostic
	// is returned. The data must be valid JSON and UTF-8 safe or an error
	// diagnostic is returned.
	//
	// The naming of keys only matters in context of a single resource,
	// however care should be taken that any historical keys are not reused
	// without accounting for older resource instances that may still have
	// older data at the key.
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

const clusterVersionPrivateDataKey = "cluster-version"

type clusterVersionPrivateData struct {
	Version string
}

func (r *Resource) saveClusterVersion(ctx context.Context, client *clients.ApiClient, priv privateData) diag.Diagnostics {
	version, sdkDiags := client.ServerVersion(ctx)
	diags := utils.FrameworkDiagsFromSDK(sdkDiags)
	if diags.HasError() {
		return diags
	}

	data, err := json.Marshal(clusterVersionPrivateData{Version: version.String()})
	if err != nil {
		diags.AddError("failed to marshal cluster version data", err.Error())
		return diags
	}

	diags.Append(priv.SetKey(ctx, clusterVersionPrivateDataKey, data)...)
	return diags
}

func (r *Resource) clusterVersionOfLastRead(ctx context.Context, priv privateData) (*version.Version, diag.Diagnostics) {
	versionBytes, diags := priv.GetKey(ctx, clusterVersionPrivateDataKey)
	if diags.HasError() {
		return nil, diags
	}

	if versionBytes == nil {
		return nil, nil
	}

	var data clusterVersionPrivateData
	err := json.Unmarshal(versionBytes, &data)
	if err != nil {
		diags.AddError("failed to parse private data json", err.Error())
		return nil, diags
	}

	v, err := version.NewVersion(data.Version)
	if err != nil {
		diags.AddError("failed to parse stored cluster version", err.Error())
		return nil, diags
	}

	return v, diags
}
