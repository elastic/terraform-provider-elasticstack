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

package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// populateFiltersFromAPI converts a slice of kbapi.LensPanelFilters_Item into models.ChartFilterJSONModel
// values, appending any errors to diags.
func populateFiltersFromAPI(filters []kbapi.LensPanelFilters_Item, diags *diag.Diagnostics) []models.ChartFilterJSONModel {
	return lenscommon.PopulateFiltersFromAPI(filters, diags)
}

// buildFiltersForAPI converts the model filter slice into the kbapi type, appending errors to diags.
// The returned slice is always non-nil (empty API payload is []kbapi.LensPanelFilters_Item{}).
func buildFiltersForAPI(filters []models.ChartFilterJSONModel, diags *diag.Diagnostics) []kbapi.LensPanelFilters_Item {
	return lenscommon.BuildFiltersForAPI(filters, diags)
}

// marshalToNormalized delegates to lenscommon.MarshalToNormalized (canonical implementation).
func marshalToNormalized(bytes []byte, err error, fieldName string, diags *diag.Diagnostics) (jsontypes.Normalized, bool) {
	return lenscommon.MarshalToNormalized(bytes, err, fieldName, diags)
}

// preservePriorNormalizedWithDefaultsIfEquivalent delegates to panelkit.PreservePriorNormalizedWithDefaultsIfEquivalent.
func preservePriorNormalizedWithDefaultsIfEquivalent[T any](ctx context.Context, prior, current jsontypes.Normalized, defaults func(T) T, diags *diag.Diagnostics) jsontypes.Normalized {
	return panelkit.PreservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior, current, defaults, diags)
}

// marshalToJSONWithDefaults stores the already-marshaled bytes as a JSONWithDefaultsValue,
// or adds an error to diags and returns (zero, false) on failure.
func marshalToJSONWithDefaults[T any](bytes []byte, err error, fieldName string, defaults func(T) T, diags *diag.Diagnostics) (customtypes.JSONWithDefaultsValue[T], bool) {
	return lenscommon.MarshalToJSONWithDefaults(bytes, err, fieldName, defaults, diags)
}

func preservePriorJSONWithDefaultsIfEquivalent[T any](ctx context.Context, prior, current customtypes.JSONWithDefaultsValue[T], diags *diag.Diagnostics) customtypes.JSONWithDefaultsValue[T] {
	return panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, prior, current, diags)
}

// lensESQLNumberFormatJSONFromAPI marshals a Lens ES|QL dimension `format` union
// value to a normalized Terraform string. Empty or null JSON is replaced with the
// default number-format payload so Terraform state matches what Kibana echoes.
func lensESQLNumberFormatJSONFromAPI(format any, errLabel string, diags *diag.Diagnostics) (jsontypes.Normalized, bool) {
	return lenscommon.LensESQLNumberFormatJSONFromAPI(format, errLabel, diags)
}

// lensQueryESQLMode returns whether a Lens chart's optional `query` object selects
// ES|QL mode (i.e. `query` is omitted, or both `expression` and `language` are
// null). ok is false when the configuration is still unknown and validation
// should defer.
func lensQueryESQLMode(ctx context.Context, config tfsdk.Config, attrPath path.Path, diags *diag.Diagnostics) (esqlMode bool, ok bool) {
	var queryObj types.Object
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query"), &queryObj)...)
	if diags.HasError() {
		return false, false
	}
	if queryObj.IsUnknown() {
		return false, false
	}
	if queryObj.IsNull() {
		return true, true
	}

	var lang, expr types.String
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query").AtName("language"), &lang)...)
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query").AtName("expression"), &expr)...)
	if diags.HasError() {
		return false, false
	}
	return lang.IsNull() && expr.IsNull(), true
}
