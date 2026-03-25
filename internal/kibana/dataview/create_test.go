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

package dataview

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/require"
)

func TestCreateOrReconcileManagedDataView(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name             string
		body             kbapi.DataViewsCreateDataViewRequestObject
		getStatusCode    int
		expectRecovered  bool
		expectCreateFail bool
		expectGetCalls   int32
	}

	tests := []testCase{
		{
			name: "reconciles managed create after error response",
			body: kbapi.DataViewsCreateDataViewRequestObject{
				DataView: kbapi.DataViewsCreateDataViewRequestObjectInner{
					Id:    new("managed-id"),
					Title: "logs-*",
					Name:  new("logs"),
				},
			},
			getStatusCode:    http.StatusOK,
			expectRecovered:  true,
			expectCreateFail: false,
			expectGetCalls:   1,
		},
		{
			name: "surfaces original create error when recovery read misses",
			body: kbapi.DataViewsCreateDataViewRequestObject{
				DataView: kbapi.DataViewsCreateDataViewRequestObjectInner{
					Id:    new("managed-id"),
					Title: "logs-*",
				},
			},
			getStatusCode:    http.StatusNotFound,
			expectRecovered:  false,
			expectCreateFail: true,
			expectGetCalls:   1,
		},
		{
			name: "does not attempt reconciliation without explicit id",
			body: kbapi.DataViewsCreateDataViewRequestObject{
				DataView: kbapi.DataViewsCreateDataViewRequestObjectInner{
					Title: "logs-*",
				},
			},
			getStatusCode:    http.StatusOK,
			expectRecovered:  false,
			expectCreateFail: true,
			expectGetCalls:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var getCalls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("Content-Type", "application/json")

				switch {
				case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/api/data_views/data_view"):
					rw.WriteHeader(http.StatusBadRequest)
					if _, err := rw.Write([]byte(`{"error":"synthetic create failure"}`)); err != nil {
						t.Fatalf("write create error response: %v", err)
					}
				case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/api/data_views/data_view/managed-id"):
					getCalls.Add(1)
					rw.WriteHeader(tt.getStatusCode)
					if tt.getStatusCode == http.StatusOK {
						if _, err := rw.Write([]byte(`{"data_view":{"id":"managed-id","title":"logs-*","name":"logs"}}`)); err != nil {
							t.Fatalf("write recovered data view response: %v", err)
						}
					}
				default:
					t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
				}
			}))
			defer server.Close()

			client, err := kibanaoapi.NewClient(kibanaoapi.Config{
				URL:      server.URL,
				Username: "elastic",
				Password: "password",
			})
			require.NoError(t, err)

			dataView, diags := createOrReconcileManagedDataView(t.Context(), client, "recovery-space", tt.body)

			require.Equal(t, tt.expectGetCalls, getCalls.Load())
			require.Equal(t, tt.expectCreateFail, diags.HasError())
			if tt.expectRecovered {
				require.NotNil(t, dataView)
				require.Equal(t, "managed-id", *dataView.DataView.Id)
				require.Equal(t, "logs-*", *dataView.DataView.Title)
				return
			}

			require.Nil(t, dataView)
			require.Len(t, diags, 1)
			require.Contains(t, diags[0].Summary(), "Unexpected status code from server: got HTTP 400")
		})
	}
}
