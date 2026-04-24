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

package integrationpolicy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchema_SpaceIDs_AtMostOne verifies the schema-level constraint that an
// integration policy can carry at most one space_id. The Fleet
// `package_policies` API does not support assigning a package policy to more
// than one space (the request body has no `space_ids` field); previously the
// provider silently dropped all but the first element, leaving state
// inconsistent with reality. This test locks in the plan-time rejection.
func TestSchema_SpaceIDs_AtMostOne(t *testing.T) {
	t.Parallel()

	r := &integrationPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "Schema must not return diagnostics: %v", resp.Diagnostics)

	a, ok := resp.Schema.Attributes["space_ids"]
	require.True(t, ok, "space_ids attribute must exist in the schema")

	setAttr, ok := a.(schema.SetAttribute)
	require.True(t, ok, "space_ids must be declared as SetAttribute, got %T", a)
	require.NotEmpty(t, setAttr.Validators, "space_ids must have at least one validator after the fix")

	zero := types.SetValueMust(types.StringType, nil)
	one := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("space-a")})
	two := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("space-a"),
		types.StringValue("space-b"),
	})

	cases := []struct {
		name      string
		value     types.Set
		wantError bool
	}{
		{name: "zero elements is allowed", value: zero, wantError: false},
		{name: "one element is allowed", value: one, wantError: false},
		{name: "two elements is rejected", value: two, wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vreq := validator.SetRequest{
				Path:        path.Root("space_ids"),
				ConfigValue: tc.value,
			}
			vresp := &validator.SetResponse{}
			for _, v := range setAttr.Validators {
				v.ValidateSet(context.Background(), vreq, vresp)
			}
			if tc.wantError {
				assert.True(t, vresp.Diagnostics.HasError(),
					"expected validator to reject %v", tc.value)
			} else {
				assert.False(t, vresp.Diagnostics.HasError(),
					"expected validator to accept %v; got diags: %v", tc.value, vresp.Diagnostics)
			}
		})
	}
}
