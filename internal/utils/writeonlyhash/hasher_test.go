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

package writeonlyhash_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHasher_Compute_Matches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		resourceTypeName string
		value            string
	}{
		{
			name:             "roundtrip",
			resourceTypeName: "elasticsearch_connector",
			value:            "x",
		},
		{
			name:             "non-trivial secret",
			resourceTypeName: "fleet_cloud_connector",
			value:            "super-secret-token-12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := writeonlyhash.New(tt.resourceTypeName)
			hash, err := h.Compute(tt.value)
			require.NoError(t, err)
			require.NotEmpty(t, hash)

			err = bcrypt.CompareHashAndPassword(hash, []byte(tt.resourceTypeName+":"+tt.value))
			require.NoError(t, err, "hash must be valid bcrypt output")

			assert.True(t, h.Matches(tt.value, hash))
		})
	}
}

func TestHasher_Matches_mismatch(t *testing.T) {
	t.Parallel()

	h := writeonlyhash.New("elasticsearch_connector")
	hash, err := h.Compute("x")
	require.NoError(t, err)

	assert.False(t, h.Matches("y", hash))
}

func TestHasher_saltIsolationAcrossResourceTypes(t *testing.T) {
	t.Parallel()

	const value = "x"

	ha := writeonlyhash.New("resource_a")
	hb := writeonlyhash.New("resource_b")

	hashA, err := ha.Compute(value)
	require.NoError(t, err)

	hashB, err := hb.Compute(value)
	require.NoError(t, err)

	assert.NotEqual(t, hashA, hashB)
	assert.True(t, ha.Matches(value, hashA))
	assert.False(t, ha.Matches(value, hashB))
	assert.True(t, hb.Matches(value, hashB))
	assert.False(t, hb.Matches(value, hashA))
}

func TestHasher_PrivateStateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attributePath string
		want          string
	}{
		{
			name:          "dot path",
			attributePath: "aws.external_id",
			want:          "secret_hash:aws.external_id",
		},
		{
			name:          "map-style path with quotes and brackets",
			attributePath: `configuration_values["password"].secret_value`,
			want:          `secret_hash:configuration_values["password"].secret_value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h1 := writeonlyhash.New("elasticsearch_connector")
			h2 := writeonlyhash.New("fleet_cloud_connector")

			got1 := h1.PrivateStateKey(tt.attributePath)
			got2 := h2.PrivateStateKey(tt.attributePath)

			assert.Equal(t, tt.want, got1)
			assert.Equal(t, got1, got2, "key must be stable across Hasher instances")
		})
	}
}

func TestHasher_Compute_customCostRoundtrip(t *testing.T) {
	t.Parallel()

	// Verifies that a caller-supplied Cost is used by Compute and that the
	// hash still roundtrips through Matches.
	h := &writeonlyhash.Hasher{Salt: []byte("elasticsearch_connector"), Cost: 4}
	hash, err := h.Compute("secret")
	require.NoError(t, err)
	cost, err := bcrypt.Cost(hash)
	require.NoError(t, err)
	assert.Equal(t, 4, cost)
	assert.True(t, h.Matches("secret", hash))
}

func TestHasher_Compute_defaultCost(t *testing.T) {
	t.Parallel()

	t.Run("default cost is 10", func(t *testing.T) {
		t.Parallel()

		h := writeonlyhash.New("elasticsearch_connector")
		hash, err := h.Compute("secret")
		require.NoError(t, err)
		cost, err := bcrypt.Cost(hash)
		require.NoError(t, err)
		assert.Equal(t, 10, cost)
	})
}

func TestHasher_Matches_malformedStoredHash(t *testing.T) {
	t.Parallel()

	t.Run("Matches handles malformed storedHash without panic", func(t *testing.T) {
		t.Parallel()

		h := writeonlyhash.New("elasticsearch_connector")
		cases := []struct {
			name string
			hash []byte
		}{
			{"nil", nil},
			{"empty", []byte{}},
			{"garbage", []byte("not-bcrypt-bytes")},
			{"truncated bcrypt prefix", []byte("$2a$10$tooshort")},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assert.False(t, h.Matches("x", tc.hash))
			})
		}
	})
}

func TestHasher_Compute_errorDoesNotLeakInput(t *testing.T) {
	t.Parallel()

	const secret = "must-not-appear-in-error"

	h := &writeonlyhash.Hasher{
		Salt: []byte("elasticsearch_connector"),
		Cost: 100,
	}

	_, err := h.Compute(secret)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), secret)
	assert.Contains(t, err.Error(), "bcrypt cost out of range")
}
