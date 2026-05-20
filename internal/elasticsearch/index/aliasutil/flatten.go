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

package aliasutil

import (
	"encoding/json"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NormalizeAliasFilterMap applies NormalizeQueryFilter to an already-decoded map
// and serializes the result to a jsontypes.Normalized.
// Returns jsontypes.NewNormalizedNull() when filterMap is empty or nil.
func NormalizeAliasFilterMap(filterMap map[string]any) (jsontypes.Normalized, diag.Diagnostics) {
	if len(filterMap) == 0 {
		return jsontypes.NewNormalizedNull(), nil
	}
	if nm, ok := elasticsearch.NormalizeQueryFilter(filterMap).(map[string]any); ok {
		filterMap = nm
	}
	normalizedBytes, err := json.Marshal(filterMap)
	if err != nil {
		return jsontypes.NewNormalizedNull(), diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
		}
	}
	return jsontypes.NewNormalizedValue(string(normalizedBytes)), nil
}

// NormalizeAliasFilterAnyToMap marshals v to JSON, unmarshals to map[string]any,
// then applies NormalizeQueryFilter. Returns (nil, nil) when v is nil.
func NormalizeAliasFilterAnyToMap(v any) (map[string]any, diag.Diagnostics) {
	if v == nil {
		return nil, nil
	}
	filterBytes, err := json.Marshal(v)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
		}
	}
	var filterMap map[string]any
	if err := json.Unmarshal(filterBytes, &filterMap); err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to unmarshal alias filter", err.Error()),
		}
	}
	if nm, ok := elasticsearch.NormalizeQueryFilter(filterMap).(map[string]any); ok {
		return nm, nil
	}
	return filterMap, nil
}

// NormalizeAliasFilterFromAny serializes v to a canonical jsontypes.Normalized alias filter.
// Returns jsontypes.NewNormalizedNull() when v is nil.
func NormalizeAliasFilterFromAny(v any) (jsontypes.Normalized, diag.Diagnostics) {
	if v == nil {
		return jsontypes.NewNormalizedNull(), nil
	}
	filterMap, diags := NormalizeAliasFilterAnyToMap(v)
	if diags.HasError() {
		return jsontypes.NewNormalizedNull(), diags
	}
	return NormalizeAliasFilterMap(filterMap)
}

// NewAliasModelFromAPI constructs an AliasModel from an API Alias response.
func NewAliasModelFromAPI(name string, apiModel estypes.Alias) (AliasModel, diag.Diagnostics) {
	tfAlias := AliasModel{
		Name:          types.StringValue(name),
		IndexRouting:  types.StringValue(typeutils.Deref(apiModel.IndexRouting)),
		IsHidden:      types.BoolValue(typeutils.Deref(apiModel.IsHidden)),
		IsWriteIndex:  types.BoolValue(typeutils.Deref(apiModel.IsWriteIndex)),
		Routing:       types.StringValue(typeutils.Deref(apiModel.Routing)),
		SearchRouting: types.StringValue(typeutils.Deref(apiModel.SearchRouting)),
	}

	filter, diags := NormalizeAliasFilterFromAny(apiModel.Filter)
	if diags.HasError() {
		return AliasModel{}, diags
	}
	tfAlias.Filter = filter

	return tfAlias, nil
}
