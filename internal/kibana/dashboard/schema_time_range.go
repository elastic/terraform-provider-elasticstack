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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// timeRangeSchemaAttributes returns the inner attributes for dashboard/panel `time_range` objects (`from`, `to`
// required strings; optional `mode`). Matches the dashboard-root `time_range` shape.
func timeRangeSchemaAttributes() map[string]schema.Attribute {
	attrs := panelkit.TimeRangeAttributes()
	// Preserve richer dashboard-root documentation for `mode` (GET omission preservation semantics).
	if mode, ok := attrs["mode"].(schema.StringAttribute); ok {
		mode.MarkdownDescription = "Time range mode. Valid values are `absolute` or `relative`. When the GET API omits `mode`, " +
			"the provider preserves the prior `time_range.mode` from configuration or state."
		attrs["mode"] = mode
	}
	return attrs
}

// timeRangeSingleNestedAttribute builds a SingleNestedAttribute wrapping timeRangeSchemaAttributes.
// When required is false the attribute is Optional (for panel-level `time_range`); when true it matches dashboard-root usage.
func timeRangeSingleNestedAttribute(markdownDescription string, required bool) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: markdownDescription,
		Required:            required,
		Optional:            !required,
		Attributes:          timeRangeSchemaAttributes(),
	}
}
