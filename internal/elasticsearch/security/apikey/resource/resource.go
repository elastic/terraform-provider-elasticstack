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

package resource

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ fwresource.Resource                 = newResource()
	_ fwresource.ResourceWithConfigure    = newResource()
	_ fwresource.ResourceWithUpgradeState = newResource()
)

// Resource embeds ElasticsearchResource[apikey.TfModel] to inherit Configure,
// Metadata, Schema, Read, Update (via writeAPIKey), Delete, and PostRead
// cluster-version caching. Create is defined on the concrete type to bypass
// the envelope write path.
type Resource struct {
	*entitycore.ElasticsearchResource[apikey.TfModel]
}

func schemaFactory(_ context.Context) rschema.Schema {
	return getSchema(apikey.CurrentSchemaVersion)
}

func newResource() *Resource {
	placeholder := entitycore.PlaceholderElasticsearchWriteCallback[apikey.TfModel]()
	return &Resource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[apikey.TfModel]("security_api_key", entitycore.ElasticsearchResourceOptions[apikey.TfModel]{
			Schema:   schemaFactory,
			Read:     readAPIKey,
			Delete:   deleteAPIKey,
			Create:   placeholder,
			Update:   writeAPIKey,
			PostRead: postReadPersistClusterVersion,
		}),
	}
}

func NewResource() fwresource.Resource {
	return newResource()
}

// privateData mirrors the GetKey/SetKey surface of *privatestate.ProviderData
// so the envelope can hand a typed handle to PostRead without importing the
// framework's internal package. See the framework docs for full key semantics.
type privateData interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

const clusterVersionPrivateDataKey = "cluster-version"

type clusterVersionPrivateData struct {
	Version string
}

// saveClusterVersion persists the current Elasticsearch cluster version in
// private state so it can be retrieved on subsequent plan computations.
func saveClusterVersion(ctx context.Context, client *clients.ElasticsearchScopedClient, priv privateData) diag.Diagnostics {
	var diags diag.Diagnostics

	ver, verDiags := client.ServerVersion(ctx)
	diags.Append(verDiags...)
	if diags.HasError() {
		return diags
	}

	data, err := json.Marshal(clusterVersionPrivateData{Version: ver.String()})
	if err != nil {
		diags.AddError("failed to marshal cluster version data", err.Error())
		return diags
	}

	diags.Append(priv.SetKey(ctx, clusterVersionPrivateDataKey, data)...)
	return diags
}

func postReadPersistClusterVersion(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	_ apikey.TfModel,
	privateState any,
) diag.Diagnostics {
	priv, ok := privateState.(privateData)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError(
			"Elasticsearch envelope configuration error",
			"security_api_key PostRead requires private state implementing GetKey and SetKey.",
		)
		return diags
	}
	return saveClusterVersion(ctx, client, priv)
}

// clusterVersionOfLastRead retrieves the cached Elasticsearch cluster version.
func clusterVersionOfLastRead(ctx context.Context, priv privateData) (*version.Version, diag.Diagnostics) {
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
