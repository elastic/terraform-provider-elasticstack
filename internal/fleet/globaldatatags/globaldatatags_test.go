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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchema_attributesAndValidators(t *testing.T) {
	t.Parallel()

	attribute := Schema(nil)

	assert.True(t, attribute.Optional)
	assert.False(t, attribute.Computed)
	assert.Nil(t, attribute.Default)
	assert.Equal(t, descriptionMarkdown, attribute.MarkdownDescription)

	stringAttr, ok := attribute.NestedObject.Attributes[StringValueAttr].(frameworkschema.StringAttribute)
	require.True(t, ok)
	assert.True(t, stringAttr.Optional)
	assert.Equal(t, stringValueDescription, stringAttr.MarkdownDescription)
	assert.Len(t, stringAttr.Validators, 2)

	numberAttr, ok := attribute.NestedObject.Attributes[NumberValueAttr].(frameworkschema.Float32Attribute)
	require.True(t, ok)
	assert.True(t, numberAttr.Optional)
	assert.Equal(t, numberValueDescription, numberAttr.MarkdownDescription)
	assert.Len(t, numberAttr.Validators, 2)
}

func TestSchema_withDefault(t *testing.T) {
	t.Parallel()

	attribute := Schema(map[string]attr.Value{})

	assert.True(t, attribute.Optional)
	assert.True(t, attribute.Computed)
	assert.NotNil(t, attribute.Default)
}
