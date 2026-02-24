// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package apikey

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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
	MinVersionWithRestriction          = version.Must(version.NewVersion("8.9.0"))  // Enabled in 8.0
	MinVersionWithCrossCluster         = version.Must(version.NewVersion("8.10.0")) // Cross-cluster API keys enabled in 8.10
)

type Resource struct {
	client *clients.APIClient
}

var configuredResources = []*Resource{}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
	configuredResources = append(configuredResources, r)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
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

func (r *Resource) saveClusterVersion(ctx context.Context, client *clients.APIClient, priv privateData) diag.Diagnostics {
	version, sdkDiags := client.ServerVersion(ctx)
	diags := diagutil.FrameworkDiagsFromSDK(sdkDiags)
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
