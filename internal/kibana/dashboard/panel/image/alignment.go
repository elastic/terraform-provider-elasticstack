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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
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

	lenscommon.PreserveKnownTfValueIfStateNull(plan.AltText, &state.AltText)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.BackgroundColor, &state.BackgroundColor)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Title, &state.Title)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Description, &state.Description)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.HideTitle, &state.HideTitle)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.HideBorder, &state.HideBorder)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.ObjectFit, &state.ObjectFit)

	if plan.Src.URL != nil {
		if state.Src.URL == nil {
			url := *plan.Src.URL
			state.Src.URL = &url
		} else {
			lenscommon.PreserveKnownTfValueIfStateNull(plan.Src.URL.URL, &state.Src.URL.URL)
		}
	}
	if plan.Src.File != nil {
		if state.Src.File == nil {
			file := *plan.Src.File
			state.Src.File = &file
		} else {
			lenscommon.PreserveKnownTfValueIfStateNull(plan.Src.File.FileID, &state.Src.File.FileID)
		}
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
