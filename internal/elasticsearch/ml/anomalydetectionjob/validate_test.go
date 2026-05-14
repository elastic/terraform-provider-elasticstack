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

package anomalydetectionjob

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestCustomRuleMissingScopeAndConditions(t *testing.T) {
	t.Parallel()

	conditionElemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"applies_to": types.StringType,
			"operator":   types.StringType,
			"value":      types.Float64Type,
		},
	}
	scopeElemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"filter_id":   types.StringType,
			"filter_type": types.StringType,
		},
	}

	cases := []struct {
		name        string
		cond        types.List
		scope       types.Map
		wantMissing bool
	}{
		{
			name:        "both null",
			cond:        types.ListNull(conditionElemType),
			scope:       types.MapNull(scopeElemType),
			wantMissing: true,
		},
		{
			name:        "unknown conditions skips",
			cond:        types.ListUnknown(conditionElemType),
			scope:       types.MapNull(scopeElemType),
			wantMissing: false,
		},
		{
			name:        "unknown scope skips",
			cond:        types.ListNull(conditionElemType),
			scope:       types.MapUnknown(scopeElemType),
			wantMissing: false,
		},
		{
			name: "non-empty scope",
			cond: types.ListNull(conditionElemType),
			scope: types.MapValueMust(scopeElemType, map[string]attr.Value{
				"partition": types.ObjectValueMust(scopeElemType.AttributeTypes(), map[string]attr.Value{
					"filter_id":   types.StringValue("filter-1"),
					"filter_type": types.StringNull(),
				}),
			}),
			wantMissing: false,
		},
		{
			name: "non-empty conditions",
			cond: types.ListValueMust(conditionElemType, []attr.Value{
				types.ObjectValueMust(conditionElemType.AttributeTypes(), map[string]attr.Value{
					"applies_to": types.StringValue("actual"),
					"operator":   types.StringValue("gt"),
					"value":      types.Float64Value(1.5),
				}),
			}),
			scope:       types.MapNull(scopeElemType),
			wantMissing: false,
		},
		{
			name: "non-empty scope and non-empty conditions together",
			cond: types.ListValueMust(conditionElemType, []attr.Value{
				types.ObjectValueMust(conditionElemType.AttributeTypes(), map[string]attr.Value{
					"applies_to": types.StringValue("actual"),
					"operator":   types.StringValue("lt"),
					"value":      types.Float64Value(10),
				}),
			}),
			scope: types.MapValueMust(scopeElemType, map[string]attr.Value{
				"clientip": types.ObjectValueMust(scopeElemType.AttributeTypes(), map[string]attr.Value{
					"filter_id":   types.StringValue("flt-1"),
					"filter_type": types.StringValue("include"),
				}),
			}),
			wantMissing: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := customRuleMissingScopeAndConditions(tc.cond, tc.scope)
			require.Equal(t, tc.wantMissing, got)
		})
	}
}
