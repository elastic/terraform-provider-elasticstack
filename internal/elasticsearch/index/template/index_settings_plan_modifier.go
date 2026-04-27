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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// indexSettingsCanonicalModifier rewrites planned template.settings to the nested
// {"index":{...}} shape Elasticsearch returns, so plan/state match the SDK contract
// while StringSemanticEquals still treats dotted and non-nested forms as equivalent.
type indexSettingsCanonicalModifier struct{}

func indexSettingsCanonicalPlanModifier() planmodifier.String {
	return indexSettingsCanonicalModifier{}
}

func (m indexSettingsCanonicalModifier) Description(_ context.Context) string {
	return "Normalizes template.settings JSON to the canonical Elasticsearch nested index object shape."
}

func (m indexSettingsCanonicalModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m indexSettingsCanonicalModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() {
		resp.PlanValue = basetypes.NewStringNull()
		return
	}
	if req.ConfigValue.IsUnknown() {
		return
	}
	raw := req.ConfigValue.ValueString()
	if raw == "" {
		resp.PlanValue = basetypes.NewStringValue("")
		return
	}
	trimmed := strings.TrimSpace(raw)
	var probe any
	if err := json.Unmarshal([]byte(trimmed), &probe); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid template.settings JSON", err.Error())
		return
	}
	if probe == nil {
		resp.PlanValue = basetypes.NewStringNull()
		return
	}
	canonical, err := customtypes.CanonicalIndexSettingsJSON(trimmed)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid template.settings JSON", err.Error())
		return
	}
	resp.PlanValue = basetypes.NewStringValue(canonical)
}
