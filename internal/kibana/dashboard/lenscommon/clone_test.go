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

package lenscommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneModel_Nil(t *testing.T) {
	t.Parallel()

	var model *struct{ Value string }
	cloned := CloneModel(model)
	assert.Nil(t, cloned)
}

func TestCloneModel_ShallowCopiesFields(t *testing.T) {
	t.Parallel()

	type sample struct {
		Value  string
		Nested *int
	}
	n := 5
	model := &sample{Value: "original", Nested: &n}

	cloned := CloneModel(model)

	require.NotNil(t, cloned)
	assert.NotSame(t, model, cloned)
	assert.Equal(t, model.Value, cloned.Value)
	assert.Same(t, model.Nested, cloned.Nested)

	cloned.Value = "changed"
	assert.Equal(t, "original", model.Value)
}
