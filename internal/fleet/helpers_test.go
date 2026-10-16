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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newKibanaScopedClientForTest returns a KibanaScopedClient backed by an
// httptest.Server that reports the given version and build flavor on
// GET /api/status, which is what EnforceMinVersion/EnforceVersionCheck query.
func newKibanaScopedClientForTest(t *testing.T, stackVersion, buildFlavor string) *clients.KibanaScopedClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":%q,"build_flavor":%q}}`, stackVersion, buildFlavor)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	kibOapi, err := kibanaoapi.NewClient(kibanaoapi.Config{URL: srv.URL})
	require.NoError(t, err)

	return clients.NewKibanaScopedClientForTest(kibOapi)
}

func TestEnforceMinVersionInto_Satisfied(t *testing.T) {
	t.Parallel()
	client := newKibanaScopedClientForTest(t, "8.15.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	var target bool
	var diags diag.Diagnostics
	EnforceMinVersionInto(t.Context(), client, &target, &diags, minVer)

	require.False(t, diags.HasError())
	assert.True(t, target)
}

func TestEnforceMinVersionInto_NotSatisfied(t *testing.T) {
	t.Parallel()
	client := newKibanaScopedClientForTest(t, "8.10.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	var target bool
	var diags diag.Diagnostics
	EnforceMinVersionInto(t.Context(), client, &target, &diags, minVer)

	require.False(t, diags.HasError())
	assert.False(t, target)
}

func TestEnforceMinVersionInto_AppendsErrorDiagnostics(t *testing.T) {
	t.Parallel()
	client := newKibanaScopedClientForTest(t, "not-a-version", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	var target bool
	diags := diag.Diagnostics{diag.NewWarningDiagnostic("pre-existing", "warning")}
	EnforceMinVersionInto(t.Context(), client, &target, &diags, minVer)

	require.True(t, diags.HasError())
	assert.False(t, target)
	assert.Len(t, diags, 2, "pre-existing diagnostics must be preserved alongside the new error")
}

func TestEnforceVersionCheckInto_Satisfied(t *testing.T) {
	t.Parallel()
	client := newKibanaScopedClientForTest(t, "8.15.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	var target bool
	var diags diag.Diagnostics
	EnforceVersionCheckInto(t.Context(), client, &target, &diags, func(v *version.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})

	require.False(t, diags.HasError())
	assert.True(t, target)
}

func TestEnforceVersionCheckInto_NotSatisfied(t *testing.T) {
	t.Parallel()
	client := newKibanaScopedClientForTest(t, "8.10.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	var target bool
	var diags diag.Diagnostics
	EnforceVersionCheckInto(t.Context(), client, &target, &diags, func(v *version.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})

	require.False(t, diags.HasError())
	assert.False(t, target)
}

func TestEnforceVersionCheckInto_ServerlessShortCircuit(t *testing.T) {
	t.Parallel()
	client := newKibanaScopedClientForTest(t, "8.10.0", "serverless")

	var target bool
	var diags diag.Diagnostics
	EnforceVersionCheckInto(t.Context(), client, &target, &diags, func(v *version.Version) bool {
		return false
	})

	require.False(t, diags.HasError())
	assert.True(t, target, "serverless must short-circuit to true regardless of the check function")
}
