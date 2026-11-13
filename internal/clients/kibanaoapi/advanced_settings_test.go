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
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const emptySettingsBody = `{"settings":{}}`

func newAdvancedSettingsTestClient(t *testing.T, status int, body string, inspect func(*http.Request)) *Client {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if inspect != nil {
			inspect(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return newTestClient(t, srv)
}

func newClosedAdvancedSettingsTestClient(t *testing.T) *Client {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	client := newTestClient(t, srv)
	srv.Close()
	return client
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	return string(body)
}

func TestGetAdvancedSettings_UsesSpaceAwarePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spaceID  string
		wantPath string
	}{
		{name: "default space is unprefixed", spaceID: "default", wantPath: "/api/kibana/settings"},
		{name: "non-default space is prefixed", spaceID: "ops", wantPath: "/s/ops/api/kibana/settings"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newAdvancedSettingsTestClient(t, http.StatusOK, `{"settings":{"dateFormat:tz":{"userValue":"UTC"}}}`, func(r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.wantPath, r.URL.Path)
			})

			settings, diags := GetAdvancedSettings(context.Background(), client, tt.spaceID)

			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			assert.Equal(t, map[string]json.RawMessage{"dateFormat:tz": json.RawMessage(`"UTC"`)}, settings)
		})
	}
}

func TestGetAdvancedSettings_ReturnsMixedValueTypesExactly(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusOK, `{"settings":{
		"custom:string":{"userValue":"x"},
		"custom:boolean":{"userValue":true},
		"custom:integer":{"userValue":12345678901234567890},
		"custom:decimal":{"userValue":1.50},
		"custom:array":{"userValue":[1,"two",null]},
		"custom:object":{"userValue":{"nested":{"k":1}}},
		"custom:pinned":{"userValue":false,"isOverridden":true}
	}}`, nil)

	settings, diags := GetAdvancedSettings(context.Background(), client, "default")

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, map[string]json.RawMessage{
		"custom:string":  json.RawMessage(`"x"`),
		"custom:boolean": json.RawMessage(`true`),
		"custom:integer": json.RawMessage(`12345678901234567890`),
		"custom:decimal": json.RawMessage(`1.50`),
		"custom:array":   json.RawMessage(`[1,"two",null]`),
		"custom:object":  json.RawMessage(`{"nested":{"k":1}}`),
		"custom:pinned":  json.RawMessage(`false`),
	}, settings)
}

func TestGetAdvancedSettings_NotFoundReturnsNilWithoutDiagnostics(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusNotFound, `{"statusCode":404,"error":"Not Found","message":"Not Found"}`, nil)

	settings, diags := GetAdvancedSettings(context.Background(), client, "ops")

	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Nil(t, settings)
}

func TestGetAdvancedSettings_SendsInternalOriginHeader(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusOK, emptySettingsBody, func(r *http.Request) {
		assert.Equal(t, "Kibana", r.Header.Get("x-elastic-internal-origin"))
	})

	_, diags := GetAdvancedSettings(context.Background(), client, "default")

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}

func TestGetAdvancedSettings_FailureBecomesErrorDiagnosticWithKibanaMessage(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusForbidden, `{"statusCode":403,"error":"Forbidden","message":"Not allowed to read settings"}`, nil)

	settings, diags := GetAdvancedSettings(context.Background(), client, "default")

	assert.Nil(t, settings)
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "Failed to read Kibana advanced settings", diags.Errors()[0].Summary())
	assert.Equal(t, "Not allowed to read settings", diags.Errors()[0].Detail())
}

func TestGetAdvancedSettings_TransportErrorBecomesErrorDiagnostic(t *testing.T) {
	t.Parallel()

	settings, diags := GetAdvancedSettings(context.Background(), newClosedAdvancedSettingsTestClient(t), "default")

	assert.Nil(t, settings)
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "Failed to read Kibana advanced settings", diags.Errors()[0].Summary())
	assert.Contains(t, diags.Errors()[0].Detail(), "connection refused")
}

