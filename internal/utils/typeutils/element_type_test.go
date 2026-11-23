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

package typeutils_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestElementTypeOf(t *testing.T) {
	t.Parallel()

	t.Run("list type returns its element type", func(t *testing.T) {
		t.Parallel()
		got := typeutils.ElementTypeOf(types.ListType{ElemType: types.StringType})
		require.Equal(t, types.StringType, got)
	})

	t.Run("set type returns its element type", func(t *testing.T) {
		t.Parallel()
		got := typeutils.ElementTypeOf(types.SetType{ElemType: types.Int64Type})
		require.Equal(t, types.Int64Type, got)
	})

	t.Run("nested list of objects returns the object type", func(t *testing.T) {
		t.Parallel()
		objType := types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType}}
		got := typeutils.ElementTypeOf(types.ListType{ElemType: objType})
		require.Equal(t, objType, got)
	})

	t.Run("type without an element type panics", func(t *testing.T) {
		t.Parallel()
		require.Panics(t, func() {
			typeutils.ElementTypeOf(types.StringType)
		})
	})
}
