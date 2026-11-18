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

package serverhostds

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

const listPath = "/api/fleet/fleet_server_hosts"

func TestRead_ByHostID(t *testing.T) {
	var gotPath string
	srv := newHostServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, hostGetBody(hostObject("host-1", "Primary", `"is_default":false,"proxy_id":"proxy-1"`)))
	})

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "ignored-envelope-id", "", serverHostModel{
		HostID: types.StringValue("host-1"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, listPath+"/host-1", gotPath)
	require.Equal(t, "host-1", got.ID.ValueString())
	require.Equal(t, "host-1", got.HostID.ValueString())
	require.Equal(t, "Primary", got.Name.ValueString())
	require.Equal(t, types.BoolValue(false), got.Default)
	require.Equal(t, "proxy-1", got.ProxyID.ValueString())
	require.Equal(t, []string{"https://fleet-server:8220"}, listStrings(t, got.Hosts))
	require.True(t, got.Ssl.IsNull())
}

func TestRead_ByHostIDUsesConfiguredSpace(t *testing.T) {
	var gotPath string
	srv := newHostServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, hostGetBody(hostObject("host-1", "Primary", "")))
	})

	_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "ops", serverHostModel{
		HostID: types.StringValue("host-1"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, "/s/ops"+listPath+"/host-1", gotPath)
}

func TestRead_ByHostIDNotFound(t *testing.T) {
	srv := newHostServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"not found"}`)
	})

	_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", serverHostModel{
		HostID: types.StringValue("does-not-exist"),
	})

	require.False(t, found)
	require.False(t, diags.HasError(), diags)
}

func TestRead_ByName(t *testing.T) {
	var gotPaths []string
	srv := newHostServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		fmt.Fprint(w, hostListBody(3,
			hostObject("host-1", "Primary", `"is_default":true`),
			hostObject("host-2", "Secondary", `"is_default":false`),
			hostObject("host-3", "Secondary EU", `"is_default":false`),
		))
	})

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "ops", serverHostModel{
		Name: types.StringValue("Secondary"),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, []string{"/s/ops" + listPath}, gotPaths)
	require.Equal(t, "host-2", got.ID.ValueString())
	require.Equal(t, "host-2", got.HostID.ValueString())
	require.Equal(t, "Secondary", got.Name.ValueString())
	require.Equal(t, types.BoolValue(false), got.Default)
}

func TestRead_ByDefault(t *testing.T) {
	srv := newHostServer(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, hostListBody(3,
			hostObject("host-1", "Primary", `"is_default":false`),
			hostObject("host-2", "Secondary", `"is_default":true`),
			hostObject("host-3", "Legacy", ""),
		))
	})

	got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", serverHostModel{
		Default: types.BoolValue(true),
	})

	require.False(t, diags.HasError(), diags)
	require.True(t, found)
	require.Equal(t, "host-2", got.HostID.ValueString())
	require.Equal(t, "Secondary", got.Name.ValueString())
	require.Equal(t, types.BoolValue(true), got.Default)
}

func TestRead_ListLookupErrors(t *testing.T) {
	tests := []struct {
		name     string
		config   serverHostModel
		listBody string
		want     string
	}{
		{
			name:     "name matches no host",
			config:   serverHostModel{Name: types.StringValue("does-not-exist")},
			listBody: hostListBody(1, hostObject("host-1", "Primary", `"is_default":true`)),
			want:     `Fleet server host not found no Fleet server host found matching name "does-not-exist"`,
		},
		{
			name:   "name matches several hosts",
			config: serverHostModel{Name: types.StringValue("Shared")},
			listBody: hostListBody(3,
				hostObject("a", "Shared", ""),
				hostObject("b", "Other", ""),
				hostObject("c", "Shared", ""),
			),
			want: `Multiple Fleet server hosts match multiple Fleet server hosts match name "Shared": host_id "a", host_id "c"`,
		},
		{
			name:     "no default host",
			config:   serverHostModel{Default: types.BoolValue(true)},
			listBody: hostListBody(2, hostObject("a", "One", `"is_default":false`), hostObject("b", "Two", "")),
			want:     "Fleet server host not found no default Fleet server host found",
		},
		{
			name:   "several default hosts",
			config: serverHostModel{Default: types.BoolValue(true)},
			listBody: hostListBody(2,
				hostObject("a", "One", `"is_default":true`),
				hostObject("b", "Two", `"is_default":true`),
			),
			want: `Multiple Fleet server hosts match multiple Fleet server hosts are marked as default: host_id "a", host_id "b"`,
		},
		{
			name:     "truncated list",
			config:   serverHostModel{Name: types.StringValue("Primary")},
			listBody: hostListBody(2, hostObject("host-1", "Primary", "")),
			want:     "Incomplete Fleet server host list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newHostServer(t, func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, tt.listBody)
			})

			_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", tt.config)

			require.False(t, found)
			require.True(t, diags.HasError())
			require.Contains(t, diagnosticText(diags), tt.want)
		})
	}
}

func TestRead_APIErrorIsReported(t *testing.T) {
	for name, config := range map[string]serverHostModel{
		"by host_id": {HostID: types.StringValue("host-1")},
		"by name":    {Name: types.StringValue("Primary")},
		"by default": {Default: types.BoolValue(true)},
	} {
		t.Run(name, func(t *testing.T) {
			srv := newHostServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`)
			})

			_, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", config)

			require.False(t, found)
			require.True(t, diags.HasError())
			require.Contains(t, diagnosticText(diags), "boom")
		})
	}
}

