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
	"go/types"
	"reflect"
	"strings"
)

// jsonField is a struct field addressed by its JSON name, with the struct type that declares it.
type jsonField struct {
	field *types.Var
	owner types.Type
}

// jsonFields maps JSON property names to the Go fields of a struct type, flattening embedded
// structs the way encoding/json does. Fields hidden from JSON ("-") and unexported fields are skipped.
func jsonFields(t types.Type) map[string]jsonField {
	out := make(map[string]jsonField)
	depths := make(map[string]int)
	collectJSONFields(t, out, depths, 0)
	return out
}

// collectJSONFields follows encoding/json precedence: when embedded structs expose the same
// JSON name, the shallowest field wins.
func collectJSONFields(t types.Type, out map[string]jsonField, depths map[string]int, depth int) {
	t = derefType(t)
	st, ok := t.Underlying().(*types.Struct)
	if !ok {
		return
	}
	for i := range st.NumFields() {
		field := st.Field(i)
		name, hidden := jsonName(st.Tag(i))
		if hidden {
			continue
		}
		if field.Anonymous() && name == "" {
			collectJSONFields(field.Type(), out, depths, depth+1)
			continue
		}
		if !field.Exported() {
			continue
		}
		if name == "" {
			name = field.Name()
		}
		if existing, seen := depths[name]; seen && existing <= depth {
			continue
		}
		depths[name] = depth
		out[name] = jsonField{field: field, owner: t}
	}
}

// jsonName returns the JSON (or, for query parameter structs, form) name from a struct tag.
func jsonName(tag string) (name string, hidden bool) {
	structTag := reflect.StructTag(tag)
	for _, key := range []string{"json", "form"} {
		value, ok := structTag.Lookup(key)
		if !ok {
			continue
		}
		name, _, _ = strings.Cut(value, ",")
		if name == "-" && !strings.Contains(value, ",") {
			return "", true
		}
		if name != "" {
			return name, false
		}
	}
	return "", false
}

func derefType(t types.Type) types.Type {
	t = types.Unalias(t)
	for {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			return t
		}
		t = types.Unalias(ptr.Elem())
	}
}
