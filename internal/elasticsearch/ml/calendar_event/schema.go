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

package calendar_event

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages scheduled events for a Machine Learning calendar. " +
			"See the [ML post calendar events API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-post-calendar-events) for more details. " +
			"**Import** id format: `<cluster_uuid>/<calendar_id>/<event_id>` (the same value as the computed `id` attribute).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal composite identifier of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"calendar_id": schema.StringAttribute{
				MarkdownDescription: "The identifier for the calendar that owns the event. Must contain lowercase alphanumeric characters " +
					"(a-z and 0-9), hyphens, or underscores. Must start and end with an alphanumeric character.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`),
						"must contain lowercase alphanumeric characters, hyphens, and underscores, "+
							"and must start and end with alphanumeric characters",
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the scheduled event.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"start_time": schema.StringAttribute{
				MarkdownDescription: "The start time of the scheduled event in RFC 3339 format.",
				Required:            true,
				CustomType:          timetypes.RFC3339Type{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"end_time": schema.StringAttribute{
				MarkdownDescription: "The end time of the scheduled event in RFC 3339 format.",
				Required:            true,
				CustomType:          timetypes.RFC3339Type{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"skip_result": schema.BoolAttribute{
				MarkdownDescription: "If true, results are not generated for buckets that fall inside the event period. " +
					"When omitted, the request does not send this field and Elasticsearch applies its default behavior. " +
					"Explicit values require Elasticsearch **8.16** or newer. Maps to `skip_result` in the Elasticsearch API.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"skip_model_update": schema.BoolAttribute{
				MarkdownDescription: "If true, model updates are not generated for buckets that fall inside the event period. " +
					"When omitted, the request does not send this field and Elasticsearch applies its default behavior. " +
					"Explicit values require Elasticsearch **8.16** or newer. Maps to `skip_model_update` in the Elasticsearch API.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"force_time_shift": schema.StringAttribute{
				MarkdownDescription: "When set, changes the duration of the event to the specified value in seconds (decimal digits as a string; the API uses a JSON number). " +
					"Requires Elasticsearch **8.16** or newer. Maps to `force_time_shift` in the Elasticsearch API.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"event_id": schema.StringAttribute{
				MarkdownDescription: "The server-generated identifier for the event.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
