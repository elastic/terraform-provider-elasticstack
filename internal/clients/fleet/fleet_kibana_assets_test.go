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

package fleet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallKibanaAssets_RequestShapeAndSpacePath(t *testing.T) {
	t.Parallel()

	var gotMethod, gotPath string
	var gotHeaderXSRF string
	var gotContentType string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotHeaderXSRF = r.Header.Get("kbn-xsrf")
		gotContentType = r.Header.Get("Content-Type")

		// Avoid require.* in handlers (different goroutine) to satisfy testifylint.
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("failed decoding request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{URL: server.URL})
	require.NoError(t, err)

	diags := InstallKibanaAssets(t.Context(), client, "tcp", "1.16.0", "my-space", true)
	require.False(t, diags.HasError())

	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, "/s/my-space/api/fleet/epm/packages/tcp/1.16.0/kibana_assets", gotPath)
	require.Equal(t, "true", gotHeaderXSRF)
	require.Contains(t, gotContentType, "application/json")
	require.Equal(t, true, gotBody["force"])
}

func TestInstallKibanaAssets_ReportsNonOKAsDiagnostic(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client, err := NewClient(Config{URL: server.URL})
	require.NoError(t, err)

	diags := InstallKibanaAssets(t.Context(), client, "tcp", "1.16.0", "", true)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "Unexpected status code")
	require.Contains(t, diags[0].Detail(), "boom")
}
