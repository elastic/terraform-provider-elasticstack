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

package typeutils

import (
	"encoding/json"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FormatStrictDateTime formats a time.Time as ISO 8601 strict date-time with milliseconds.
func FormatStrictDateTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000Z")
}

// TimeToStringValue formats a time.Time to ISO 8601 format and returns a types.StringValue.
func TimeToStringValue(t time.Time) types.String {
	return types.StringValue(FormatStrictDateTime(t))
}

// ElasticDateTimeToMillis converts any Elastic DateTime representation to Unix milliseconds.
// Handles float64, int64, int, uint64, json.Number, and RFC3339 strings.
// Returns (0, false) for nil, empty, or unparseable values.
//
// estypes.DateTime is defined as `type DateTime any` (a named interface); its concrete value is
// unwrapped by Go's interface-boxing rules before the type switch, so DateTime(float64) is handled
// by the float64 case and DateTime(string) by the string case.
func ElasticDateTimeToMillis(v any) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	case uint64:
		return int64(x), true
	case json.Number:
		i, err := x.Int64()
		if err == nil {
			return i, true
		}
		f, err := x.Float64()
		if err != nil {
			return 0, false
		}
		return int64(f), true
	case string:
		if x == "" {
			return 0, false
		}
		t, err := time.Parse(time.RFC3339, x)
		if err != nil {
			return 0, false
		}
		return t.UnixMilli(), true
	default:
		return 0, false
	}
}

// ElasticDateTimeToStringValue converts an Elastic DateTime (passed as any) to a types.String in
// ISO 8601 format. Returns types.StringNull() for nil, zero, empty, or unparseable values.
// Accepts any value; callers typically pass estypes.DateTime which is `type DateTime any`.
func ElasticDateTimeToStringValue(v any) types.String {
	ms, ok := ElasticDateTimeToMillis(v)
	if !ok || ms == 0 {
		return types.StringNull()
	}
	return TimeToStringValue(time.UnixMilli(ms).UTC())
}
