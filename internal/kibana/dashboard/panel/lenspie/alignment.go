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

package lenspie

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func alignPieConfigStateFromPlan(ctx context.Context, plan, state *models.PieChartConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DataSourceJSON, &state.DataSourceJSON, "time_field", "name")
	// Kibana materializes label_position="outside" when the practitioner omits it.
	lenscommon.PreserveNullIfStateEquals(plan.LabelPosition, &state.LabelPosition, types.StringValue("outside"))
	// Pie group_by/metrics config_json are re-emitted with default keys (color,
	// rank_by, limit) added by Kibana. PreservePlanJSONWithDefaults handles the
	// JSONWithDefaults type via semantic-equality.
	var diags diag.Diagnostics
	m := min(len(plan.Metrics), len(state.Metrics))
	for i := range m {
		state.Metrics[i].Config = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, plan.Metrics[i].Config, state.Metrics[i].Config, &diags)
	}
	g := min(len(plan.GroupBy), len(state.GroupBy))
	for i := range g {
		state.GroupBy[i].Config = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, plan.GroupBy[i].Config, state.GroupBy[i].Config, &diags)
	}
}
