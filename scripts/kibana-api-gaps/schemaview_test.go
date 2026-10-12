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

const viewDoc = `
paths: {}
components:
  schemas:
    Base:
      type: object
      required: [id]
      properties:
        id: {type: string}
    Merged:
      allOf:
        - $ref: '#/components/schemas/Base'
        - type: object
          required: [extra]
          properties:
            extra: {type: string}
            id: {type: integer}
    Own:
      type: object
      required: [a]
      properties:
        a: {type: string}
        b: {type: string}
    Choice:
      oneOf:
        - {type: string}
        - {type: integer}
    Either:
      allOf:
        - $ref: '#/components/schemas/Choice'
    Negated:
      not: {type: string}
    ReadOnlyRef: {$ref: '#/components/schemas/Own', readOnly: true}
`

func viewOf(t *testing.T, name string) *schemaView {
	t.Helper()
	doc := oasFromYAML(t, viewDoc)
	pointer := "/components/schemas/" + name
	node, err := doc.lookup(pointer)
	require.NoError(t, err)
	v, err := doc.view(schemaRef{node: node, pointer: pointer})
	require.NoError(t, err)
	return v
}

func propertyNames(v *schemaView) []string {
	var names []string
	for _, p := range v.properties {
		names = append(names, p.name)
	}
	return names
}

func TestViewMergesAllOfMembersAndLetsLaterMembersOverride(t *testing.T) {
	v := viewOf(t, "Merged")

	assert.Equal(t, []string{"id", "extra"}, propertyNames(v))
	assert.Equal(t, "/components/schemas/Merged/allOf/1/properties/id", v.properties[0].ref.pointer)
	assert.Empty(t, v.unsupported)
}

func TestViewRecordsRequiredPerProperty(t *testing.T) {
	v := viewOf(t, "Own")

	require.Len(t, v.properties, 2)
	assert.True(t, v.properties[0].required)
	assert.False(t, v.properties[1].required)
}

func TestViewCollectsOneOfBranches(t *testing.T) {
	v := viewOf(t, "Choice")

	require.Len(t, v.oneOf, 2)
	assert.Equal(t, "/components/schemas/Choice/oneOf/1", v.oneOf[1].pointer)
}

func TestViewFlagsAllOfMembersCarryingCombinators(t *testing.T) {
	assert.Contains(t, viewOf(t, "Either").unsupported, "allOf")
}

func TestViewFlagsNotAsUnsupported(t *testing.T) {
	assert.Equal(t, "not", viewOf(t, "Negated").unsupported)
}

func TestViewReadsFlagsFromTheReferencingNode(t *testing.T) {
	assert.True(t, viewOf(t, "ReadOnlyRef").readOnly)
}
