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

package indexsettings

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateIndexSettings(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	indexName := plan.Index.ValueString()

	planSettings, planDiags := declaredSettingsPayload(plan)
	diags.Append(planDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	priorSettings, priorDiags := declaredSettingsPayload(*req.Prior)
	diags.Append(priorDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	// Canonical key spellings and JSON-encoded values make the diff
	// independent of `index.` prefix differences and of the numeric types
	// produced by typed attributes vs. settings_json values.
	planCanonical := canonicalSettingsMap(planSettings)
	priorCanonical := canonicalSettingsMap(priorSettings)

	if reflect.DeepEqual(planCanonical, priorCanonical) {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	// Settings removed from the declaration must be sent as null so
	// Elasticsearch resets them to their default, mirroring
	// elasticstack_elasticsearch_index's updateSettings behavior.
	for key := range priorSettings {
		if _, ok := planCanonical[normalizeSettingsKey(key)]; !ok {
			planSettings[key] = nil
		}
	}

	diags.Append(elasticsearch.UpdateIndexSettings(ctx, client, indexName, planSettings)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}

func canonicalSettingsMap(m map[string]any) map[string]string {
	canonical := make(map[string]string, len(m))
	for key, value := range m {
		encoded, _ := json.Marshal(value)
		canonical[normalizeSettingsKey(key)] = string(encoded)
	}
	return canonical
}
