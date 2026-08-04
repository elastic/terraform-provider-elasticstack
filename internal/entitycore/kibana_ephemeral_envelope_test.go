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

package entitycore

import (
	"context"
	"maps"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

type testKibanaEphemeralModel struct {
	KibanaConnectionField
	Value types.String `tfsdk:"value"`
}

type testKibanaEphemeralCloseState struct {
	Value string
}

func testKibanaEphemeralSchema(_ context.Context) eschema.Schema {
	return eschema.Schema{
		Attributes: map[string]eschema.Attribute{
			"value": eschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func configureKibanaEphemeral(t *testing.T, r ephemeral.EphemeralResource, factory *clients.ProviderClientFactory) {
	t.Helper()
	var cfgResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(context.Background(), ephemeral.ConfigureRequest{
		ProviderData: factory,
	}, &cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError(), "Configure must not produce errors: %v", cfgResp.Diagnostics)
}

func buildKibanaEphemeralOpenConfig(t *testing.T, schema eschema.Schema, values map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	attrTypes := make(map[string]tftypes.Type, len(schema.Attributes)+1)
	for name := range schema.Attributes {
		attrTypes[name] = tftypes.String
	}
	attrTypes["kibana_connection"] = kibanaConnectionBlockType()

	objValues := make(map[string]tftypes.Value, len(values)+1)
	maps.Copy(objValues, values)
	if _, ok := objValues["kibana_connection"]; !ok {
		objValues["kibana_connection"] = tftypes.NewValue(kibanaConnectionBlockType(), nil)
	}

	return tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, objValues),
		Schema: eschema.Schema{
			Attributes: schema.Attributes,
			Blocks: map[string]eschema.Block{
				"kibana_connection": schema.Blocks["kibana_connection"],
			},
		},
	}
}

func TestKibanaConnectionSnapshot_inEnvelopeContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	connData, diags := encodeKibanaConnection(ctx, providerschema.KibanaConnectionNullList())
	require.False(t, diags.HasError())

	restored, restoreDiags := decodeKibanaConnection(ctx, connData)
	require.False(t, restoreDiags.HasError())
	require.True(t, restored.IsNull())
}
