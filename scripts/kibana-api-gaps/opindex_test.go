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

func TestOpIndexMapsClientMethodsToOperations(t *testing.T) {
	mod := loadFixture(t, "testdata/fixture")
	idx, err := buildOpIndex(mod.pkg(t, fixtureKbapi))
	require.NoError(t, err)

	for _, method := range []string{"CreateWidget", "CreateWidgetWithResponse"} {
		op, ok := idx.method(method)
		require.True(t, ok, method)
		assert.Equal(t, "POST", op.Method)
		assert.Equal(t, "/api/widgets", op.Path)
	}
	op, ok := idx.method("GetWidgetWithResponse")
	require.True(t, ok)
	assert.Equal(t, "GET", op.Method)
	assert.Equal(t, "/api/widgets/%s", op.Path)
}

func TestOpIndexExposesConstructorRequestTypes(t *testing.T) {
	mod := loadFixture(t, "testdata/fixture")
	idx, err := buildOpIndex(mod.pkg(t, fixtureKbapi))
	require.NoError(t, err)

	op, ok := idx.method("CreateWidgetWithResponse")
	require.True(t, ok)
	require.NotNil(t, op.ParamsType)
	require.NotNil(t, op.BodyType)
	assert.Equal(t, "fixture/kbapi.CreateWidgetParams", op.ParamsType.String())
	assert.Equal(t, "fixture/kbapi.Widget", op.BodyType.String())

	get, ok := idx.method("GetWidget")
	require.True(t, ok)
	assert.Nil(t, get.ParamsType)
	assert.Nil(t, get.BodyType)
}

func TestOpIndexReportsClientMethodWithoutConstructor(t *testing.T) {
	mod := loadFixture(t, "testdata/badclient")
	_, err := buildOpIndex(mod.pkg(t, "badclient/kbapi"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Orphan")
}

func TestOpIndexReportsConstructorDelegationCycles(t *testing.T) {
	mod := loadFixture(t, "testdata/badctor")
	_, err := buildOpIndex(mod.pkg(t, "badctor/cycle"))
	require.Error(t, err)
}

func TestOpIndexReportsConstructorsWithoutLiteralPath(t *testing.T) {
	mod := loadFixture(t, "testdata/badctor")
	_, err := buildOpIndex(mod.pkg(t, "badctor/literalless"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NewOpRequest")
}
