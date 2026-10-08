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

package security_entity_store

import (
	"context"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ basetypes.StringValuableWithSemanticEquals = tfModel{}.StatusJSON
	_ basetypes.StringValuableWithSemanticEquals = dsModel{}.StatusJSON
)

func TestStatusJSONCustomTypeWiring(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	resourceAttr, ok := getSchema(ctx).Attributes["status_json"]
	require.True(t, ok, "resource schema must define status_json")
	resourceStringAttr, ok := resourceAttr.(rschema.StringAttribute)
	require.True(t, ok, "resource status_json must be a StringAttribute")
	assert.Equal(t, StatusJSONType{}, resourceStringAttr.CustomType)

	dsAttr, ok := getDataSourceSchema(ctx).Attributes["status_json"]
	require.True(t, ok, "data source schema must define status_json")
	dsStringAttr, ok := dsAttr.(dsschema.StringAttribute)
	require.True(t, ok, "data source status_json must be a StringAttribute")
	assert.Equal(t, StatusJSONType{}, dsStringAttr.CustomType)
}

func TestStatusJSONEngineOrderIgnoredAtResourceBoundary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userFirst := `{"status":"running","engines":[{"type":"user","indexPattern":".entities-user-v1"},{"type":"generic","indexPattern":".entities-generic-v1"}]}`
	genericFirst := `{"status":"running","engines":[{"type":"generic","indexPattern":".entities-generic-v1"},{"type":"user","indexPattern":".entities-user-v1"}]}`

	model := tfModel{}
	model.StatusJSON = NewStatusJSONValue(userFirst)
	other := tfModel{}
	other.StatusJSON = NewStatusJSONValue(genericFirst)

	eq, diags := model.StatusJSON.StringSemanticEquals(ctx, other.StatusJSON)
	require.False(t, diags.HasError(), "%v", diags)
	assert.True(t, eq)

	assert.Equal(t, userFirst, model.StatusJSON.ValueString())
	assert.Equal(t, genericFirst, other.StatusJSON.ValueString())
}
