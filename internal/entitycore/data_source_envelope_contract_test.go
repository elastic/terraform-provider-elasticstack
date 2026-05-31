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

type kibanaDSIdentityModel struct {
	KibanaConnectionField
	ID      types.String `tfsdk:"id"`
	SkillID types.String `tfsdk:"skill_id"`
	SpaceID types.String `tfsdk:"space_id"`
	Result  types.String `tfsdk:"result"`
}

func (m kibanaDSIdentityModel) GetID() types.String         { return m.ID }
func (m kibanaDSIdentityModel) GetResourceID() types.String { return m.SkillID }
func (m kibanaDSIdentityModel) GetSpaceID() types.String    { return m.SpaceID }

type esDSIdentityModel struct {
	ElasticsearchConnectionField
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Result types.String `tfsdk:"result"`
}

func (m esDSIdentityModel) GetID() types.String         { return m.ID }
func (m esDSIdentityModel) GetResourceID() types.String { return m.Name }

func TestElasticsearchDataSource_Read_invalidIdentity(t *testing.T) {
	ctx := context.Background()

	ds := NewElasticsearchDataSource[esDSIdentityModel](ComponentElasticsearch, "test_entity", ElasticsearchDataSourceOptions[esDSIdentityModel]{
		Schema: func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{Attributes: map[string]dsschema.Attribute{
				"name":   dsschema.StringAttribute{Optional: true, Computed: true},
				"id":     dsschema.StringAttribute{Computed: true},
				"result": dsschema.StringAttribute{Computed: true},
			}}
		},
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model esDSIdentityModel) (esDSIdentityModel, bool, diag.Diagnostics) {
			return model, true, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchDataSource(t, ds, factory)

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{"elasticsearch_connection": providerschema.GetEsFWConnectionBlock()},
		Attributes: map[string]dsschema.Attribute{
			"name":   dsschema.StringAttribute{Optional: true, Computed: true},
			"id":     dsschema.StringAttribute{Computed: true},
			"result": dsschema.StringAttribute{Computed: true},
		},
	}
	connBlockType := elasticsearchConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"elasticsearch_connection": connBlockType,
			"name":                     tftypes.String,
			"id":                       tftypes.String,
			"result":                   tftypes.String,
		},
	}
	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
				"elasticsearch_connection": tftypes.NewValue(connBlockType, nil),
				"name":                     tftypes.NewValue(tftypes.String, nil),
				"id":                       tftypes.NewValue(tftypes.String, nil),
				"result":                   tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: schema,
		},
	}

	var resp datasource.ReadResponse
	ds.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
}

func TestElasticsearchDataSource_Read_notFound(t *testing.T) {
	ctx := context.Background()

	ds := NewElasticsearchDataSource[esDSIdentityModel](ComponentElasticsearch, "test_entity", ElasticsearchDataSourceOptions[esDSIdentityModel]{
		Schema: func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{Attributes: map[string]dsschema.Attribute{
				"name":   dsschema.StringAttribute{Required: true},
				"id":     dsschema.StringAttribute{Computed: true},
				"result": dsschema.StringAttribute{Computed: true},
			}}
		},
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, resourceID string, model esDSIdentityModel) (esDSIdentityModel, bool, diag.Diagnostics) {
			model.Name = types.StringValue(resourceID)
			return model, false, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchDataSource(t, ds, factory)

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{"elasticsearch_connection": providerschema.GetEsFWConnectionBlock()},
		Attributes: map[string]dsschema.Attribute{
			"name":   dsschema.StringAttribute{Required: true},
			"id":     dsschema.StringAttribute{Computed: true},
			"result": dsschema.StringAttribute{Computed: true},
		},
	}
	req := buildReadRequestForElasticsearchSchema(schema)

	var resp datasource.ReadResponse
	ds.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "elasticsearch_test_entity not found")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `elasticsearch_test_entity "test" was not found`)
}

