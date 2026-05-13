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

package dashboard

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// URLDrilldownNestedOpts configures the shared URL drilldown list-element schema used by typed panel `drilldowns`.
//
// When AllowedTriggers contains exactly one value, no `trigger` attribute is exposed; callers fix trigger in the
// model layer to that constant (matches existing slo_burn_rate / slo_overview panels). When more than one value is
// allowed (e.g. image panel `url_drilldown`), a required `trigger` attribute is added with stringvalidator.OneOf.
type URLDrilldownNestedOpts struct {
	AllowedTriggers []string

	// TriggerMarkdownDescription is required when len(AllowedTriggers) > 1.
	TriggerMarkdownDescription string

	URLMarkdownDescription          string
	LabelMarkdownDescription        string
	EncodeURLMarkdownDescription    string
	OpenInNewTabMarkdownDescription string
}

// urlDrilldownNestedAttributeObject returns the NestedAttributeObject placed inside a ListNestedAttribute `drilldowns`.
func urlDrilldownNestedAttributeObject(opts URLDrilldownNestedOpts) schema.NestedAttributeObject {
	if len(opts.AllowedTriggers) < 1 {
		panic("dashboard: urlDrilldownNestedAttributeObject requires a non-empty AllowedTriggers slice")
	}
	if len(opts.AllowedTriggers) > 1 && opts.TriggerMarkdownDescription == "" {
		panic("dashboard: TriggerMarkdownDescription is required when multiple triggers are allowed")
	}

	attrs := map[string]schema.Attribute{
		"url": schema.StringAttribute{
			MarkdownDescription: opts.URLMarkdownDescription,
			Required:            true,
		},
		"label": schema.StringAttribute{
			MarkdownDescription: opts.LabelMarkdownDescription,
			Required:            true,
		},
		"encode_url": schema.BoolAttribute{
			MarkdownDescription: opts.EncodeURLMarkdownDescription,
			Optional:            true,
		},
		"open_in_new_tab": schema.BoolAttribute{
			MarkdownDescription: opts.OpenInNewTabMarkdownDescription,
			Optional:            true,
		},
	}

	if len(opts.AllowedTriggers) > 1 {
		attrs["trigger"] = schema.StringAttribute{
			MarkdownDescription: opts.TriggerMarkdownDescription,
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(opts.AllowedTriggers...),
			},
		}
	}

	return schema.NestedAttributeObject{
		Attributes: attrs,
	}
}
