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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// testModel embeds KibanaConnectionField for envelope tests.
type testModel struct {
	KibanaConnectionField
	ID types.String `tfsdk:"id"`
}

func getTestSchema() dsschema.Schema {
	return dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func testReadFunc(_ context.Context, _ *clients.KibanaScopedClient, model testModel) (testModel, diag.Diagnostics) {
	model.ID = types.StringValue("result")
	return model, nil
}

func kibanaConnectionBlockType() tftypes.Type {
	nestedObjType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key":      tftypes.String,
			"bearer_token": tftypes.String,
			"username":     tftypes.String,
			"password":     tftypes.String,
			"endpoints":    tftypes.List{ElementType: tftypes.String},
			"ca_certs":     tftypes.List{ElementType: tftypes.String},
			"insecure":     tftypes.Bool,
		},
	}
	return tftypes.List{ElementType: nestedObjType}
}

func TestNewKibanaDataSource_typeAssertions(t *testing.T) {
	t.Parallel()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)
	require.NotNil(t, ds)
	require.Implements(t, (*datasource.DataSource)(nil), ds)
	require.Implements(t, (*datasource.DataSourceWithConfigure)(nil), ds)
}

func TestNewElasticsearchDataSource_typeAssertions(t *testing.T) {
	t.Parallel()
	ds := NewElasticsearchDataSource[struct {
		ElasticsearchConnectionField
	}](ComponentElasticsearch, "test_entity", func() dsschema.Schema {
		return dsschema.Schema{}
	}, func(_ context.Context, _ *clients.ElasticsearchScopedClient, model struct {
		ElasticsearchConnectionField
	}) (struct {
		ElasticsearchConnectionField
	}, diag.Diagnostics) {
		return model, nil
	})
	require.NotNil(t, ds)
	require.Implements(t, (*datasource.DataSource)(nil), ds)
	require.Implements(t, (*datasource.DataSourceWithConfigure)(nil), ds)
}

func TestNewKibanaDataSource_schemaInjection(t *testing.T) {
	t.Parallel()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "kibana_connection")
	require.Contains(t, resp.Schema.Attributes, "id")
}

func TestNewElasticsearchDataSource_schemaInjection(t *testing.T) {
	t.Parallel()
	ds := NewElasticsearchDataSource[struct {
		ElasticsearchConnectionField
	}](ComponentElasticsearch, "test_entity", func() dsschema.Schema {
		return dsschema.Schema{}
	}, func(_ context.Context, _ *clients.ElasticsearchScopedClient, model struct {
		ElasticsearchConnectionField
	}) (struct {
		ElasticsearchConnectionField
	}, diag.Diagnostics) {
		return model, nil
	})

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "elasticsearch_connection")
}

func TestNewKibanaDataSource_schemaDefensiveClone(t *testing.T) {
	t.Parallel()
	originalSchema := getTestSchema()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", func() dsschema.Schema {
		return originalSchema
	}, testReadFunc)

	// Call Schema twice; the second call should still inject a fresh block.
	var resp1 datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp1)
	require.False(t, resp1.Diagnostics.HasError())

	var resp2 datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp2)
	require.False(t, resp2.Diagnostics.HasError())

	// The original schema's Blocks map should remain nil because the factory
	// is called once per Schema invocation and the constructor clones the map.
	require.Nil(t, originalSchema.Blocks)
}

func TestNewKibanaDataSource_Configure(t *testing.T) {
	ctx := context.Background()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	t.Run("nil_provider_data", func(t *testing.T) {
		t.Parallel()
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
			ProviderData: nil,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("valid_factory", func(t *testing.T) {
		t.Parallel()
		f := nonNilTestFactory()
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
			ProviderData: f,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("invalid_provider_data", func(t *testing.T) {
		t.Parallel()
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
			ProviderData: "wrong-type",
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestNewKibanaDataSource_Metadata(t *testing.T) {
	t.Parallel()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: testProviderTypeName,
	}, &resp)

	require.Equal(t, "elasticstack_kibana_test_entity", resp.TypeName)
}

func TestNewKibanaDataSource_Read_unconfiguredFactory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	// Configure with a zero-value factory (no default client).
	var cfgResp datasource.ConfigureResponse
	ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
		ProviderData: new(clients.ProviderClientFactory),
	}, &cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError())

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, nil),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})

	fullSchema := getTestSchema()
	fullSchema.Blocks = map[string]dsschema.Block{
		"kibana_connection": providerschema.GetKbFWConnectionBlock(),
	}

	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    objValue,
			Schema: fullSchema,
		},
	}

	var resp datasource.ReadResponse
	ds.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
}

func TestKibanaConnectionField_configDecode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}

	config := tfsdk.Config{
		Raw:    objValue,
		Schema: schema,
	}

	var model struct {
		KibanaConnectionField
	}
	diags := config.Get(ctx, &model)
	require.False(t, diags.HasError())
	require.True(t, model.KibanaConnection.IsNull())
}

func TestKibanaConnectionField_stateRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"kibana_connection": connBlockType,
		},
	}

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}

	state := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: schema,
	}

	model := struct {
		KibanaConnectionField
	}{
		KibanaConnectionField: KibanaConnectionField{
			KibanaConnection: providerschema.KibanaConnectionNullList(),
		},
	}

	diags := state.Set(ctx, &model)
	require.False(t, diags.HasError())

	var result struct {
		KibanaConnectionField
	}
	diags = state.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.True(t, result.KibanaConnection.IsNull())
}