func TestGetAdvancedSettings_MalformedBodyBecomesErrorDiagnostic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "truncated JSON", contentType: "application/json", body: `{"settings":`},
		{name: "non-JSON content type", contentType: "text/html", body: `<html>login</html>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)

			settings, diags := GetAdvancedSettings(context.Background(), newTestClient(t, srv), "default")

			assert.Nil(t, settings)
			require.Len(t, diags.Errors(), 1)
			assert.Equal(t, "Failed to read Kibana advanced settings", diags.Errors()[0].Summary())
		})
	}
}

func TestApplyAdvancedSettings_SendsChangesInOneBulkPost(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	client := newAdvancedSettingsTestClient(t, http.StatusOK, emptySettingsBody, func(r *http.Request) {
		requests.Add(1)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.JSONEq(t, `{"changes":{"custom:a":"x","custom:b":true}}`, readBody(t, r))
	})

	notFound, diags := ApplyAdvancedSettings(context.Background(), client, "default", map[string]any{
		"custom:a": "x",
		"custom:b": true,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.False(t, notFound)
	assert.Equal(t, int32(1), requests.Load())
}

func TestApplyAdvancedSettings_UsesSpaceAwarePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spaceID  string
		wantPath string
	}{
		{name: "default space is unprefixed", spaceID: "default", wantPath: "/api/kibana/settings"},
		{name: "non-default space is prefixed", spaceID: "ops", wantPath: "/s/ops/api/kibana/settings"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newAdvancedSettingsTestClient(t, http.StatusOK, emptySettingsBody, func(r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, tt.wantPath, r.URL.Path)
			})

			_, diags := ApplyAdvancedSettings(context.Background(), client, tt.spaceID, map[string]any{"custom:a": "x"})

			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		})
	}
}

func TestApplyAdvancedSettings_NilValueResetsKey(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusOK, emptySettingsBody, func(r *http.Request) {
		assert.JSONEq(t, `{"changes":{"custom:reset":null,"custom:set":"x"}}`, readBody(t, r))
	})

	_, diags := ApplyAdvancedSettings(context.Background(), client, "default", map[string]any{
		"custom:reset": nil,
		"custom:set":   "x",
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}

func TestApplyAdvancedSettings_PreservesExactNumbers(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusOK, emptySettingsBody, func(r *http.Request) {
		body := readBody(t, r)
		assert.Contains(t, body, `"custom:decimal":1.50`)
		assert.Contains(t, body, `"custom:integer":12345678901234567890`)
		assert.Contains(t, body, `"custom:list":[1.0,2]`)
	})

	_, diags := ApplyAdvancedSettings(context.Background(), client, "default", map[string]any{
		"custom:integer": json.Number("12345678901234567890"),
		"custom:decimal": json.RawMessage(`1.50`),
		"custom:list":    json.RawMessage(`[1.0,2]`),
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}

func TestApplyAdvancedSettings_SendsInternalOriginHeader(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusOK, emptySettingsBody, func(r *http.Request) {
		assert.Equal(t, "Kibana", r.Header.Get("x-elastic-internal-origin"))
	})

	_, diags := ApplyAdvancedSettings(context.Background(), client, "default", map[string]any{"custom:a": "x"})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}

func TestApplyAdvancedSettings_NotFoundIsReportedWithoutDiagnostics(t *testing.T) {
	t.Parallel()

	client := newAdvancedSettingsTestClient(t, http.StatusNotFound, `{"statusCode":404,"error":"Not Found","message":"Not Found"}`, nil)

	notFound, diags := ApplyAdvancedSettings(context.Background(), client, "ops", map[string]any{"custom:a": nil})

	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.True(t, notFound)
}

func TestApplyAdvancedSettings_FailuresBecomeErrorDiagnosticsWithKibanaMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		status      int
		body        string
		wantMessage string
	}{
		{
			name:        "invalid value for a registered setting",
			status:      http.StatusBadRequest,
			body:        `{"statusCode":400,"error":"Bad Request","message":"[validation [dateFormat:tz]]: expected value of type [string] but got [number]"}`,
			wantMessage: "[validation [dateFormat:tz]]: expected value of type [string] but got [number]",
		},
		{
			name:        "setting pinned by Kibana server configuration",
			status:      http.StatusBadRequest,
			body:        `{"statusCode":400,"error":"Bad Request","message":"Unable to update \"observability:streamsEnableQueryStreams\" because it is overridden"}`,
			wantMessage: `Unable to update "observability:streamsEnableQueryStreams" because it is overridden`,
		},
		{
			name:   "missing privilege",
			status: http.StatusForbidden,
			body: `{"statusCode":403,"error":"Forbidden","message":"API [POST /api/kibana/settings] is unauthorized for user, ` +
				`this action is granted by the Kibana privileges [manage_advanced_settings]"}`,
			wantMessage: "API [POST /api/kibana/settings] is unauthorized for user, this action is granted by the Kibana privileges [manage_advanced_settings]",
		},
		{
			name:        "unexpected server error",
			status:      http.StatusInternalServerError,
			body:        `{"statusCode":500,"error":"Internal Server Error","message":"An internal server error occurred."}`,
			wantMessage: "An internal server error occurred.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newAdvancedSettingsTestClient(t, tt.status, tt.body, nil)

			notFound, diags := ApplyAdvancedSettings(context.Background(), client, "default", map[string]any{"dateFormat:tz": 1})

			assert.False(t, notFound)
			require.Len(t, diags.Errors(), 1)
			assert.Equal(t, "Failed to apply Kibana advanced settings", diags.Errors()[0].Summary())
			assert.Equal(t, tt.wantMessage, diags.Errors()[0].Detail())
		})
	}
}

func TestApplyAdvancedSettings_UnstructuredErrorBodyIsSurfacedVerbatim(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream unavailable"))
	}))
	t.Cleanup(srv.Close)

	notFound, diags := ApplyAdvancedSettings(context.Background(), newTestClient(t, srv), "default", map[string]any{"custom:a": "x"})

	assert.False(t, notFound)
	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Summary(), "502")
	assert.Equal(t, "upstream unavailable", diags.Errors()[0].Detail())
}

func TestApplyAdvancedSettings_TransportErrorBecomesErrorDiagnostic(t *testing.T) {
	t.Parallel()

	notFound, diags := ApplyAdvancedSettings(context.Background(), newClosedAdvancedSettingsTestClient(t), "default", map[string]any{"custom:a": "x"})

	assert.False(t, notFound)
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "Failed to apply Kibana advanced settings", diags.Errors()[0].Summary())
	assert.Contains(t, diags.Errors()[0].Detail(), "connection refused")
}
