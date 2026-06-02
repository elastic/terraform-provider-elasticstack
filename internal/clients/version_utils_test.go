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
	"testing"

	goversion "github.com/hashicorp/go-version"
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

	check := func(sv *goversion.Version) bool { return sv.GreaterThanOrEqual(minVer) }

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
