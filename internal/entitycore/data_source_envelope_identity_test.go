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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// kibanaDSPlainModel does NOT opt in to composite GetResourceID parsing.
type kibanaDSPlainModel struct {
	KibanaConnectionField
	ID         types.String `tfsdk:"id"`
	ResourceID types.String `tfsdk:"resource_id"`
	SpaceID    types.String `tfsdk:"space_id"`
	Result     types.String `tfsdk:"result"`
}

func (m kibanaDSPlainModel) GetID() types.String         { return m.ID }
func (m kibanaDSPlainModel) GetResourceID() types.String { return m.ResourceID }
func (m kibanaDSPlainModel) GetSpaceID() types.String    { return m.SpaceID }

// esDSOverrideModel implements WithReadResourceID.
type esDSOverrideModel struct {
	ElasticsearchConnectionField
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Result types.String `tfsdk:"result"`
}

func (m esDSOverrideModel) GetID() types.String         { return m.ID }
func (m esDSOverrideModel) GetResourceID() types.String { return m.Name }
func (esDSOverrideModel) GetReadResourceID() string     { return "override-id" }

func kibanaPlainAttrs() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"id":          dsschema.StringAttribute{Optional: true, Computed: true},
		"resource_id": dsschema.StringAttribute{Optional: true},
		"space_id":    dsschema.StringAttribute{Optional: true, Computed: true},
		"result":      dsschema.StringAttribute{Computed: true},
	}
}

func kibanaPlainSchema() dsschema.Schema {
	return dsschema.Schema{
		Blocks:     map[string]dsschema.Block{"kibana_connection": providerschema.GetKbFWConnectionBlock()},
		Attributes: kibanaPlainAttrs(),
	}
}

func strVal(s *string) tftypes.Value {
	if s == nil {
		return tftypes.NewValue(tftypes.String, nil)
	}
	return tftypes.NewValue(tftypes.String, *s)
}

func ptr(s string) *string { return &s }

// readKibanaPlain runs a Kibana data source read with the given config values
// (nil means null; use unknownSpace for an unknown space_id).
func readKibanaPlain(
	t *testing.T,
	opts KibanaDataSourceOptions[kibanaDSPlainModel],
	configure bool,
	id, resourceID, spaceID *string,
	unknownSpace bool,
) *datasource.ReadResponse {
	t.Helper()
	ds := NewKibanaDataSource[kibanaDSPlainModel](ComponentKibana, "test_entity", opts)
	if configure {
		configureDataSource(t, ds, newKibanaFactoryMinimal(t))
	}
	schema := kibanaPlainSchema()
	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id": tftypes.String, "resource_id": tftypes.String, "space_id": tftypes.String,
		"result": tftypes.String, "kibana_connection": connBlockType,
	}}
	space := strVal(spaceID)
	if unknownSpace {
		space = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	}
	req := datasource.ReadRequest{Config: tfsdk.Config{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"id": strVal(id), "resource_id": strVal(resourceID), "space_id": space,
			"result": tftypes.NewValue(tftypes.String, nil), "kibana_connection": tftypes.NewValue(connBlockType, nil),
		}),
		Schema: schema,
	}}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), req, resp)
	return resp
}

func kibanaPlainOpts(read func(ctx context.Context, c *clients.KibanaScopedClient, resourceID, spaceID string, m kibanaDSPlainModel) (kibanaDSPlainModel, bool, diag.Diagnostics)) KibanaDataSourceOptions[kibanaDSPlainModel] {
	return KibanaDataSourceOptions[kibanaDSPlainModel]{
		Schema: func(_ context.Context) dsschema.Schema { return dsschema.Schema{Attributes: kibanaPlainAttrs()} },
		Read:   read,
	}
}

func recordingRead(gotResource, gotSpace *string, calls *int) func(context.Context, *clients.KibanaScopedClient, string, string, kibanaDSPlainModel) (kibanaDSPlainModel, bool, diag.Diagnostics) {
	return func(_ context.Context, _ *clients.KibanaScopedClient, resourceID, spaceID string, m kibanaDSPlainModel) (kibanaDSPlainModel, bool, diag.Diagnostics) {
		*gotResource, *gotSpace = resourceID, spaceID
		*calls++
		m.Result = types.StringValue("read")
		return m, true, nil
	}
}

