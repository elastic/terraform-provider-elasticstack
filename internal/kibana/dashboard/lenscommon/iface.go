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

// Package lenscommon holds shared Lens visualization converter contracts and helpers.
package lenscommon

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Resolver abstracts dashboard-level dependencies so Lens converters do not import the dashboard package.
type Resolver interface {
	ResolveChartTimeRange(chartLevel *models.TimeRangeModel) kbapi.KbnEsQueryServerTimeRangeSchema
	// DashboardLensComparableTimeRange returns the dashboard-level time range used when comparing
	// chart-root API time_range for Terraform null-preservation. ok is false when no comparable range exists.
	DashboardLensComparableTimeRange() (kbapi.KbnEsQueryServerTimeRangeSchema, bool)
}

// VizConverter converts one Lens chart kind between Terraform models and kbapi vis config.
type VizConverter interface {
	VizType() string
	HandlesBlocks(blocks *models.LensByValueChartBlocks) bool

	// SchemaAttribute returns the SingleNestedAttribute for this chart kind inside vis_config.by_value
	// or lens_dashboard_app_config.by_value.
	SchemaAttribute() schema.Attribute

	// PopulateFromAttributes reads the typed API chart payload from attrs and writes the result into the
	// matching blocks.<Chart>Config field. The caller MUST seed blocks.<Chart>Config from prior state
	// (for example existing Terraform state) before invoking the converter so prior-dependent merge logic
	// (drilldowns, presentation, JSON defaults preservation) sees the pre-existing values. After this
	// method returns successfully, blocks.<Chart>Config points to the freshly populated model.
	// Implementations may copy blocks.<Chart>Config into a local prior variable before reconstruction.
	PopulateFromAttributes(ctx context.Context, resolver Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics
	BuildAttributes(blocks *models.LensByValueChartBlocks, resolver Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics)
	AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks)
	PopulateJSONDefaults(attrs map[string]any) map[string]any
}
