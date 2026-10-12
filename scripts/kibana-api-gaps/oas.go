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
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const jsonContentType = "application/json"

// schemaRef is a YAML node together with its JSON pointer inside the OpenAPI document.
type schemaRef struct {
	node    *yaml.Node
	pointer string
}

type oasParam struct {
	name       string
	deprecated bool
	ref        schemaRef
}

type oasOperation struct {
	Method     string
	Path       string
	Deprecated bool
	XState     string
	Body       *schemaRef
	Query      []oasParam
}

type oasDoc struct {
	root       *yaml.Node
	operations map[string]*oasOperation
}

func loadOAS(path string) (*oasDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(document.Content) == 0 {
		return nil, fmt.Errorf("parse %s: empty document", path)
	}
	doc := &oasDoc{root: document.Content[0], operations: make(map[string]*oasOperation)}
	if err := doc.indexOperations(); err != nil {
		return nil, err
	}
	return doc, nil
}

func (d *oasDoc) operation(method, pathTemplate string) (*oasOperation, bool) {
	op, ok := d.operations[operationKey(method, pathTemplate)]
	return op, ok
}

func operationKey(method, pathTemplate string) string {
	return strings.ToUpper(method) + " " + normalizePathTemplate(pathTemplate)
}

// normalizePathTemplate makes OpenAPI ("{id}") and Go format ("%s") path templates comparable.
func normalizePathTemplate(path string) string {
	var out strings.Builder
	for i := 0; i < len(path); i++ {
		switch {
		case path[i] == '{':
			end := strings.IndexByte(path[i:], '}')
			if end < 0 {
				out.WriteString(path[i:])
				return out.String()
			}
			out.WriteString("{}")
			i += end
		case path[i] == '%' && i+1 < len(path) && path[i+1] == 's':
			out.WriteString("{}")
			i++
		default:
			out.WriteByte(path[i])
		}
	}
	return out.String()
}

var httpMethods = []string{"get", "put", "post", "delete", "patch", "head", "options"}

func (d *oasDoc) indexOperations() error {
	paths := mappingValue(d.root, "paths")
	if paths == nil {
		return fmt.Errorf("openapi document has no paths")
	}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		template := paths.Content[i].Value
		pathItem := paths.Content[i+1]
		pathPointer := "/paths/" + escapePointerToken(template)
		for _, method := range httpMethods {
			opNode := mappingValue(pathItem, method)
			if opNode == nil {
				continue
			}
			op, err := d.buildOperation(strings.ToUpper(method), template, pathPointer, pathItem, opNode)
			if err != nil {
				return err
			}
			d.operations[operationKey(method, template)] = op
		}
	}
	return nil
}

func (d *oasDoc) buildOperation(method, template, pathPointer string, pathItem, opNode *yaml.Node) (*oasOperation, error) {
	opPointer := pathPointer + "/" + strings.ToLower(method)
	op := &oasOperation{
		Method:     method,
		Path:       template,
		Deprecated: boolValue(opNode, "deprecated"),
		XState:     stringValue(opNode, "x-state"),
	}
	for _, source := range []struct {
		node    *yaml.Node
		pointer string
	}{
		{mappingValue(pathItem, "parameters"), pathPointer + "/parameters"},
		{mappingValue(opNode, "parameters"), opPointer + "/parameters"},
	} {
		if source.node == nil {
			continue
		}
		for i, paramNode := range source.node.Content {
			param, err := d.resolve(schemaRef{node: paramNode, pointer: fmt.Sprintf("%s/%d", source.pointer, i)})
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", method, template, err)
			}
			if stringValue(param.node, "in") != "query" {
				continue
			}
			op.Query = append(op.Query, oasParam{
				name:       stringValue(param.node, "name"),
				deprecated: boolValue(param.node, "deprecated"),
				ref:        param,
			})
		}
	}
	body, err := d.requestBodySchema(opNode, opPointer+"/requestBody")
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, template, err)
	}
	op.Body = body
	return op, nil
}

func (d *oasDoc) requestBodySchema(opNode *yaml.Node, pointer string) (*schemaRef, error) {
	bodyNode := mappingValue(opNode, "requestBody")
	if bodyNode == nil {
		return nil, nil
	}
	body, err := d.resolve(schemaRef{node: bodyNode, pointer: pointer})
	if err != nil {
		return nil, err
	}
	schema := mappingValue(mappingValue(mappingValue(body.node, "content"), jsonContentType), "schema")
	if schema == nil {
		return nil, nil
	}
	resolved, err := d.resolve(schemaRef{
		node:    schema,
		pointer: body.pointer + "/content/" + escapePointerToken(jsonContentType) + "/schema",
	})
	if err != nil {
		return nil, err
	}
	return &resolved, nil
}

// resolve follows $ref chains and returns the referenced node with its own pointer.
func (d *oasDoc) resolve(ref schemaRef) (schemaRef, error) {
	for range 64 {
		target := stringValue(ref.node, "$ref")
		if target == "" {
			return ref, nil
		}
		pointer, ok := strings.CutPrefix(target, "#")
		if !ok {
			return schemaRef{}, fmt.Errorf("unsupported non-local $ref %q", target)
		}
		node, err := d.lookup(pointer)
		if err != nil {
			return schemaRef{}, err
		}
		ref = schemaRef{node: node, pointer: pointer}
	}
	return schemaRef{}, fmt.Errorf("$ref chain too deep at %s", ref.pointer)
}

func (d *oasDoc) lookup(pointer string) (*yaml.Node, error) {
	node := d.root
	for token := range strings.SplitSeq(strings.TrimPrefix(pointer, "/"), "/") {
		token = unescapePointerToken(token)
		if node.Kind == yaml.SequenceNode {
			var index int
			if _, err := fmt.Sscanf(token, "%d", &index); err != nil || index < 0 || index >= len(node.Content) {
				return nil, fmt.Errorf("pointer %q: bad index %q", pointer, token)
			}
			node = node.Content[index]
			continue
		}
		next := mappingValue(node, token)
		if next == nil {
			return nil, fmt.Errorf("pointer %q: %q not found", pointer, token)
		}
		node = next
	}
	return node, nil
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func stringValue(node *yaml.Node, key string) string {
	value := mappingValue(node, key)
	if value == nil || value.Kind != yaml.ScalarNode {
		return ""
	}
	return value.Value
}

func boolValue(node *yaml.Node, key string) bool {
	return stringValue(node, key) == "true"
}

func escapePointerToken(token string) string {
	return strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
}

func unescapePointerToken(token string) string {
	return strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
}
