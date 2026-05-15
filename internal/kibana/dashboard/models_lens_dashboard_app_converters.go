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
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func defaultOpaqueRootJSON(v any) any { return v }

// configPriorForLensRead returns the last known lens_dashboard_app_config from
// plan/state. tfPanel is the panel at the same index from the prior model (nil on
// import or when a panel is new); it is the authoritative source when present.
func configPriorForLensRead(tfPanel, pm *models.PanelModel) *models.LensDashboardAppConfigModel {
	if tfPanel != nil && tfPanel.LensDashboardAppConfig != nil {
		return tfPanel.LensDashboardAppConfig
	}
	if pm != nil {
		return pm.LensDashboardAppConfig
	}
	return nil
}

// lensConfigClass identifies how a lens-dashboard-app panel `config` JSON should be
// represented in Terraform after read, using raw shape before trusting the
// generated union helper alone (see AsKbnDashboardPanelTypeLensDashboardAppConfig1,
// which is json.Unmarshal, not a oneOf discriminant in the real wire payload).
type lensConfigClass int

const (
	// The payload has a non-empty string at top-level "type" — by-value inline chart
	// config from the Kibana Lens chart union. Such configs may also include
	// time_range, references, and ref_id for chart needs; we still treat the panel
	// as by_value (REQ-035 / design: chart discriminator wins over ref_id-only cues).
	lensConfigClassByValueChart lensConfigClass = iota
	// The payload is missing a chart "type" at the root, and has ref_id with a
	// time_range object with non-empty from/to — the by-reference (Config1) shape.
	lensConfigClassByReference
	// Neither a chart payload nor a complete by-reference shape (e.g. incomplete
	// or unexpected JSON). The caller may preserve prior by_reference state instead
	// of falling back to by_value (see populateLensDashboardAppFromAPI).
	lensConfigClassAmbiguous
)

// classifyLensDashboardAppConfigFromRoot classifies the raw API config object. It must
// not be the only check for by-reference, because unmarshaling the generated Config1
// struct alone would accept mixed keys from by-value and by-reference wire shapes.
func classifyLensDashboardAppConfigFromRoot(root map[string]any) lensConfigClass {
	if hasLensByValueChartTypeAtRoot(root) {
		return lensConfigClassByValueChart
	}
	if hasLensByReferenceShapeAtRoot(root) {
		return lensConfigClassByReference
	}
	return lensConfigClassAmbiguous
}

func hasLensByValueChartTypeAtRoot(m map[string]any) bool {
	if m == nil {
		return false
	}
	v, ok := m["type"]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s != ""
}

func hasLensByReferenceShapeAtRoot(m map[string]any) bool {
	if m == nil {
		return false
	}
	ref, ok := m["ref_id"]
	if !ok {
		return false
	}
	refS, ok := ref.(string)
	if !ok || refS == "" {
		return false
	}
	trAny, ok := m["time_range"]
	if !ok {
		return false
	}
	tr, ok := trAny.(map[string]any)
	if !ok {
		return false
	}
	from, fOK := tr["from"].(string)
	to, tOK := tr["to"].(string)
	return fOK && tOK && from != "" && to != ""
}

// lensDashboardAPIGrid is the wire shape used by dashboard panel toAPI.
type lensDashboardAPIGrid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}

