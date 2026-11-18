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

package elasticsearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupOrNotFoundDiag(t *testing.T) {
	t.Run("returns the value when the key is present", func(t *testing.T) {
		m := map[string]string{"foo": "bar"}
		value, diags := LookupOrNotFoundDiag(m, "foo", "widget")
		assert.False(t, diags.HasError())
		assert.Equal(t, "bar", *value)
	})

	t.Run("returns a not-found diagnostic when the key is absent", func(t *testing.T) {
		m := map[string]string{"foo": "bar"}
		value, diags := LookupOrNotFoundDiag(m, "missing", "widget")
		assert.Nil(t, value)
		assert.True(t, diags.HasError())
		assert.Equal(t, "Unable to find widget in the cluster", diags[0].Summary())
		assert.Equal(t, `Unable to find "missing" widget in the cluster`, diags[0].Detail())
	})

	t.Run("works with a named map type", func(t *testing.T) {
		type namedMap map[string]int
		m := namedMap{"foo": 1}
		value, diags := LookupOrNotFoundDiag(m, "foo", "widget")
		assert.False(t, diags.HasError())
		assert.Equal(t, 1, *value)
	})
}

func TestSingleOrNotFoundDiag(t *testing.T) {
	t.Run("returns the sole element", func(t *testing.T) {
		value, diags := SingleOrNotFoundDiag([]string{"bar"}, "foo", "widget")
		assert.False(t, diags.HasError())
		assert.Equal(t, "bar", *value)
	})

	t.Run("returns a not-found diagnostic when empty", func(t *testing.T) {
		value, diags := SingleOrNotFoundDiag([]string{}, "foo", "widget")
		assert.Nil(t, value)
		assert.True(t, diags.HasError())
		assert.Equal(t, "Unable to find widget in the cluster", diags[0].Summary())
		assert.Equal(t, `Unable to find "foo" widget in the cluster`, diags[0].Detail())
	})

	t.Run("returns a not-found diagnostic when more than one element", func(t *testing.T) {
		value, diags := SingleOrNotFoundDiag([]string{"bar", "baz"}, "foo", "widget")
		assert.Nil(t, value)
		assert.True(t, diags.HasError())
	})
}
