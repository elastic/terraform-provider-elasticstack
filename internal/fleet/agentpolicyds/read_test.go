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

package agentpolicyds

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestRead_ConfiguredPolicyIDPlaceholderIsLookedUp(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		if !strings.HasSuffix(r.URL.Path, "/agent_policies/_") {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"not found"}`)
			return
		}
		fmt.Fprint(w, fleetPolicyGetBody("_", "underscore"))
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "ignored-envelope-id", "", agentPolicyModel{
		PolicyID: types.StringValue("_"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, "/api/fleet/agent_policies/_", gotPath)
	require.Equal(t, "_", got.ID.ValueString())
}

func TestRead_PopulatesIDOnByIDLookup(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.HasSuffix(r.URL.Path, "/agent_policies/abc123") {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"not found"}`)
			return
		}
		fmt.Fprint(w, fleetPolicyGetBody("abc123", "k8s-metrics-pernode"))
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", agentPolicyModel{
		PolicyID: types.StringValue("abc123"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, "abc123", got.GetID().ValueString())
}

func TestRead_ByNameLookupPopulatesPolicyID(t *testing.T) {
	const policyName = "k8s-metrics-pernode"
	const policyID = "policy-9"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, fleetPolicyListBody(1, fleetPolicyObject(policyID, policyName)))
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "default", agentPolicyModel{
		Name: types.StringValue(policyName),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, policyID, got.PolicyID.ValueString())
	require.Equal(t, policyName, got.Name.ValueString())
}

