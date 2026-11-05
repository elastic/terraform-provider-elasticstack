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
	"reflect"
	"strings"
	"testing"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func executeRoundTrip(t *testing.T, tr *transport, req *http.Request) {
	t.Helper()

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func TestConfig_carriesCustomHeaders(t *testing.T) {
	headers, ok := reflect.TypeFor[Config]().FieldByName("Headers")

	if !ok {
		t.Fatal("Config does not carry custom headers")
	}
	if headers.Type != reflect.TypeFor[map[string]string]() {
		t.Fatalf("Config.Headers is %v, want map[string]string", headers.Type)
	}
}

func TestTransport_RoundTrip_sendsProviderConfiguredHeaders(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			tr := &transport{
				Headers: map[string]string{"X-Proxy-Token": "secret"},
				next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					if got := req.Header.Get("X-Proxy-Token"); got != "secret" {
						t.Errorf("X-Proxy-Token = %q, want %q", got, "secret")
					}
					if got := req.Header.Get("kbn-xsrf"); got != "" {
						t.Errorf("kbn-xsrf = %q, want empty", got)
					}
					return &http.Response{StatusCode: http.StatusOK, Request: req}, nil
				}),
			}

			req := httptest.NewRequest(method, "http://example.com", nil)
			executeRoundTrip(t, tr, req)
		})
	}
}

func TestTransport_RoundTrip_usesCustomAuthorizationWithoutNativeAuthentication(t *testing.T) {
	tr := &transport{
		Headers: map[string]string{"Authorization": "Proxy secret"},
		next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.Header.Get("Authorization"); got != "Proxy secret" {
				t.Errorf("Authorization = %q, want %q", got, "Proxy secret")
			}
			return &http.Response{StatusCode: http.StatusOK, Request: req}, nil
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	executeRoundTrip(t, tr, req)
}

func TestTransport_RoundTrip_sendsCustomHeaderAlongsideKbnXsrf(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			tr := &transport{
				Headers: map[string]string{"cf-access-token": "xyz"},
				next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					if got := req.Header.Get("cf-access-token"); got != "xyz" {
						t.Errorf("cf-access-token = %q, want %q", got, "xyz")
					}
					if got := req.Header.Get("kbn-xsrf"); got != "true" {
						t.Errorf("kbn-xsrf = %q, want %q", got, "true")
					}
					return &http.Response{StatusCode: http.StatusOK, Request: req}, nil
				}),
			}

			req := httptest.NewRequest(method, "http://example.com", nil)
			executeRoundTrip(t, tr, req)
		})
	}
}

func TestTransport_RoundTrip_omittedHeadersPreserveExistingBehavior(t *testing.T) {
	tr := &transport{
		Username: "user",
		Password: "pass",
		next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.Header.Get("X-Existing"); got != "existing" {
				t.Errorf("X-Existing = %q, want %q", got, "existing")
			}
			if username, password, ok := req.BasicAuth(); !ok || username != "user" || password != "pass" {
				t.Errorf("BasicAuth = %q, %q, %t; want %q, %q, true", username, password, ok, "user", "pass")
			}
			return &http.Response{StatusCode: http.StatusOK, Request: req}, nil
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Existing", "existing")
	executeRoundTrip(t, tr, req)
}

func TestTransport_RoundTrip_nativeAuthenticationReplacesCustomAuthorization(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		wantAuth string
	}{
		{
			name: "username and password",
			cfg: Config{
				Username: "user",
				Password: "pass",
			},
			wantAuth: "Basic ",
		},
		{
			name:     "API key",
			cfg:      Config{APIKey: "key"},
			wantAuth: "ApiKey key",
		},
		{
			name:     "bearer token",
			cfg:      Config{BearerToken: "token"},
			wantAuth: "Bearer token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.Headers = map[string]string{"Authorization": "Proxy secret"}
			tr := &transport{
				Config: tt.cfg,
				next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					values := req.Header.Values("Authorization")
					if len(values) != 1 {
						t.Errorf("Authorization values = %v, want exactly one", values)
					}
					auth := req.Header.Get("Authorization")
					if strings.HasPrefix(tt.wantAuth, "Basic ") {
						if !strings.HasPrefix(auth, "Basic ") {
							t.Errorf("Authorization = %q, want Basic authentication", auth)
						}
					} else if auth != tt.wantAuth {
						t.Errorf("Authorization = %q, want %q", auth, tt.wantAuth)
					}
					return &http.Response{StatusCode: http.StatusOK, Request: req}, nil
				}),
			}

			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			executeRoundTrip(t, tr, req)
		})
	}
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
				_ = resp.Body.Close()
			}
		})
	}
}
