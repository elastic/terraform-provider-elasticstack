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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
	assert.Equal(t, "/paths/~1api~1widgets/post/parameters/dry_run", op.Query[0].ref.pointer)
}

func TestOASMatchesPathTemplatesIgnoringParameterNames(t *testing.T) {
	doc, err := loadOAS(fixtureOAS)
	require.NoError(t, err)

	op, ok := doc.operation("GET", "/api/widgets/%s")
	require.True(t, ok)
	assert.Nil(t, op.Body)
	assert.Empty(t, op.Query)
}

func oasFromYAML(t *testing.T, text string) *oasDoc {
	t.Helper()
	path := filepath.Join(t.TempDir(), "oas.yaml")
	require.NoError(t, os.WriteFile(path, []byte(text), 0o600))
	doc, err := loadOAS(path)
	require.NoError(t, err)
	return doc
}

func TestOASExposesOperationLevelDeprecationAndXState(t *testing.T) {
	doc, err := loadOAS(fixtureOAS)
	require.NoError(t, err)

	op, ok := doc.operation("POST", "/api/gadgets")
	require.True(t, ok)
	assert.True(t, op.Deprecated)
	assert.Equal(t, "Added in 9.1.0", op.XState)
}

func TestOASResolveReportsBrokenReferences(t *testing.T) {
	doc := oasFromYAML(t, "paths: {}\ncomponents:\n  schemas:\n    A: {type: string}\n")

	_, err := doc.resolve(schemaRef{node: mustNode(t, "$ref: '#/components/schemas/Missing'")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing")

	_, err = doc.resolve(schemaRef{node: mustNode(t, "$ref: 'other.yaml#/A'")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-local")

	_, err = doc.resolve(schemaRef{node: mustNode(t, "$ref: '#/components/schemas/A/0'")})
	require.Error(t, err)
}

func TestOASResolveHandlesEscapedPointerTokens(t *testing.T) {
	doc := oasFromYAML(t, "paths: {}\ncomponents:\n  schemas:\n    'a/b~c': {type: string}\n")

	resolved, err := doc.resolve(schemaRef{node: mustNode(t, "$ref: '#/components/schemas/a~1b~0c'")})
	require.NoError(t, err)
	assert.Equal(t, "/components/schemas/a~1b~0c", resolved.pointer)
	assert.Equal(t, "string", stringValue(resolved.node, "type"))
}

func TestOASResolveRejectsReferenceCycles(t *testing.T) {
	doc := oasFromYAML(t, "paths: {}\ncomponents:\n  schemas:\n    A: {$ref: '#/components/schemas/B'}\n    B: {$ref: '#/components/schemas/A'}\n")

	_, err := doc.resolve(schemaRef{node: mustNode(t, "$ref: '#/components/schemas/A'")})
	require.Error(t, err)
}

func mustNode(t *testing.T, text string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(text), &doc))
	return doc.Content[0]
}

func TestLoadOASReportsHowToGenerateAMissingFile(t *testing.T) {
	_, err := loadOAS(filepath.Join(t.TempDir(), "oas-filtered.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "make -C generated/kbapi transform")
}
