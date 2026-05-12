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

// deleteClusterSettings is the deleteFunc callback for the entitycore envelope.
// It nulls out all tracked settings so Elasticsearch reverts them to defaults.
func deleteClusterSettings(ctx context.Context, client *clients.ElasticsearchScopedClient, _ string, state tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	configuredSettings, ds := getConfiguredSettings(ctx, state)
	diags.Append(ds...)
	if diags.HasError() {
		return diags
	}

	pSettings := make(map[string]any)
	if v := configuredSettings["persistent"]; v != nil {
		for k := range v.(map[string]any) {
			pSettings[k] = nil
		}
	}

	tSettings := make(map[string]any)
	if v := configuredSettings["transient"]; v != nil {
		for k := range v.(map[string]any) {
			tSettings[k] = nil
		}
	}

	apiSettings := map[string]any{
		"persistent": pSettings,
		"transient":  tSettings,
	}

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSettings(ctx, client, apiSettings))...)
	return diags
}
