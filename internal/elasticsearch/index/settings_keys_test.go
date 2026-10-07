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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDynamicSettingAttributes_coversAllDynamicSettingsKeys(t *testing.T) {
	attrs := GetDynamicSettingAttributes()

	require.Len(t, attrs, len(DynamicSettingsKeys))
	for _, key := range DynamicSettingsKeys {
		attr, ok := attrs[typeutils.ConvertSettingsKeyToTFFieldKey(key)]
		require.True(t, ok, "missing attribute for %s", key)
		assert.True(t, attr.IsOptional(), key)
		assert.False(t, attr.IsRequired(), key)
		assert.NotEmpty(t, attr.GetDescription(), key)
	}
}

func TestGetDynamicSettingAttributes_matchesIndexResourceAttributeShapes(t *testing.T) {
	attrs := GetDynamicSettingAttributes()

	assert.Equal(t, types.Int64Type, attrs["number_of_replicas"].GetType())
	assert.Equal(t, "Number of shard replicas.", attrs["number_of_replicas"].GetDescription())

	assert.Equal(t, types.StringType, attrs["refresh_interval"].GetType())
	assert.Equal(t, "How often to perform a refresh operation, which makes recent changes to the index visible to search. Can be set to `-1` to disable refresh.",
		attrs["refresh_interval"].GetDescription())

	assert.Equal(t, types.BoolType, attrs["blocks_read_only"].GetType())
	assert.Equal(t, "Set to `true` to make the index and index metadata read only, `false` to allow writes and metadata changes.", attrs["blocks_read_only"].GetDescription())

	assert.Equal(t, types.SetType{ElemType: types.StringType}, attrs["query_default_field"].GetType())
}
