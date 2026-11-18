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

package agentdownloadsource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_sslAtLeastOneSubattrValidator(t *testing.T) {
	ctx := context.Background()
	attrTypes := getSslAttrTypes(ctx)

	tests := []struct {
		name    string
		obj     types.Object
		wantErr bool
	}{
		{
			name: "null object is allowed",
			obj:  types.ObjectNull(attrTypes),
		},
		{
			name: "empty object is rejected",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate":             types.StringNull(),
				"certificate_authorities": types.ListNull(types.StringType),
				"key":                     types.StringNull(),
				"key_wo":                  types.StringNull(),
			}),
			wantErr: true,
		},
		{
			name: "certificate set is allowed",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate":             types.StringValue("cert"),
				"certificate_authorities": types.ListNull(types.StringType),
				"key":                     types.StringNull(),
				"key_wo":                  types.StringNull(),
			}),
		},
		{
			name: "certificate_authorities set is allowed",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate":             types.StringNull(),
				"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
				"key":                     types.StringNull(),
				"key_wo":                  types.StringNull(),
			}),
		},
		{
			name: "key set is allowed",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate":             types.StringNull(),
				"certificate_authorities": types.ListNull(types.StringType),
				"key":                     types.StringValue("key"),
				"key_wo":                  types.StringNull(),
			}),
		},
		{
			name: "key_wo only is allowed",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate":             types.StringNull(),
				"certificate_authorities": types.ListNull(types.StringType),
				"key":                     types.StringNull(),
				"key_wo":                  types.StringValue("secret"),
			}),
		},
		{
			name: "unknown key_wo defers validation",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate":             types.StringNull(),
				"certificate_authorities": types.ListNull(types.StringType),
				"key":                     types.StringNull(),
				"key_wo":                  types.StringUnknown(),
			}),
		},
		{
			name: "fully unknown object defers validation",
			obj:  types.ObjectUnknown(attrTypes),
		},
		{
			name: "unknown child defers validation",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate":             types.StringUnknown(),
				"certificate_authorities": types.ListNull(types.StringType),
				"key":                     types.StringNull(),
				"key_wo":                  types.StringNull(),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp validator.ObjectResponse
			sslAtLeastOneSubattrValidator{}.ValidateObject(ctx, validator.ObjectRequest{
				ConfigValue: tt.obj,
				Path:        path.Root("ssl"),
			}, &resp)
			assert.Equal(t, tt.wantErr, resp.Diagnostics.HasError(), resp.Diagnostics)
			if tt.wantErr {
				assert.Equal(t, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("ssl"), sslAtLeastOneSubattrSummary, sslAtLeastOneSubattrMessage)}, resp.Diagnostics)
			}
		})
	}
}
