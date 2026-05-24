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

package entitycore

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const ephemeralUserStateKey = "entitycore.ephemeral.user_state"

const pluginFrameworkPkgPrefix = "github.com/hashicorp/terraform-plugin-framework"

func encodeUserCloseState[S any](s S) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	data, err := json.Marshal(s)
	if err != nil {
		diags.AddError("Failed to marshal ephemeral close state", err.Error())
		return nil, diags
	}
	return data, diags
}

func decodeUserCloseState[S any](data []byte) (S, diag.Diagnostics) {
	var diags diag.Diagnostics
	var state S
	if err := json.Unmarshal(data, &state); err != nil {
		diags.AddError("Failed to parse ephemeral close state", err.Error())
		return state, diags
	}
	return state, diags
}

func mustBePlainGoCloseState[S any]() {
	var zero S
	rootName := rootCloseStateTypeName(reflect.TypeOf(zero))
	walkCloseStateType(reflect.TypeOf(zero), rootName, "", map[reflect.Type]bool{})
}

func walkCloseStateType(t reflect.Type, rootName, path string, visited map[reflect.Type]bool) {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return
	}
	if visited[t] {
		return
	}
	visited[t] = true

	if isForbiddenPluginFrameworkType(t) {
		fieldPath := path
		if fieldPath == "" {
			fieldPath = t.Name()
		}
		panic(fmt.Sprintf(
			"entitycore: ephemeral close state %s has field %s of plugin-framework type %s/%s; Close state must be plain Go types only",
			rootName,
			fieldPath,
			t.PkgPath(),
			t.Name(),
		))
	}

	switch t.Kind() {
	case reflect.Struct:
		for sf := range t.Fields() {
			fieldPath := sf.Name
			if path != "" {
				fieldPath = path + "." + sf.Name
			}
			walkCloseStateType(sf.Type, rootName, fieldPath, visited)
		}
	case reflect.Slice, reflect.Array:
		walkCloseStateType(t.Elem(), rootName, appendPath(path, "[]"), visited)
	case reflect.Map:
		walkCloseStateType(t.Key(), rootName, appendPath(path, "<key>"), visited)
		walkCloseStateType(t.Elem(), rootName, appendPath(path, "<value>"), visited)
	}
}

func rootCloseStateTypeName(t reflect.Type) string {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return "unknown"
	}
	if name := t.Name(); name != "" {
		return name
	}
	return t.String()
}

func appendPath(prefix, suffix string) string {
	if prefix == "" {
		return strings.TrimPrefix(suffix, ".")
	}
	return prefix + suffix
}

func isForbiddenPluginFrameworkType(t reflect.Type) bool {
	pkgPath := t.PkgPath()
	return strings.HasPrefix(pkgPath, pluginFrameworkPkgPrefix)
}
