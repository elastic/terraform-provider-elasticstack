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

package componenttemplate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func upgradeStateV0ToV1() resource.StateUpgrader {
	return resource.StateUpgrader{
		StateUpgrader: migrateComponentTemplateStateV0ToV1,
	}
}

// migrateComponentTemplateStateV0ToV1 collapses SDK list-shaped MaxItems:1 blocks to Plugin Framework
// SingleNestedBlock object/null shape.
func migrateComponentTemplateStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	resp.Diagnostics.Append(collapseListPath(stateMap, "template", "template")...)
	if resp.Diagnostics.HasError() {
		return
	}

	if tmpl, ok := stateMap["template"].(map[string]any); ok {
		ensureTemplateObjectKeysForV1(tmpl)
		normalizeTemplateAliasesInV1State(tmpl)
	}

	// SDKv2 may persist version = 0 when the field is omitted in HCL; Elasticsearch readback and
	// the Plugin Framework schema treat that as unset (null). Drop the key so migrated state matches.
	if v, ok := stateMap["version"]; ok && jsonNumberish(v) == 0 {
		delete(stateMap, "version")
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: stateJSON,
	}
}

// collapseListPath applies the v0→v1 singleton-list collapse rule at m[key].
// Returns a diagnostic when the value is an array with 2+ elements.
func collapseListPath(m map[string]any, key, pathLabel string) diag.Diagnostics {
	v, ok := m[key]
	if !ok {
		return nil
	}
	if v == nil {
		return nil
	}
	list, ok := v.([]any)
	if !ok {
		return nil
	}
	switch len(list) {
	case 0:
		m[key] = nil
	case 1:
		m[key] = list[0]
	default:
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"State upgrade error",
				fmt.Sprintf(`unexpected multi-element array at path %q`, pathLabel),
			),
		}
	}
	return nil
}

func jsonNumberish(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	default:
		return 0
	}
}

func ensureTemplateObjectKeysForV1(tmpl map[string]any) {
	if _, ok := tmpl["alias"]; !ok {
		// Empty nested sets are null in Terraform JSON state, not [].
		tmpl["alias"] = nil
	}
	if _, ok := tmpl["mappings"]; !ok {
		tmpl["mappings"] = nil
	}
	if _, ok := tmpl["settings"]; !ok {
		tmpl["settings"] = nil
	}
}

// normalizeTemplateAliasesInV1State collapses SDK-style echoed index_routing/search_routing (same
// non-empty value, routing empty or equal) into the Plugin Framework routing-only shape so migrated
// state matches configuration and avoids spurious refresh plans after upgrade.
func normalizeTemplateAliasesInV1State(tmpl map[string]any) {
	av, ok := tmpl["alias"]
	if !ok || av == nil {
		return
	}
	list, ok := av.([]any)
	if !ok {
		return
	}
	for i, el := range list {
		am, ok := el.(map[string]any)
		if !ok {
			continue
		}
		// Plugin Framework jsontypes.Normalized rejects ""; SDK state may store the literal empty string.
		if fv, ok := am["filter"]; ok {
			if s, ok := fv.(string); ok && s == "" {
				am["filter"] = nil
			}
		}
		ir := stringishJSONState(am["index_routing"])
		sr := stringishJSONState(am["search_routing"])
		rt := stringishJSONState(am["routing"])
		if ir != "" && ir == sr && (rt == "" || rt == ir) {
			am["routing"] = ir
			am["index_routing"] = ""
			am["search_routing"] = ""
			list[i] = am
		}
	}
	tmpl["alias"] = list
}

func stringishJSONState(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}
