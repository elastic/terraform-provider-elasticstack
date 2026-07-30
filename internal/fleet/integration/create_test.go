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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
