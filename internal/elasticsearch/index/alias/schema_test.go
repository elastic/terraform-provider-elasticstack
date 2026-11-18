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

package alias

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSchema_readIndicesIncludesComputedConcreteIndices(t *testing.T) {
	t.Parallel()

	resourceSchema := getSchemaFactory(context.Background())
	readIndices := resourceSchema.Attributes["read_indices"].(schema.SetNestedAttribute)
	concreteIndices := readIndices.NestedObject.Attributes["concrete_indices"].(schema.SetAttribute)

	require.True(t, concreteIndices.Computed)
	require.False(t, concreteIndices.Optional)
}

func TestReadIndexAttrTypes_includeConcreteIndices(t *testing.T) {
	t.Parallel()

	attrType, ok := getReadIndexAttrTypes(context.Background())["concrete_indices"]

	require.True(t, ok)
	require.Equal(t, types.SetType{ElemType: types.StringType}, attrType)
}
