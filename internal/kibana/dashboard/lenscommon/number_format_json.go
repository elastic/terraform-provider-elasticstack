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

package lenscommon

import (
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	jsonLensFormatNullJSON      = "null"
	DefaultLensNumberFormatJSON = `{"type":"number"}`
)

// NormalizeKibanaLensNumberFormatJSONString trims Lens number-format defaults Kibana adds on read
// (decimals: 2, compact: false) so state matches compact Terraform jsonencode like {"type":"number"}.
func NormalizeKibanaLensNumberFormatJSONString(jsonStr string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return jsonStr
	}
	typ, _ := m["type"].(string)
	if typ != "number" {
		return jsonStr
	}
	if jsonNumericEqualsLoose(m["decimals"], 2) {
		delete(m, "decimals")
	}
	if b, ok := m["compact"].(bool); ok && !b {
		delete(m, "compact")
	}
	sorted := SortJSONMapKeysRecursive(m)
	out, err := json.Marshal(sorted)
	if err != nil {
		return jsonStr
	}
	return string(out)
}

func jsonNumericEqualsLoose(v any, want float64) bool {
	switch x := v.(type) {
	case float64:
		return x == want
	case float32:
		return float64(x) == want
	case int:
		return float64(x) == want
	case int64:
		return float64(x) == want
	case json.Number:
		f, err := x.Float64()
		return err == nil && f == want
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return err == nil && f == want
	default:
		return false
	}
}

// LensESQLNumberFormatJSONFromAPI marshals a Lens ES|QL dimension `format` union
// value to a normalized Terraform string. Empty or null JSON is replaced with the
// default number-format payload so Terraform state matches what Kibana echoes.
func LensESQLNumberFormatJSONFromAPI(format any, errLabel string, diags *diag.Diagnostics) (jsontypes.Normalized, bool) {
	bytes, err := json.Marshal(format)
	if err != nil {
		diags.AddError("Failed to marshal "+errLabel, err.Error())
		return jsontypes.Normalized{}, false
	}
	if len(bytes) == 0 || string(bytes) == jsonLensFormatNullJSON {
		bytes = []byte(DefaultLensNumberFormatJSON)
	}
	return jsontypes.NewNormalizedValue(NormalizeKibanaLensNumberFormatJSONString(string(bytes))), true
}
