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

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func upgradeStateV0ToV1() resource.StateUpgrader {
	return resource.StateUpgrader{
		StateUpgrader: migrateComponentTemplateStateV0ToV1,
	}
}

// migrateComponentTemplateStateV0ToV1 collapses SDK list-shaped MaxItems:1 blocks to Plugin Framework
// SingleNestedBlock object/null shape.
func migrateComponentTemplateStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(stateutil.CollapseListPath(stateMap, attrTemplate, attrTemplate)...)
	if resp.Diagnostics.HasError() {
		return
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

	stateutil.MarshalStateMap(stateMap, resp)
}

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
	if _, ok := tmpl[attrDataStreamOptions]; !ok {
		tmpl[attrDataStreamOptions] = nil
	}
}
