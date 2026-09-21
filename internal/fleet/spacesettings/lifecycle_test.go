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

package spacesettings

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

const settingsPath = "/api/fleet/space_settings"

type fakeKibana struct {
	*httptest.Server
	mu              sync.Mutex
	requests        []string
	settingsHandler http.HandlerFunc
}

func newFakeKibana(t *testing.T, stackVersion string, settingsHandler http.HandlerFunc) *fakeKibana {
	t.Helper()

	fake := &fakeKibana{settingsHandler: settingsHandler}
	fake.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		fake.requests = append(fake.requests, r.Method+" "+r.URL.Path)
		fake.mu.Unlock()

		if r.URL.Path == "/api/status" {
			writeJSON(w, http.StatusOK, fmt.Sprintf(`{"version":{"number":%q,"build_flavor":"traditional"}}`, stackVersion))
			return
		}
		if strings.HasSuffix(r.URL.Path, settingsPath) {
			fake.settingsHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(fake.Close)
	return fake
}

func (f *fakeKibana) requestsExcludingStatus() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var recorded []string
	for _, request := range f.requests {
		if !strings.HasSuffix(request, " /api/status") {
			recorded = append(recorded, request)
		}
	}
	return recorded
}

func (f *fakeKibana) settingsRequests() []string {
	var methods []string
	for _, request := range f.requestsExcludingStatus() {
		method, path, _ := strings.Cut(request, " ")
		if strings.HasSuffix(path, settingsPath) {
			methods = append(methods, method)
		}
	}
	return methods
}

func respondWithPrefixes(prefixes string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, fmt.Sprintf(`{"item":{"allowed_namespace_prefixes":%s}}`, prefixes))
	}
}

func newConfiguredResource(t *testing.T, providerEndpoint string) resource.Resource {
	t.Helper()

	factory, diags := clients.NewProviderClientFactoryFromFramework(context.Background(), config.ProviderConfiguration{
		Kibana: []config.KibanaConnection{{
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{types.StringValue(providerEndpoint)}),
			Username:  types.StringValue("elastic"),
			Password:  types.StringValue("password"),
			CACerts:   types.ListNull(types.StringType),
		}},
	}, "test")
	require.False(t, diags.HasError(), "%v", diags)

	res := NewResource()
	configureResp := &resource.ConfigureResponse{}
	res.(resource.ResourceWithConfigure).Configure(context.Background(), resource.ConfigureRequest{ProviderData: factory}, configureResp)
	require.False(t, configureResp.Diagnostics.HasError(), "%v", configureResp.Diagnostics)
	return res
}

func resourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()

	resp := resource.SchemaResponse{}
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	return resp
}

func planFor(t *testing.T, model spaceSettingsModel) tfsdk.Plan {
	t.Helper()

	s := resourceSchema(t).Schema
	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
	timeoutsType := s.Attributes["timeouts"].GetType().(timeouts.Type)
	model.Timeouts = timeouts.Value{Object: types.ObjectNull(timeoutsType.AttrTypes)}
	diags := plan.Set(context.Background(), &model)
	require.False(t, diags.HasError(), "%v", diags)
	return plan
}

func planModel(spaceID string, prefixes ...string) spaceSettingsModel {
	return spaceSettingsModel{
		ID:                       types.StringUnknown(),
		KibanaConnection:         providerschema.KibanaConnectionNullList(),
		SpaceID:                  types.StringValue(spaceID),
		AllowedNamespacePrefixes: prefixList(prefixes...),
		ManagedBy:                types.StringUnknown(),
	}
}

func stateFor(t *testing.T, spaceID string, prefixes ...string) tfsdk.State {
	t.Helper()

	model := planModel(spaceID, prefixes...)
	model.ID = types.StringValue(spaceID)
	model.ManagedBy = types.StringNull()
	plan := planFor(t, model)
	return tfsdk.State(plan)
}

func scopedConnectionTo(t *testing.T, endpoint string) types.List {
	t.Helper()

	list, diags := types.ListValueFrom(context.Background(), providerschema.KibanaConnectionObjectType(), []config.KibanaConnection{{
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{types.StringValue(endpoint)}),
		Username:  types.StringValue("elastic"),
		Password:  types.StringValue("password"),
		CACerts:   types.ListNull(types.StringType),
	}})
	require.False(t, diags.HasError(), "%v", diags)
	return list
}

func createResource(t *testing.T, res resource.Resource, model spaceSettingsModel) *resource.CreateResponse {
	t.Helper()

	plan := planFor(t, model)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema, Raw: tftypes.NewValue(plan.Schema.Type().TerraformType(context.Background()), nil)}}
	res.Create(context.Background(), resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, resp)
	return resp
}

func TestCreate_serverOlderThan910FailsWithoutSettingsCall(t *testing.T) {
	kibana := newFakeKibana(t, "9.0.0", respondWithPrefixes(`["team_a"]`))
	res := newConfiguredResource(t, kibana.URL)

	resp := createResource(t, res, planModel("team-a", "team_a"))

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, fmt.Sprint(resp.Diagnostics), "9.1.0")
	require.Empty(t, kibana.settingsRequests())
}

