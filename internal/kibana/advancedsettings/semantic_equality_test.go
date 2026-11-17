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

package advancedsettings

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reformattingProvider struct{}

func (reformattingProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "test"
}

func (reformattingProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {
}

func (reformattingProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

func (reformattingProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }

func (reformattingProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{func() resource.Resource { return reformattingResource{} }}
}

// reformattingResource exposes the real json_settings attribute and always
// reads back the same compact, key-sorted JSON object for the key "setting".
type reformattingResource struct{}

const reformattedValue = `{"a":[1,2],"b":1}`

func (reformattingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reformatting"
}

func (reformattingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"json_settings": getSchema(ctx).Attributes["json_settings"],
	}}
}

func (reformattingResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
}

func (reformattingResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

func (reformattingResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (reformattingResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	readBack := types.MapValueMust(jsontypes.NormalizedType{}, map[string]attr.Value{
		"setting": jsontypes.NewNormalizedValue(reformattedValue),
	})
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("json_settings"), readBack)...)
}

func TestJSONSettings_mapElementsApplySemanticEquality(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		priorValue string
		want       string
	}{
		"formatting and key order differences keep the prior value": {
			priorValue: `{ "b": 1,   "a": [1, 2] }`,
			want:       `{ "b": 1,   "a": [1, 2] }`,
		},
		"a different value replaces the prior value": {
			priorValue: `{ "b": 2,   "a": [1, 2] }`,
			want:       reformattedValue,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, readBackSetting(t, tt.priorValue))
		})
	}
}

func readBackSetting(t *testing.T, priorValue string) string {
	t.Helper()
	ctx := t.Context()
	server, err := providerserver.NewProtocol6WithError(reformattingProvider{})()
	require.NoError(t, err)
	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	stateType := schemaResp.ResourceSchemas["test_reformatting"].ValueType()

	priorState, err := tfprotov6.NewDynamicValue(stateType, tftypes.NewValue(stateType, map[string]tftypes.Value{
		"json_settings": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"setting": tftypes.NewValue(tftypes.String, priorValue),
		}),
	}))
	require.NoError(t, err)

	readResp, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: "test_reformatting", CurrentState: &priorState})
	require.NoError(t, err)
	require.Empty(t, readResp.Diagnostics)
	newState, err := readResp.NewState.Unmarshal(stateType)
	require.NoError(t, err)

	var newAttrs map[string]tftypes.Value
	require.NoError(t, newState.As(&newAttrs))
	var settings map[string]tftypes.Value
	require.NoError(t, newAttrs["json_settings"].As(&settings))
	var got string
	require.NoError(t, settings["setting"].As(&got))
	return got
}
