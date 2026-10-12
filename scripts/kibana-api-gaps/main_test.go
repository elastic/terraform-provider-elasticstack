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

func TestSelectResourcesRejectsUnknownNames(t *testing.T) {
	all := []Resource{{Name: "a"}, {Name: "b"}}

	got, err := selectResources(all, []string{"b"})
	require.NoError(t, err)
	assert.Equal(t, []Resource{{Name: "b"}}, got)

	got, err = selectResources(all, []string{"b", "a", "b"})
	require.NoError(t, err)
	assert.Equal(t, all, got, "duplicates collapse and canonical order is kept")

	_, err = selectResources(all, []string{"typo"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "typo")

	got, err = selectResources(all, nil)
	require.NoError(t, err)
	assert.Equal(t, all, got)
}

func TestFindPackageBySuffixRequiresExactlyOneMatch(t *testing.T) {
	mod := loadFixture(t, "testdata/fixture")

	path, err := findPackageBySuffix(mod, "/kbapi")
	require.NoError(t, err)
	assert.Equal(t, fixtureKbapi, path)

	_, err = findPackageBySuffix(mod, "/missing")
	require.Error(t, err)
	_, err = findPackageBySuffix(mod, "s")
	require.Error(t, err)
}
