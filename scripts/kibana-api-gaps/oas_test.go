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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixtureOAS = "testdata/fixture/oas.yaml"

func TestOASExposesOperationBodyQueryAndFlags(t *testing.T) {
	doc, err := loadOAS(fixtureOAS)
	require.NoError(t, err)

	op, ok := doc.operation("POST", "/api/widgets")
	require.True(t, ok)

	assert.False(t, op.Deprecated)
	require.NotNil(t, op.Body)
	assert.Equal(t, "/components/schemas/Widget", op.Body.pointer)
	require.Len(t, op.Query, 2)
	assert.Equal(t, "dry_run", op.Query[0].name)
	assert.Equal(t, "/paths/~1api~1widgets/post/parameters/0", op.Query[0].ref.pointer)
}

func TestOASMatchesPathTemplatesIgnoringParameterNames(t *testing.T) {
	doc, err := loadOAS(fixtureOAS)
	require.NoError(t, err)

	op, ok := doc.operation("GET", "/api/widgets/%s")
	require.True(t, ok)
	assert.Nil(t, op.Body)
	assert.Empty(t, op.Query)
}
