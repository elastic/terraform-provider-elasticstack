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

package calendar

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *calendarResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Machine Learning calendars. " +
			"See the [ML Calendar API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-calendar.html) for more details.",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"calendar_id": schema.StringAttribute{
				MarkdownDescription: "A string that uniquely identifies a calendar. Must contain lowercase alphanumeric characters " +
					"(a-z and 0-9), hyphens, or underscores. Must start and end with an alphanumeric character.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$|^[a-z0-9]$`),
						"must contain lowercase alphanumeric characters, hyphens, and underscores, "+
							"and must start and end with alphanumeric characters",
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the calendar.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"job_ids": schema.SetAttribute{
				MarkdownDescription: "A list of anomaly detection job identifiers.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}
