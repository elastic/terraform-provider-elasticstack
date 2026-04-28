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

package template

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestApplyAllowCustomRouting8xWorkaround_priorTrue_configOmitsAttr(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(true),
		"allow_custom_routing": types.BoolValue(true),
	})
	require.False(t, diags.HasError(), "%v", diags)
	configDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(false),
		"allow_custom_routing": types.BoolNull(),
	})
	require.False(t, diags.HasError(), "%v", diags)

	prior := Model{DataStream: priorDS}
	config := Model{DataStream: configDS}
	api := &models.IndexTemplate{}
	applyAllowCustomRouting8xWorkaround(ctx, prior, config, api)
	require.NotNil(t, api.DataStream)
	require.NotNil(t, api.DataStream.AllowCustomRouting)
	require.False(t, *api.DataStream.AllowCustomRouting)
}

func TestApplyAllowCustomRouting8xWorkaround_priorTrue_configExplicitFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(true),
		"allow_custom_routing": types.BoolValue(true),
	})
	require.False(t, diags.HasError(), "%v", diags)
	configDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(true),
		"allow_custom_routing": types.BoolValue(false),
	})
	require.False(t, diags.HasError(), "%v", diags)

	prior := Model{DataStream: priorDS}
	config := Model{DataStream: configDS}
	api := &models.IndexTemplate{}
	applyAllowCustomRouting8xWorkaround(ctx, prior, config, api)
	require.NotNil(t, api.DataStream)
	require.NotNil(t, api.DataStream.AllowCustomRouting)
	require.False(t, *api.DataStream.AllowCustomRouting)
}

func TestApplyAllowCustomRouting8xWorkaround_priorTrue_configOmitsAttr_planWouldMatchPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(true),
		"allow_custom_routing": types.BoolValue(true),
	})
	require.False(t, diags.HasError(), "%v", diags)
	configDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(true),
		"allow_custom_routing": types.BoolNull(),
	})
	require.False(t, diags.HasError(), "%v", diags)

	prior := Model{DataStream: priorDS}
	config := Model{DataStream: configDS}
	api := &models.IndexTemplate{}
	applyAllowCustomRouting8xWorkaround(ctx, prior, config, api)
	require.NotNil(t, api.DataStream)
	require.NotNil(t, api.DataStream.AllowCustomRouting)
	require.False(t, *api.DataStream.AllowCustomRouting)
}

func TestApplyAllowCustomRouting8xWorkaround_configExplicitTrue_noOverwrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(true),
		"allow_custom_routing": types.BoolValue(true),
	})
	require.False(t, diags.HasError(), "%v", diags)
	configDS, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(true),
		"allow_custom_routing": types.BoolValue(true),
	})
	require.False(t, diags.HasError(), "%v", diags)

	prior := Model{DataStream: priorDS}
	config := Model{DataStream: configDS}
	tTrue := true
	api := &models.IndexTemplate{DataStream: &models.DataStreamSettings{AllowCustomRouting: &tTrue}}
	applyAllowCustomRouting8xWorkaround(ctx, prior, config, api)
	require.NotNil(t, api.DataStream.AllowCustomRouting)
	require.True(t, *api.DataStream.AllowCustomRouting)
}
