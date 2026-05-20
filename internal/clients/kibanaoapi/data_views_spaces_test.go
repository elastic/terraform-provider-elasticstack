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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateDataViewNamespaces(t *testing.T) {
	t.Parallel()

	const dataViewID = "dv"

	tests := []struct {
		name          string
		spaceID       string
		oldNS         []string
		newNS         []string
		responseBody  string
		wantPath      string
		wantError     bool
		wantNoRequest bool
		errorContains []string
		diagSummary   string
	}{
		{
			name:         "happy path default space",
			spaceID:      "default",
			oldNS:        []string{"default"},
			newNS:        []string{"default", "other"},
			responseBody: `{"objects":[{"id":"dv","type":"index-pattern"}]}`,
			wantPath:     "/api/spaces/_update_objects_spaces",
			wantError:    false,
		},
		{
			name:         "happy path non-default space",
			spaceID:      "test-space",
			oldNS:        []string{"test-space"},
			newNS:        []string{"test-space", "other"},
			responseBody: `{"objects":[{"id":"dv","type":"index-pattern"}]}`,
			wantPath:     "/s/test-space/api/spaces/_update_objects_spaces",
			wantError:    false,
		},
		{
			name:     "per-object error surfaced",
			spaceID:  "test-space",
			oldNS:    []string{"test-space"},
			newNS:    []string{"test-space", "target"},
			wantPath: "/s/test-space/api/spaces/_update_objects_spaces",
			responseBody: `{"objects":[{"id":"dv","type":"index-pattern","error":` +
				`{"statusCode":404,"error":"Not Found","message":"Saved object [index-pattern/dv] not found"}}]}`,
			wantError: true,
			errorContains: []string{
				"dv",
				"index-pattern",
				"Saved object [index-pattern/dv] not found",
			},
		},
		{
			name:     "multiple per-object errors",
			spaceID:  "test-space",
			oldNS:    []string{"test-space"},
			newNS:    []string{"test-space", "a", "b"},
			wantPath: "/s/test-space/api/spaces/_update_objects_spaces",
			responseBody: `{"objects":[` +
				`{"id":"dv1","type":"index-pattern","error":{"statusCode":404,"error":"Not Found","message":"missing dv1"}},` +
				`{"id":"dv2","type":"index-pattern","error":{"statusCode":404,"error":"Not Found","message":"missing dv2"}}` +
				`]}`,
			wantError: true,
			errorContains: []string{
				"missing dv1",
				"missing dv2",
			},
		},
		{
			name:          "empty objects array",
			spaceID:       "default",
			oldNS:         []string{"default"},
			newNS:         []string{"default", "other"},
			responseBody:  `{"objects":[]}`,
			wantPath:      "/api/spaces/_update_objects_spaces",
			wantError:     true,
			diagSummary:   "Unexpected response from data view namespace update",
			errorContains: []string{`{"objects":[]}`},
		},
		{
			name:          "malformed JSON body",
			spaceID:       "default",
			oldNS:         []string{"default"},
			newNS:         []string{"default", "other"},
			responseBody:  `not-json`,
			wantPath:      "/api/spaces/_update_objects_spaces",
			wantError:     true,
			errorContains: []string{"failed to parse response body"},
		},
		{
			name:          "no-op short circuit",
			spaceID:       "default",
			oldNS:         []string{"default", "a"},
			newNS:         []string{"default", "a"},
			wantNoRequest: true,
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var requestCount atomic.Int32
			var gotPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount.Add(1)
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, tt.responseBody)
			}))
			t.Cleanup(srv.Close)

			client := newTestClient(t, srv)
			diags := UpdateDataViewNamespaces(
				context.Background(),
				client,
				tt.spaceID,
				dataViewID,
				tt.oldNS,
				tt.newNS,
			)

			if tt.wantNoRequest {
				assert.Equal(t, int32(0), requestCount.Load())
				assert.False(t, diags.HasError())
				return
			}

			require.Equal(t, int32(1), requestCount.Load(), "expected exactly one HTTP request")
			assert.Equal(t, tt.wantPath, gotPath)

			if tt.wantError {
				require.True(t, diags.HasError())
				if tt.diagSummary != "" {
					assert.Equal(t, tt.diagSummary, diags.Errors()[0].Summary())
				}
				detail := strings.Join(func() []string {
					parts := make([]string, len(diags.Errors()))
					for i, d := range diags.Errors() {
						parts[i] = d.Detail()
					}
					return parts
				}(), " ")
				for _, substr := range tt.errorContains {
					assert.Contains(t, detail, substr)
				}
				return
			}

			assert.False(t, diags.HasError())
		})
	}
}
