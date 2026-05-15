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

// Package iface defines the Terraform dashboard panel Handler contract used by registry-driven routing.
package iface

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Handler binds a dashboard panel discriminator to schema, conversions, validation, and pinning.
type Handler interface {
	PanelType() string
	SchemaAttribute() schema.Attribute
	FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics
	ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics)
	ValidatePanelConfig(ctx context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics
	AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel)
	ClassifyJSON(config map[string]any) bool
	PopulateJSONDefaults(config map[string]any) map[string]any
	PinnedHandler() PinnedHandler
}

// PinnedHandler converts control panels pinned in the dashboard control bar (optional).
type PinnedHandler interface {
	FromAPI(ctx context.Context, prior *models.PinnedPanelModel, raw kbapi.DashboardPinnedPanels_Item) (models.PinnedPanelModel, diag.Diagnostics)
	ToAPI(ppm models.PinnedPanelModel) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics)
}
