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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func testKibanaResourceObjectTypeWithTimeouts() tftypes.Type {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": kibanaConnectionBlockType(),
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
}

func TestKibanaResource_Schema_injectsTimeoutsAttribute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	require.Contains(t, resp.Schema.Attributes, attrTimeouts)
	nested, ok := resp.Schema.Attributes[attrTimeouts].(rschema.SingleNestedAttribute)
	require.True(t, ok)
	require.Contains(t, nested.Attributes, "create")
	require.Contains(t, nested.Attributes, "read")
	require.Contains(t, nested.Attributes, "update")
	require.Contains(t, nested.Attributes, "delete")
}

func TestKibanaResource_Create_validateSpaceIDBeforeClientUnderTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		createCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, ""))

	var resp resource.CreateResponse
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.False(t, createCalled)
}

type kbTimeoutsVersionModel struct {
	ResourceTimeoutsField
	KibanaConnectionField
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	SpaceID types.String `tfsdk:"space_id"`
}

func (m kbTimeoutsVersionModel) GetID() types.String         { return m.ID }
func (m kbTimeoutsVersionModel) GetResourceID() types.String { return m.Name }
func (m kbTimeoutsVersionModel) GetSpaceID() types.String    { return m.SpaceID }

func (kbTimeoutsVersionModel) GetVersionRequirements(_ context.Context) ([]VersionRequirement, diag.Diagnostics) {
	return []VersionRequirement{{
		MinVersion:   *version.Must(version.NewVersion("8.0.0")),
		ErrorMessage: "requires Kibana 8.0.0+",
	}}, nil
}

func TestKibanaResource_Read_versionCheckTimesOutBeforeCallback(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
			time.Sleep(3 * time.Second)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	readCalled := false
	r := NewKibanaResource[kbTimeoutsVersionModel](ComponentKibana, "test_entity", KibanaResourceOptions[kbTimeoutsVersionModel]{
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":       rschema.StringAttribute{Computed: true},
					"name":     rschema.StringAttribute{Optional: true},
					"space_id": rschema.StringAttribute{Optional: true},
				},
			}
		},
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ kbTimeoutsVersionModel) (kbTimeoutsVersionModel, bool, diag.Diagnostics) {
			readCalled = true
			return kbTimeoutsVersionModel{}, false, nil
		},
		Delete: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ kbTimeoutsVersionModel) diag.Diagnostics {
			return nil
		},
		Create: func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[kbTimeoutsVersionModel]) (KibanaWriteResult[kbTimeoutsVersionModel], diag.Diagnostics) {
			return KibanaWriteResult[kbTimeoutsVersionModel]{}, nil
		},
		Update: func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[kbTimeoutsVersionModel]) (KibanaWriteResult[kbTimeoutsVersionModel], diag.Diagnostics) {
			return KibanaWriteResult[kbTimeoutsVersionModel]{}, nil
		},
		Timeouts: ResourceTimeouts{Read: 200 * time.Millisecond},
	})
	factory := newKibanaFactoryForURL(t, srv.URL)
	r.client = factory

	objType := testKibanaResourceObjectTypeWithTimeouts()
	stateValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "user1"),
		"name":              tftypes.NewValue(tftypes.String, "user1"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Raw: stateValue, Schema: schemaResp.Schema}
	var resp resource.ReadResponse
	resp.State = state
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.False(t, readCalled)
}

func TestKibanaResource_Schema_silentlyOverwritesFactoryTimeouts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sentinel := rschema.StringAttribute{Description: "sentinel"}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":         rschema.StringAttribute{Computed: true},
					"name":       rschema.StringAttribute{Optional: true},
					"space_id":   rschema.StringAttribute{Optional: true},
					attrTimeouts: sentinel,
				},
			}
		},
		Read:   testKibanaReadFuncFound,
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	var resp resource.SchemaResponse
	require.NotPanics(t, func() {
		r.Schema(ctx, resource.SchemaRequest{}, &resp)
	})
	require.Equal(t, fmt.Sprintf("%T", timeouts.AttributesAll(ctx)), fmt.Sprintf("%T", resp.Schema.Attributes[attrTimeouts]))
}
