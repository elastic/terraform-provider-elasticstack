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

package debugutils

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogHTTPRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://example.com/path", strings.NewReader(`{"foo":"bar"}`))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Authorization", "secret-token")

	// Should not panic and must leave the request body intact for the real
	// round trip that follows the dump.
	LogHTTPRequest(context.Background(), "test", req)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body after logging: %v", err)
	}
	if string(body) != `{"foo":"bar"}` {
		t.Errorf("request body was mutated by LogHTTPRequest, got %q", string(body))
	}
}

func TestLogHTTPResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusOK)
	_, _ = rec.WriteString(`{"foo":"bar"}`)
	resp := rec.Result()

	// Should not panic and must leave the response body intact for the
	// caller that still needs to read it.
	LogHTTPResponse(context.Background(), "test", resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body after logging: %v", err)
	}
	if string(body) != `{"foo":"bar"}` {
		t.Errorf("response body was mutated by LogHTTPResponse, got %q", string(body))
	}
}

func TestNewDebugTransport_RoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"ping":"pong"}` {
			t.Errorf("server received unexpected body: %q", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := &http.Client{Transport: NewDebugTransport("test", http.DefaultTransport)}

	req, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(`{"ping":"pong"}`))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("expected status %d, got %d", http.StatusTeapot, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("unexpected response body: %q", string(body))
	}
}
