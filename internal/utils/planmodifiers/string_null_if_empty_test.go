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

package planmodifiers_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestStringNullIfEmpty(t *testing.T) {
	t.Parallel()

	modifier := planmodifiers.StringNullIfEmpty()
	ctx := context.Background()

	cases := []struct {
		name        string
		configValue types.String
		planValue   types.String
		wantNull    bool
		wantValue   string
	}{
		{
			name:        "empty string becomes null",
			configValue: types.StringValue(""),
			planValue:   types.StringValue(""),
			wantNull:    true,
		},
		{
			name:        "non-empty string unchanged",
			configValue: types.StringValue("40mb"),
			planValue:   types.StringValue("40mb"),
			wantNull:    false,
			wantValue:   "40mb",
		},
		{
			name:        "null config unchanged",
			configValue: types.StringNull(),
			planValue:   types.StringNull(),
			wantNull:    true,
		},
		{
			name:        "unknown config unchanged",
			configValue: types.StringUnknown(),
			planValue:   types.StringUnknown(),
			wantNull:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.StringRequest{
				ConfigValue: tc.configValue,
				PlanValue:   tc.planValue,
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tc.planValue,
			}

			modifier.PlanModifyString(ctx, req, resp)

			if tc.wantNull {
				require.True(t, resp.PlanValue.IsNull(), "expected null plan value")
			} else if tc.configValue.IsUnknown() {
				require.True(t, resp.PlanValue.IsUnknown(), "expected unknown plan value")
			} else {
				require.Equal(t, tc.wantValue, resp.PlanValue.ValueString())
			}
		})
	}
}
