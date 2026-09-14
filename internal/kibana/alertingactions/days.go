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

// Package alertingactions provides shared conversion helpers for the
// alerts_filter.timeframe.days field of a Kibana alerting rule action, used
// by both the generic alertingrule resource and the security_detection_rule
// resource. Both resources wrap the same underlying Kibana alerting actions
// API, but decode it via different clients: alertingrule's hand-rolled JSON
// structs already yield typed integers, while security_detection_rule's
// generated OpenAPI client surfaces the field as untyped map[string]any,
// where JSON numbers can arrive as float64 or json.Number depending on the
// decode path. DaysFromAPI accepts either shape so both callers get the same
// defensive numeric handling.
package alertingactions

import (
	"encoding/json"
	"fmt"
)

// DaysFromAPI normalizes a Kibana API "days" value (already-typed integer
// slices, or the untyped []any produced by decoding JSON into map[string]any)
// into a []int64 suitable for a Terraform list.
func DaysFromAPI(raw any) ([]int64, error) {
	switch v := raw.(type) {
	case []int64:
		return v, nil
	case []int32:
		days := make([]int64, len(v))
		for i, d := range v {
			days[i] = int64(d)
		}
		return days, nil
	case []int:
		days := make([]int64, len(v))
		for i, d := range v {
			days[i] = int64(d)
		}
		return days, nil
	case []any:
		days := make([]int64, 0, len(v))
		for _, d := range v {
			day, err := dayFromAPI(d)
			if err != nil {
				return nil, err
			}
			days = append(days, day)
		}
		return days, nil
	default:
		return nil, fmt.Errorf("days must be an array, got %T", raw)
	}
}

// dayFromAPI converts a single decoded JSON number into an int64, handling
// the concrete types the standard library's encoding/json can produce for a
// number depending on the decode path (float64 for interface{} targets,
// json.Number when UseNumber is set), plus already-typed integers.
func dayFromAPI(raw any) (int64, error) {
	switch v := raw.(type) {
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	default:
		return 0, fmt.Errorf("unexpected day value type %T", raw)
	}
}

// Int32FromInt64 converts a Terraform-decoded []int64 into the []int32 shape
// expected by alertingrule's typed API model.
func Int32FromInt64(days []int64) []int32 {
	out := make([]int32, len(days))
	for i, d := range days {
		out[i] = int32(d)
	}
	return out
}

// IntFromInt64 converts a Terraform-decoded []int64 into the []int shape
// expected by security_detection_rule's map[string]any API payload.
func IntFromInt64(days []int64) []int {
	out := make([]int, len(days))
	for i, d := range days {
		out[i] = int(d)
	}
	return out
}

// Int32FromInt converts a decoded []int (kibanaoapi's JSON response shape)
// into the []int32 shape expected by models.AlertsFilterTimeframe.
func Int32FromInt(days []int) []int32 {
	out := make([]int32, len(days))
	for i, d := range days {
		out[i] = int32(d)
	}
	return out
}

// IntFromInt32 converts a models.AlertsFilterTimeframe []int32 into the
// []int shape expected by kibanaoapi's request body.
func IntFromInt32(days []int32) []int {
	out := make([]int, len(days))
	for i, d := range days {
		out[i] = int(d)
	}
	return out
}
