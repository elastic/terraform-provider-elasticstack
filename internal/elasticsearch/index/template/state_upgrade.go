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

package template

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func upgradeStateV0ToV1() resource.StateUpgrader {
	return resource.StateUpgrader{
		StateUpgrader: migrateIndexTemplateStateV0ToV1,
	}
}

// migrateIndexTemplateStateV0ToV1 collapses SDK list/set-shaped MaxItems:1 blocks to Plugin Framework SingleNestedBlock object/null shape.
func migrateIndexTemplateStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	resp.Diagnostics.Append(aliasutil.CollapseListPath(stateMap, attrDataStream, attrDataStream)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(aliasutil.CollapseListPath(stateMap, attrTemplate, attrTemplate)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, ok := stateMap[attrTemplate].(map[string]any)
	if ok {
		resp.Diagnostics.Append(aliasutil.CollapseListPath(tmpl, attrLifecycle, "template.lifecycle")...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(aliasutil.CollapseListPath(tmpl, attrDataStreamOptions, "template.data_stream_options")...)
		if resp.Diagnostics.HasError() {
			return
		}

		dso, ok := tmpl[attrDataStreamOptions].(map[string]any)
		if ok {
			resp.Diagnostics.Append(aliasutil.CollapseListPath(dso, attrFailureStore, "template.data_stream_options.failure_store")...)
			if resp.Diagnostics.HasError() {
				return
			}
			fs, ok := dso[attrFailureStore].(map[string]any)
			if ok {
				resp.Diagnostics.Append(aliasutil.CollapseListPath(fs, attrLifecycle, "template.data_stream_options.failure_store.lifecycle")...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}
	}

	if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok {
		ensureTemplateObjectKeysForV1(tmpl)
		aliasutil.NormalizeTemplateAliasesInV1State(tmpl)
	}

	// SDKv2 may persist version = 0 when the field is omitted in HCL; Elasticsearch readback and
	// the Plugin Framework schema treat that as unset (null). Drop the key so migrated state matches.
	if v, ok := stateMap["version"]; ok && aliasutil.JSONNumberish(v) == 0 {
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

// ensureTemplateObjectKeysForV1 fills keys the Plugin Framework v1 schema expects on the template
// object so RawState JSON decodes after upgrade. Plugin SDK state may omit optional empty blocks.
func ensureTemplateObjectKeysForV1(tmpl map[string]any) {
	if _, ok := tmpl[attrAlias]; !ok {
		// Empty nested sets are null in Terraform JSON state, not [].
		tmpl[attrAlias] = nil
	}
	if _, ok := tmpl[attrMappings]; !ok {
		tmpl[attrMappings] = nil
	}
	if _, ok := tmpl[attrSettings]; !ok {
		tmpl[attrSettings] = nil
	}
	if _, ok := tmpl[attrLifecycle]; !ok {
		tmpl[attrLifecycle] = nil
	}
	if _, ok := tmpl[attrDataStreamOptions]; !ok {
		tmpl[attrDataStreamOptions] = nil
	}
}
