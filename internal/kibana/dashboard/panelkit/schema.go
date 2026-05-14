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

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// URLDrilldownSchema returns the NestedAttributeObject for typed panel URL drilldown list elements
// that fix trigger and type constants in the model layer (matching dashboard URLDrilldownNestedOpts).
func URLDrilldownSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Templated URL for the drilldown.",
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Display label shown in the drilldown menu.",
				Required:            true,
			},
			"encode_url": schema.BoolAttribute{
				MarkdownDescription: "When true, the URL is percent-encoded. Omit to use the API default.",
				Optional:            true,
			},
			"open_in_new_tab": schema.BoolAttribute{
				MarkdownDescription: "When true, the URL opens in a new browser tab. Omit to use the API default.",
				Optional:            true,
			},
		},
	}
}
