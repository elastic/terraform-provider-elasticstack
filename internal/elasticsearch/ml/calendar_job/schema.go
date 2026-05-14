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

package calendar_job

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	calendarIDAllowedCharsMessage = "must contain lowercase alphanumeric characters, hyphens, and underscores, " +
		"and must start and end with alphanumeric characters"
	jobIDAllowedCharsMessage = "must contain lowercase alphanumeric characters (a-z and 0-9), hyphens, and underscores. " +
		"It must start and end with alphanumeric characters"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Assigns a single anomaly detection **job or job group** to an ML calendar using " +
			"`PUT _ml/calendars/{calendar_id}/jobs/{job_id}` (and removes it on destroy). " +
			"The `job_id` value is the same path parameter Elasticsearch accepts: a job identifier or a job group name (see the Elasticsearch REST API operation `ml.put_calendar_job`). " +
			"This resource models **one** identifier per instance (comma-separated lists in the API are not valid for the Terraform `job_id` attribute). " +
			"The computed `id` is `<cluster_uuid>/<calendar_id>|<job_id>` (a pipe separates calendar and job because the composite ID only allows one slash).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal composite identifier of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"calendar_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the ML calendar. Must contain lowercase alphanumeric characters " +
					"(a-z and 0-9), hyphens, or underscores. Must start and end with an alphanumeric character.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`),
						calendarIDAllowedCharsMessage,
					),
				},
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "Anomaly detection **job identifier** or **job group name** to attach to the calendar, matching Elasticsearch `PUT .../jobs/{job_id}` (one value per resource; not a comma-separated list).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`),
						jobIDAllowedCharsMessage,
					),
				},
			},
		},
	}
}
