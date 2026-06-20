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

package entity

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Common attribute keys used throughout schema and helpers
const (
	attrName                = "name"
	attrType                = "type"
	attrRisk                = "risk"
	attrAsset               = "asset"
	attrCalculatedLevel     = "calculated_level"
	attrCalculatedScore     = "calculated_score"
	attrCalculatedScoreNorm = "calculated_score_norm"
	attrDomain              = "domain"
	attrEmail               = "email"
	attrProvider            = "provider"
	attrReason              = "reason"
	attrValue               = "value"

	descCalculatedLevel     = "The calculated risk level."
	descCalculatedScore     = "The raw numeric value of the given entity's risk score."
	descCalculatedScoreNorm = "The normalized numeric value of the given entity's risk score."

	// Attribute keys used in maps (to satisfy goconst)
	attrTimestamp           = "@timestamp"
	attrEntity              = "entity"
	attrHost                = "host"
	attrUser                = "user"
	attrService             = "service"
	attrCloud               = "cloud"
	attrOrchestrator        = "orchestrator"
	attrEvent               = "event"
	attrLabels              = "labels"
	attrTags                = "tags"
	attrDocumentJSON        = "document_json"
	attrAttributes          = "attributes"
	attrBehaviors           = "behaviors"
	attrLifecycle           = "lifecycle"
	attrRelationships       = "relationships"
	attrSubType             = "sub_type"
	attrID                  = "id"
	attrSource              = "source"
	attrCriticalityFeedback = "criticality_feedback"
	attrOwner               = "owner"
	attrOs                  = "os"
)

// canonicalJSON normalizes a Go value to canonical JSON (sorted keys).
func canonicalJSON(v any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, err := json.Marshal(v)
	if err != nil {
		diags.AddError("JSON marshal error", err.Error())
		return "", diags
	}
	var tmp any
	if err := json.Unmarshal(b, &tmp); err != nil {
		diags.AddError("JSON unmarshal error", err.Error())
		return "", diags
	}
	b, err = json.Marshal(tmp)
	if err != nil {
		diags.AddError("JSON marshal error", err.Error())
		return "", diags
	}
	return string(b), diags
}

// canonicalMapJSON returns canonical JSON for a map[string]any, or empty string for nil.
func canonicalMapJSON(m map[string]any) string {
	if m == nil {
		return ""
	}
	s, diags := canonicalJSON(m)
	if diags.HasError() {
		return ""
	}
	return s
}

// getStringSetValue converts a []any of strings to a types.Set of strings.
func getStringSetValue(m map[string]any, key string) types.Set {
	if m == nil {
		return types.SetNull(types.StringType)
	}
	raw, ok := m[key]
	if !ok {
		return types.SetNull(types.StringType)
	}
	arr, ok := raw.([]any)
	if !ok {
		return types.SetNull(types.StringType)
	}
	vals := make([]attr.Value, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			vals = append(vals, types.StringValue(s))
		}
	}
	set, _ := types.SetValue(types.StringType, vals)
	return set
}

// appendStringSetToMap appends a types.Set of strings to a map as []string if non-empty.
func appendStringSetToMap(m map[string]any, key string, set types.Set) {
	if set.IsNull() || set.IsUnknown() || len(set.Elements()) == 0 {
		return
	}
	vals := make([]string, 0, len(set.Elements()))
	for _, v := range set.Elements() {
		if s, ok := v.(types.String); ok {
			vals = append(vals, s.ValueString())
		}
	}
	if len(vals) > 0 {
		m[key] = vals
	}
}