// lensDashboardAppToAPI converts a lens-dashboard-app panel to the Kibana API model.
// parentDashboard is the enclosing dashboard resource model (required for typed by_value charts
// so chart roots resolve time_range from dashboard-level defaults; by_reference does not use it).
func lensDashboardAppToAPI(pm models.PanelModel, grid lensDashboardAPIGrid, panelID *string, parentDashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.LensDashboardAppConfig
	if cfg == nil {
		diags.AddError("Missing `lens_dashboard_app_config`", "The `lens_dashboard_app_config` block is required for `lens-dashboard-app` panels.")
		return kbapi.DashboardPanelItem{}, diags
	}
	switch {
	case cfg.ByValue != nil:
		return lensDashboardAppByValueToAPI(*cfg.ByValue, grid, panelID, parentDashboard)
	case cfg.ByReference != nil:
		return lensDashboardAppByReferenceToAPI(*cfg.ByReference, grid, panelID)
	default:
		diags.AddError("Invalid `lens_dashboard_app_config`", "Exactly one of `by_value` or `by_reference` must be set inside `lens_dashboard_app_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}
}

// parentDashboard is the enclosing dashboard resource model when converting typed by_value charts.
// Callers that build payloads in isolation (e.g. unit tests) pass nil; models.PanelModel.toAPI passes
// the real enclosing dashboard model.
func lensDashboardAppByValueToAPI(
	byValue models.LensDashboardAppByValueModel,
	grid lensDashboardAPIGrid,
	panelID *string,
	parentDashboard *models.DashboardModel,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	if blocks, ok := lensByValueChartBlocksForTypedLensApp(byValue); ok {
		conv, okConv := lenscommon.FirstForBlocks(blocks)
		if !okConv {
			diags.AddError("Invalid `by_value` for lens-dashboard-app", "The typed by-value chart block could not be resolved to a Lens visualization converter.")
			return kbapi.DashboardPanelItem{}, diags
		}
		vis0, d := conv.BuildAttributes(blocks, lensChartResolver(parentDashboard))
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		config, err := lensByValueConfigFromVisConfig0(vis0)
		if err != nil {
			diags.AddError("Invalid typed by-value config for lens-dashboard-app", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
			Config: config,
			Grid: kbapi.KbnDashboardPanelGrid{
				H: grid.H,
				W: grid.W,
				X: grid.X,
				Y: grid.Y,
			},
			Id: panelID,
		}
		var panelItem kbapi.DashboardPanelItem
		if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
			diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
		}
		return panelItem, diags
	}
	if !typeutils.IsKnown(byValue.ConfigJSON) {
		diags.AddError(
			"Invalid `by_value.config_json` for lens-dashboard-app",
			"by_value.config_json is unknown. Ensure it is set to a non-null JSON value when using `config_json` as the by-value source.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}
	var config kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	if err := config.UnmarshalJSON([]byte(byValue.ConfigJSON.ValueString())); err != nil {
		diags.AddError("Invalid `by_value.config_json` for lens-dashboard-app", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
		Config: config,
		Grid: kbapi.KbnDashboardPanelGrid{
			H: grid.H,
			W: grid.W,
			X: grid.X,
			Y: grid.Y,
		},
		Id: panelID,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
		diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
	}
	return panelItem, diags
}

func lensDashboardAppByReferenceToAPI(
	byRef models.LensDashboardAppByReferenceModel,
	grid lensDashboardAPIGrid,
	panelID *string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	api1, d := lenscommon.LensDashboardAppByReferenceModelToAPIConfig1(byRef, "by_reference.references_json")
	diags.Append(d...)
	if d.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}
	var config kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	if err := config.FromKbnDashboardPanelTypeLensDashboardAppConfig1(api1); err != nil {
		diags.AddError("Failed to set lens-dashboard-app by_reference config", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
		Config: config,
		Grid: kbapi.KbnDashboardPanelGrid{
			H: grid.H,
			W: grid.W,
			X: grid.X,
			Y: grid.Y,
		},
		Id: panelID,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
		diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
	}
	return panelItem, diags
}

// populateLensDashboardAppFromAPI maps an API lens-dashboard-app panel to the TF model in pm.
// pm is seeded from the prior plan/state when tfPanel is set; tfPanel is the panel at
// the same index from the prior model (used for prior lens mode and optional seeding).
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
	// REQ-035: align read mode with the wire object: generated AsConfig1 is
	// Unmarshal, not a real JSON oneOf, so we inspect the raw object first. Inline
	// by-value chart configs from the Kibana union carry a string `type` at the root
	// (e.g. metricNoESQL). By-reference Config1 has ref_id and time_range but not
	// that chart discriminator; prefer by-reference only when the chart
	// discriminator is absent and the ref_id + time_range shape is present.
	//
	// When classification is ambiguous and the practitioner previously used
	// by_reference, we do not fall back to by_value: that would silently flip modes
	// (REQ-009-style preservation). The stronger type/ref heuristic above makes
	// ref_id+time_range without a root chart `type` resolve to by_reference; the
	// ambiguous case is for incomplete/odd payloads where we keep prior by_reference.
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
	default: // lensConfigClassAmbiguous
		if prior != nil && prior.ByReference != nil {
			// Avoid silently switching a prior by_reference panel to by_value when
			// the response is not clearly a by-value chart and not a full by-reference shape.
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

// preservePriorLensByValueConfigJSON returns practitioner `by_value.config_json` in state
// when the API read would otherwise diverge (REQ-035): (1) full semantic JSON equality, or
// (2) the API `config` is a strict value-expansion of the prior object — same value at
// every key path the practitioner set, with extra keys/entries allowed in the response
// (Kibana re-keying, defaults, ordering). (2) applies only when the prior object has a
// non-empty string top-level `type` (the by-value chart discriminator) so we never
// vacuously match `{}` against a full chart. Empty prior arrays with non-empty API arrays
// are not treated as embedded (would risk wiping list-shaped fields on the next write).
func preservePriorLensByValueConfigJSON(
	ctx context.Context,
	prior, fromAPI jsontypes.Normalized,
	diags *diag.Diagnostics,
) jsontypes.Normalized {
	after := preservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior, fromAPI, defaultOpaqueRootJSON, diags)
	embedded, err := jsonValuePriorEmbeddedInExpandedCurrent(prior.ValueString(), fromAPI.ValueString())
	if err != nil {
		return after
	}
	if embedded {
		return prior
	}
	return after
}

// jsonValuePriorEmbeddedInExpandedCurrent is true when prior unmarshals to an object
// with a by-value chart `type` and every value path in prior is equal in the current
// object, allowing extra keys and trailing array elements in current.
func jsonValuePriorEmbeddedInExpandedCurrent(priorJSON, currentJSON string) (bool, error) {
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

// jsonValueSubsumedByCurrentObject returns whether every key path in prior exists in
// current with a matching value, allowing extra keys in current. For the top-level
// panel `config` object only (`isRoot`), the `styling` key is ignored: Kibana may
// replace the full styling subtree (REQ-035).
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

// isOmissibleDefaultKqlQuery reports whether the practitioner `query` object is the
// usual Kibana default so that a missing `query` key on read can still match.
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
				// Kibana often omits optional empty string fields.
				continue
			}
			if k == "query" {
				if qm, y := pv.(map[string]any); y && isOmissibleDefaultKqlQuery(qm) {
					continue
				}
			}
			return false
		}
		// Kibana can omit a key but also return an empty list instead.
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
			// Kibana can serialize optional lists as `null` instead of omitting the key.
			return true
		}
		c, ok := current.([]any)
		if !ok {
			return false
		}
		if len(p) == 0 {
			// A non-empty list in the response when the user sent [] is not treated as
			// an embed; the next write from preserved [] could strip API data.
			return len(c) == 0
		}
		// Trailing elements in `current` beyond `len(prior)` are allowed (API may append);
		// indices 0..len(p)-1 must match. Reordering or prepending is not a subset match.
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

// populateLensDashboardAppByValueFromAPI stores by_value from a by-value chart API read.
// When prior state used raw `config_json`, preservation rules for that string are unchanged.
// When prior state used a typed chart block, the same block is repopulated when the API
// response decodes to that chart type via the vis converter; otherwise the read falls
// back to `by_value.config_json`.
func populateLensDashboardAppByValueFromAPI(
	ctx context.Context,
	dashboard *models.DashboardModel,
	prior *models.LensDashboardAppConfigModel,
	configBytes []byte,
	pm *models.PanelModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	norm, okNorm := marshalToNormalized(configBytes, nil, "by_value.config_json", &diags)

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
