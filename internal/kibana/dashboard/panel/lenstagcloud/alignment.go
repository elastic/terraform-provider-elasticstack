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

package lenstagcloud

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

func alignTagcloudStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	alignTagcloudConfigStateFromPlan(ctx, plan.TagcloudConfig, state.TagcloudConfig)
}

func alignTagcloudConfigStateFromPlan(ctx context.Context, plan, state *models.TagcloudConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.TagByJSON.Normalized, &state.TagByJSON.Normalized, "rank_by", "color")
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.TagByJSON, &state.TagByJSON)
	if plan.EsqlMetric != nil && state.EsqlMetric != nil {
		lenscommon.PreserveNormalizedJSONSemanticEquality(plan.EsqlMetric.FormatJSON, &state.EsqlMetric.FormatJSON)
	}
	if plan.EsqlTagBy != nil && state.EsqlTagBy != nil {
		lenscommon.PreserveNormalizedJSONSemanticEquality(plan.EsqlTagBy.FormatJSON, &state.EsqlTagBy.FormatJSON)
		lenscommon.PreserveNormalizedJSONSemanticEquality(plan.EsqlTagBy.ColorJSON, &state.EsqlTagBy.ColorJSON)
	}
}
