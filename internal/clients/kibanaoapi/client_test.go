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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSchemeRedirectTestClient(t *testing.T, source, destination string, destinationRequests *atomic.Int32) *Client {
	t.Helper()

	client, err := NewClient(Config{
		URL:     source,
		Headers: map[string]string{"cf-access-token": "xyz"},
	})
	require.NoError(t, err)
	client.HTTP.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/redirected" {
			destinationRequests.Add(1)
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       http.NoBody,
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusFound,
			Body:       io.NopCloser(strings.NewReader("redirecting")),
			Header:     http.Header{"Location": []string{destination + "/redirected"}},
			Request:    req,
		}, nil
	})
	return client
}

func TestClient_CustomHeadersFollowSameOriginRedirect(t *testing.T) {
	const headerName = "cf-access-token"
	const headerValue = "xyz"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/redirected", http.StatusFound)
		case "/redirected":
			assert.Equal(t, headerValue, r.Header.Get(headerName))
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		URL:     server.URL,
		Headers: map[string]string{headerName: headerValue},
	})
	require.NoError(t, err)
	require.NotNil(t, client.HTTP.CheckRedirect)

	response, err := client.HTTP.Get(server.URL + "/start")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestClient_CustomHeadersRejectHostnameChangeBeforeDestinationRequest(t *testing.T) {
	var destinationRequests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			destination := strings.Replace(server.URL, "127.0.0.1", "localhost", 1) + "/redirected"
			http.Redirect(w, r, destination, http.StatusFound)
		case "/redirected":
			destinationRequests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		URL:     server.URL,
		Headers: map[string]string{"cf-access-token": "xyz"},
	})
	require.NoError(t, err)

	response, err := client.HTTP.Get(server.URL + "/start")
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}
	require.Error(t, err)
	require.ErrorIs(t, err, errCrossOriginRedirect)
	assert.Zero(t, destinationRequests.Load())
}

func TestClient_CustomHeadersStopAfterTenRedirects(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) > 10 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		URL:     server.URL,
		Headers: map[string]string{"cf-access-token": "xyz"},
	})
	require.NoError(t, err)

	response, err := client.HTTP.Get(server.URL + "/redirect")
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}
	require.ErrorContains(t, err, "stopped after 10 redirects")
	assert.Equal(t, int32(10), requests.Load())
}

func TestClient_CustomHeadersRejectPortChangeBeforeDestinationRequest(t *testing.T) {
	var destinationRequests atomic.Int32

	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		destinationRequests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(destination.Close)

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	t.Cleanup(source.Close)

	client, err := NewClient(Config{
		URL:     source.URL,
		Headers: map[string]string{"cf-access-token": "xyz"},
	})
	require.NoError(t, err)

	response, err := client.HTTP.Get(source.URL)
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}
	require.Error(t, err)
	assert.Zero(t, destinationRequests.Load())
}

func TestClient_CustomHeadersAllowEquivalentDefaultPorts(t *testing.T) {
	client, err := NewClient(Config{
		URL:     "http://example.com",
		Headers: map[string]string{"cf-access-token": "xyz"},
	})
	require.NoError(t, err)

	initial := httptest.NewRequest(http.MethodGet, "http://example.com/start", nil)
	redirect := httptest.NewRequest(http.MethodGet, "http://example.com:80/redirected", nil)

	assert.NoError(t, client.HTTP.CheckRedirect(redirect, []*http.Request{initial}))
}

func TestClient_CustomHeadersAllowEquivalentHTTPSDefaultPorts(t *testing.T) {
	client, err := NewClient(Config{
		URL:     "https://example.com",
		Headers: map[string]string{"cf-access-token": "xyz"},
	})
	require.NoError(t, err)

	initial := httptest.NewRequest(http.MethodGet, "https://example.com/start", nil)
	redirect := httptest.NewRequest(http.MethodGet, "https://example.com:443/redirected", nil)

	assert.NoError(t, client.HTTP.CheckRedirect(redirect, []*http.Request{initial}))
}

func TestClient_CustomHeadersRejectSchemeChangeBeforeDestinationRequest(t *testing.T) {
	var destinationRequests atomic.Int32

	client := newSchemeRedirectTestClient(
		t,
		"http://example.com:443",
		"https://example.com:443",
		&destinationRequests,
	)

	response, err := client.HTTP.Get("http://example.com:443/start")
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}
	require.Error(t, err)
	assert.Zero(t, destinationRequests.Load())
}

func TestClient_CustomHeadersRejectHTTPSDowngradeBeforeDestinationRequest(t *testing.T) {
	var destinationRequests atomic.Int32

	client := newSchemeRedirectTestClient(
		t,
		"https://example.com:80",
		"http://example.com:80",
		&destinationRequests,
	)

	response, err := client.HTTP.Get("https://example.com:80/start")
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}
	require.Error(t, err)
	assert.Zero(t, destinationRequests.Load())
}

func TestClient_WithoutCustomHeadersPreservesLegacyRedirectBehavior(t *testing.T) {
	var destinationRequests atomic.Int32

	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		destinationRequests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(destination.Close)

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	t.Cleanup(source.Close)

	client, err := NewClient(Config{URL: source.URL})
	require.NoError(t, err)
	assert.Nil(t, client.HTTP.CheckRedirect)

	response, err := client.HTTP.Get(source.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	assert.Equal(t, int32(1), destinationRequests.Load())
}
