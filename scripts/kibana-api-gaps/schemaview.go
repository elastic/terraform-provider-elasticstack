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
	"fmt"

	"gopkg.in/yaml.v3"
)

type schemaProperty struct {
	name     string
	ref      schemaRef
	required bool
}

// schemaView is a resolved schema node with $ref followed and allOf members merged.
type schemaView struct {
	pointer    string
	node       *yaml.Node
	properties []schemaProperty
	oneOf      []schemaRef
	items      *schemaRef
	deprecated bool
	readOnly   bool
	xState     string
	// unsupported names a schema shape the walk cannot follow, empty when the shape is supported.
	unsupported string
}

// view resolves ref and merges allOf members into one property list. Shapes that cannot be
// walked faithfully (not, or allOf members carrying combinators) are flagged, never guessed.
func (d *oasDoc) view(ref schemaRef) (*schemaView, error) {
	resolved, err := d.resolve(ref)
	if err != nil {
		return nil, err
	}
	v := &schemaView{
		pointer:    resolved.pointer,
		node:       resolved.node,
		deprecated: boolValue(ref.node, "deprecated") || boolValue(resolved.node, "deprecated"),
		readOnly:   boolValue(ref.node, "readOnly") || boolValue(resolved.node, "readOnly"),
		xState:     firstNonEmpty(stringValue(ref.node, "x-state"), stringValue(resolved.node, "x-state")),
	}
	if mappingValue(resolved.node, "not") != nil {
		v.unsupported = "not"
		return v, nil
	}
	if err := d.addProperties(v, resolved); err != nil {
		return nil, err
	}
	addCombinators(v, resolved)
	if items := mappingValue(resolved.node, "items"); items != nil {
		v.items = &schemaRef{node: items, pointer: resolved.pointer + "/items"}
	}
	return v, nil
}

func (d *oasDoc) addProperties(v *schemaView, resolved schemaRef) error {
	required := make(map[string]bool)
	for _, name := range sortedStrings(orEmpty(mappingValue(resolved.node, "required"))) {
		required[name] = true
	}
	if props := mappingValue(resolved.node, "properties"); props != nil {
		for i := 0; i+1 < len(props.Content); i += 2 {
			name := props.Content[i].Value
			v.properties = append(v.properties, schemaProperty{
				name:     name,
				ref:      schemaRef{node: props.Content[i+1], pointer: resolved.pointer + "/properties/" + escapePointerToken(name)},
				required: required[name],
			})
		}
	}
	members := mappingValue(resolved.node, "allOf")
	if members == nil {
		return nil
	}
	for i, member := range members.Content {
		sub, err := d.view(schemaRef{node: member, pointer: fmt.Sprintf("%s/allOf/%d", resolved.pointer, i)})
		if err != nil {
			return err
		}
		if sub.unsupported != "" || len(sub.oneOf) > 0 || sub.items != nil {
			v.unsupported = "allOf member with " + firstNonEmpty(sub.unsupported, "combinator or items")
			return nil
		}
		v.properties = mergeProperties(v.properties, sub.properties)
	}
	return nil
}

func addCombinators(v *schemaView, resolved schemaRef) {
	for _, key := range []string{"oneOf", "anyOf"} {
		list := mappingValue(resolved.node, key)
		if list == nil {
			continue
		}
		if len(v.oneOf) > 0 {
			v.unsupported = "both oneOf and anyOf"
			return
		}
		for i, item := range list.Content {
			v.oneOf = append(v.oneOf, schemaRef{node: item, pointer: fmt.Sprintf("%s/%s/%d", resolved.pointer, key, i)})
		}
	}
}

func mergeProperties(existing, added []schemaProperty) []schemaProperty {
	for _, prop := range added {
		replaced := false
		for i := range existing {
			if existing[i].name == prop.name {
				existing[i] = prop
				replaced = true
				break
			}
		}
		if !replaced {
			existing = append(existing, prop)
		}
	}
	return existing
}

func orEmpty(node *yaml.Node) *yaml.Node {
	if node == nil {
		return &yaml.Node{}
	}
	return node
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
