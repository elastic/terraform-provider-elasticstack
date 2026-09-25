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

package spacesettings

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func newKibanaClientForServer(t *testing.T, handler http.HandlerFunc) *clients.KibanaScopedClient {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	t.Setenv("KIBANA_ENDPOINT", srv.URL)

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)
	return client
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

func TestReadSpaceSettings_notFoundRemovesResource(t *testing.T) {
	client := newKibanaClientForServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"statusCode":404,"error":"Not Found","message":"space not found"}`)
	})

	_, found, diags := readSpaceSettings(context.Background(), client, "team-a", "team-a", spaceSettingsModel{})

	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, found)
}

func TestReadSpaceSettings_apiErrorSurfacesDiagnostic(t *testing.T) {
	client := newKibanaClientForServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`)
	})
	prior := spaceSettingsModel{SpaceID: types.StringValue("team-a")}

	model, found, diags := readSpaceSettings(context.Background(), client, "team-a", "team-a", prior)

	require.True(t, diags.HasError())
	require.False(t, found)
	require.Equal(t, prior, model)
}
