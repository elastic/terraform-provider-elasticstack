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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newStatusServer creates an httptest.Server that serves the given body and
// status code for every request. The returned server must be closed by the
// caller.
func newStatusServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(statusCode)
		_, _ = rw.Write([]byte(body))
	}))
}

func TestGetKibanaStatus_NormalResponse(t *testing.T) {
	srv := newStatusServer(http.StatusOK, `{
		"version": {
			"number": "8.14.0",
			"build_flavor": "default"
		}
	}`)
	defer srv.Close()

	c := newTestClient(t, srv)
	ver, flavor, diags := GetKibanaStatus(t.Context(), c.API)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, "8.14.0", ver)
	assert.Equal(t, "default", flavor)
}

func TestGetKibanaStatus_MissingBuildFlavor(t *testing.T) {
	srv := newStatusServer(http.StatusOK, `{
		"version": {
			"number": "7.17.0"
		}
	}`)
	defer srv.Close()

	c := newTestClient(t, srv)
	ver, flavor, diags := GetKibanaStatus(t.Context(), c.API)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, "7.17.0", ver)
	assert.Empty(t, flavor, "build_flavor should be empty string when absent")
}

func TestGetKibanaStatus_ServerlessFlavor(t *testing.T) {
	srv := newStatusServer(http.StatusOK, `{
		"version": {
			"number": "8.999.0",
			"build_flavor": "serverless"
		}
	}`)
	defer srv.Close()

	c := newTestClient(t, srv)
	ver, flavor, diags := GetKibanaStatus(t.Context(), c.API)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, "8.999.0", ver)
	assert.Equal(t, "serverless", flavor)
}

func TestGetKibanaStatus_Non2xxResponse(t *testing.T) {
	srv := newStatusServer(http.StatusInternalServerError, `{"error":"something went wrong"}`)
	defer srv.Close()

	c := newTestClient(t, srv)
	ver, flavor, diags := GetKibanaStatus(t.Context(), c.API)

	assert.True(t, diags.HasError(), "expected error diagnostics for non-2xx response")
	assert.Empty(t, ver)
	assert.Empty(t, flavor)
}

func TestGetKibanaStatus_MalformedJSON(t *testing.T) {
	srv := newStatusServer(http.StatusOK, `not-valid-json`)
	defer srv.Close()

	c := newTestClient(t, srv)
	ver, flavor, diags := GetKibanaStatus(t.Context(), c.API)

	assert.True(t, diags.HasError(), "expected error diagnostics for malformed JSON")
	assert.Empty(t, ver)
	assert.Empty(t, flavor)
}

func TestGetKibanaStatus_MissingVersionNumber(t *testing.T) {
	srv := newStatusServer(http.StatusOK, `{
		"version": {
			"build_flavor": "default"
		}
	}`)
	defer srv.Close()

	c := newTestClient(t, srv)
	ver, flavor, diags := GetKibanaStatus(t.Context(), c.API)

	assert.True(t, diags.HasError(), "expected error diagnostics when version.number is absent")
	assert.Empty(t, ver)
	assert.Empty(t, flavor)
}
