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

package ccr

import (
	"fmt"
	"math"
	"strconv"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NarrowInt64ToInt converts v to an int, returning a diagnostic if it overflows.
func NarrowInt64ToInt(field string, v int64) (int, diag.Diagnostics) {
	if v > math.MaxInt || v < math.MinInt {
		return 0, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Integer overflow",
				fmt.Sprintf("%s value %d exceeds the range of a signed int", field, v),
			),
		}
	}
	return int(v), nil
}

// OptIntFromInt64 returns a pointer to the int value when the Terraform int64 is
// known, narrowing safely from int64 to int. Returns nil for unknown/null.
func OptIntFromInt64(field string, v types.Int64) (*int, diag.Diagnostics) {
	if !typeutils.IsKnown(v) {
		return nil, nil
	}
	narrowed, diags := NarrowInt64ToInt(field, v.ValueInt64())
	if diags.HasError() {
		return nil, diags
	}
	return &narrowed, nil
}

// ByteSizeFromString converts a known Terraform string to an estypes.ByteSize.
func ByteSizeFromString(v types.String) estypes.ByteSize {
	if !typeutils.IsKnown(v) {
		return nil
	}
	return estypes.ByteSize(v.ValueString())
}

// ByteSizeToString converts an estypes.ByteSize to a Terraform string.
func ByteSizeToString(v estypes.ByteSize) types.String {
	if v == nil {
		return types.StringNull()
	}
	switch x := v.(type) {
	case string:
		return types.StringValue(x)
	case int64:
		return types.StringValue(strconv.FormatInt(x, 10))
	case int:
		return types.StringValue(strconv.Itoa(x))
	case float64:
		return types.StringValue(strconv.FormatInt(int64(x), 10))
	default:
		return types.StringValue(fmt.Sprint(v))
	}
}
