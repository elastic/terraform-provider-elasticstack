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

func variantGroup(resource, schema, field, fingerprint string, members ...string) Group {
	pointer := "/components/schemas/" + schema + "/properties/" + field
	g := Group{Key: resource + "#" + pointer, Resource: resource, Field: field, Pointer: pointer, Fingerprint: fingerprint}
	for _, m := range members {
		g.Members = append(g.Members, Member{Field: m, Pointer: pointer + "/properties/" + m})
	}
	return g
}

func TestMergeGroupsCollapsesSameResourceAndFieldAcrossBodiesUsingTheSmallestPointer(t *testing.T) {
	create := variantGroup("r", "Create", "settings", "fc", "settings.x")
	update := variantGroup("r", "Update", "settings", "fu", "settings.x", "settings.y")

	merged := mergeGroups([]Group{update, create})

	require.Len(t, merged, 1)
	assert.Equal(t, create.Key, merged[0].Key)
	assert.Equal(t, create.Pointer, merged[0].Pointer)
	assert.Equal(t, "settings", merged[0].Field)
	assert.Len(t, merged[0].Members, 3, "members keep their own locations")
	assert.Equal(t, mergeGroups([]Group{create, update}), merged, "input order does not matter")
}

func TestMergeGroupsKeepsSingleGroupsUnchanged(t *testing.T) {
	g := variantGroup("r", "Create", "settings", "fc", "settings.x")
	assert.Equal(t, []Group{g}, mergeGroups([]Group{g}))
}

func TestMergeGroupsFingerprintFollowsEveryConstituent(t *testing.T) {
	create := variantGroup("r", "Create", "settings", "fc", "settings.x")
	update := variantGroup("r", "Update", "settings", "fu", "settings.x")
	base := mergeGroups([]Group{create, update})[0].Fingerprint

	changed := update
	changed.Fingerprint = "fu2"
	assert.NotEqual(t, base, mergeGroups([]Group{create, changed})[0].Fingerprint)

	third := variantGroup("r", "Other", "settings", "fo", "settings.x")
	assert.NotEqual(t, base, mergeGroups([]Group{create, update, third})[0].Fingerprint)
	assert.NotEqual(t, create.Fingerprint, base)
}

func TestMergeGroupsKeepsDifferentResourcesAndFieldsApart(t *testing.T) {
	groups := []Group{
		variantGroup("r", "Create", "settings", "f", "a"),
		variantGroup("s", "Create", "settings", "f", "a"),
		variantGroup("r", "Create", "labels", "f", "a"),
	}
	assert.Len(t, mergeGroups(groups), 3)
}

func TestMergeGroupsReturnsGroupsSortedByKey(t *testing.T) {
	merged := mergeGroups([]Group{
		variantGroup("b", "Create", "x", "f", "a"),
		variantGroup("a", "Create", "x", "f", "a"),
	})
	assert.Equal(t, "a", merged[0].Resource)
}