func TestRead_PopulatesIDOnByNameLookup(t *testing.T) {
	const policyName = "k8s-metrics-pernode"
	const policyID = "policy-9"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(r.URL.Path, "/agent_policies") || r.URL.Query().Get("kuery") == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"not found"}`)
			return
		}
		fmt.Fprint(w, fleetPolicyListBody(1, fleetPolicyObject(policyID, policyName)))
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "default", agentPolicyModel{
		Name: types.StringValue(policyName),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, policyID, got.GetID().ValueString())
}

func TestRead_ByIDLookupPopulatesComputedAttributes(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"item": {
				"id": "abc123",
				"name": "k8s-metrics-pernode",
				"namespace": "metrics",
				"description": "node metrics",
				"data_output_id": "data-out",
				"monitoring_output_id": "mon-out",
				"fleet_server_host_id": "host-1",
				"download_source_id": "dl-1",
				"monitoring_enabled": ["logs", "metrics"],
				"is_protected": true,
				"agent_features": [{"name": "fqdn", "enabled": true}],
				"supports_agentless": true,
				"inactivity_timeout": 120,
				"unenroll_timeout": 3600,
				"global_data_tags": [
					{"name": "env", "value": "prod"},
					{"name": "count", "value": 2}
				],
				"space_ids": ["default", "other"],
				"required_versions": [{"version": "8.15.0", "percentage": 50}],
				"advanced_settings": {
					"agent_logging_level": "debug",
					"agent_logging_to_files": true,
					"agent_logging_files_interval": "10s",
					"agent_logging_files_keepfiles": 3,
					"agent_logging_files_rotateeverybytes": 2048,
					"agent_logging_metrics_period": "15s",
					"agent_limits_go_max_procs": 2,
					"agent_download_timeout": "30m",
					"agent_download_target_directory": "/tmp/agent",
					"agent_monitoring_runtime_experimental": "process",
					"agent_features_disable_policy_change_acks_enabled": true,
					"agent_features_include_tags_in_events_enabled": false,
					"agent_internal": "key: value"
				},
				"monitoring_http": {"enabled": true, "host": "0.0.0.0", "port": 6792, "buffer": {"enabled": true}},
				"monitoring_pprof_enabled": true,
				"monitoring_diagnostics": {
					"limit": {"interval": "2m", "burst": 4},
					"uploader": {"init_dur": "2s", "max_dur": "3m", "max_retries": 5}
				},
				"status": "active",
				"revision": 1,
				"updated_at": "2020-01-01T00:00:00.000Z",
				"updated_by": "elastic"
			}
		}`)
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "ops", agentPolicyModel{
		PolicyID: types.StringValue("abc123"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, "/s/ops/api/fleet/agent_policies/abc123", gotPath)
	require.Equal(t, "abc123", got.ID.ValueString())
	require.Equal(t, "k8s-metrics-pernode", got.Name.ValueString())
	require.Equal(t, "metrics", got.Namespace.ValueString())
	require.Equal(t, "node metrics", got.Description.ValueString())
	require.Equal(t, "data-out", got.DataOutputID.ValueString())
	require.Equal(t, "mon-out", got.MonitoringOutputID.ValueString())
	require.Equal(t, "host-1", got.FleetServerHostID.ValueString())
	require.Equal(t, "dl-1", got.DownloadSourceID.ValueString())
	require.True(t, got.MonitorLogs.ValueBool())
	require.True(t, got.MonitorMetrics.ValueBool())
	require.True(t, got.IsProtected.ValueBool())
	require.Equal(t, "fqdn", got.HostNameFormat.ValueString())
	require.True(t, got.SupportsAgentless.ValueBool())
	require.Equal(t, "2m0s", got.InactivityTimeout.ValueString())
	require.Equal(t, "1h0m0s", got.UnenrollmentTimeout.ValueString())

	env := got.GlobalDataTags.Elements()["env"].(types.Object)
	require.Equal(t, "prod", env.Attributes()["string_value"].(types.String).ValueString())
	count := got.GlobalDataTags.Elements()["count"].(types.Object)
	require.InEpsilon(t, float32(2), count.Attributes()["number_value"].(types.Float32).ValueFloat32(), 0.001)

	var spaceIDs []string
	require.False(t, got.SpaceIDs.ElementsAs(t.Context(), &spaceIDs, false).HasError())
	require.ElementsMatch(t, []string{"default", "other"}, spaceIDs)
	require.Equal(t, int32(50), got.RequiredVersions.Elements()["8.15.0"].(types.Int32).ValueInt32())

	settings := got.AdvancedSettings.Attributes()
	require.Equal(t, "debug", settings["logging_level"].(types.String).ValueString())
	require.True(t, settings["logging_to_files"].(types.Bool).ValueBool())
	require.Equal(t, "10s", settings["logging_files_interval"].(customtypes.Duration).ValueString())
	require.Equal(t, int32(3), settings["logging_files_keepfiles"].(types.Int32).ValueInt32())
	require.Equal(t, int64(2048), settings["logging_files_rotateeverybytes"].(types.Int64).ValueInt64())
	require.Equal(t, "15s", settings["logging_metrics_period"].(customtypes.Duration).ValueString())
	require.Equal(t, int32(2), settings["go_max_procs"].(types.Int32).ValueInt32())
	require.Equal(t, "30m", settings["download_timeout"].(customtypes.Duration).ValueString())
	require.Equal(t, "/tmp/agent", settings["download_target_directory"].(types.String).ValueString())
	require.Equal(t, "process", settings["monitoring_runtime_experimental"].(types.String).ValueString())
	require.True(t, settings["features_disable_policy_change_acks_enabled"].(types.Bool).ValueBool())
	require.False(t, settings["features_include_tags_in_events_enabled"].(types.Bool).ValueBool())
	require.Equal(t, "key: value", settings["internal"].(customtypes.NormalizedYamlValue).ValueString())

	httpEndpoint := got.AdvancedMonitoringOptions.Attributes()["http_monitoring_endpoint"].(types.Object).Attributes()
	require.True(t, httpEndpoint["enabled"].(types.Bool).ValueBool())
	require.Equal(t, "0.0.0.0", httpEndpoint["host"].(types.String).ValueString())
	require.Equal(t, int32(6792), httpEndpoint["port"].(types.Int32).ValueInt32())
	require.True(t, httpEndpoint["buffer_enabled"].(types.Bool).ValueBool())
	require.True(t, httpEndpoint["pprof_enabled"].(types.Bool).ValueBool())

	diagnostics := got.AdvancedMonitoringOptions.Attributes()["diagnostics"].(types.Object).Attributes()
	rateLimits := diagnostics["rate_limits"].(types.Object).Attributes()
	require.Equal(t, "2m", rateLimits["interval"].(customtypes.Duration).ValueString())
	require.Equal(t, int32(4), rateLimits["burst"].(types.Int32).ValueInt32())
	uploader := diagnostics["file_uploader"].(types.Object).Attributes()
	require.Equal(t, "2s", uploader["init_duration"].(customtypes.Duration).ValueString())
	require.Equal(t, "3m", uploader["backoff_duration"].(customtypes.Duration).ValueString())
	require.Equal(t, int32(5), uploader["max_retries"].(types.Int32).ValueInt32())
}

