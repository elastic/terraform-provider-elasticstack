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

package settings

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// readClusterSettings is the readFunc callback for the entitycore envelope.
// It retrieves cluster settings from Elasticsearch and populates only the keys
// that are tracked in the current Terraform state.
func readClusterSettings(ctx context.Context, client *clients.ElasticsearchScopedClient, _ string, state tfModel) (tfModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	clusterSettings, sdkDiags := elasticsearch.GetSettings(ctx, client)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	configuredSettings, ds := getConfiguredSettings(ctx, state)
	diags.Append(ds...)
	if diags.HasError() {
		return state, false, diags
	}

	persistent, ds := flattenSettings(ctx, "persistent", configuredSettings, clusterSettings)
	diags.Append(ds...)
	if diags.HasError() {
		return state, false, diags
	}

	transient, ds := flattenSettings(ctx, "transient", configuredSettings, clusterSettings)
	diags.Append(ds...)
	if diags.HasError() {
		return state, false, diags
	}

	result := state
	result.Persistent = persistent
	result.Transient = transient

	return result, true, diags
}
