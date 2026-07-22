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

package clients

import (
	"context"
	"testing"

	goversion "github.com/hashicorp/go-version"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyVersionConstraint_ServerlessShortCircuit(t *testing.T) {
	t.Parallel()
	ok, diags := applyVersionConstraint(ServerlessFlavor, "8.10.0", func(_ *goversion.Version) bool {
		return false
	})
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must short-circuit to true regardless of check result")
}

func TestApplyVersionConstraint_StatefulCheckTrue(t *testing.T) {
	t.Parallel()
	ok, diags := applyVersionConstraint("default", "8.15.0", func(_ *goversion.Version) bool {
		return true
	})
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestApplyVersionConstraint_StatefulCheckFalse(t *testing.T) {
	t.Parallel()
	ok, diags := applyVersionConstraint("default", "8.10.0", func(_ *goversion.Version) bool {
		return false
	})
	require.False(t, diags.HasError())
	assert.False(t, ok)
}

func TestApplyVersionConstraint_MalformedVersion(t *testing.T) {
	t.Parallel()
	ok, diags := applyVersionConstraint("default", "not-a-version", func(_ *goversion.Version) bool {
		return true
	})
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestApplyVersionConstraint_VersionComparison(t *testing.T) {
	t.Parallel()

	minVer, err := goversion.NewVersion("8.15.0")
	require.NoError(t, err)

	check := func(sv *goversion.Version) bool { return versionAtLeastRelease(sv, minVer) }

	tests := []struct {
		rawVersion string
		want       bool
	}{
		{"7.17.0", false},
		{"8.14.9", false},
		{"8.15.0", true},
		{"8.15.1", true},
		{"9.0.0", true},
	}

	for _, tc := range tests {
		ok, diags := applyVersionConstraint("default", tc.rawVersion, check)
		require.False(t, diags.HasError(), "version %s", tc.rawVersion)
		assert.Equal(t, tc.want, ok, "version %s", tc.rawVersion)
	}
}

// TestKibanaScopedClient_EnforceMinVersion_NilVersion verifies that the nil
// guard added to KibanaScopedClient.EnforceMinVersion matches the existing
// ElasticsearchScopedClient behaviour: passing nil returns (true, nil) without
// contacting the server.
func TestKibanaScopedClient_EnforceMinVersion_NilVersion(t *testing.T) {
	t.Parallel()
	// Use a client with no configured server — if it contacts the server it
	// will fail, proving the nil guard short-circuits before the network call.
	sc := &KibanaScopedClient{}
	ok, diags := sc.EnforceMinVersion(t.Context(), nil)
	require.False(t, diags.HasError())
	assert.True(t, ok, "nil minVersion must short-circuit to true without a network call")
}

// TestElasticsearchScopedClient_EnforceMinVersion_NilVersion verifies the
// existing nil guard in ElasticsearchScopedClient.EnforceMinVersion.
func TestElasticsearchScopedClient_EnforceMinVersion_NilVersion(t *testing.T) {
	t.Parallel()
	sc := &ElasticsearchScopedClient{}
	ok, diags := sc.EnforceMinVersion(t.Context(), nil)
	require.False(t, diags.HasError())
	assert.True(t, ok, "nil minVersion must short-circuit to true without a network call")
}

// --- shared helper tests ---

func makeFetcher(rawVersion, flavor string) versionFetcher {
	return func(_ context.Context) (string, string, fwdiag.Diagnostics) {
		return rawVersion, flavor, nil
	}
}

func makeErrorFetcher() versionFetcher {
	return func(_ context.Context) (string, string, fwdiag.Diagnostics) {
		return "", "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("fetch error", "simulated fetch failure")}
	}
}

func TestEnforceMinVersion_NilVersion(t *testing.T) {
	t.Parallel()
	// fetch must never be called when minVersion is nil.
	ok, diags := enforceMinVersion(t.Context(), nil, makeErrorFetcher(), versionAtLeastRelease)
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestEnforceMinVersion_ServerlessShortCircuit(t *testing.T) {
	t.Parallel()
	minVer := goversion.Must(goversion.NewVersion("99.0.0"))
	ok, diags := enforceMinVersion(t.Context(), minVer, makeFetcher("8.10.0", ServerlessFlavor), versionAtLeastRelease)
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must satisfy any version gate")
}

func TestEnforceMinVersion_Satisfied(t *testing.T) {
	t.Parallel()
	minVer := goversion.Must(goversion.NewVersion("8.0.0"))
	ok, diags := enforceMinVersion(t.Context(), minVer, makeFetcher("8.19.0", "default"), versionAtLeastRelease)
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestEnforceMinVersion_NotSatisfied(t *testing.T) {
	t.Parallel()
	minVer := goversion.Must(goversion.NewVersion("8.0.0"))
	ok, diags := enforceMinVersion(t.Context(), minVer, makeFetcher("7.17.0", "default"), versionAtLeastRelease)
	require.False(t, diags.HasError())
	assert.False(t, ok)
}

func TestEnforceMinVersion_FetchError(t *testing.T) {
	t.Parallel()
	minVer := goversion.Must(goversion.NewVersion("8.0.0"))
	ok, diags := enforceMinVersion(t.Context(), minVer, makeErrorFetcher(), versionAtLeastRelease)
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestEnforceVersionCheck_ServerlessShortCircuit(t *testing.T) {
	t.Parallel()
	ok, diags := enforceVersionCheck(t.Context(), func(_ *goversion.Version) bool { return false }, makeFetcher("8.10.0", ServerlessFlavor))
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must short-circuit to true even when check returns false")
}

func TestEnforceVersionCheck_CheckTrue(t *testing.T) {
	t.Parallel()
	ok, diags := enforceVersionCheck(t.Context(), func(_ *goversion.Version) bool { return true }, makeFetcher("8.15.0", "default"))
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestEnforceVersionCheck_CheckFalse(t *testing.T) {
	t.Parallel()
	ok, diags := enforceVersionCheck(t.Context(), func(_ *goversion.Version) bool { return false }, makeFetcher("8.15.0", "default"))
	require.False(t, diags.HasError())
	assert.False(t, ok)
}

func TestEnforceVersionCheck_FetchError(t *testing.T) {
	t.Parallel()
	ok, diags := enforceVersionCheck(t.Context(), func(_ *goversion.Version) bool { return true }, makeErrorFetcher())
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestKibanaVersionAtLeastRelease_SnapshotBuildOnReleaseFloor(t *testing.T) {
	t.Parallel()

	minVer := goversion.Must(goversion.NewVersion("9.5.0"))

	tests := []struct {
		server string
		want   bool
		name   string
	}{
		{"9.4.0-SNAPSHOT", false, "snapshot below core"},
		{"9.5.0-SNAPSHOT", true, "snapshot same core as release floor"},
		{"9.5.0", true, "release equal"},
		{"9.5.1", true, "release above"},
		{"9.5.0-beta.1", false, "non-SNAPSHOT prerelease on same core"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sv, err := goversion.NewVersion(tc.server)
			require.NoError(t, err)
			assert.Equal(t, tc.want, kibanaVersionAtLeastRelease(sv, minVer), "server %q", tc.server)
		})
	}
}

func TestSnapshotBuildSatisfiesReleaseMinimum_MinimumAlsoPrerelease(t *testing.T) {
	t.Parallel()
	server := goversion.Must(goversion.NewVersion("9.5.0-SNAPSHOT"))
	minBeta := goversion.Must(goversion.NewVersion("9.5.0-beta.1"))
	assert.False(t, snapshotBuildSatisfiesReleaseMinimum(server, minBeta))
}

func TestEnforceMinVersion_KibanaSnapshotBuildMeetsReleaseMinimum(t *testing.T) {
	t.Parallel()
	minVer := goversion.Must(goversion.NewVersion("9.5.0"))
	ok, diags := enforceMinVersion(t.Context(), minVer, makeFetcher("9.5.0-SNAPSHOT", "default"), kibanaVersionAtLeastRelease)
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestEnforceMinVersion_ElasticsearchSnapshotBelowReleaseMinimum(t *testing.T) {
	t.Parallel()
	minVer := goversion.Must(goversion.NewVersion("9.5.0"))
	ok, diags := enforceMinVersion(t.Context(), minVer, makeFetcher("9.5.0-SNAPSHOT", "default"), versionAtLeastRelease)
	require.False(t, diags.HasError())
	assert.False(t, ok)
}