func TestRead_serverOlderThan910FailsWithoutSettingsCall(t *testing.T) {
	kibana := newFakeKibana(t, "9.0.0", respondWithPrefixes(`["team_a"]`))
	res := newConfiguredResource(t, kibana.URL)
	state := stateFor(t, "team-a", "team_a")

	resp := &resource.ReadResponse{State: state}
	res.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, fmt.Sprint(resp.Diagnostics), "9.1.0")
	require.Empty(t, kibana.settingsRequests())
}

func TestCreate_server910ProceedsWithSettingsWrite(t *testing.T) {
	kibana := newFakeKibana(t, "9.1.0", respondWithPrefixes(`["team_a"]`))
	res := newConfiguredResource(t, kibana.URL)

	resp := createResource(t, res, planModel("team-a", "team_a"))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, []string{http.MethodPut, http.MethodGet}, kibana.settingsRequests())
}

func TestDelete_serverOlderThan910FailsWithoutSettingsCall(t *testing.T) {
	kibana := newFakeKibana(t, "9.0.0", respondWithPrefixes(`[]`))
	res := newConfiguredResource(t, kibana.URL)
	state := stateFor(t, "team-a", "team_a")

	resp := &resource.DeleteResponse{State: state}
	res.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, fmt.Sprint(resp.Diagnostics), "9.1.0")
	require.Empty(t, kibana.settingsRequests())
}

func TestCreate_withoutKibanaConnectionUsesProviderClient(t *testing.T) {
	providerKibana := newFakeKibana(t, "9.1.0", respondWithPrefixes(`["team_a"]`))
	otherKibana := newFakeKibana(t, "9.1.0", respondWithPrefixes(`["team_a"]`))
	res := newConfiguredResource(t, providerKibana.URL)

	resp := createResource(t, res, planModel("team-a", "team_a"))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.NotEmpty(t, providerKibana.settingsRequests())
	require.Empty(t, otherKibana.settingsRequests())
}

func TestCreate_withKibanaConnectionUsesScopedClient(t *testing.T) {
	providerKibana := newFakeKibana(t, "9.1.0", respondWithPrefixes(`["team_a"]`))
	scopedKibana := newFakeKibana(t, "9.1.0", respondWithPrefixes(`["team_a"]`))
	res := newConfiguredResource(t, providerKibana.URL)
	model := planModel("team-a", "team_a")
	model.KibanaConnection = scopedConnectionTo(t, scopedKibana.URL)

	resp := createResource(t, res, model)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, []string{http.MethodPut, http.MethodGet}, scopedKibana.settingsRequests())
	require.Empty(t, providerKibana.settingsRequests())
}

func TestCreate_missingSpaceSurfacesAPIError(t *testing.T) {
	kibana := newFakeKibana(t, "9.1.0", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"statusCode":404,"error":"Not Found","message":"Saved object [space/missing] not found"}`)
	})
	res := newConfiguredResource(t, kibana.URL)

	resp := createResource(t, res, planModel("missing", "team_a"))

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, fmt.Sprint(resp.Diagnostics), "Saved object [space/missing] not found")
	require.True(t, resp.State.Raw.IsNull())
}

func TestCreate_insufficientPrivilegesSurfacesAPIError(t *testing.T) {
	kibana := newFakeKibana(t, "9.1.0", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusForbidden, `{"statusCode":403,"error":"Forbidden","message":"missing fleet-settings-all privilege"}`)
	})
	res := newConfiguredResource(t, kibana.URL)

	resp := createResource(t, res, planModel("team-a", "team_a"))

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, fmt.Sprint(resp.Diagnostics), "missing fleet-settings-all privilege")
	require.True(t, resp.State.Raw.IsNull())
}

func TestUpdate_insufficientPrivilegesSurfacesAPIError(t *testing.T) {
	kibana := newFakeKibana(t, "9.1.0", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusForbidden, `{"statusCode":403,"error":"Forbidden","message":"missing fleet-settings-all privilege"}`)
	})
	res := newConfiguredResource(t, kibana.URL)
	prior := stateFor(t, "team-a", "team_a")
	plan := planFor(t, planModel("team-a", "team_a", "shared"))

	resp := &resource.UpdateResponse{State: prior}
	res.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: prior, Config: tfsdk.Config(plan)}, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, fmt.Sprint(resp.Diagnostics), "missing fleet-settings-all privilege")
}

func TestCreate_doesNotCreateOrCheckSpace(t *testing.T) {
	kibana := newFakeKibana(t, "9.1.0", respondWithPrefixes(`["team_a"]`))
	res := newConfiguredResource(t, kibana.URL)

	resp := createResource(t, res, planModel("missing", "team_a"))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, []string{
		"PUT /s/missing/api/fleet/space_settings",
		"GET /s/missing/api/fleet/space_settings",
	}, kibana.requestsExcludingStatus())
}
