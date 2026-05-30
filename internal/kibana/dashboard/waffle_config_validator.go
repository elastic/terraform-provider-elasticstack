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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenswaffle"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = waffleConfigModeValidator{}

// waffleConfigModeValidator enforces consistency between non-ES|QL and ES|QL waffle fields,
// matching heatmap-style ES|QL detection (omit `query` or leave both `query.expression` and `query.language` unset).
type waffleConfigModeValidator struct{}

func (waffleConfigModeValidator) Description(_ context.Context) string {
	return "Ensures waffle_config uses `metrics`/`group_by` for non-ES|QL mode and `esql_metrics`/`esql_group_by` for ES|QL mode."
}

func (v waffleConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func waffleModeListStateFromTF(list types.List) lenswaffle.WaffleModeListState {
	if list.IsUnknown() {
		return lenswaffle.WaffleModeListState{Unknown: true}
	}
	if list.IsNull() {
		return lenswaffle.WaffleModeListState{Count: 0}
	}
	return lenswaffle.WaffleModeListState{Count: len(list.Elements())}
}

func (v waffleConfigModeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var queryObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query"), &queryObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	esqlMode := queryObj.IsNull()
	if !esqlMode {
		var lang, qStr types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query").AtName("language"), &lang)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query").AtName("expression"), &qStr)...)
		if resp.Diagnostics.HasError() {
			return
		}
		esqlMode = lang.IsNull() && qStr.IsNull()
	}

	var metrics, groupBy, esqlMetrics, esqlGroupBy types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("metrics"), &metrics)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("group_by"), &groupBy)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("esql_metrics"), &esqlMetrics)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("esql_group_by"), &esqlGroupBy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(lenswaffle.WaffleConfigModeValidateDiags(esqlMode,
		waffleModeListStateFromTF(metrics),
		waffleModeListStateFromTF(groupBy),
		waffleModeListStateFromTF(esqlMetrics),
		waffleModeListStateFromTF(esqlGroupBy),
	)...)
}
