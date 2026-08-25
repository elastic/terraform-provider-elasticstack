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

package globaldatatags

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNestedObject_attributesAndValidators(t *testing.T) {
	t.Parallel()

	obj := NestedObject("a string description", "a number description", false)

	stringAttr, ok := obj.Attributes[StringValueAttr].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, stringAttr.Optional)
	assert.Equal(t, "a string description", stringAttr.Description)
	assert.Empty(t, stringAttr.MarkdownDescription)
	assert.Len(t, stringAttr.Validators, 2)

	numberAttr, ok := obj.Attributes[NumberValueAttr].(schema.Float32Attribute)
	require.True(t, ok)
	assert.True(t, numberAttr.Optional)
	assert.Equal(t, "a number description", numberAttr.Description)
	assert.Empty(t, numberAttr.MarkdownDescription)
	assert.Len(t, numberAttr.Validators, 2)
}

func TestNestedObject_markdownDescription(t *testing.T) {
	t.Parallel()

	obj := NestedObject("a string description", "a number description", true)

	stringAttr, ok := obj.Attributes[StringValueAttr].(schema.StringAttribute)
	require.True(t, ok)
	assert.Empty(t, stringAttr.Description)
	assert.Equal(t, "a string description", stringAttr.MarkdownDescription)

	numberAttr, ok := obj.Attributes[NumberValueAttr].(schema.Float32Attribute)
	require.True(t, ok)
	assert.Empty(t, numberAttr.Description)
	assert.Equal(t, "a number description", numberAttr.MarkdownDescription)
}
