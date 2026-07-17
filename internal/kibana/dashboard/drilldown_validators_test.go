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

package dashboard

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_structuredDrilldown_urlTriggerStringValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	listAttr := panelkit.StructuredDrilldownsAttribute().(schema.ListNestedAttribute)
	urlAttr, ok := listAttr.NestedObject.Attributes["url"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	triggerAttr, ok := urlAttr.Attributes["trigger"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, triggerAttr.Required)
	require.NotEmpty(t, triggerAttr.Validators)

	t.Run("rejects_invalid", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{Path: path.Root("trigger"), ConfigValue: types.StringValue("nope")}
		var resp validator.StringResponse
		for _, m := range triggerAttr.Validators {
			m.ValidateString(ctx, req, &resp)
		}
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("allows_known", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{Path: path.Root("trigger"), ConfigValue: types.StringValue("on_click_row")}
		var resp validator.StringResponse
		for _, m := range triggerAttr.Validators {
			m.ValidateString(ctx, req, &resp)
		}
		require.False(t, resp.Diagnostics.HasError())
	})
}
