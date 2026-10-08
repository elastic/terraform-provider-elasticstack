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
	"context"
	"testing"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The index resource schema sources its dynamic-setting attributes from the
// shared indexparent.GetDynamicSettingAttributes() map, so the two schemas
// cannot drift apart (REQ-UNCHANGED in the elasticsearch-index-settings change).
func Test_getSchema_dynamicSettingAttributesMatchShared(t *testing.T) {
	t.Parallel()

	schemaAttributes := getSchema(context.Background()).Attributes
	sharedAttributes := indexparent.GetDynamicSettingAttributes()

	require.NotEmpty(t, sharedAttributes)
	for name, sharedAttr := range sharedAttributes {
		require.Contains(t, schemaAttributes, name)
		require.True(t, schemaAttributes[name].IsOptional(), name)
		require.False(t, schemaAttributes[name].IsRequired(), name)
		assert.Equal(t, sharedAttr, schemaAttributes[name], name)
	}
}
