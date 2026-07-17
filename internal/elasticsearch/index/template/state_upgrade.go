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

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func upgradeStateV0ToV1() resource.StateUpgrader {
	return resource.StateUpgrader{
		StateUpgrader: migrateIndexTemplateStateV0ToV1,
	}
}

// migrateIndexTemplateStateV0ToV1 collapses SDK list/set-shaped MaxItems:1 blocks to Plugin Framework SingleNestedBlock object/null shape.
func migrateIndexTemplateStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(stateutil.CollapseListPath(stateMap, attrDataStream, attrDataStream)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(stateutil.CollapseListPath(stateMap, attrTemplate, attrTemplate)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, ok := stateMap[attrTemplate].(map[string]any)
	if ok {
		resp.Diagnostics.Append(stateutil.CollapseListPath(tmpl, attrLifecycle, "template.lifecycle")...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(stateutil.CollapseListPath(tmpl, attrDataStreamOptions, "template.data_stream_options")...)
		if resp.Diagnostics.HasError() {
			return
		}

		dso, ok := tmpl[attrDataStreamOptions].(map[string]any)
		if ok {
			resp.Diagnostics.Append(stateutil.CollapseListPath(dso, attrFailureStore, "template.data_stream_options.failure_store")...)
			if resp.Diagnostics.HasError() {
				return
			}
			fs, ok := dso[attrFailureStore].(map[string]any)
			if ok {
				resp.Diagnostics.Append(stateutil.CollapseListPath(fs, attrLifecycle, "template.data_stream_options.failure_store.lifecycle")...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}
	}

	aliasutil.NormalizeTemplateObjectInV1State(stateMap, attrLifecycle, attrDataStreamOptions)

	stateutil.NullifyEmptyString(stateMap, "metadata")

	aliasutil.NormalizeVersionZero(stateMap)

	stateutil.MarshalStateMap(stateMap, resp)
}
