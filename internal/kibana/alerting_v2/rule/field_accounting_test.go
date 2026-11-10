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

package rule

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

// waivedPutBodyPaths are PUT body fields deliberately left out of the schema.
// metadata.builder marks the Kibana UI form that authored a rule; Terraform
// manages ES|QL rules only (design D3), so it is never sent.
var waivedPutBodyPaths = []string{"metadata.builder"}

// serverOwnedAttributes are schema attributes with no counterpart in the PUT body.
var serverOwnedAttributes = map[string]bool{
	"id": true, "rule_id": true, "space_id": true, "version": true, "created_at": true, "updated_at": true,
}

var modelledUnionVariants = map[string][]string{
	"recovery": {"Condition", "Manual", "NoBreach", "Query"},
	"no_data":  {"Alert", "Ignore", "KeepLast", "Resolve"},
}

func TestPutBodyFieldAccounting_schemaModelsEveryWriteField(t *testing.T) {
	t.Parallel()

	walked := &walkedPutBody{paths: map[string]bool{}, variants: map[string][]string{}}
	walked.walkStruct(reflect.TypeFor[kbapi.KibanaHTTPAPIsAlertingPutRule](), "")

	modelled := map[string]bool{}
	collectSchemaJSONPaths(getSchema(context.Background()).Attributes, "", modelled)

	var missingFromSchema, missingFromAPI []string
	for p := range walked.paths {
		if !modelled[p] && !isWaived(p) {
			missingFromSchema = append(missingFromSchema, p)
		}
	}
	for p := range modelled {
		if !walked.paths[p] {
			missingFromAPI = append(missingFromAPI, p)
		}
	}
	sort.Strings(missingFromSchema)
	sort.Strings(missingFromAPI)
	require.Empty(t, missingFromSchema, "PUT body fields not modelled in the schema (model them or add them to waivedPutBodyPaths)")
	require.Empty(t, missingFromAPI, "schema attributes with no PUT body field")

	for union, variants := range walked.variants {
		sort.Strings(variants)
		require.Equalf(t, modelledUnionVariants[union], variants, "union %s variants differ from the modelled set", union)
	}
	require.Len(t, walked.variants, len(modelledUnionVariants), "modelled unions no longer present in the PUT body")
}

func isWaived(path string) bool {
	for _, waived := range waivedPutBodyPaths {
		if path == waived || strings.HasPrefix(path, waived+".") {
			return true
		}
	}
	return false
}

type walkedPutBody struct {
	paths    map[string]bool
	variants map[string][]string
}

func (w *walkedPutBody) walkStruct(structType reflect.Type, prefix string) {
	for field := range structType.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		w.walkField(field.Type, joinPath(prefix, name))
	}
}

func (w *walkedPutBody) walkField(fieldType reflect.Type, path string) {
	for fieldType.Kind() == reflect.Pointer || fieldType.Kind() == reflect.Slice {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() != reflect.Struct {
		w.paths[path] = true
		return
	}
	if isGeneratedUnion(fieldType) {
		w.walkUnion(fieldType, path)
		return
	}
	if fieldType.NumField() == 0 {
		w.paths[path] = true
		return
	}
	w.walkStruct(fieldType, path)
}

func isGeneratedUnion(structType reflect.Type) bool {
	if structType.NumField() != 1 {
		return false
	}
	field := structType.Field(0)
	return field.Name == "union" && field.Type == reflect.TypeFor[json.RawMessage]()
}

// walkUnion enumerates the As<Variant>() accessors the generator emits on a
// union type and walks every variant struct, so a new upstream strategy shows
// up both as a variant name and as any new field it carries.
func (w *walkedPutBody) walkUnion(unionType reflect.Type, path string) {
	accessorPrefix := "As" + unionType.Name()
	for method := range unionType.Methods() {
		if !strings.HasPrefix(method.Name, accessorPrefix) {
			continue
		}
		w.variants[path] = append(w.variants[path], strings.TrimPrefix(method.Name, accessorPrefix))
		w.walkStruct(method.Type.Out(0), path)
	}
}

func collectSchemaJSONPaths(attributes map[string]schema.Attribute, prefix string, into map[string]bool) {
	for name, attribute := range attributes {
		if prefix == "" && serverOwnedAttributes[name] {
			continue
		}
		path := joinPath(prefix, name)
		switch nested := attribute.(type) {
		case schema.SingleNestedAttribute:
			collectSchemaJSONPaths(nested.Attributes, path, into)
		case schema.ListNestedAttribute:
			collectSchemaJSONPaths(nested.NestedObject.Attributes, path, into)
		default:
			into[path] = true
		}
	}
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
