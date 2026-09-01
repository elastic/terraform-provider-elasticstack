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

package lenscommon

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// PopulateJSONWithDefaultsSlice marshals each element of items and wraps it as a
// customtypes.JSONWithDefaultsValue via populateDefaults, writing the result into a new []TFModel
// through configOf. When prior is non-empty, the index-aligned entry is compared via
// panelkit.PreservePriorJSONWithDefaultsIfEquivalent so state doesn't churn when Kibana echoes
// back semantically-unchanged defaults; pass a nil prior to skip that comparison entirely. A
// marshal failure on one item is recorded on diags and leaves that index's config at its zero
// value, matching the per-package loops this replaces.
func PopulateJSONWithDefaultsSlice[APIItem, TFModel, TModel any](
	ctx context.Context,
	items []APIItem,
	prior []TFModel,
	configOf func(*TFModel) *customtypes.JSONWithDefaultsValue[TModel],
	populateDefaults customtypes.PopulateDefaultsFunc[TModel],
	fieldName string,
	diags *diag.Diagnostics,
) []TFModel {
	dest := make([]TFModel, len(items))
	for i, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			diags.AddError("Failed to marshal "+fieldName, err.Error())
			continue
		}
		cfg := customtypes.NewJSONWithDefaultsValue(string(b), populateDefaults)
		if i < len(prior) {
			cfg = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, *configOf(&prior[i]), cfg, diags)
		}
		*configOf(&dest[i]) = cfg
	}
	return dest
}

// PopulateNormalizedJSONSlice marshals each element of items and wraps it as a jsontypes.Normalized
// via WrapNormalizedJSON, writing the result into a new []TFModel through configOf. It returns
// (nil, false) as soon as a marshal/wrap failure occurs, matching the early-return-on-error
// behavior of the panel FromAPI converters this replaces.
func PopulateNormalizedJSONSlice[APIItem, TFModel any](
	items []APIItem,
	configOf func(*TFModel) *jsontypes.Normalized,
	fieldName string,
	diags *diag.Diagnostics,
) ([]TFModel, bool) {
	dest := make([]TFModel, len(items))
	for i, item := range items {
		b, err := json.Marshal(item)
		v, ok := WrapNormalizedJSON(b, err, fieldName, diags)
		if !ok {
			return nil, false
		}
		*configOf(&dest[i]) = v
	}
	return dest, true
}
