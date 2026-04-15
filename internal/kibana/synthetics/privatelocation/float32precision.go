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

package privatelocation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// float32PrecisionPlanModifier normalizes a float64 value through float32 precision.
// The Kibana API stores geo coordinates as float32. When a user provides a float64
// value (e.g. 42.42), it is transmitted as float32 to the API and returned as float32
// on read (e.g. 42.41999816894531). This modifier ensures the planned value uses
// float32 precision so that it is consistent with the API response, avoiding
// "Provider produced inconsistent result after apply" errors.
type float32PrecisionPlanModifier struct{}

// Float32Precision returns a plan modifier that normalizes float64 values to float32 precision.
func Float32Precision() planmodifier.Float64 {
	return float32PrecisionPlanModifier{}
}

func (m float32PrecisionPlanModifier) Description(_ context.Context) string {
	return "Normalizes float64 values to float32 precision to match API storage."
}

func (m float32PrecisionPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes float64 values to float32 precision to match API storage."
}

func (m float32PrecisionPlanModifier) PlanModifyFloat64(_ context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	// Normalize through float32: float64(float32(v)) produces the same value the API returns.
	normalized := float64(float32(req.ConfigValue.ValueFloat64()))
	resp.PlanValue = types.Float64Value(normalized)
}
