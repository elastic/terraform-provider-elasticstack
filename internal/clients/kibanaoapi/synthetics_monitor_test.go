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

package kibanaoapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(Config{URL: srv.URL})
	require.NoError(t, err)
	return c
}
func httpMonitorRequest(t *testing.T, name, url string) kbapi.SyntheticsMonitorRequest {
	t.Helper()
	req := kbapi.SyntheticsMonitorRequest{}
	err := req.FromSyntheticsHttpMonitorFields(kbapi.SyntheticsHttpMonitorFields{
		Name:   name,
		Type:   kbapi.SyntheticsHttpMonitorFieldsType(kbapi.SyntheticsMonitorTypeHttp),
		Url:    url,
		Labels: &map[string]string{},
	})
	require.NoError(t, err)
	return req
}

func tcpMonitorRequest(t *testing.T, name, host string) kbapi.SyntheticsMonitorRequest {
	t.Helper()
	req := kbapi.SyntheticsMonitorRequest{}
	err := req.FromSyntheticsTcpMonitorFields(kbapi.SyntheticsTcpMonitorFields{
		Name:   name,
		Type:   kbapi.SyntheticsTcpMonitorFieldsType(kbapi.SyntheticsMonitorTypeTcp),
		Host:   host,
		Labels: &map[string]string{},
	})
	require.NoError(t, err)
	return req
}

func icmpMonitorRequest(t *testing.T, name, host string) kbapi.SyntheticsMonitorRequest {
	t.Helper()
	req := kbapi.SyntheticsMonitorRequest{}
	err := req.FromSyntheticsIcmpMonitorFields(kbapi.SyntheticsIcmpMonitorFields{
		Name:   name,
		Type:   kbapi.SyntheticsIcmpMonitorFieldsType(kbapi.SyntheticsMonitorTypeIcmp),
		Host:   host,
		Labels: &map[string]string{},
	})
	require.NoError(t, err)
	return req
}

func browserMonitorRequest(t *testing.T, name, script string) kbapi.SyntheticsMonitorRequest {
	t.Helper()
	req := kbapi.SyntheticsMonitorRequest{}
	err := req.FromSyntheticsBrowserMonitorFields(kbapi.SyntheticsBrowserMonitorFields{
		Name:         name,
		Type:         kbapi.SyntheticsBrowserMonitorFieldsType(kbapi.SyntheticsMonitorTypeBrowser),
		InlineScript: script,
		Labels:       &map[string]string{},
	})
	require.NoError(t, err)
	return req
}

func TestGetMonitor200(t *testing.T) {
	monitor := kbapi.SyntheticsMonitor{
		Id:        new("abc123"),
		Name:      new("my-http-monitor"),
		Type:      new(kbapi.SyntheticsMonitorTypeHttp),
		Namespace: new("default"),
		Enabled:   new(true),
		Schedule: &kbapi.SyntheticsMonitorSchedule{
			Number: new("5"),
			Unit:   new("m"),
		},
		Url: new("https://example.com"),
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "abc123")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(monitor)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := GetMonitor(context.Background(), client, "default", "abc123")
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
	assert.Equal(t, "abc123", *result.Id)
	assert.Equal(t, "my-http-monitor", *result.Name)
	assert.Equal(t, kbapi.SyntheticsMonitorTypeHttp, *result.Type)
}

func TestGetMonitor404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := GetMonitor(context.Background(), client, "default", "nonexistent")
	assert.False(t, diags.HasError(), diags)
	assert.Nil(t, result)
}

func TestCreateMonitor200(t *testing.T) {
	req := httpMonitorRequest(t, "new-monitor", "https://example.com")
	expectedResponse := kbapi.SyntheticsMonitor{
		Id:   new("created-id"),
		Name: new("new-monitor"),
		Type: new(kbapi.SyntheticsMonitorTypeHttp),
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var body map[string]any
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "http", body["type"])
		assert.Equal(t, "new-monitor", body["name"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := CreateMonitor(context.Background(), client, "default", req)
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
	assert.Equal(t, "created-id", *result.Id)
	assert.Equal(t, kbapi.SyntheticsMonitorTypeHttp, *result.Type)
}

func TestUpdateMonitor200(t *testing.T) {
	req := httpMonitorRequest(t, "updated-monitor", "https://updated.example.com")
	expectedResponse := kbapi.SyntheticsMonitor{
		Id:       new("monitor-id"),
		Name:     new("updated-monitor"),
		Type:     new(kbapi.SyntheticsMonitorTypeHttp),
		Url:      new("https://updated.example.com"),
		ProxyUrl: new("http://localhost"),
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "monitor-id")
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPut:
			var body map[string]any
			err := json.NewDecoder(r.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, "updated-monitor", body["name"])
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"warnings": []map[string]any{},
			})
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedResponse)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := UpdateMonitor(context.Background(), client, "default", "monitor-id", req)
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
	assert.Equal(t, "monitor-id", *result.Id)
	assert.Equal(t, "updated-monitor", *result.Name)
	assert.Equal(t, "http://localhost", *result.ProxyUrl)
}

func TestDeleteMonitor200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/synthetics/monitors", r.URL.Path)
		var body map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		ids, _ := body["ids"].([]any)
		assert.Len(t, ids, 1)
		assert.Equal(t, "monitor-id", ids[0])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"monitor-id","deleted":true}]`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	diags := DeleteMonitor(context.Background(), client, "default", "monitor-id")
	assert.False(t, diags.HasError(), diags)
}

func TestDeleteMonitorDeletedFalse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"monitor-id","deleted":false}]`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	diags := DeleteMonitor(context.Background(), client, "default", "monitor-id")
	assert.True(t, diags.HasError())
}

func TestCreateMonitorSpaceAwarePath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/s/my-space/")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(kbapi.SyntheticsMonitor{
			Id:   new("space-monitor-id"),
			Type: new(kbapi.SyntheticsMonitorTypeHttp),
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	req := httpMonitorRequest(t, "space-monitor", "https://example.com")
	result, diags := CreateMonitor(context.Background(), client, "my-space", req)
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
}

func TestSyntheticsMonitorRequestMarshalTypeDiscriminator(t *testing.T) {
	testcases := []struct {
		name     string
		req      kbapi.SyntheticsMonitorRequest
		wantType string
		wantKey  string
	}{
		{name: "HTTP", req: httpMonitorRequest(t, "http-mon", "https://example.com"), wantType: "http", wantKey: "url"},
		{name: "TCP", req: tcpMonitorRequest(t, "tcp-mon", "example.com:9200"), wantType: "tcp", wantKey: "host"},
		{name: "ICMP", req: icmpMonitorRequest(t, "icmp-mon", "8.8.8.8"), wantType: "icmp", wantKey: "host"},
		{name: "Browser", req: browserMonitorRequest(t, "browser-mon", "step('go', () => page.goto('https://example.com'))"), wantType: "browser", wantKey: "inline_script"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.req)
			require.NoError(t, err)

			var m map[string]any
			require.NoError(t, json.Unmarshal(data, &m))

			assert.Equal(t, tc.wantType, m["type"])
			assert.Contains(t, m, tc.wantKey)
			assert.Contains(t, m, "labels")
		})
	}
}
