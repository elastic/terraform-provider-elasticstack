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

package index

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// GetFieldValueByTagValue returns the attr.Value of the tfsdk-tagged field with
// the given tagName from the struct value v, along with whether the field was found.
// v must be a struct (not a pointer).
func GetFieldValueByTagValue(v reflect.Value, t reflect.Type, tagName string) (attr.Value, bool) {
	numField := t.NumField()
	for i := range numField {
		field := t.Field(i)
		if field.Tag.Get("tfsdk") == tagName {
			return v.Field(i).Interface().(attr.Value), true
		}
	}
	return nil, false
}

// SetFieldValueByTagValue sets the tfsdk-tagged field with the given tagName on
// the struct pointed to by ptr to value. Returns true if the field was found and set.
// ptr must be a pointer to a struct.
func SetFieldValueByTagValue(ptr reflect.Value, t reflect.Type, tagName string, value attr.Value) bool {
	numField := t.NumField()
	for i := range numField {
		field := t.Field(i)
		if field.Tag.Get("tfsdk") == tagName {
			ptr.Elem().Field(i).Set(reflect.ValueOf(value))
			return true
		}
	}
	return false
}
