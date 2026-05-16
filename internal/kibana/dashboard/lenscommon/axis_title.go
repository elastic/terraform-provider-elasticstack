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

package lenscommon

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AxisTitleFromAPI maps optional axis title API structs into Terraform models.
func AxisTitleFromAPI(m *models.AxisTitleModel, apiTitle *struct {
	Text    *string `json:"text,omitempty"`
	Visible *bool   `json:"visible,omitempty"`
}) {
	if apiTitle == nil {
		return
	}
	m.Value = types.StringPointerValue(apiTitle.Text)
	m.Visible = types.BoolPointerValue(apiTitle.Visible)
}

// AxisTitleToAPI writes Terraform axis title models back into the kbapi anonymous title shape.
func AxisTitleToAPI(m *models.AxisTitleModel) *struct {
	Text    *string `json:"text,omitempty"`
	Visible *bool   `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}

	title := &struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	}{}

	if typeutils.IsKnown(m.Value) {
		title.Text = new(m.Value.ValueString())
	}
	if typeutils.IsKnown(m.Visible) {
		title.Visible = new(m.Visible.ValueBool())
	}

	return title
}