func TestRead_Ssl(t *testing.T) {
	tests := []struct {
		name string
		ssl  string
		want *serverHostSslModel
	}{
		{name: "absent", ssl: ""},
		{name: "empty object", ssl: `"ssl":{}`},
		{name: "only fields that are not exposed", ssl: `"ssl":{"key":"private","es_certificate":"es-cert","certificate_authorities":[]}`},
		{
			name: "all exposed fields",
			ssl:  `"ssl":{"certificate":"cert","certificate_authorities":["ca-1","ca-2"],"client_auth":"required","key":"private"}`,
			want: &serverHostSslModel{
				Certificate:            types.StringValue("cert"),
				CertificateAuthorities: stringList("ca-1", "ca-2"),
				ClientAuth:             types.StringValue("required"),
			},
		},
		{
			name: "certificate only",
			ssl:  `"ssl":{"certificate":"cert","key":"private"}`,
			want: &serverHostSslModel{
				Certificate:            types.StringValue("cert"),
				CertificateAuthorities: types.ListNull(types.StringType),
				ClientAuth:             types.StringNull(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newHostServer(t, func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, hostGetBody(hostObject("host-1", "Primary", tt.ssl)))
			})

			got, found, diags := readDataSource(t.Context(), newReadTestClient(t, srv), "_", "", serverHostModel{
				HostID: types.StringValue("host-1"),
			})

			require.False(t, diags.HasError(), diags)
			require.True(t, found)
			require.NotContains(t, got.Ssl.String(), "private")
			if tt.want == nil {
				require.True(t, got.Ssl.IsNull())
				return
			}
			var ssl serverHostSslModel
			require.False(t, got.Ssl.As(t.Context(), &ssl, objectAsOptions).HasError())
			require.Equal(t, *tt.want, ssl)
		})
	}
}

func TestDataSourceSchema(t *testing.T) {
	schema := getDataSourceSchema(t.Context())

	hostID := schema.Attributes["host_id"].(dsschema.StringAttribute)
	name := schema.Attributes["name"].(dsschema.StringAttribute)
	isDefault := schema.Attributes["default"].(dsschema.BoolAttribute)
	spaceID := schema.Attributes["space_id"].(dsschema.StringAttribute)
	for _, selector := range []dsschema.Attribute{hostID, name, isDefault, spaceID} {
		require.True(t, selector.IsOptional())
		require.True(t, selector.IsComputed())
	}
	require.Len(t, hostID.Validators, 2)
	require.Len(t, name.Validators, 1)
	require.Len(t, isDefault.Validators, 1)

	for _, output := range []string{"id", "hosts", "proxy_id", "ssl"} {
		require.True(t, schema.Attributes[output].IsComputed(), output)
		require.False(t, schema.Attributes[output].IsOptional(), output)
	}

	ssl := schema.Attributes["ssl"].(dsschema.SingleNestedAttribute)
	require.ElementsMatch(t, []string{"certificate", "certificate_authorities", "client_auth"}, mapKeys(ssl.Attributes))
	_, hasSpaceIDs := schema.Attributes["space_ids"]
	require.False(t, hasSpaceIDs)
}

func TestNewDataSource_TypeName(t *testing.T) {
	ds := NewDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(t.Context(), datasource.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
	require.Equal(t, "elasticstack_fleet_server_host", resp.TypeName)
}

func TestGetResourceID_UsesPlaceholderWhenHostIDUnset(t *testing.T) {
	require.Equal(t, "_", serverHostModel{}.GetResourceID().ValueString())
	require.Equal(t, "_", serverHostModel{Name: types.StringValue("Primary")}.GetResourceID().ValueString())
	require.Equal(t, "host-1", serverHostModel{HostID: types.StringValue("host-1")}.GetResourceID().ValueString())
}

var objectAsOptions = basetypes.ObjectAsOptions{}

func stringList(values ...string) types.List {
	elems := make([]attr.Value, len(values))
	for i, value := range values {
		elems[i] = types.StringValue(value)
	}
	return types.ListValueMust(types.StringType, elems)
}

func listStrings(t *testing.T, list types.List) []string {
	t.Helper()
	var values []string
	require.False(t, list.ElementsAs(t.Context(), &values, false).HasError())
	return values
}

func mapKeys(attributes map[string]dsschema.Attribute) []string {
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	return keys
}

func hostObject(id, name, extra string) string {
	fields := fmt.Sprintf(`"id":%q,"name":%q,"host_urls":["https://fleet-server:8220"]`, id, name)
	if extra != "" {
		fields += "," + extra
	}
	return "{" + fields + "}"
}

func hostGetBody(host string) string {
	return `{"item":` + host + `}`
}

func hostListBody(total int, hosts ...string) string {
	return fmt.Sprintf(`{"items":[%s],"page":1,"perPage":10000,"total":%d}`, strings.Join(hosts, ","), total)
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

func newHostServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newReadTestClient(t *testing.T, srv *httptest.Server) *clients.KibanaScopedClient {
	t.Helper()
	fleetClient, err := fleet.NewClient(fleet.Config{URL: srv.URL})
	require.NoError(t, err)
	return clients.NewKibanaScopedClientWithFleetForTest(fleetClient)
}
