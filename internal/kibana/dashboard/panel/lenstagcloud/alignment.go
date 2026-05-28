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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

// Kibana's tagcloud renderer materializes a fixed font size range when the
// practitioner omits font_size. Preserving the null plan in that case avoids
// "Provider produced inconsistent result after apply" diagnostics.
const (
	tagcloudDefaultFontSizeMin float64 = 18
	tagcloudDefaultFontSizeMax float64 = 72
)

func fontSizeMatchesKibanaDefault(m *models.FontSizeModel) bool {
	if m == nil {
		return false
	}
	return typeutils.IsKnown(m.Min) && m.Min.ValueFloat64() == tagcloudDefaultFontSizeMin &&
		typeutils.IsKnown(m.Max) && m.Max.ValueFloat64() == tagcloudDefaultFontSizeMax
}

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
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DataSourceJSON, &state.DataSourceJSON, "time_field")
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.TagByJSON.Normalized, &state.TagByJSON.Normalized, "rank_by", "color")
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.TagByJSON, &state.TagByJSON)
	// Kibana materializes server-side defaults when the practitioner omits these fields.
	lenscommon.PreserveNullStringIfStateEquals(plan.Orientation, &state.Orientation, "horizontal")
	if plan.FontSize == nil && state.FontSize != nil && fontSizeMatchesKibanaDefault(state.FontSize) {
		state.FontSize = nil
	}
	if plan.EsqlMetric != nil && state.EsqlMetric != nil {
		lenscommon.PreserveNormalizedJSONSemanticEquality(plan.EsqlMetric.FormatJSON, &state.EsqlMetric.FormatJSON)
	}
	if plan.EsqlTagBy != nil && state.EsqlTagBy != nil {
		lenscommon.PreserveNormalizedJSONSemanticEquality(plan.EsqlTagBy.FormatJSON, &state.EsqlTagBy.FormatJSON)
		lenscommon.PreserveNormalizedJSONSemanticEquality(plan.EsqlTagBy.ColorJSON, &state.EsqlTagBy.ColorJSON)
	}
}
