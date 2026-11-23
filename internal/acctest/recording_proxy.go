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

package acctest

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"sync"
	"testing"
)

// RecordingProxy forwards requests to a live test service and records headers.
type RecordingProxy struct {
	server *httptest.Server

	mu       sync.Mutex
	requests []http.Header
}

// NewRecordingProxy creates a local reverse proxy that records request headers
// before forwarding each request to upstreamEndpoint.
func NewRecordingProxy(t *testing.T, upstreamEndpoint string) *RecordingProxy {
	t.Helper()

	upstream, err := url.Parse(upstreamEndpoint)
	if err != nil {
		t.Fatalf("parse recording proxy upstream endpoint: %v", err)
	}

	proxy := &RecordingProxy{}
	reverseProxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.mu.Lock()
		proxy.requests = append(proxy.requests, r.Header.Clone())
		proxy.mu.Unlock()
		reverseProxy.ServeHTTP(w, r)
	}))
	t.Cleanup(proxy.server.Close)

	return proxy
}

// URL returns the local proxy endpoint.
func (p *RecordingProxy) URL() string {
	return p.server.URL
}

// HeaderValues returns all values observed for name across recorded requests.
func (p *RecordingProxy) HeaderValues(name string) []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var values []string
	for _, headers := range p.requests {
		values = append(values, headers.Values(name)...)
	}
	return values
}

// RequestCount returns the number of requests received by the proxy.
func (p *RecordingProxy) RequestCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.requests)
}
