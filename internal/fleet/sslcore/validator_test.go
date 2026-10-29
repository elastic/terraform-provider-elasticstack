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

package sslcore

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtLeastOneSubattrValidator(t *testing.T) {
	t.Parallel()

	const summary = "Empty configuration"
	const message = "At least one of a or b must be set."

	v := AtLeastOneSubattrValidator([]string{"a", "b"}, summary, message)

	require.Equal(t, "At least one of a or b must be set", v.Description(t.Context()))
	require.Equal(t, v.Description(t.Context()), v.MarkdownDescription(t.Context()))

	attrTypes := map[string]attr.Type{"a": types.StringType, "b": types.StringType}

	tests := []struct {
		name    string
		obj     types.Object
		wantErr bool
	}{
		{name: "null object is allowed", obj: types.ObjectNull(attrTypes)},
		{name: "unknown object is allowed", obj: types.ObjectUnknown(attrTypes)},
		{
			name: "all-null object is rejected",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"a": types.StringNull(),
				"b": types.StringNull(),
			}),
			wantErr: true,
		},
		{
			name: "one subattr set is allowed",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"a": types.StringValue("x"),
				"b": types.StringNull(),
			}),
		},
		{
			name: "unknown child defers validation",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"a": types.StringUnknown(),
				"b": types.StringNull(),
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.ObjectResponse{}
			v.ValidateObject(t.Context(), validator.ObjectRequest{
				Path:        path.Root("ssl"),
				ConfigValue: tc.obj,
			}, resp)
			assert.Equal(t, tc.wantErr, resp.Diagnostics.HasError())
			if tc.wantErr {
				assert.Equal(t, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("ssl"), summary, message)}, resp.Diagnostics)
			}
		})
	}
}
