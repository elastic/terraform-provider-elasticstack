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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fingerprintBase = `
paths: {}
components:
  schemas:
    Thing:
      type: object
      description: a thing
      required: [id]
      properties:
        id: {type: string, description: identifier, example: abc}
        mode: {type: string, enum: [fast, slow], x-state: Added in 8.1.0}
        child: {$ref: '#/components/schemas/Child'}
    Child:
      type: object
      properties:
        size: {type: integer}
`

func fingerprintOf(t *testing.T, yamlText string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "oas.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yamlText), 0o600))
	doc, err := loadOAS(path)
	require.NoError(t, err)
	node, err := doc.lookup("/components/schemas/Thing")
	require.NoError(t, err)
	fp, err := doc.fingerprint(schemaRef{node: node, pointer: "/components/schemas/Thing"})
	require.NoError(t, err)
	return fp
}

func TestFingerprintIgnoresDescriptionsExamplesAndExtensions(t *testing.T) {
	base := fingerprintOf(t, fingerprintBase)
	assert.Equal(t, base, fingerprintOf(t, withChange("description: identifier, example: abc", "description: changed text, example: xyz")))
	assert.Equal(t, base, fingerprintOf(t, withChange("x-state: Added in 8.1.0", "x-state: Added in 9.0.0")))
	assert.Equal(t, base, fingerprintOf(t, withChange("description: a thing", "description: another thing\n      title: Thing")))
}

func TestFingerprintChangesWithStructure(t *testing.T) {
	base := fingerprintOf(t, fingerprintBase)
	cases := map[string]string{
		"enum value added":      withChange("enum: [fast, slow]", "enum: [fast, slow, turbo]"),
		"property added":        withChange("        child:", "        extra: {type: string}\n        child:"),
		"required changed":      withChange("required: [id]", "required: [id, mode]"),
		"type changed":          withChange("id: {type: string,", "id: {type: integer,"),
		"deprecated":            withChange("mode: {type: string,", "mode: {type: string, deprecated: true,"),
		"referenced schema":     withChange("size: {type: integer}", "size: {type: integer}\n        depth: {type: integer}"),
		"format added":          withChange("id: {type: string,", "id: {type: string, format: uuid,"),
		"additional properties": withChange("      description: a thing\n", "      description: a thing\n      additionalProperties: {type: string}\n"),
	}
	for name, yamlText := range cases {
		t.Run(name, func(t *testing.T) {
			assert.NotEqual(t, base, fingerprintOf(t, yamlText))
		})
	}
}

func TestFingerprintHandlesRecursiveSchemas(t *testing.T) {
	recursive := withChange("        child: {$ref: '#/components/schemas/Child'}", "        child: {$ref: '#/components/schemas/Thing'}")
	assert.NotEmpty(t, fingerprintOf(t, recursive))
}

func withChange(oldText, newText string) string {
	return strings.Replace(fingerprintBase, oldText, newText, 1)
}
