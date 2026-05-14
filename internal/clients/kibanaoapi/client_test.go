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

func Test_transport_RoundTrip_customHeaders(t *testing.T) {
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		URL: server.URL,
		Headers: map[string]string{
			"CF-Access-Client-Id":     "client-id",
			"CF-Access-Client-Secret": "client-secret",
		},
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)

	resp, err := client.HTTP.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "client-id", capturedHeaders.Get("CF-Access-Client-Id"))
	assert.Equal(t, "client-secret", capturedHeaders.Get("CF-Access-Client-Secret"))
}

func Test_transport_RoundTrip_authHeadersOverrideCustomHeaders(t *testing.T) {
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// If someone mistakenly sets Authorization as a custom header, the explicit
	// auth config (BearerToken) should override it.
	cfg := Config{
		URL:         server.URL,
		BearerToken: "real-token",
		Headers: map[string]string{
			"Authorization": "Bearer wrong-token",
		},
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)

	resp, err := client.HTTP.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "Bearer real-token", capturedHeaders.Get("Authorization"))
}
