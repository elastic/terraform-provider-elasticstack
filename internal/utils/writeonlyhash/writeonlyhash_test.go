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
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("defaults cost to bcrypt default", func(t *testing.T) {
		t.Parallel()

		hasher := writeonlyhash.New("elasticstack_fleet_cloud_connector")
		require.Equal(t, bcrypt.DefaultCost, hasher.Cost)
	})

	t.Run("accepts empty resource type name", func(t *testing.T) {
		t.Parallel()

		hasher := writeonlyhash.New("")
		require.NotNil(t, hasher)

		hash, err := hasher.Compute("secret")
		require.NoError(t, err)
		assert.True(t, hasher.Matches("secret", hash))
	})
}

func TestComputeAndMatches(t *testing.T) {
	t.Parallel()

	hasher := writeonlyhash.New("elasticstack_fleet_cloud_connector")
	secret := "super-secret-token"

	t.Run("roundtrip matches stored hash", func(t *testing.T) {
		t.Parallel()

		hash, err := hasher.Compute(secret)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		_, err = bcrypt.Cost(writeonlyhash.DecodeStoredHash(hash))
		require.NoError(t, err)

		assert.True(t, hasher.Matches(secret, hash))
	})

	t.Run("different value does not match", func(t *testing.T) {
		t.Parallel()

		hash, err := hasher.Compute(secret)
		require.NoError(t, err)

		assert.False(t, hasher.Matches("other-secret-token", hash))
	})

	t.Run("nil stored hash returns false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, hasher.Matches(secret, nil))
	})

	t.Run("empty stored hash returns false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, hasher.Matches(secret, []byte{}))
	})

	t.Run("empty value roundtrip", func(t *testing.T) {
		t.Parallel()

		hash, err := hasher.Compute("")
		require.NoError(t, err)
		assert.True(t, hasher.Matches("", hash))
	})

	t.Run("invalid stored hash returns false without panic", func(t *testing.T) {
		t.Parallel()

		assert.False(t, hasher.Matches("value", []byte("not-a-bcrypt-hash")))
	})
}

func TestPerResourceTypeSeparation(t *testing.T) {
	t.Parallel()

	value := "shared-secret"
	fleetHasher := writeonlyhash.New("elasticstack_fleet_cloud_connector")
	connectorHasher := writeonlyhash.New("elasticstack_kibana_action_connector")

	fleetHash, err := fleetHasher.Compute(value)
	require.NoError(t, err)

	connectorHash, err := connectorHasher.Compute(value)
	require.NoError(t, err)

	t.Run("same value produces different hashes across resource types", func(t *testing.T) {
		t.Parallel()

		assert.NotEqual(t, fleetHash, connectorHash)
	})

	t.Run("hash from one resource type does not match on another", func(t *testing.T) {
		t.Parallel()

		assert.False(t, connectorHasher.Matches(value, fleetHash))
		assert.False(t, fleetHasher.Matches(value, connectorHash))
	})
}

func TestComputeCost(t *testing.T) {
	t.Parallel()

	t.Run("custom cost is reflected in hash", func(t *testing.T) {
		t.Parallel()

		hasher := writeonlyhash.New("elasticstack_fleet_cloud_connector")
		hasher.Cost = bcrypt.MinCost

		hash, err := hasher.Compute("secret")
		require.NoError(t, err)

		cost, err := bcrypt.Cost(writeonlyhash.DecodeStoredHash(hash))
		require.NoError(t, err)
		assert.Equal(t, bcrypt.MinCost, cost)
		assert.True(t, hasher.Matches("secret", hash))
	})

	t.Run("zero cost falls back to default", func(t *testing.T) {
		t.Parallel()

		hasher := writeonlyhash.New("elasticstack_fleet_cloud_connector")
		hasher.Cost = 0

		hash, err := hasher.Compute("secret")
		require.NoError(t, err)

		cost, err := bcrypt.Cost(writeonlyhash.DecodeStoredHash(hash))
		require.NoError(t, err)
		assert.Equal(t, bcrypt.DefaultCost, cost)
	})
}

func TestComputeErrors(t *testing.T) {
	t.Parallel()

	secret := "super-secret-token"

	t.Run("out of range cost returns error without leaking input", func(t *testing.T) {
		t.Parallel()

		hasher := writeonlyhash.New("elasticstack_fleet_cloud_connector")
		hasher.Cost = 100

		_, err := hasher.Compute(secret)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bcrypt cost out of range")
		assert.NotContains(t, err.Error(), secret)
	})

	t.Run("long secret succeeds without length error", func(t *testing.T) {
		t.Parallel()

		hasher := writeonlyhash.New("elasticstack_fleet_cloud_connector")
		longSecret := strings.Repeat("x", 200)

		hash, err := hasher.Compute(longSecret)
		require.NoError(t, err)
		assert.True(t, hasher.Matches(longSecret, hash))
	})
}

func TestPrivateStateKey(t *testing.T) {
	t.Parallel()

	hasherA := writeonlyhash.New("elasticstack_fleet_cloud_connector")
	hasherB := writeonlyhash.New("elasticstack_kibana_action_connector")
	path := "aws.external_id"

	t.Run("returns stable prefixed key", func(t *testing.T) {
		t.Parallel()

		key := hasherA.PrivateStateKey(path)
		assert.Equal(t, "secret_hash:aws.external_id", key)
		assert.Equal(t, key, hasherA.PrivateStateKey(path))
	})

	t.Run("is independent of resource type", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, hasherA.PrivateStateKey(path), hasherB.PrivateStateKey(path))
	})

	t.Run("preserves attribute path verbatim", func(t *testing.T) {
		t.Parallel()

		nestedPath := `vars["external_id"].secret_value`
		assert.True(t, strings.HasPrefix(hasherA.PrivateStateKey(nestedPath), "secret_hash:"))
		assert.Equal(t, "secret_hash:"+nestedPath, hasherA.PrivateStateKey(nestedPath))
	})
}