func TestKibanaDataSource_Read_postRead_invokedOnceOnFound(t *testing.T) {
	var postCalls int
	var postClient *clients.KibanaScopedClient
	var postModel kibanaDSPlainModel
	var r, s string
	var calls int
	opts := kibanaPlainOpts(recordingRead(&r, &s, &calls))
	opts.PostRead = func(_ context.Context, c *clients.KibanaScopedClient, m kibanaDSPlainModel) diag.Diagnostics {
		postCalls++
		postClient, postModel = c, m
		return nil
	}
	resp := readKibanaPlain(t, opts, true, nil, ptr("x"), nil, false)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, 1, postCalls)
	require.NotNil(t, postClient)
	require.Equal(t, "read", postModel.Result.ValueString())

	var got types.String
	require.False(t, resp.State.GetAttribute(context.Background(), path.Root("result"), &got).HasError())
	require.Equal(t, "read", got.ValueString(), "state must be set before PostRead")
}

func TestKibanaDataSource_Read_postRead_errorSurfaces_stateStillSet(t *testing.T) {
	var r, s string
	var calls int
	opts := kibanaPlainOpts(recordingRead(&r, &s, &calls))
	opts.PostRead = func(context.Context, *clients.KibanaScopedClient, kibanaDSPlainModel) diag.Diagnostics {
		return diag.Diagnostics{diag.NewErrorDiagnostic("post read failed", "boom")}
	}
	resp := readKibanaPlain(t, opts, true, nil, ptr("x"), nil, false)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "post read failed", resp.Diagnostics.Errors()[0].Summary())

	// Pinned behavior: PostRead runs after state is set, so state persists.
	var got types.String
	require.False(t, resp.State.GetAttribute(context.Background(), path.Root("result"), &got).HasError())
	require.Equal(t, "read", got.ValueString())
}

func TestElasticsearchDataSource_Read_postRead_errorSurfaces_stateStillSet(t *testing.T) {
	ds := NewElasticsearchDataSource[esDSIdentityModel](ComponentElasticsearch, "test_entity", ElasticsearchDataSourceOptions[esDSIdentityModel]{
		Schema: func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{Attributes: map[string]dsschema.Attribute{
				"name": dsschema.StringAttribute{Required: true}, "id": dsschema.StringAttribute{Computed: true}, "result": dsschema.StringAttribute{Computed: true},
			}}
		},
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, m esDSIdentityModel) (esDSIdentityModel, bool, diag.Diagnostics) {
			m.Result = types.StringValue("read")
			return m, true, nil
		},
		PostRead: func(context.Context, *clients.ElasticsearchScopedClient, esDSIdentityModel) diag.Diagnostics {
			return diag.Diagnostics{diag.NewErrorDiagnostic("post read failed", "boom")}
		},
	})
	configureElasticsearchDataSource(t, ds, newElasticsearchFactoryMinimal(t))
	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{"elasticsearch_connection": providerschema.GetEsFWConnectionBlock()},
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{Required: true}, "id": dsschema.StringAttribute{Computed: true}, "result": dsschema.StringAttribute{Computed: true},
		},
	}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), buildReadRequestForElasticsearchSchema(schema), &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "post read failed", resp.Diagnostics.Errors()[0].Summary())

	var got types.String
	require.False(t, resp.State.GetAttribute(context.Background(), path.Root("result"), &got).HasError())
	require.Equal(t, "read", got.ValueString())
}

func TestKibanaDataSource_Read_noOptIn_slashResourceIDNotSplit(t *testing.T) {
	var r, s string
	var calls int
	resp := readKibanaPlain(t, kibanaPlainOpts(recordingRead(&r, &s, &calls)), true, nil, ptr("a/b"), nil, false)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "a/b", r)
	require.Equal(t, "", s)
}

