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

package dashboard

import (
	"context"
	"encoding/json"
	"maps"
	"strconv"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func titleOnlyModel() *models.DashboardModel {
	return &models.DashboardModel{Title: types.StringValue("title only")}
}

func Test_dashboardToAPIRequests_omitNilRootBlocks(t *testing.T) {
	ctx := context.Background()

	var diags diag.Diagnostics
	createBody, err := json.Marshal(dashboardToAPICreateRequest(ctx, titleOnlyModel(), &diags))
	require.NoError(t, err)
	require.False(t, diags.HasError())

	updateBody, err := json.Marshal(dashboardToAPIUpdateRequest(ctx, titleOnlyModel(), &diags))
	require.NoError(t, err)
	require.False(t, diags.HasError())

	for name, body := range map[string][]byte{"create": createBody, "update": updateBody} {
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got), name)
		require.Equal(t, "title only", got["title"], name)
		for _, key := range []string{"time_range", "refresh_interval", "query", "options"} {
			require.NotContains(t, got, key, "%s request must omit %s", name, key)
		}
	}
}

func Test_dashboardToAPIRequests_sendSetRootBlocks(t *testing.T) {
	ctx := context.Background()
	m := testDashboardPlanModel(types.StringNull())
	m.Query.Expression = types.StringValue("response:200")
	m.Options = &models.OptionsModel{HidePanelBorders: types.BoolValue(true)}

	var diags diag.Diagnostics
	createBody, err := json.Marshal(dashboardToAPICreateRequest(ctx, &m, &diags))
	require.NoError(t, err)
	updateBody, err := json.Marshal(dashboardToAPIUpdateRequest(ctx, &m, &diags))
	require.NoError(t, err)
	require.False(t, diags.HasError())

	for name, body := range map[string][]byte{"create": createBody, "update": updateBody} {
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got), name)
		require.Equal(t, map[string]any{"from": "now-15m", "to": "now"}, got["time_range"], name)
		require.Equal(t, map[string]any{"pause": true, "value": float64(90000)}, got["refresh_interval"], name)
		require.Equal(t, "response:200", got["query"].(map[string]any)["expression"], name)
		require.Contains(t, got, "options", name)
	}
}

type validateTestProvider struct{}

func (validateTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "elasticstack"
}
func (validateTestProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {
}
func (validateTestProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (validateTestProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }
func (validateTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewResource}
}

// nullObject returns a value map with every attribute of objType set to a typed null.
func nullObject(objType tftypes.Object) map[string]tftypes.Value {
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(typ, nil)
	}
	return vals
}

func requireNoErrors(t *testing.T, diags []*tfprotov6.Diagnostic) {
	t.Helper()
	for _, d := range diags {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
}

func validateDashboardConfig(t *testing.T, overrides map[string]tftypes.Value) []*tfprotov6.Diagnostic {
	t.Helper()
	ctx := context.Background()
	server := providerserver.NewProtocol6(validateTestProvider{})()

	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	res, ok := schemaResp.ResourceSchemas["elasticstack_kibana_dashboard"]
	require.True(t, ok)
	objType, ok := res.ValueType().(tftypes.Object)
	require.True(t, ok)

	vals := nullObject(objType)
	maps.Copy(vals, overrides)
	cfg, err := tfprotov6.NewDynamicValue(objType, tftypes.NewValue(objType, vals))
	require.NoError(t, err)

	resp, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{
		TypeName: "elasticstack_kibana_dashboard",
		Config:   &cfg,
	})
	require.NoError(t, err)
	return resp.Diagnostics
}

func Test_resourceSchema_rootBlocksOptional(t *testing.T) {
	s := getSchema()
	for _, name := range []string{"time_range", "refresh_interval", "query"} {
		attr := s.Attributes[name]
		require.True(t, attr.IsOptional(), name)
		require.False(t, attr.IsComputed(), name)
	}
}

func Test_resourceSchema_titleOnlyValidates(t *testing.T) {
	diags := validateDashboardConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "title only"),
	})
	requireNoErrors(t, diags)
}

// rootBlockValue builds an object value for the named root block from the
// resource schema, with every nested attribute null except the given overrides.
func rootBlockValue(t *testing.T, block string, overrides map[string]string) tftypes.Value {
	t.Helper()
	objType, ok := getSchema().Attributes[block].GetType().TerraformType(context.Background()).(tftypes.Object)
	require.True(t, ok)
	vals := nullObject(objType)
	for name, v := range overrides {
		typ := objType.AttributeTypes[name]
		switch {
		case typ.Equal(tftypes.Bool):
			vals[name] = tftypes.NewValue(tftypes.Bool, v == "true")
		case typ.Equal(tftypes.Number):
			n, err := strconv.ParseInt(v, 10, 64)
			require.NoError(t, err)
			vals[name] = tftypes.NewValue(tftypes.Number, n)
		default:
			vals[name] = tftypes.NewValue(typ, v)
		}
	}
	return tftypes.NewValue(objType, vals)
}

func hasErrorAt(diags []*tfprotov6.Diagnostic, p *tftypes.AttributePath) bool {
	for _, d := range diags {
		if d.Severity == tfprotov6.DiagnosticSeverityError && d.Attribute != nil && d.Attribute.Equal(p) {
			return true
		}
	}
	return false
}

func Test_resourceSchema_rootBlockNestedValidation(t *testing.T) {
	title := tftypes.NewValue(tftypes.String, "t")
	at := func(block, attr string) *tftypes.AttributePath {
		return tftypes.NewAttributePath().WithAttributeName(block).WithAttributeName(attr)
	}

	tests := []struct {
		name     string
		block    string
		nested   map[string]string
		errBlock string
		errAttr  string
		wantErr  bool
	}{
		{"query missing expression", "query", map[string]string{"language": "kql"}, "query", "expression", true},
		{"query invalid language", "query", map[string]string{"language": "sql", "expression": "a"}, "query", "language", true},
		{"query missing language", "query", map[string]string{"expression": "a"}, "query", "language", true},
		{"query expression valid", "query", map[string]string{"language": "kql", "expression": "a"}, "", "", false},
		{"query empty expression valid", "query", map[string]string{"language": "kql", "expression": ""}, "", "", false},
		{"refresh_interval missing pause", "refresh_interval", map[string]string{"value": "1000"}, "refresh_interval", "pause", true},
		{"refresh_interval missing value", "refresh_interval", map[string]string{"pause": "true"}, "refresh_interval", "value", true},
		{"refresh_interval valid", "refresh_interval", map[string]string{"pause": "true", "value": "0"}, "", "", false},
		{"time_range missing to", "time_range", map[string]string{"from": "now-7d"}, "time_range", "to", true},
		{"time_range invalid mode", "time_range", map[string]string{"from": "a", "to": "b", "mode": "bogus"}, "time_range", "mode", true},
		{"time_range valid mode", "time_range", map[string]string{"from": "a", "to": "b", "mode": "relative"}, "", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateDashboardConfig(t, map[string]tftypes.Value{
				"title":  title,
				tc.block: rootBlockValue(t, tc.block, tc.nested),
			})
			if tc.wantErr {
				require.True(t, hasErrorAt(diags, at(tc.errBlock, tc.errAttr)), "expected error at %s.%s, got %v", tc.errBlock, tc.errAttr, diags)
				return
			}
			requireNoErrors(t, diags)
		})
	}
}
