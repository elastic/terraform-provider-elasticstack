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

package panelkit

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

// PanelPresentationAttributes returns the standard optional presentation attributes shared by
// dashboard panels (`title`, `description`, `hide_title`, `hide_border`). Panel handlers compose
// this map with their type-specific attributes via map literal spread or `maps.Copy`.
func PanelPresentationAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "Optional panel title shown in the panel header.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Optional panel description.",
			Optional:            true,
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, hides the panel title.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, hides the panel border.",
			Optional:            true,
		},
	}
}
