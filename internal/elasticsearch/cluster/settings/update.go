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
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// updateClusterSettings implements the envelope's Update callback. It diffs
// the prior and planned settings to PUT both new values and explicit nulls for
// removed keys, then carries the composite ID forward from the prior state.
func updateClusterSettings(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	plan := req.Plan
	prior := *req.Prior

	oldSettings, g := getConfiguredSettings(ctx, prior)
	diags.Append(g...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	newSettings, g := getConfiguredSettings(ctx, plan)
	diags.Append(g...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	apiSettings := make(map[string]any)
	maps.Copy(apiSettings, newSettings)
	for _, category := range []string{"persistent", "transient"} {
		oldCat, _ := oldSettings[category].(map[string]any)
		newCat, _ := newSettings[category].(map[string]any)
		if oldCat == nil {
			oldCat = make(map[string]any)
		}
		if newCat == nil {
			newCat = make(map[string]any)
		}
		updateRemovedSettings(category, oldCat, newCat, apiSettings)
	}

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSettings(ctx, client, apiSettings))...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	plan.ID = prior.ID
	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}
