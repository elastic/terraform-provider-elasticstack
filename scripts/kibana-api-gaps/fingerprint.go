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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

var (
	fingerprintScalarKeys = []string{"type", "format", "deprecated"}
	fingerprintMapKeys    = []string{"properties"}
	fingerprintSchemaKeys = []string{"items", "not"}
	fingerprintListKeys   = []string{"oneOf", "anyOf", "allOf"}
)

// fingerprint hashes the structure of the resolved subtree at ref. Descriptions, titles,
// examples, and x- extensions never take part, so documentation edits keep the fingerprint.
func (d *oasDoc) fingerprint(ref schemaRef) (string, error) {
	projection, err := d.project(ref, map[string]bool{})
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		return "", fmt.Errorf("fingerprint %s: %w", ref.pointer, err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func (d *oasDoc) project(ref schemaRef, active map[string]bool) (any, error) {
	resolved, err := d.resolve(ref)
	if err != nil {
		return nil, err
	}
	if active[resolved.pointer] {
		return map[string]any{"$cycle": resolved.pointer}, nil
	}
	active[resolved.pointer] = true
	defer delete(active, resolved.pointer)

	node := resolved.node
	out := make(map[string]any)
	for _, key := range fingerprintScalarKeys {
		if value := mappingValue(node, key); value != nil {
			out[key] = decodeNode(value)
		}
	}
	if value := mappingValue(node, "enum"); value != nil {
		out["enum"] = sortedStrings(value)
	}
	if value := mappingValue(node, "required"); value != nil {
		out["required"] = sortedStrings(value)
	}
	if err := d.projectChildren(resolved, out, active); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *oasDoc) projectChildren(resolved schemaRef, out map[string]any, active map[string]bool) error {
	node := resolved.node
	for _, key := range fingerprintMapKeys {
		value := mappingValue(node, key)
		if value == nil {
			continue
		}
		children := make(map[string]any)
		for i := 0; i+1 < len(value.Content); i += 2 {
			child, err := d.project(schemaRef{node: value.Content[i+1], pointer: resolved.pointer + "/" + key + "/" + escapePointerToken(value.Content[i].Value)}, active)
			if err != nil {
				return err
			}
			children[value.Content[i].Value] = child
		}
		out[key] = children
	}
	for _, key := range fingerprintSchemaKeys {
		if value := mappingValue(node, key); value != nil {
			child, err := d.project(schemaRef{node: value, pointer: resolved.pointer + "/" + key}, active)
			if err != nil {
				return err
			}
			out[key] = child
		}
	}
	for _, key := range fingerprintListKeys {
		value := mappingValue(node, key)
		if value == nil {
			continue
		}
		children := make([]any, 0, len(value.Content))
		for i, item := range value.Content {
			child, err := d.project(schemaRef{node: item, pointer: fmt.Sprintf("%s/%s/%d", resolved.pointer, key, i)}, active)
			if err != nil {
				return err
			}
			children = append(children, child)
		}
		out[key] = children
	}
	if value := mappingValue(node, "additionalProperties"); value != nil {
		if value.Kind == yaml.MappingNode {
			child, err := d.project(schemaRef{node: value, pointer: resolved.pointer + "/additionalProperties"}, active)
			if err != nil {
				return err
			}
			out["additionalProperties"] = child
		} else {
			out["additionalProperties"] = decodeNode(value)
		}
	}
	return nil
}

func decodeNode(node *yaml.Node) any {
	var value any
	if err := node.Decode(&value); err != nil {
		return node.Value
	}
	return value
}

func sortedStrings(node *yaml.Node) []string {
	values := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		values = append(values, item.Value)
	}
	sort.Strings(values)
	return values
}