func TestElasticsearchDataSource_Read_postRead(t *testing.T) {
	ctx := context.Background()

	postReadCalled := false
	ds := NewElasticsearchDataSource[esDSIdentityModel](ComponentElasticsearch, "test_entity", ElasticsearchDataSourceOptions[esDSIdentityModel]{
		Schema: func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{Attributes: map[string]dsschema.Attribute{
				"name":   dsschema.StringAttribute{Required: true},
				"id":     dsschema.StringAttribute{Computed: true},
				"result": dsschema.StringAttribute{Computed: true},
			}}
		},
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, resourceID string, model esDSIdentityModel) (esDSIdentityModel, bool, diag.Diagnostics) {
			model.ID = types.StringValue("cluster/" + resourceID)
			model.Result = types.StringValue("read")
			return model, true, nil
		},
		PostRead: func(_ context.Context, _ *clients.ElasticsearchScopedClient, model esDSIdentityModel) diag.Diagnostics {
			postReadCalled = true
			require.Equal(t, "read", model.Result.ValueString())
			return nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchDataSource(t, ds, factory)

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{"elasticsearch_connection": providerschema.GetEsFWConnectionBlock()},
		Attributes: map[string]dsschema.Attribute{
			"name":   dsschema.StringAttribute{Required: true},
			"id":     dsschema.StringAttribute{Computed: true},
			"result": dsschema.StringAttribute{Computed: true},
		},
	}
	req := buildReadRequestForElasticsearchSchema(schema)

	var resp datasource.ReadResponse
	resp.State = tfsdk.State{Schema: schema}
	ds.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, postReadCalled)
}

func TestKibanaDataSource_Read_compositeResourceID(t *testing.T) {
	ctx := context.Background()

	var gotResourceID, gotSpaceID string
	ds := NewKibanaDataSource[kibanaDSIdentityModel](ComponentKibana, "test_entity", KibanaDataSourceOptions[kibanaDSIdentityModel]{
		Schema: func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{Attributes: map[string]dsschema.Attribute{
				"skill_id": dsschema.StringAttribute{Required: true},
				"space_id": dsschema.StringAttribute{Optional: true, Computed: true},
				"id":       dsschema.StringAttribute{Computed: true},
				"result":   dsschema.StringAttribute{Computed: true},
			}}
		},
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, resourceID, spaceID string, model kibanaDSIdentityModel) (kibanaDSIdentityModel, bool, diag.Diagnostics) {
			gotResourceID = resourceID
			gotSpaceID = spaceID
			model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}).String())
			return model, true, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureDataSource(t, ds, factory)

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{"kibana_connection": providerschema.GetKbFWConnectionBlock()},
		Attributes: map[string]dsschema.Attribute{
			"skill_id": dsschema.StringAttribute{Required: true},
			"space_id": dsschema.StringAttribute{Optional: true, Computed: true},
			"id":       dsschema.StringAttribute{Computed: true},
			"result":   dsschema.StringAttribute{Computed: true},
		},
	}
	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"skill_id":          tftypes.String,
			"space_id":          tftypes.String,
			"id":                tftypes.String,
			"result":            tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
				"skill_id":          tftypes.NewValue(tftypes.String, "custom/my-skill"),
				"space_id":          tftypes.NewValue(tftypes.String, nil),
				"id":                tftypes.NewValue(tftypes.String, nil),
				"result":            tftypes.NewValue(tftypes.String, nil),
				"kibana_connection": tftypes.NewValue(connBlockType, nil),
			}),
			Schema: schema,
		},
	}

	var resp datasource.ReadResponse
	resp.State = tfsdk.State{Schema: schema}
	ds.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "my-skill", gotResourceID)
	require.Equal(t, "custom", gotSpaceID)
}

func TestKibanaDataSource_Read_notFound_skipsPostRead(t *testing.T) {
	ctx := context.Background()

	postReadCalled := false
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", KibanaDataSourceOptions[testModel]{
		Schema: getTestSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testModel) (testModel, bool, diag.Diagnostics) {
			return model, false, nil
		},
		PostRead: func(_ context.Context, _ *clients.KibanaScopedClient, _ testModel) diag.Diagnostics {
			postReadCalled = true
			return nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureDataSource(t, ds, factory)

	schemaWithConn := getTestSchema(context.Background())
	schemaWithConn.Blocks = map[string]dsschema.Block{
		"kibana_connection": providerschema.GetKbFWConnectionBlock(),
	}
	req := buildReadRequestForSchema(schemaWithConn)

	var resp datasource.ReadResponse
	ds.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	var found bool
	for _, e := range resp.Diagnostics.Errors() {
		if e.Summary() == "kibana_test_entity not found" {
			found = true
		}
	}
	require.True(t, found, "expected not-found diagnostic, got: %v", resp.Diagnostics.Errors())
	require.False(t, postReadCalled)
}
