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

// Package syntheticscommon contains schema helpers shared by Synthetics-flavoured
// dashboard panel handlers (`synthetics_monitors`, `synthetics_stats_overview`).
package syntheticscommon

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

// FilterAttributeOptions configures FilterAttribute.
//
// Each category description field is optional; an empty string falls back to a
// sensible default. The top-level BlockMarkdownDescription is required since
// the wrapping block's purpose varies between Synthetics panels.
type FilterAttributeOptions struct {
	BlockMarkdownDescription string
	ProjectsDescription      string
	TagsDescription          string
	MonitorIDsDescription    string
	LocationsDescription     string
	MonitorTypesDescription  string
}

// FilterAttribute returns the shared `filters` SingleNestedAttribute used by both
// `synthetics_monitors_config` and `synthetics_stats_overview_config`. The five
// category lists (`projects`, `tags`, `monitor_ids`, `locations`, `monitor_types`)
// share an identical `{label, value}` element schema.
func FilterAttribute(opts FilterAttributeOptions) schema.Attribute {
	item := filterItem()
	return schema.SingleNestedAttribute{
		MarkdownDescription: opts.BlockMarkdownDescription,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				MarkdownDescription: nz(opts.ProjectsDescription, "Filter by Synthetics project."),
				Optional:            true,
				NestedObject:        item,
			},
			"tags": schema.ListNestedAttribute{
				MarkdownDescription: nz(opts.TagsDescription, "Filter by monitor tag."),
				Optional:            true,
				NestedObject:        item,
			},
			"monitor_ids": schema.ListNestedAttribute{
				MarkdownDescription: nz(opts.MonitorIDsDescription, "Filter by monitor ID. The API accepts up to 5000 entries."),
				Optional:            true,
				NestedObject:        item,
			},
			"locations": schema.ListNestedAttribute{
				MarkdownDescription: nz(opts.LocationsDescription, "Filter by monitor location."),
				Optional:            true,
				NestedObject:        item,
			},
			"monitor_types": schema.ListNestedAttribute{
				MarkdownDescription: nz(opts.MonitorTypesDescription, "Filter by monitor type (e.g. `browser`, `http`)."),
				Optional:            true,
				NestedObject:        item,
			},
		},
	}
}

func filterItem() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"label": schema.StringAttribute{
				MarkdownDescription: "Display label for the filter option.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value for the filter option.",
				Required:            true,
			},
		},
	}
}

func nz(s, def string) string {
	if s != "" {
		return s
	}
	return def
}
