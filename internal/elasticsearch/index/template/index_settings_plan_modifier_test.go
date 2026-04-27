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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestIndexSettingsCanonicalPlanModifier_jsonNullConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := indexSettingsCanonicalModifier{}
	resp := planmodifier.StringResponse{}
	mod.PlanModifyString(ctx, planmodifier.StringRequest{
		ConfigValue: basetypes.NewStringValue("null"),
		Path:        path.Root("template").AtName("settings"),
	}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, resp.PlanValue.IsNull())
}

func TestIndexSettingsCanonicalPlanModifier_jsonNullWithWhitespace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := indexSettingsCanonicalModifier{}
	resp := planmodifier.StringResponse{}
	mod.PlanModifyString(ctx, planmodifier.StringRequest{
		ConfigValue: basetypes.NewStringValue("  null\n"),
		Path:        path.Root("template").AtName("settings"),
	}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, resp.PlanValue.IsNull())
}
