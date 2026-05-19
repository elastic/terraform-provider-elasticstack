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

package lensdashboardapp

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func defaultOpaqueRootJSON(v any) any { return v }

func populateLensDashboardAppFromAPI(
	ctx context.Context,
	dashboard *models.DashboardModel,
	pm *models.PanelModel,
	tfPanel *models.PanelModel,
	api kbapi.KbnDashboardPanelTypeLensDashboardApp,
) diag.Diagnostics {
	var diags diag.Diagnostics
	prior := configPriorForLensRead(tfPanel, pm)

	configBytes, err := api.Config.MarshalJSON()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var root map[string]any
	if err := json.Unmarshal(configBytes, &root); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch classifyLensDashboardAppConfigFromRoot(root) {
	case lensConfigClassByValueChart:
		return populateLensDashboardAppByValueFromAPI(ctx, dashboard, prior, configBytes, pm)
	case lensConfigClassByReference:
		cfg1, err1 := api.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
		if err1 != nil {
			diags.AddError("Invalid lens-dashboard-app config on read", err1.Error())
			return diags
		}
		return populateLensDashboardAppByReferenceFromAPI(ctx, prior, pm, cfg1)
	default:
		if prior != nil && prior.ByReference != nil {
			return diags
		}
		return populateLensDashboardAppByValueFromAPI(ctx, dashboard, prior, configBytes, pm)
	}
}

func populateLensDashboardAppByReferenceFromAPI(
	ctx context.Context,
	prior *models.LensDashboardAppConfigModel,
	pm *models.PanelModel,
	cfg1 kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1,
) diag.Diagnostics {
	var priorBR *models.LensDashboardAppByReferenceModel
	if prior != nil {
		priorBR = prior.ByReference
	}
	by, diags := lenscommon.PopulateLensByReferenceTFModelFromLensAppConfig1(ctx, cfg1, priorBR)
	if diags.HasError() {
		return diags
	}
	pm.LensDashboardAppConfig = &models.LensDashboardAppConfigModel{
		ByReference: &by,
	}
	return diags
}

func preservePriorLensByValueConfigJSON(
	ctx context.Context,
	prior, fromAPI jsontypes.Normalized,
	diags *diag.Diagnostics,
) jsontypes.Normalized {
	after := panelkit.PreservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior, fromAPI, defaultOpaqueRootJSON, diags)
	embedded, err := JSONValuePriorEmbeddedInExpandedCurrent(prior.ValueString(), fromAPI.ValueString())
	if err != nil {
		return after
	}
	if embedded {
		return prior
	}
	return after
}

func JSONValuePriorEmbeddedInExpandedCurrent(priorJSON, currentJSON string) (bool, error) {
	var priorObj map[string]any
	if err := json.Unmarshal([]byte(priorJSON), &priorObj); err != nil {
		return false, err
	}
	if !hasLensByValueChartTypeAtRoot(priorObj) {
		return false, nil
	}
	var currentObj map[string]any
	if err := json.Unmarshal([]byte(currentJSON), &currentObj); err != nil {
		return false, err
	}
	return jsonValueSubsumedByCurrentObject(priorObj, currentObj, true), nil
}

func isEmptyJSONSlice(prior any) bool {
	if prior == nil {
		return true
	}
	if pArr, ok := prior.([]any); ok && len(pArr) == 0 {
		return true
	}
	return false
}

func isEmptyJSONMap(prior any) bool {
	if prior == nil {
		return true
	}
	if pMap, ok := prior.(map[string]any); ok && len(pMap) == 0 {
		return true
	}
	return false
}

func isOmissibleDefaultKqlQuery(m map[string]any) bool {
	if len(m) == 0 {
		return true
	}
	lang, hasLang := m["language"]
	expr, hasExpr := m["expression"]
	switch {
	case hasLang && lang == "kql" && !hasExpr && len(m) == 1:
		return true
	case hasLang && lang == "kql" && hasExpr && expr == "" && len(m) == 2:
		return true
	default:
		return false
	}
}

func jsonValueSubsumedByCurrentObject(prior, current map[string]any, isRoot bool) bool {
	for k, pv := range prior {
		if isRoot && k == "styling" {
			continue
		}
		cv, ok := current[k]
		if !ok {
			if isEmptyJSONSlice(pv) || isEmptyJSONMap(pv) {
				continue
			}
			if s, y := pv.(string); y && s == "" {
				continue
			}
			if k == "query" {
				if qm, y := pv.(map[string]any); y && isOmissibleDefaultKqlQuery(qm) {
					continue
				}
			}
			return false
		}
		if isEmptyJSONSlice(pv) {
			if isEmptyJSONSlice(cv) {
				continue
			}
			return false
		}
		if !jsonValueSubsumedByCurrentAny(pv, cv) {
			return false
		}
	}
	return true
}

func jsonValueSubsumedByCurrentAny(prior, current any) bool {
	switch p := prior.(type) {
	case nil:
		return current == nil
	case bool:
		c, ok := current.(bool)
		return ok && c == p
	case float64:
		c, ok := current.(float64)
		return ok && c == p
	case string:
		c, ok := current.(string)
		return ok && c == p
	case []any:
		if isEmptyJSONSlice(prior) && (current == nil) {
			return true
		}
		c, ok := current.([]any)
		if !ok {
			return false
		}
		if len(p) == 0 {
			return len(c) == 0
		}
		if len(p) > len(c) {
			return false
		}
		for i := range p {
			if !jsonValueSubsumedByCurrentAny(p[i], c[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		c, ok := current.(map[string]any)
		if !ok {
			return false
		}
		return jsonValueSubsumedByCurrentObject(p, c, false)
	default:
		return false
	}
}

func populateLensDashboardAppByValueFromAPI(
	ctx context.Context,
	dashboard *models.DashboardModel,
	prior *models.LensDashboardAppConfigModel,
	configBytes []byte,
	pm *models.PanelModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	norm, okNorm := lenscommon.MarshalToNormalized(configBytes, nil, "by_value.config_json", &diags)

	if prior != nil && prior.ByValue != nil && !lensByValueModelHasAnyTypedChartBlock(prior.ByValue) {
		if okNorm {
			if typeutils.IsKnown(prior.ByValue.ConfigJSON) {
				norm = preservePriorLensByValueConfigJSON(ctx, prior.ByValue.ConfigJSON, norm, &diags)
			}
			pm.LensDashboardAppConfig = &models.LensDashboardAppConfigModel{
				ByValue: &models.LensDashboardAppByValueModel{ConfigJSON: norm},
			}
		}
		return diags
	}

	if tryPopulateTypedLensByValueFromAPI(ctx, dashboard, prior, configBytes, pm, &diags) {
		return diags
	}

	if okNorm {
		if prior != nil && prior.ByValue != nil && typeutils.IsKnown(prior.ByValue.ConfigJSON) {
			norm = preservePriorLensByValueConfigJSON(ctx, prior.ByValue.ConfigJSON, norm, &diags)
		}
		pm.LensDashboardAppConfig = &models.LensDashboardAppConfigModel{
			ByValue: &models.LensDashboardAppByValueModel{ConfigJSON: norm},
		}
	}
	return diags
}
