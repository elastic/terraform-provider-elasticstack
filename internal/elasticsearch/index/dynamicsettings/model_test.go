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

package dynamicsettings_test

import (
	"reflect"
	"testing"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/dynamicsettings"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelFieldsMatchDynamicSettingAttributes(t *testing.T) {
	attrs := indexparent.GetDynamicSettingAttributes()
	modelType := reflect.TypeFor[dynamicsettings.Model]()

	tagsByField := make(map[string]string, modelType.NumField())
	for field := range modelType.Fields() {
		tag := field.Tag.Get("tfsdk")
		require.NotEmptyf(t, tag, "field %s has no tfsdk tag", field.Name)
		tagsByField[tag] = field.Name

		attr, ok := attrs[tag]
		require.Truef(t, ok, "field %s tags unknown attribute %s", field.Name, tag)

		fieldType := field.Type
		switch attrType := attr.GetType().(type) {
		case basetypes.Int64Typable:
			require.Equal(t, types.Int64Type, attrType, tag)
			assert.Equal(t, reflect.TypeFor[types.Int64](), fieldType, tag)
		case basetypes.StringTypable:
			require.Equal(t, types.StringType, attrType, tag)
			assert.Equal(t, reflect.TypeFor[types.String](), fieldType, tag)
		case basetypes.BoolTypable:
			require.Equal(t, types.BoolType, attrType, tag)
			assert.Equal(t, reflect.TypeFor[types.Bool](), fieldType, tag)
		case basetypes.SetTypable:
			require.Equal(t, types.SetType{ElemType: types.StringType}, attrType, tag)
			assert.Equal(t, reflect.TypeFor[types.Set](), fieldType, tag)
		default:
			t.Fatalf("unexpected attribute type %T for %s", attrType, tag)
		}
	}

	assert.Len(t, tagsByField, len(attrs))
}
