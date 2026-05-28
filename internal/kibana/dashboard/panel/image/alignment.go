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

package image

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func alignImageStateFromPlan(plan, state *models.PanelModel) {
	if plan == nil || state == nil {
		return
	}
	if plan.ImageConfig == nil {
		return
	}
	if state.ImageConfig == nil {
		state.ImageConfig = cloneImagePanelConfigModel(plan.ImageConfig)
		return
	}
	alignImagePanelConfigFromPlan(plan.ImageConfig, state.ImageConfig)
}

func alignImagePanelConfigFromPlan(plan, state *models.ImagePanelConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveKnownStringIfStateNull(plan.AltText, &state.AltText)
	preserveKnownStringIfStateNull(plan.BackgroundColor, &state.BackgroundColor)
	preserveKnownStringIfStateNull(plan.Title, &state.Title)
	preserveKnownStringIfStateNull(plan.Description, &state.Description)
	preserveKnownBoolIfStateNull(plan.HideTitle, &state.HideTitle)
	preserveKnownBoolIfStateNull(plan.HideBorder, &state.HideBorder)
	preserveKnownStringIfStateNull(plan.ObjectFit, &state.ObjectFit)

	if plan.Src.URL != nil {
		if state.Src.URL == nil {
			url := *plan.Src.URL
			state.Src.URL = &url
		} else {
			preserveKnownStringIfStateNull(plan.Src.URL.URL, &state.Src.URL.URL)
		}
	}
	if plan.Src.File != nil {
		if state.Src.File == nil {
			file := *plan.Src.File
			state.Src.File = &file
		} else {
			preserveKnownStringIfStateNull(plan.Src.File.FileID, &state.Src.File.FileID)
		}
	}
}

func preserveKnownStringIfStateNull(plan types.String, state *types.String) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveKnownBoolIfStateNull(plan types.Bool, state *types.Bool) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func cloneImagePanelConfigModel(model *models.ImagePanelConfigModel) *models.ImagePanelConfigModel {
	if model == nil {
		return nil
	}
	cloned := *model
	if model.Src.URL != nil {
		url := *model.Src.URL
		cloned.Src.URL = &url
	}
	if model.Src.File != nil {
		file := *model.Src.File
		cloned.Src.File = &file
	}
	return &cloned
}