func TestKibanaDataSource_Read_compositeIDAndExplicitSpace(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		resourceID   *string
		space        *string
		unknownSpace bool
		wantResource string
		wantSpace    string
	}{
		{name: "explicit space wins", id: "custom/skill", space: ptr("explicit"), wantResource: "skill", wantSpace: "explicit"},
		{name: "empty space falls through", id: "custom/skill", space: ptr(""), wantResource: "skill", wantSpace: "custom"},
		{name: "unknown space falls through", id: "custom/skill", unknownSpace: true, wantResource: "skill", wantSpace: "custom"},
		{name: "leading slash", id: "/skill", wantResource: "skill", wantSpace: ""},
		{name: "trailing slash falls back to resource_id", id: "space/", resourceID: ptr("fallback"), wantResource: "fallback", wantSpace: ""},
		{name: "multiple slashes", id: "a/b/c", wantResource: "b/c", wantSpace: "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r, s string
			var calls int
			resp := readKibanaPlain(t, kibanaPlainOpts(recordingRead(&r, &s, &calls)), true, ptr(tt.id), tt.resourceID, tt.space, tt.unknownSpace)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			require.Equal(t, 1, calls)
			require.Equal(t, tt.wantResource, r)
			require.Equal(t, tt.wantSpace, s)
		})
	}
}

func TestKibanaDataSource_Read_nilReadOption_diagnosticNoPanic(t *testing.T) {
	opts := kibanaPlainOpts(nil)
	var resp *datasource.ReadResponse
	require.NotPanics(t, func() { resp = readKibanaPlain(t, opts, true, nil, ptr("x"), nil, false) })
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "envelope configuration error")
}

func TestKibanaDataSource_Read_clientFailure_readNotCalled(t *testing.T) {
	var r, s string
	var calls int
	// Not configured: GetKibanaClient fails.
	resp := readKibanaPlain(t, kibanaPlainOpts(recordingRead(&r, &s, &calls)), false, nil, ptr("x"), nil, false)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, 0, calls)
}

func TestDataSourceNotFoundDiagnostic(t *testing.T) {
	d := dataSourceNotFoundDiagnostic(ComponentKibana, "thing", "id1", "space1")
	require.Len(t, d, 1)
	require.Equal(t, "kibana_thing not found", d.Errors()[0].Summary())
	require.Equal(t, `kibana_thing "id1" in space "space1" was not found`, d.Errors()[0].Detail())

	d = dataSourceNotFoundDiagnostic(ComponentKibana, "thing", "id1", "")
	require.Equal(t, `kibana_thing "id1" was not found`, d.Errors()[0].Detail())

	d = dataSourceNotFoundDiagnostic(ComponentElasticsearch, "thing", "id2", "")
	require.Equal(t, "elasticsearch_thing not found", d.Errors()[0].Summary())
	require.Equal(t, `elasticsearch_thing "id2" was not found`, d.Errors()[0].Detail())
}

func TestElasticsearchDataSource_Read_withReadResourceIDOverride(t *testing.T) {
	var got string
	ds := NewElasticsearchDataSource[esDSOverrideModel](ComponentElasticsearch, "test_entity", ElasticsearchDataSourceOptions[esDSOverrideModel]{
		Schema: func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{Attributes: map[string]dsschema.Attribute{
				"name": dsschema.StringAttribute{Required: true}, "id": dsschema.StringAttribute{Computed: true}, "result": dsschema.StringAttribute{Computed: true},
			}}
		},
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, resourceID string, m esDSOverrideModel) (esDSOverrideModel, bool, diag.Diagnostics) {
			got = resourceID
			return m, true, nil
		},
	})
	configureElasticsearchDataSource(t, ds, newElasticsearchFactoryMinimal(t))
	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{"elasticsearch_connection": providerschema.GetEsFWConnectionBlock()},
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{Required: true}, "id": dsschema.StringAttribute{Computed: true}, "result": dsschema.StringAttribute{Computed: true},
		},
	}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), buildReadRequestForElasticsearchSchema(schema), &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "override-id", got)
}
