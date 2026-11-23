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

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestSchemaSsl(t *testing.T) {
	ctx := context.Background()

	sslAttr, ok := getSchema(ctx).Attributes["ssl"].(schema.SingleNestedAttribute)
	require.True(t, ok, "ssl attribute must exist as a single nested attribute")
	require.True(t, sslAttr.Optional, "ssl must be Optional")
	require.Equal(t, []validator.Object{sslAtLeastOneSubattrValidator{}}, sslAttr.Validators)

	cas, ok := sslAttr.Attributes["certificate_authorities"].(schema.ListAttribute)
	require.True(t, ok)
	require.Equal(t, types.StringType, cas.ElementType)
	require.True(t, cas.Optional)
	require.Equal(t, []validator.List{
		listvalidator.SizeAtLeast(1),
		listvalidator.SizeAtMost(10),
		listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
	}, cas.Validators)

	cert, ok := sslAttr.Attributes["certificate"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, cert.Optional)
	require.Equal(t, []validator.String{stringvalidator.LengthAtLeast(1)}, cert.Validators)

	key, ok := sslAttr.Attributes["key"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, key.Optional)
	require.True(t, key.Sensitive)
	require.Equal(t, []validator.String{stringvalidator.LengthAtLeast(1)}, key.Validators)

	keyWO, ok := sslAttr.Attributes["key_wo"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, keyWO.Optional)
	require.Len(t, keyWO.Validators, 2, "key_wo must have a length and a conflicts-with-key validator")

	_, hasVerificationMode := sslAttr.Attributes["verification_mode"]
	require.False(t, hasVerificationMode, "ssl must not expose verification_mode")
}

func TestSchemaSslKeyWOConflictsWithKey(t *testing.T) {
	ctx := context.Background()
	sslAttr := getSchema(ctx).Attributes["ssl"].(schema.SingleNestedAttribute)
	keyWO := sslAttr.Attributes["key_wo"].(schema.StringAttribute)

	validate := func(key, keyWOValue types.String) diag.Diagnostics {
		obj := types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate":             types.StringNull(),
			"certificate_authorities": types.ListNull(types.StringType),
			"key":                     key,
			"key_wo":                  keyWOValue,
		})
		tfObj, err := obj.ToTerraformValue(ctx)
		require.NoError(t, err)
		cfgSchema := schema.Schema{Attributes: map[string]schema.Attribute{"ssl": sslAttr}}
		cfg := tfsdk.Config{
			Schema: cfgSchema,
			Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"ssl": tfObj.Type()}},
				map[string]tftypes.Value{"ssl": tfObj}),
		}
		var resp validator.StringResponse
		for _, v := range keyWO.Validators {
			v.ValidateString(ctx, validator.StringRequest{
				Path:           path.Root("ssl").AtName("key_wo"),
				PathExpression: path.MatchRoot("ssl").AtName("key_wo"),
				Config:         cfg,
				ConfigValue:    keyWOValue,
			}, &resp)
		}
		return resp.Diagnostics
	}

	require.False(t, validate(types.StringNull(), types.StringValue("secret")).HasError())
	require.True(t, validate(types.StringValue("plain"), types.StringValue("secret")).HasError(), "key and key_wo together must be rejected")
}
