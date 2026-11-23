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

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitForFleetIntegrationInstalled_routesToScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    spaceScope
		wantPath string
	}{
		{
			name:     "empty scope uses default path",
			scope:    spaceScope{},
			wantPath: "/api/fleet/epm/packages/system/1.0.0",
		},
		{
			name:     "literal default uses default path",
			scope:    spaceScope{id: "default"},
			wantPath: "/api/fleet/epm/packages/system/1.0.0",
		},
		{
			name:     "custom scope uses space path",
			scope:    spaceScope{id: "my-space"},
			wantPath: "/s/my-space/api/fleet/epm/packages/system/1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"item":{"assets":{},"name":"system","title":"System","version":"1.0.0","status":"installed"}}`)
			}))
			defer srv.Close()

			err := waitForFleetIntegrationInstalled(
				t.Context(),
				newTestFleetClient(t, srv),
				testPackageName,
				testPackageVersion,
				tt.scope,
			)

			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, gotPath)
		})
	}
}

func TestWriteIntegration_allowOutdatedVersionRequestBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value types.Bool
		want  *bool // nil means the key must be omitted
	}{
		{name: "true is sent", value: types.BoolValue(true), want: new(true)},
		{name: "false is sent", value: types.BoolValue(false), want: new(false)},
		{name: "null omits key", value: types.BoolNull()},
		{name: "unknown omits key", value: types.BoolUnknown()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var captured map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/api/fleet/epm/packages/system/1.0.0") {
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
					fmt.Fprint(w, `{"items":[]}`)
					return
				}
				fmt.Fprint(w, `{"item":{"assets":{},"name":"system","title":"System","version":"1.0.0","status":"installed"}}`)
			}))
			defer srv.Close()

			client := clients.NewKibanaScopedClientWithFleetForTest(newTestFleetClient(t, srv))
			_, diags := writeIntegration(t.Context(), client, entitycore.KibanaWriteRequest[integrationModel]{
				Plan: integrationModel{
					Name:                 types.StringValue(testPackageName),
					Version:              types.StringValue(testPackageVersion),
					SpaceID:              types.StringNull(),
					AllowOutdatedVersion: tt.value,
				},
			})
			require.False(t, diags.HasError(), "%v", diags)
			require.NotNil(t, captured, "install request was not captured")

			if tt.want == nil {
				assert.NotContains(t, captured, "allow_outdated_version")
				return
			}
			assert.Equal(t, *tt.want, captured["allow_outdated_version"])
		})
	}
}
