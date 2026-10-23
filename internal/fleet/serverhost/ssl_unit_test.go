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

package serverhost

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaSsl(t *testing.T) {
	t.Parallel()

	sslAttr, ok := getSchema(context.Background()).Attributes["ssl"].(schema.SingleNestedAttribute)
	require.True(t, ok, "ssl attribute must exist as a single nested attribute")
	require.True(t, sslAttr.Optional)
	require.False(t, sslAttr.Computed)
	require.Equal(t, []validator.Object{sslAtLeastOneSubattrValidator()}, sslAttr.Validators)
	require.Len(t, sslAttr.Attributes, 4)
	require.Equal(t, sslAttrTypes, sslAttr.GetType().(attr.TypeWithAttributeTypes).AttributeTypes())

	cert, ok := sslAttr.Attributes["certificate"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, cert.Optional)
	require.False(t, cert.Sensitive)
	require.Equal(t, []validator.String{stringvalidator.LengthAtLeast(1)}, cert.Validators)

	key, ok := sslAttr.Attributes["key"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, key.Optional)
	require.True(t, key.Sensitive)
	require.Equal(t, []validator.String{stringvalidator.LengthAtLeast(1)}, key.Validators)

	cas, ok := sslAttr.Attributes["certificate_authorities"].(schema.ListAttribute)
	require.True(t, ok)
	require.True(t, cas.Optional)
	require.Equal(t, types.StringType, cas.ElementType)
	require.Equal(t, []validator.List{
		listvalidator.SizeAtLeast(1),
		listvalidator.SizeAtMost(10),
		listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
	}, cas.Validators)

	clientAuth, ok := sslAttr.Attributes["client_auth"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, clientAuth.Optional)
	require.False(t, clientAuth.Computed, "Kibana returns no client_auth default, so it must not be computed")
	require.Equal(t, []validator.String{stringvalidator.OneOf("none", "optional", "required")}, clientAuth.Validators)
}

func TestSslAtLeastOneSubattrValidator(t *testing.T) {
	t.Parallel()

	attrTypes := sslAttrTypes
	nullAttrs := func() map[string]attr.Value {
		return map[string]attr.Value{
			"certificate":             types.StringNull(),
			"key":                     types.StringNull(),
			"certificate_authorities": types.ListNull(types.StringType),
			"client_auth":             types.StringNull(),
		}
	}
	with := func(name string, value attr.Value) types.Object {
		attrs := nullAttrs()
		attrs[name] = value
		return types.ObjectValueMust(attrTypes, attrs)
	}

	tests := []struct {
		name    string
		obj     types.Object
		wantErr bool
	}{
		{name: "null object is allowed", obj: types.ObjectNull(attrTypes)},
		{name: "unknown object is allowed", obj: types.ObjectUnknown(attrTypes)},
		{name: "all-null object is rejected", obj: types.ObjectValueMust(attrTypes, nullAttrs()), wantErr: true},
		{name: "certificate alone", obj: with("certificate", types.StringValue("c"))},
		{name: "key alone", obj: with("key", types.StringValue("k"))},
		{name: "certificate_authorities alone", obj: with("certificate_authorities", types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}))},
		{name: "client_auth alone", obj: with("client_auth", types.StringValue("none"))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.ObjectResponse{}
			sslAtLeastOneSubattrValidator().ValidateObject(t.Context(), validator.ObjectRequest{
				Path:        path.Root("ssl"),
				ConfigValue: tc.obj,
			}, resp)
			assert.Equal(t, tc.wantErr, resp.Diagnostics.HasError())
		})
	}
}

type versionRequirementsTestClient struct {
	version *version.Version
}

func (c versionRequirementsTestClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	return !c.version.LessThan(minVersion), nil
}

func sslDiagnostics(diags diag.Diagnostics) []diag.DiagnosticWithPath {
	var out []diag.DiagnosticWithPath
	for _, d := range diags {
		if withPath, ok := d.(diag.DiagnosticWithPath); ok && withPath.Path().Equal(path.Root("ssl")) {
			out = append(out, withPath)
		}
	}
	return out
}

func TestGetVersionRequirements_Ssl(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nullSSL := types.ObjectNull(sslAttrTypes)
	configured := sslObjectValue(t, new("cert"), nil, nil, nil)
	unknownSSL := types.ObjectUnknown(sslAttrTypes)

	tests := []struct {
		name     string
		ssl      types.Object
		stack    string
		wantDiag bool
		wantReqs int
	}{
		{name: "absent on 9.0.0", ssl: nullSSL, stack: "9.0.0", wantReqs: 0},
		{name: "absent on 8.0.0", ssl: nullSSL, stack: "8.0.0", wantReqs: 0},
		{name: "unknown adds no requirement", ssl: unknownSSL, stack: "9.0.0", wantReqs: 0},
		{name: "configured below floor", ssl: configured, stack: "9.0.0", wantDiag: true, wantReqs: 1},
		{name: "configured at floor", ssl: configured, stack: "9.1.0", wantReqs: 1},
		{name: "configured above floor", ssl: configured, stack: "9.5.5", wantReqs: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := baseSSLModel(tc.ssl)
			reqs, diags := m.GetVersionRequirements(ctx)
			require.False(t, diags.HasError())
			require.Len(t, reqs, tc.wantReqs)

			client := versionRequirementsTestClient{version: version.Must(version.NewVersion(tc.stack))}
			enforced := entitycore.EnforceVersionRequirements(ctx, client, m)
			require.Equal(t, tc.wantDiag, enforced.HasError())
			if tc.wantDiag {
				sslDiags := sslDiagnostics(enforced)
				require.Len(t, sslDiags, 1, "expected exactly one attribute-scoped diagnostic on ssl")
				require.Contains(t, sslDiags[0].Detail(), "ssl is only supported in Elastic Stack 9.1.0 and above")
			} else {
				require.Empty(t, sslDiagnostics(enforced))
			}
		})
	}
}