func TestDataSourceSchema_SelectorsAreOptionalComputedAndWriteOnlyFlagsAreAbsent(t *testing.T) {
	schema := getDataSourceSchema(t.Context())

	policyID := schema.Attributes["policy_id"].(dsschema.StringAttribute)
	name := schema.Attributes["name"].(dsschema.StringAttribute)
	require.True(t, policyID.Optional)
	require.True(t, policyID.Computed)
	require.Len(t, policyID.Validators, 1)
	require.True(t, name.Optional)
	require.True(t, name.Computed)
	require.Len(t, name.Validators, 1)

	inactivity := schema.Attributes["inactivity_timeout"].(dsschema.StringAttribute)
	unenroll := schema.Attributes["unenrollment_timeout"].(dsschema.StringAttribute)
	require.Equal(t, customtypes.DurationType{}, inactivity.CustomType)
	require.Equal(t, customtypes.DurationType{}, unenroll.CustomType)
	require.True(t, inactivity.Computed)
	require.False(t, inactivity.Optional)

	_, hasSkipDestroy := schema.Attributes["skip_destroy"]
	_, hasSysMonitoring := schema.Attributes["sys_monitoring"]
	require.False(t, hasSkipDestroy)
	require.False(t, hasSysMonitoring)
}

func TestNewDataSource_TypeName(t *testing.T) {
	ds := NewDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(t.Context(), datasource.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
	require.Equal(t, "elasticstack_fleet_agent_policy", resp.TypeName)
}

func TestGetVersionRequirements_ReturnsNoEntries(t *testing.T) {
	reqs, diags := agentPolicyModel{
		PolicyID:    types.StringValue("abc123"),
		IsProtected: types.BoolValue(true),
	}.GetVersionRequirements(t.Context())

	require.False(t, diags.HasError(), diags)
	require.Empty(t, reqs)
}

func TestRead_MonitorLogsReflectsCurrentAPIState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, fleetPolicyGetBody("abc123", "named", `"monitoring_enabled":[]`))
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", agentPolicyModel{
		PolicyID:    types.StringValue("abc123"),
		MonitorLogs: types.BoolValue(true),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.False(t, got.MonitorLogs.ValueBool())
}

func TestRead_NullDescriptionStaysNull(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, fleetPolicyGetBody("abc123", "named"))
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", agentPolicyModel{
		PolicyID:    types.StringValue("abc123"),
		Description: types.StringValue("stale"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.True(t, got.Description.IsNull())
}

func TestRead_OmittedDescriptionIsEmptyString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, fleetPolicyGetBody("abc123", "named", `"description":""`))
	}))
	t.Cleanup(srv.Close)

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", agentPolicyModel{
		PolicyID: types.StringValue("abc123"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.False(t, got.Description.IsNull())
	require.Empty(t, got.Description.ValueString())
}

