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

package kibanaoapi

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
)

func TestOsqueryPackShardsFromMap(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		assert.Nil(t, osqueryPackShardsFromMap(nil))
	})

	t.Run("returns nil for empty map", func(t *testing.T) {
		empty := kbapi.SecurityOsqueryAPIShards{}
		assert.Nil(t, osqueryPackShardsFromMap(&empty))
	})

	t.Run("converts float32 map to float64", func(t *testing.T) {
		shards := kbapi.SecurityOsqueryAPIShards{
			"policy-a": 50,
			"policy-b": 100,
		}

		result := osqueryPackShardsFromMap(&shards)

		assert.Equal(t, OsqueryPackShards{"policy-a": 50, "policy-b": 100}, result)
	})
}

func TestOsqueryPackShardsFromCreateArray(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		assert.Nil(t, osqueryPackShardsFromCreateArray(nil))
	})

	t.Run("returns nil for empty array", func(t *testing.T) {
		empty := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{}
		assert.Nil(t, osqueryPackShardsFromCreateArray(&empty))
	})

	t.Run("converts key-value array to map", func(t *testing.T) {
		keyA := "policy-a"
		keyB := "policy-b"
		valA := float32(25)
		valB := float32(75)
		shards := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{
			{Key: &keyA, Value: &valA},
			{Key: &keyB, Value: &valB},
		}

		result := osqueryPackShardsFromCreateArray(&shards)

		assert.Equal(t, OsqueryPackShards{"policy-a": 25, "policy-b": 75}, result)
	})

	t.Run("skips entries with nil key", func(t *testing.T) {
		val := float32(50)
		shards := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{
			{Key: nil, Value: &val},
		}

		assert.Nil(t, osqueryPackShardsFromCreateArray(&shards))
	})
}
