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
	"testing"
)

func TestElasticsearchProxyAddressFromEnv(t *testing.T) {
	// Cannot use t.Parallel(): subtests call t.Setenv, which is incompatible with parallel tests.
	tests := []struct {
		name         string
		endpoints    string
		proxyAddress string
		transport    string
		want         string
		wantErr      bool
	}{
		{
			name:      "derives transport port from http endpoint host",
			endpoints: "http://localhost:9200",
			want:      "localhost:9300",
		},
		{
			name:      "preserves hostname without http port",
			endpoints: "http://elasticsearch",
			want:      "elasticsearch:9300",
		},
		{
			name:      "replaces http port with transport port",
			endpoints: "http://localhost:12345",
			want:      "localhost:9300",
		},
		{
			name:      "honors transport port override",
			endpoints: "http://localhost:9200",
			transport: "19300",
			want:      "localhost:19300",
		},
		{
			name:         "honors remote proxy address override",
			endpoints:    "http://localhost:9200",
			proxyAddress: "remote.example:9400",
			want:         "remote.example:9400",
		},
		{
			name:      "supports ipv6 host with brackets",
			endpoints: "http://[::1]:9200",
			want:      "[::1]:9300",
		},
		{
			name:    "missing endpoints",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ELASTICSEARCH_ENDPOINTS", tt.endpoints)
			t.Setenv("ELASTICSEARCH_TRANSPORT_PORT", tt.transport)
			t.Setenv("ELASTICSEARCH_REMOTE_PROXY_ADDRESS", tt.proxyAddress)

			got, err := elasticsearchProxyAddressFromEnv()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got address %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRemoteClusterConnected(t *testing.T) {
	if !remoteClusterConnected(map[string]any{"connected": true}) {
		t.Fatal("expected map payload with connected=true to be treated as connected")
	}
	if remoteClusterConnected(map[string]any{"connected": false}) {
		t.Fatal("expected map payload with connected=false to be treated as disconnected")
	}
	if remoteClusterConnected(nil) {
		t.Fatal("expected nil remote info to be disconnected")
	}
}

func TestElasticsearchProxyAddressFromEnvUsesFirstEndpoint(t *testing.T) {
	t.Setenv("ELASTICSEARCH_ENDPOINTS", "http://first:9200,http://second:9200")
	t.Setenv("ELASTICSEARCH_TRANSPORT_PORT", "")
	t.Setenv("ELASTICSEARCH_REMOTE_PROXY_ADDRESS", "")

	got, err := elasticsearchProxyAddressFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "first:9300" {
		t.Fatalf("got %q, want %q", got, "first:9300")
	}
}
