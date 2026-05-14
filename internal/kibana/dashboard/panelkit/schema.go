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

// Default URL drilldown element descriptions (typed panels that fix trigger/type in the model layer).
const (
	urlDrilldownDefaultURLDescription          = "Templated URL for the drilldown."
	urlDrilldownDefaultLabelDescription        = "Display label shown in the drilldown menu."
	urlDrilldownDefaultEncodeURLDescription    = "When true, the URL is percent-encoded. Omit to use the API default."
	urlDrilldownDefaultOpenInNewTabDescription = "When true, the URL opens in a new browser tab. Omit to use the API default."
)

// URLDrilldownOptions overrides MarkdownDescription on URL drilldown nested object attributes.
// Trigger and type are not schema fields (fixed in the model layer).
// Empty string in a field means use the default for that attribute (see Default URL drilldown constants).
type URLDrilldownOptions struct {
	URLMarkdownDescription          string
	LabelMarkdownDescription        string
	EncodeURLMarkdownDescription    string
	OpenInNewTabMarkdownDescription string
}

// URLDrilldownSchema returns the NestedAttributeObject used inside a ListNestedAttribute `drilldowns`.
func URLDrilldownSchema(opts URLDrilldownOptions) schema.NestedAttributeObject {
	urlDesc := opts.URLMarkdownDescription
	if urlDesc == "" {
		urlDesc = urlDrilldownDefaultURLDescription
	}
	labelDesc := opts.LabelMarkdownDescription
	if labelDesc == "" {
		labelDesc = urlDrilldownDefaultLabelDescription
	}
	encodeDesc := opts.EncodeURLMarkdownDescription
	if encodeDesc == "" {
		encodeDesc = urlDrilldownDefaultEncodeURLDescription
	}
	openDesc := opts.OpenInNewTabMarkdownDescription
	if openDesc == "" {
		openDesc = urlDrilldownDefaultOpenInNewTabDescription
	}
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: urlDesc,
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: labelDesc,
				Required:            true,
			},
			"encode_url": schema.BoolAttribute{
				MarkdownDescription: encodeDesc,
				Optional:            true,
			},
			"open_in_new_tab": schema.BoolAttribute{
				MarkdownDescription: openDesc,
				Optional:            true,
			},
		},
	}
}
