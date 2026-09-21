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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteSpaceSettings_writesEmptyList(t *testing.T) {
	var method, path, body string
	client := newKibanaClientForServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		method, path, body = r.Method, r.URL.Path, string(raw)
		writeJSON(w, http.StatusOK, `{"item":{"allowed_namespace_prefixes":[]}}`)
	})

	diags := deleteSpaceSettings(context.Background(), client, "team-a", "team-a", spaceSettingsModel{})

	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, http.MethodPut, method)
	require.Equal(t, "/s/team-a/api/fleet/space_settings", path)
	require.JSONEq(t, `{"allowed_namespace_prefixes":[]}`, body)
}

func TestDeleteSpaceSettings_spaceAlreadyGoneSucceeds(t *testing.T) {
	client := newKibanaClientForServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"statusCode":404,"error":"Not Found","message":"space not found"}`)
	})

	diags := deleteSpaceSettings(context.Background(), client, "team-a", "team-a", spaceSettingsModel{})

	require.False(t, diags.HasError(), "%v", diags)
}

func TestDeleteSpaceSettings_apiErrorSurfacesDiagnostic(t *testing.T) {
	client := newKibanaClientForServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusForbidden, `{"statusCode":403,"error":"Forbidden","message":"missing privilege"}`)
	})

	diags := deleteSpaceSettings(context.Background(), client, "team-a", "team-a", spaceSettingsModel{})

	require.True(t, diags.HasError())
}
