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

package index

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type mappingsPlanModifier struct{}

func (p mappingsPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !typeutils.IsKnown(req.StateValue) {
		return
	}

	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}

	stateStr := req.StateValue.ValueString()
	cfgStr := req.ConfigValue.ValueString()

	var stateMappings map[string]any
	var cfgMappings map[string]any

	// No error checking, schema validation ensures this is valid json
	_ = json.Unmarshal([]byte(stateStr), &stateMappings)
	_ = json.Unmarshal([]byte(cfgStr), &cfgMappings)

	result := compareMappingsForPlan(stateMappings, cfgMappings)
	resp.RequiresReplace = result.RequiresReplace
	resp.Diagnostics.Append(result.Diags...)
}

func (p mappingsPlanModifier) Description(_ context.Context) string {
	return "Detects incompatible mapping changes that require replacement"
}

func (p mappingsPlanModifier) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}
