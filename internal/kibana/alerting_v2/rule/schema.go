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

package rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alerting_v2/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	kindAlert  = "alert"
	kindSignal = "signal"

	minScheduleEvery = "5s"

	queryAttribute = "query"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier of the rule in the form `<space_id>/<rule_id>`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"rule_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the rule within the space: 1-150 characters of letters, digits, underscores, or hyphens. A UUID is generated when omitted. Changing it replaces the resource.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{utils.EntityIDValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": kbschema.ResourceSpaceIDAttributeRequiresReplaceOnly(),
			"kind": schema.StringAttribute{
				MarkdownDescription: "Whether the rule creates alerts (`alert`) or only stores matching events (`signal`). The API treats it as immutable, so changing it replaces the resource.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf(kindAlert, kindSignal)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule is enabled. Always sent, so a rule is created in or replaced into the configured state in one request.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"time_field": schema.StringAttribute{
				MarkdownDescription: "Document field Kibana uses with `schedule.lookback` to time-filter `query.base`. Defaults to `@timestamp`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultTimeField),
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 256)},
			},
			"metadata": schema.SingleNestedAttribute{
				MarkdownDescription: "Rule metadata.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Rule name. Must be unique within the space.",
						Required:            true,
						Validators:          []validator.String{stringvalidator.LengthBetween(1, 256)},
					},
				},
			},
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: "Execution schedule.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"every": schema.StringAttribute{
						MarkdownDescription: "Execution interval as an unsigned integer followed by `ms`, `s`, `m`, `h`, `d`, or `w` (for example `5m`). " +
							"At least `5s` and at most `365d`. The deployment may enforce a higher minimum server-side.",
						Required:   true,
						Validators: []validator.String{utils.DurationValidatorAtLeast(minScheduleEvery)},
					},
					"lookback": schema.StringAttribute{
						MarkdownDescription: "Lookback window for the query in the same duration grammar (for example `5m`). At most `365d`.",
						Optional:            true,
						Validators:          []validator.String{utils.DurationValidator()},
					},
				},
			},
			queryAttribute: schema.SingleNestedAttribute{
				MarkdownDescription: "ES|QL query the rule evaluates.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"base": schema.StringAttribute{
						MarkdownDescription: "ES|QL query that specifies the data to evaluate. Must include a `FROM` clause. Kibana applies the time filter from `schedule.lookback` using `time_field`.",
						Required:            true,
						Validators:          []validator.String{stringvalidator.LengthBetween(1, 10000)},
					},
				},
			},
			"recovery": schema.SingleNestedAttribute{
				MarkdownDescription: "When an alert recovers. Required when `kind` is `alert` and not allowed when `kind` is `signal`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"strategy": schema.StringAttribute{
						MarkdownDescription: "Recovery strategy: `no_breach` recovers when the group no longer breaches, `condition` when `query.base` plus `segment` returns the group, " +
							"`query` when the separate `query` returns the group, and `manual` never recovers automatically.",
						Required: true,
						Validators: []validator.String{stringvalidator.OneOf(
							recoveryStrategyNoBreach, recoveryStrategyCondition, recoveryStrategyQuery, recoveryStrategyManual,
						)},
					},
					"segment": schema.StringAttribute{
						MarkdownDescription: "ES|QL clause appended to `query.base`, for example `WHERE avg_cpu < 0.60`. Required when `strategy` is `condition` and not allowed otherwise.",
						Optional:            true,
					},
					queryAttribute: schema.StringAttribute{
						MarkdownDescription: "Independent ES|QL query, including its own `FROM` clause. Required when `strategy` is `query` and not allowed otherwise.",
						Optional:            true,
					},
				},
			},
			"no_data": schema.SingleNestedAttribute{
				MarkdownDescription: "What the rule does when a group has no data. Required when `kind` is `alert` and not allowed when `kind` is `signal`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"strategy": schema.StringAttribute{
						MarkdownDescription: "No-data strategy: `ignore` does not check for missing data, `keep_last` holds the alert's current status, " +
							"and `resolve` closes the alert the first time the rule finds no data.",
						Required: true,
						Validators: []validator.String{stringvalidator.OneOf(
							noDataStrategyIgnore, noDataStrategyKeepLast, noDataStrategyResolve,
						)},
					},
					queryAttribute: schema.StringAttribute{
						MarkdownDescription: "Optional ES|QL query that checks whether a group has data. Not allowed when `strategy` is `ignore`.",
						Optional:            true,
					},
				},
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Configuration revision of the rule, incremented by the server on every change.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the rule's creation in ISO 8601 format.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the rule's last change in ISO 8601 format.",
				Computed:            true,
			},
		},
	}
}
