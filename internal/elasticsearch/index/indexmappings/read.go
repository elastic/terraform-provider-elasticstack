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

package indexmappings

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readIndexMappings(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state tfModel) (tfModel, bool, diag.Diagnostics) {
	indexName := resourceID

	apiIndex, diags := elasticsearch.GetIndex(ctx, client, indexName)
	if diags.HasError() {
		return state, false, diags
	}
	if apiIndex == nil {
		return state, false, nil
	}

	mappingsValue, mapDiags := mappingsFromAPI(apiIndex.Mappings)
	if mapDiags.HasError() {
		return state, false, mapDiags
	}

	if priorMappingsEmpty(state.Mappings) {
		state.Mappings = mappingsValue
		return state, true, nil
	}

	var apiMap map[string]any
	if err := json.Unmarshal([]byte(mappingsValue.ValueString()), &apiMap); err != nil {
		return state, false, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to unmarshal API mappings", err.Error()),
		}
	}

	var stateMap map[string]any
	if err := json.Unmarshal([]byte(state.Mappings.ValueString()), &stateMap); err != nil {
		return state, false, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to unmarshal state mappings", err.Error()),
		}
	}

	intersected := intersectMappings(apiMap, stateMap)
	intersectedBytes, err := json.Marshal(intersected)
	if err != nil {
		return state, false, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal intersected mappings", err.Error()),
		}
	}

	state.Mappings = index.NewMappingsValue(string(intersectedBytes))
	return state, true, nil
}

func mappingsFromAPI(apiMappings any) (index.MappingsValue, diag.Diagnostics) {
	if apiMappings == nil {
		return index.NewMappingsValue("{}"), nil
	}

	mappingBytes, err := json.Marshal(apiMappings)
	if err != nil {
		return index.MappingsValue{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal index mappings from API", err.Error()),
		}
	}
	return index.NewMappingsValue(string(mappingBytes)), nil
}

func priorMappingsEmpty(mappings index.MappingsValue) bool {
	if mappings.IsNull() || mappings.IsUnknown() {
		return true
	}
	trimmed := strings.TrimSpace(mappings.ValueString())
	return trimmed == "" || trimmed == "{}"
}
