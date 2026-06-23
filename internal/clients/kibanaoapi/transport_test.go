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
	"strings"
	"testing"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTransport_RoundTrip_AuthPriority(t *testing.T) {
	tests := []struct {
		name       string
		cfg        Config
		wantAuth   string
		wantNoAuth bool
	}{
		{
			name: "APIKey + BasicAuth set → only ApiKey header",
			cfg: Config{
				Username: "user",
				Password: "pass",
				APIKey:   "key",
			},
			wantAuth: "ApiKey key",
		},
		{
			name: "BearerToken set alongside others → only Bearer header",
			cfg: Config{
				Username:    "user",
				Password:    "pass",
				APIKey:      "key",
				BearerToken: "token",
			},
			wantAuth: "Bearer token",
		},
		{
			name: "BasicAuth only → only Basic header",
			cfg: Config{
				Username: "user",
				Password: "pass",
			},
			wantAuth: "Basic ",
		},
		{
			name:       "no auth fields set → no Authorization header",
			cfg:        Config{},
			wantNoAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &transport{
				Config: tt.cfg,
				next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					values := req.Header.Values("Authorization")
					if tt.wantNoAuth {
						if len(values) != 0 {
							t.Errorf("expected no Authorization header, got %d: %v", len(values), values)
						}
						return &http.Response{StatusCode: 200, Request: req}, nil
					}
					// Regression guard for issue #1393: exactly one Authorization
					// header must be sent regardless of how many auth fields are
					// populated on the config.
					if len(values) != 1 {
						t.Errorf("expected exactly one Authorization header, got %d: %v", len(values), values)
					}
					auth := req.Header.Get("Authorization")
					if strings.HasPrefix(tt.wantAuth, "Basic ") {
						if auth == "" || !strings.HasPrefix(auth, "Basic ") {
							t.Errorf("expected Basic auth header, got %q", auth)
						}
					} else if auth != tt.wantAuth {
						t.Errorf("expected Authorization %q, got %q", tt.wantAuth, auth)
					}
					return &http.Response{StatusCode: 200, Request: req}, nil
				}),
			}

			req := httptest.NewRequest("GET", "http://example.com", nil)
			resp, err := tr.RoundTrip(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		})
	}
}
