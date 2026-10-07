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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reflectInner struct {
	ExampleKey types.String `tfsdk:"example_key"`
}

type reflectOuter struct {
	DirectKey types.String `tfsdk:"direct_key"`
	reflectInner
}

type reflectFlat struct {
	FlatKey types.String `tfsdk:"flat_key"`
}

func TestGetFieldValueByTagValue_returnsFlatField(t *testing.T) {
	v := reflectFlat{FlatKey: types.StringValue("flat-value")}

	got, found := GetFieldValueByTagValue(reflect.ValueOf(v), reflect.TypeFor[reflectFlat](), "flat_key")

	require.True(t, found)
	assert.Equal(t, types.StringValue("flat-value"), got)
}

func TestGetFieldValueByTagValue_returnsEmbeddedStructField(t *testing.T) {
	v := reflectOuter{
		DirectKey: types.StringValue("direct-value"),
		//nolint:modernize // embedlit shorthand is rejected by the compiler
		reflectInner: reflectInner{ExampleKey: types.StringValue("embedded-value")},
	}

	got, found := GetFieldValueByTagValue(reflect.ValueOf(v), reflect.TypeFor[reflectOuter](), "example_key")

	require.True(t, found)
	assert.Equal(t, types.StringValue("embedded-value"), got)
}

func TestSetFieldValueByTagValue_setsEmbeddedStructField(t *testing.T) {
	v := reflectOuter{}
	ptr := reflect.ValueOf(&v)

	set := SetFieldValueByTagValue(ptr, reflect.TypeFor[reflectOuter](), "example_key", types.StringValue("new-value"))

	require.True(t, set)
	assert.Equal(t, types.StringValue("new-value"), v.ExampleKey)
}

func TestSetFieldValueByTagValue_setsFlatField(t *testing.T) {
	v := reflectFlat{}
	ptr := reflect.ValueOf(&v)

	set := SetFieldValueByTagValue(ptr, reflect.TypeFor[reflectFlat](), "flat_key", types.StringValue("new-value"))

	require.True(t, set)
	assert.Equal(t, types.StringValue("new-value"), v.FlatKey)
}

func TestSetFieldValueByTagValue_skipsNonAssignableValue(t *testing.T) {
	v := reflectFlat{FlatKey: types.StringValue("original")}
	ptr := reflect.ValueOf(&v)

	set := SetFieldValueByTagValue(ptr, reflect.TypeFor[reflectFlat](), "flat_key", types.Int64Value(7))

	require.False(t, set)
	assert.Equal(t, types.StringValue("original"), v.FlatKey)
}
