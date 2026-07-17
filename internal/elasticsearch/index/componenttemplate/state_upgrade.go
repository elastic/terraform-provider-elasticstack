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

	aliasutil.NormalizeTemplateObjectInV1State(stateMap, attrDataStreamOptions)

	stateutil.NullifyEmptyString(stateMap, "metadata")

	aliasutil.NormalizeVersionZero(stateMap)

	stateutil.MarshalStateMap(stateMap, resp)
}