func TestRead_NameLookupPropagatesInconsistentPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, fleetPolicyListBody(150, fleetPolicyObject("a", "k8s-metrics-pernode")))
	}))
	t.Cleanup(srv.Close)

	_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "ops", agentPolicyModel{
		Name: types.StringValue("k8s-metrics-pernode"),
	})

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diagnosticText(diags), "Inconsistent agent policy list")
}

func TestRead_NameLookupMultipleMatches(t *testing.T) {
	const policyName = "shared-name"
	const spaceID = "ops"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, fleetPolicyListBody(2, fleetPolicyObject("a", policyName), fleetPolicyObject("b", policyName)))
	}))
	t.Cleanup(srv.Close)

	_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", spaceID, agentPolicyModel{
		Name: types.StringValue(policyName),
	})

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diagnosticText(diags), `multiple agent policies match name "shared-name" in space "ops": policy_id "a", policy_id "b"`)
}

func TestRead_NameLookupNoMatch(t *testing.T) {
	const policyName = "does-not-exist"
	const spaceID = "ops"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"page":1,"perPage":100,"total":0}`)
	}))
	t.Cleanup(srv.Close)

	_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", spaceID, agentPolicyModel{
		Name: types.StringValue(policyName),
	})

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diagnosticText(diags), `no agent policy found matching name "does-not-exist" in space "ops"`)
}

func TestGetSpaceID_ReturnsConfiguredSpace(t *testing.T) {
	require.Equal(t, "ops", agentPolicyModel{SpaceID: types.StringValue("ops")}.GetSpaceID().ValueString())
	require.True(t, agentPolicyModel{}.GetSpaceID().IsNull())
}

func TestGetResourceID_UsesPlaceholderWhenPolicyIDUnset(t *testing.T) {
	require.Equal(t, "_", agentPolicyModel{}.GetResourceID().ValueString())
	require.Equal(t, "_", agentPolicyModel{PolicyID: types.StringNull()}.GetResourceID().ValueString())
	require.Equal(t, "_", agentPolicyModel{PolicyID: types.StringValue("")}.GetResourceID().ValueString())
	require.Equal(t, "abc123", agentPolicyModel{PolicyID: types.StringValue("abc123")}.GetResourceID().ValueString())
}

func TestRead_ByIDLookupNotFound(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"not found"}`)
	}))
	t.Cleanup(srv.Close)

	config := agentPolicyModel{PolicyID: types.StringValue("does-not-exist")}
	_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", config)

	require.Equal(t, "/api/fleet/agent_policies/does-not-exist", gotPath)
	require.False(t, found)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, "does-not-exist", config.GetResourceID().ValueString())
}

const policyResponseTail = `"namespace":"default","status":"active","revision":1,"updated_at":"2020-01-01T00:00:00.000Z","updated_by":"elastic","is_protected":false`

func fleetPolicyObject(id, name string, extra ...string) string {
	fields := fmt.Sprintf(`"id":%q,"name":%q,%s`, id, name, policyResponseTail)
	if len(extra) > 0 && extra[0] != "" {
		fields += "," + extra[0]
	}
	return "{" + fields + "}"
}

func fleetPolicyGetBody(id, name string, extra ...string) string {
	return `{"item":` + fleetPolicyObject(id, name, extra...) + `}`
}

func fleetPolicyListBody(total int, items ...string) string {
	return fmt.Sprintf(`{"items":[%s],"page":1,"perPage":100,"total":%d}`, strings.Join(items, ","), total)
}

func diagnosticText(diags diag.Diagnostics) string {
	var b strings.Builder
	for _, d := range diags {
		b.WriteString(d.Summary())
		b.WriteByte(' ')
		b.WriteString(d.Detail())
	}
	return b.String()
}

func newReadTestClient(t *testing.T, srv *httptest.Server) *clients.KibanaScopedClient {
	t.Helper()
	fleetClient, err := fleet.NewClient(fleet.Config{URL: srv.URL})
	require.NoError(t, err)
	return clients.NewKibanaScopedClientWithFleetForTest(fleetClient)
}
