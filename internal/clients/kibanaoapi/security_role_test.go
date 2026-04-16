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

package kibanaoapi_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestKibanaOapiClient(t *testing.T, server *httptest.Server) *kibanaoapi.Client {
	t.Helper()
	t.Setenv("ELASTICSEARCH_URL", server.URL)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)
	return oapiClient
}

func TestGetSecurityRole_200(t *testing.T) {
	roleBody := map[string]any{
		"elasticsearch": map[string]any{
			"cluster": []string{"monitor"},
			"indices": []any{
				map[string]any{
					"names":      []string{"logs-*"},
					"privileges": []string{"read"},
				},
			},
		},
		"kibana": []any{
			map[string]any{
				"base":   []string{"all"},
				"spaces": []string{"default"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(roleBody)
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)
	role, diags := kibanaoapi.GetSecurityRole(t.Context(), oapiClient, "test-role")

	require.Nil(t, diags)
	require.NotNil(t, role)
	require.NotNil(t, role.Elasticsearch.Cluster)
	assert.Equal(t, []string{"monitor"}, *role.Elasticsearch.Cluster)
}

func TestGetSecurityRole_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)
	role, diags := kibanaoapi.GetSecurityRole(t.Context(), oapiClient, "missing-role")

	assert.Nil(t, diags)
	assert.Nil(t, role)
}

func TestGetSecurityRole_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("not-valid-json"))
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)
	role, diags := kibanaoapi.GetSecurityRole(t.Context(), oapiClient, "test-role")

	assert.NotNil(t, diags)
	assert.True(t, diags.HasError())
	assert.Nil(t, role)
}

func TestGetSecurityRole_500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)
	role, diags := kibanaoapi.GetSecurityRole(t.Context(), oapiClient, "test-role")

	assert.NotNil(t, diags)
	assert.True(t, diags.HasError())
	assert.Nil(t, role)
}

func TestPutSecurityRole_200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)

	createOnly := true
	params := kbapi.PutSecurityRoleNameParams{CreateOnly: &createOnly}
	cluster := []string{"monitor"}
	body := kibanaoapi.SecurityRolePutBody{
		Elasticsearch: kibanaoapi.SecurityRoleES{
			Cluster: &cluster,
		},
	}

	diags := kibanaoapi.PutSecurityRole(t.Context(), oapiClient, "test-role", params, body)
	assert.Nil(t, diags)
}

func TestPutSecurityRole_WithBase_SerializesCorrectly(t *testing.T) {
	var capturedBody []byte
	var captureErr error

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		capturedBody, captureErr = io.ReadAll(r.Body)
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)

	spaces := []string{"default"}
	body := kibanaoapi.SecurityRolePutBody{
		Kibana: []kibanaoapi.SecurityRoleKibana{
			{
				Base:   json.RawMessage(`["all"]`),
				Spaces: &spaces,
			},
		},
		Elasticsearch: kibanaoapi.SecurityRoleES{},
	}

	params := kbapi.PutSecurityRoleNameParams{}
	diags := kibanaoapi.PutSecurityRole(t.Context(), oapiClient, "test-role", params, body)
	assert.Nil(t, diags)
	require.NoError(t, captureErr)

	// Verify kibana.base is correctly serialized in the request body
	var sent map[string]any
	require.NoError(t, json.Unmarshal(capturedBody, &sent))
	kibanaArr, ok := sent["kibana"].([]any)
	require.True(t, ok)
	require.Len(t, kibanaArr, 1)
	kibanaEntry := kibanaArr[0].(map[string]any)
	base, ok := kibanaEntry["base"].([]any)
	require.True(t, ok)
	require.Len(t, base, 1)
	assert.Equal(t, "all", base[0])
}

func TestPutSecurityRole_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)

	params := kbapi.PutSecurityRoleNameParams{}
	body := kibanaoapi.SecurityRolePutBody{}

	diags := kibanaoapi.PutSecurityRole(t.Context(), oapiClient, "test-role", params, body)
	assert.NotNil(t, diags)
	assert.True(t, diags.HasError())
}

func TestDeleteSecurityRole_200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)
	diags := kibanaoapi.DeleteSecurityRole(t.Context(), oapiClient, "test-role")
	assert.Nil(t, diags)
}

func TestDeleteSecurityRole_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)
	diags := kibanaoapi.DeleteSecurityRole(t.Context(), oapiClient, "missing-role")
	assert.Nil(t, diags)
}

func TestDeleteSecurityRole_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	oapiClient := newTestKibanaOapiClient(t, server)
	diags := kibanaoapi.DeleteSecurityRole(t.Context(), oapiClient, "test-role")
	assert.NotNil(t, diags)
	assert.True(t, diags.HasError())
}
