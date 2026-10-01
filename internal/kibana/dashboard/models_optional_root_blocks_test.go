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
	m.Query.Text = types.StringValue("response:200")
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

	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(typ, nil)
	}
	for name, v := range overrides {
		vals[name] = v
	}
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
		require.False(t, attr.IsRequired(), name)
		require.False(t, attr.IsComputed(), name)
	}
}

func Test_resourceSchema_titleOnlyValidates(t *testing.T) {
	diags := validateDashboardConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "title only"),
	})
	for _, d := range diags {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
}

func Test_resourceSchema_timeRangeMissingToRejected(t *testing.T) {
	trType := getSchema().Attributes["time_range"].GetType().TerraformType(context.Background()).(tftypes.Object)
	trVals := make(map[string]tftypes.Value, len(trType.AttributeTypes))
	for name, typ := range trType.AttributeTypes {
		trVals[name] = tftypes.NewValue(typ, nil)
	}
	trVals["from"] = tftypes.NewValue(tftypes.String, "now-7d")

	diags := validateDashboardConfig(t, map[string]tftypes.Value{
		"title":      tftypes.NewValue(tftypes.String, "bad time range"),
		"time_range": tftypes.NewValue(trType, trVals),
	})
	var found bool
	for _, d := range diags {
		if d.Severity == tfprotov6.DiagnosticSeverityError && d.Attribute != nil &&
			d.Attribute.Equal(tftypes.NewAttributePath().WithAttributeName("time_range").WithAttributeName("to")) {
			found = true
		}
	}
	require.True(t, found, "expected an error diagnostic on time_range.to, got %v", diags)
}
